package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/config"
	"github.com/echoline/echoline/backend/internal/conversation"
	"github.com/echoline/echoline/backend/internal/db"
	"github.com/echoline/echoline/backend/internal/message"
	"github.com/echoline/echoline/backend/internal/migrate"
	"github.com/echoline/echoline/backend/internal/user"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()
	if err := migrate.Up(ctx, cfg.DatabaseURL); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	userRepo := user.NewRepository(pool)
	convRepo := conversation.NewRepository(pool)
	msgRepo := message.NewRepository(pool)

	aliceName := envOr("SEED_ALICE", "alice")
	bobName := envOr("SEED_BOB", "bob")
	password := envOr("SEED_PASSWORD", "secret123")

	aliceHash, _ := auth.HashPassword(password)
	bobHash, _ := auth.HashPassword(password)

	alice, err := ensureUser(ctx, userRepo, aliceName, "Alice", aliceHash)
	if err != nil {
		log.Fatalf("seed alice: %v", err)
	}
	bob, err := ensureUser(ctx, userRepo, bobName, "Bob", bobHash)
	if err != nil {
		log.Fatalf("seed bob: %v", err)
	}

	conv, _, err := convRepo.CreateDirect(ctx, alice.ID, bob.ID)
	if err != nil {
		log.Fatalf("create direct conversation: %v", err)
	}

	msg, err := msgRepo.Create(ctx, conv.ID, alice.ID, "seed-msg-1", message.TypeText, "hello from seed")
	if err != nil {
		log.Fatalf("seed message: %v", err)
	}

	fmt.Printf("seeded users alice=%s bob=%s conversation=%s message_seq=%d\n",
		alice.ID, bob.ID, conv.ID, msg.Seq)
}

func ensureUser(ctx context.Context, repo *user.Repository, username, displayName, hash string) (*user.User, error) {
	u, err := repo.GetByUsername(ctx, username)
	if err == nil {
		return u, nil
	}
	return repo.Create(ctx, username, displayName, hash)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
