package migrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationsDirEnvOverride(t *testing.T) {
	t.Setenv("MIGRATIONS_DIR", "/custom/migrations/path")

	got := migrationsDir()
	want := filepath.Clean("/custom/migrations/path")
	if got != want {
		t.Fatalf("migrationsDir() = %q, want %q", got, want)
	}
}

func TestMigrationsDirSourceTreeFallback(t *testing.T) {
	t.Setenv("MIGRATIONS_DIR", "")

	got := migrationsDir()
	if !dirExists(got) {
		t.Fatalf("migrationsDir() = %q, expected an existing directory", got)
	}

	entries, err := os.ReadDir(got)
	if err != nil {
		t.Fatalf("read migrations dir %q: %v", got, err)
	}
	if len(entries) == 0 {
		t.Fatalf("migrations dir %q is empty", got)
	}

	found := false
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".sql") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("migrations dir %q has no .sql files", got)
	}
}
