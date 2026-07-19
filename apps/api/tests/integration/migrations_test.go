package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"retailpulse/apps/api/internal/platform/database"
)

func TestMigrationsApply(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	if err := database.RunMigrations(ctx, pool, "../../migrations"); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	var count int
	if err := pool.QueryRow(ctx, `select count(*) from roles`).Scan(&count); err != nil {
		t.Fatalf("count roles: %v", err)
	}
	if count < 4 {
		t.Fatalf("expected seeded roles, got %d", count)
	}
}
