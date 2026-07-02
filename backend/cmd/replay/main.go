// cmd/replay/main.go — DLQ replay CLI aligned with dead_letter_events schema.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type dlqEvent struct {
	ID           string          `json:"id"`
	SourceTopic  string          `json:"source_topic"`
	Payload      json.RawMessage `json:"payload"`
	ErrorMessage string          `json:"error_message"`
	Attempts     int             `json:"attempts"`
	CreatedAt    time.Time       `json:"created_at"`
}

type dlqListResp struct {
	DeadLetters []dlqEvent `json:"dead_letters"`
}

var (
	flagID     = flag.String("id", "", "DLQ event UUID to replay")
	flagAll    = flag.Bool("all", false, "Replay all DLQ events via API")
	flagList   = flag.Bool("list", false, "List DLQ events")
	flagDirect = flag.Bool("direct", false, "Requeue into outbox via DB (no API)")
	flagDry    = flag.Bool("dry-run", false, "Print actions without executing")
)

func main() {
	flag.Parse()

	baseURL := envOrDefault("ECHOLINE_BASE_URL", "http://localhost:8080")
	token := os.Getenv("ECHOLINE_TOKEN")
	dbURL := os.Getenv("DATABASE_URL")

	if !*flagList && !*flagAll && *flagID == "" {
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()

	if *flagDirect {
		if dbURL == "" {
			log.Fatal("DATABASE_URL required for --direct mode")
		}
		runDirect(ctx, dbURL)
		return
	}

	if token == "" && !*flagList {
		log.Fatal("ECHOLINE_TOKEN required (or use --direct with DATABASE_URL)")
	}

	runAPI(ctx, baseURL, token)
}

func runAPI(ctx context.Context, baseURL, token string) {
	if *flagList {
		printEvents(listEvents(ctx, baseURL, token))
		return
	}

	if *flagAll {
		events := listEvents(ctx, baseURL, token)
		if len(events) == 0 {
			fmt.Println("No events in DLQ.")
			return
		}
		ok, fail := 0, 0
		for _, e := range events {
			if *flagDry {
				fmt.Printf("[dry-run] would replay %s (%s)\n", e.ID, e.SourceTopic)
				continue
			}
			if err := replayEvent(ctx, baseURL, token, e.ID); err != nil {
				fmt.Printf("[FAIL] %s: %v\n", e.ID, err)
				fail++
			} else {
				fmt.Printf("[OK]   %s (%s)\n", e.ID, e.SourceTopic)
				ok++
			}
			time.Sleep(100 * time.Millisecond)
		}
		fmt.Printf("\nDone: %d replayed, %d failed\n", ok, fail)
		if fail > 0 {
			os.Exit(1)
		}
		return
	}

	if *flagDry {
		fmt.Printf("[dry-run] would replay event %s\n", *flagID)
		return
	}
	if err := replayEvent(ctx, baseURL, token, *flagID); err != nil {
		log.Fatalf("Replay failed: %v", err)
	}
	fmt.Printf("Event %s replay queued successfully\n", *flagID)
}

func listEvents(ctx context.Context, baseURL, token string) []dlqEvent {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/admin/dlq", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("List DLQ: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("List DLQ HTTP %d: %s", resp.StatusCode, body)
	}
	var result dlqListResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("Decode DLQ list: %v", err)
	}
	return result.DeadLetters
}

func replayEvent(ctx context.Context, baseURL, token, id string) error {
	url := fmt.Sprintf("%s/api/admin/dlq/%s/replay", baseURL, id)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if status, _ := out["status"].(string); status != "replay_queued" && status != "replayed" && status != "ok" {
		return fmt.Errorf("unexpected status %q", status)
	}
	return nil
}

func runDirect(ctx context.Context, dbURL string) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Connect DB: %v", err)
	}
	defer pool.Close()

	if *flagList {
		rows, err := pool.Query(ctx,
			`SELECT id, source_topic, error_message, attempts, created_at
			 FROM dead_letter_events ORDER BY created_at DESC LIMIT 100`)
		if err != nil {
			log.Fatalf("Query DLQ: %v", err)
		}
		defer rows.Close()
		fmt.Printf("%-36s  %-20s  %-8s  %s\n", "ID", "TOPIC", "ATTEMPTS", "ERROR")
		fmt.Println("──────────────────────────────────────────────────────────────────────────────")
		for rows.Next() {
			var e dlqEvent
			if err := rows.Scan(&e.ID, &e.SourceTopic, &e.ErrorMessage, &e.Attempts, &e.CreatedAt); err != nil {
				log.Printf("Scan: %v", err)
				continue
			}
			fmt.Printf("%-36s  %-20s  %-8d  %s\n", e.ID, e.SourceTopic, e.Attempts, truncate(e.ErrorMessage, 40))
		}
		return
	}

	requeue := func(id string) error {
		var topic, payload string
		err := pool.QueryRow(ctx,
			`SELECT source_topic, payload::text FROM dead_letter_events WHERE id=$1`, id).
			Scan(&topic, &payload)
		if err != nil {
			return err
		}
		_, err = pool.Exec(ctx,
			`INSERT INTO outbox_events (id, topic, payload, status, attempts, created_at)
			 VALUES (gen_random_uuid(), $1, $2::jsonb, 'pending', 0, NOW())`, topic, payload)
		return err
	}

	if *flagAll {
		rows, err := pool.Query(ctx, `SELECT id, source_topic FROM dead_letter_events`)
		if err != nil {
			log.Fatalf("Query DLQ: %v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var id, topic string
			if err := rows.Scan(&id, &topic); err != nil {
				continue
			}
			if *flagDry {
				fmt.Printf("[dry-run] would requeue %s (%s)\n", id, topic)
				continue
			}
			if err := requeue(id); err != nil {
				fmt.Printf("[FAIL] %s: %v\n", id, err)
			} else {
				fmt.Printf("[OK]   %s requeued\n", id)
			}
		}
		return
	}

	if *flagDry {
		fmt.Printf("[dry-run] would requeue %s into outbox\n", *flagID)
		return
	}
	if err := requeue(*flagID); err != nil {
		log.Fatalf("Requeue failed: %v", err)
	}
	fmt.Printf("Event %s requeued into outbox\n", *flagID)
}

func printEvents(events []dlqEvent) {
	if len(events) == 0 {
		fmt.Println("No events in DLQ.")
		return
	}
	fmt.Printf("%-36s  %-20s  %-8s  %s\n", "ID", "TOPIC", "ATTEMPTS", "ERROR")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────")
	for _, e := range events {
		fmt.Printf("%-36s  %-20s  %-8d  %s\n", e.ID, e.SourceTopic, e.Attempts, truncate(e.ErrorMessage, 40))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
