// cmd/replay/main.go
// DLQ replay CLI — replays one or all dead-letter events by calling the
// EchoLine admin API or directly updating the database.
//
// Usage:
//
//	go run ./cmd/replay --id <uuid>               # replay one event via API
//	go run ./cmd/replay --all                     # replay all failed events via API
//	go run ./cmd/replay --id <uuid> --direct      # direct DB update (no API)
//	go run ./cmd/replay --list                    # list DLQ events
//
// Required environment variables (API mode):
//
//	ECHOLINE_BASE_URL  base URL of the EchoLine API   (default: http://localhost:8080)
//	ECHOLINE_TOKEN     admin JWT
//
// Required environment variables (direct DB mode):
//
//	DATABASE_URL       postgres DSN
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

// ── DLQ event ────────────────────────────────────────────────────────────────

type dlqEvent struct {
	ID          string    `json:"id"`
	EventType   string    `json:"event_type"`
	Payload     string    `json:"payload"`
	Attempts    int       `json:"attempts"`
	LastError   string    `json:"last_error"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	LastAttempt time.Time `json:"last_attempt"`
}

type dlqListResp struct {
	Events []dlqEvent `json:"events"`
}

// ── flags ─────────────────────────────────────────────────────────────────────

var (
	flagID     = flag.String("id", "", "DLQ event UUID to replay")
	flagAll    = flag.Bool("all", false, "Replay all failed events")
	flagList   = flag.Bool("list", false, "List DLQ events (no replay)")
	flagDirect = flag.Bool("direct", false, "Directly update DB status (no API call)")
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

// ── API mode ──────────────────────────────────────────────────────────────────

func runAPI(ctx context.Context, baseURL, token string) {
	if *flagList {
		events := listEvents(ctx, baseURL, token)
		printEvents(events)
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
			if e.Status != "failed" && e.Status != "" {
				continue
			}
			if *flagDry {
				fmt.Printf("[dry-run] would replay %s (%s)\n", e.ID, e.EventType)
				continue
			}
			if err := replayEvent(ctx, baseURL, token, e.ID); err != nil {
				fmt.Printf("[FAIL] %s: %v\n", e.ID, err)
				fail++
			} else {
				fmt.Printf("[OK]   %s (%s)\n", e.ID, e.EventType)
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

	// single event
	if *flagDry {
		fmt.Printf("[dry-run] would replay event %s\n", *flagID)
		return
	}
	if err := replayEvent(ctx, baseURL, token, *flagID); err != nil {
		log.Fatalf("Replay failed: %v", err)
	}
	fmt.Printf("Event %s replayed successfully\n", *flagID)
}

func listEvents(ctx context.Context, baseURL, token string) []dlqEvent {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/admin/dlq", nil)
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
	return result.Events
}

func replayEvent(ctx context.Context, baseURL, token, id string) error {
	url := fmt.Sprintf("%s/admin/dlq/%s/replay", baseURL, id)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url,
		bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if status, _ := out["status"].(string); status != "replayed" && status != "ok" {
		return fmt.Errorf("unexpected status %q", status)
	}
	return nil
}

// ── Direct DB mode ────────────────────────────────────────────────────────────

func runDirect(ctx context.Context, dbURL string) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Connect DB: %v", err)
	}
	defer pool.Close()

	if *flagList {
		rows, err := pool.Query(ctx,
			`SELECT id, event_type, attempts, last_error, status, created_at
			 FROM dead_letter ORDER BY created_at DESC LIMIT 100`)
		if err != nil {
			log.Fatalf("Query DLQ: %v", err)
		}
		defer rows.Close()
		fmt.Printf("%-36s  %-20s  %-8s  %-8s  %s\n",
			"ID", "TYPE", "ATTEMPTS", "STATUS", "LAST_ERROR")
		fmt.Println("─────────────────────────────────────────────────────────────────────────────────")
		for rows.Next() {
			var e dlqEvent
			if err := rows.Scan(&e.ID, &e.EventType, &e.Attempts,
				&e.LastError, &e.Status, &e.CreatedAt); err != nil {
				log.Printf("Scan: %v", err)
				continue
			}
			fmt.Printf("%-36s  %-20s  %-8d  %-8s  %s\n",
				e.ID, e.EventType, e.Attempts, e.Status, truncate(e.LastError, 40))
		}
		return
	}

	updateStatus := func(id string) error {
		tag, err := pool.Exec(ctx,
			`UPDATE dead_letter SET status='replayed', attempts=attempts+1,
			  last_attempt=NOW() WHERE id=$1 AND status='failed'`, id)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("event %s not found or not in 'failed' state", id)
		}
		return nil
	}

	if *flagAll {
		rows, err := pool.Query(ctx,
			`SELECT id, event_type FROM dead_letter WHERE status='failed'`)
		if err != nil {
			log.Fatalf("Query failed events: %v", err)
		}
		var ids []string
		var types []string
		for rows.Next() {
			var id, et string
			if err := rows.Scan(&id, &et); err != nil {
				continue
			}
			ids = append(ids, id)
			types = append(types, et)
		}
		rows.Close()

		ok, fail := 0, 0
		for i, id := range ids {
			if *flagDry {
				fmt.Printf("[dry-run] would mark %s (%s) as replayed\n", id, types[i])
				continue
			}
			if err := updateStatus(id); err != nil {
				fmt.Printf("[FAIL] %s: %v\n", id, err)
				fail++
			} else {
				fmt.Printf("[OK]   %s (%s)\n", id, types[i])
				ok++
			}
		}
		fmt.Printf("\nDone: %d updated, %d failed\n", ok, fail)
		if fail > 0 {
			os.Exit(1)
		}
		return
	}

	if *flagDry {
		fmt.Printf("[dry-run] would mark %s as replayed in DB\n", *flagID)
		return
	}
	if err := updateStatus(*flagID); err != nil {
		log.Fatalf("Update failed: %v", err)
	}
	fmt.Printf("Event %s marked as replayed\n", *flagID)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func printEvents(events []dlqEvent) {
	if len(events) == 0 {
		fmt.Println("DLQ is empty.")
		return
	}
	fmt.Printf("%-36s  %-20s  %-8s  %-8s  %s\n",
		"ID", "TYPE", "ATTEMPTS", "STATUS", "LAST_ERROR")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────────")
	for _, e := range events {
		fmt.Printf("%-36s  %-20s  %-8d  %-8s  %s\n",
			e.ID, e.EventType, e.Attempts, e.Status, truncate(e.LastError, 40))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
