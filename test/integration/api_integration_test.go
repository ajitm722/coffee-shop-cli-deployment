//go:build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"coffee/internal/db"
	"coffee/internal/server"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// startPostgres spins up a Postgres container for tests.
func startPostgres(t *testing.T) (tc.Container, string) {
	t.Helper()
	ctx := context.Background()

	req := tc.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "coffee",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}
	pg, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres: %v", err)
	}
	ip, err := pg.ContainerIP(ctx)
	if err != nil {
		t.Fatalf("get container IP: %v", err)
	}
	dsn := fmt.Sprintf("postgres://postgres:password@%s:%d/coffee?sslmode=disable", ip, 5432)
	return pg, dsn
}

// runMigrations applies migrations to the fresh Postgres.
func runMigrations(t *testing.T, dsn string) {
	t.Helper()
	dbConn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("db connect: %v", err)
	}
	defer dbConn.Close()

	// --- wait until DB really ready ---
	for i := 0; i < 10; i++ {
		if err := dbConn.Ping(); err == nil {
			break
		}
		if i == 9 {
			t.Fatalf("db never became ready: %v", err)
		}
		time.Sleep(time.Second)
	}

	applyFile := func(path string) {
		bytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		lines := strings.Split(string(bytes), "\n")

		var cleaned []string
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" || strings.HasPrefix(l, "--") {
				continue
			}
			cleaned = append(cleaned, l)
		}
		sqlText := strings.Join(cleaned, " ")

		stmts := strings.Split(sqlText, ";")
		for _, stmt := range stmts {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := dbConn.Exec(stmt); err != nil {
				t.Fatalf("apply %s stmt %q: %v", path, stmt, err)
			}
		}
	}

	migPath := "../../db/migrations/0001_init.up.sql"
	seedPath := "../../db/seed.sql"

	applyFile(migPath)
	applyFile(seedPath)
}

// TestMenuEndpoint runs end-to-end: Postgres → migrate → server → /menu.
func TestMenuEndpoint(t *testing.T) {
	pg, dsn := startPostgres(t)
	defer pg.Terminate(context.Background())

	runMigrations(t, dsn)

	// connect and launch server
	database, err := db.Connect(dsn)
	if err != nil {
		t.Fatalf("db.Connect: %v", err)
	}
	defer database.Close()

	s := server.NewServer(
		server.WithAddr(":0"), // dynamic port not needed for httptest
		server.WithDB(database),
		server.WithGinMode(gin.TestMode), // Use TestMode for cleaner test output
	)

	// Serve via httptest
	ts := httptest.NewServer(s.Engine())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v1/menu")
	if err != nil {
		t.Fatalf("GET /menu failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// BenchmarkMenuEndpoint measures /menu latency with seeded DB.
func BenchmarkMenuEndpoint(b *testing.B) {
	pg, dsn := startPostgres(&testing.T{})
	defer pg.Terminate(context.Background())

	runMigrations(&testing.T{}, dsn)

	database, _ := db.Connect(dsn)
	defer database.Close()

	s := server.NewServer(
		server.WithDB(database),             // Configure server with database connection
		server.WithGinMode(gin.ReleaseMode), // To avoid debug logs in benchmarks
	)
	ts := httptest.NewServer(s.Engine())
	defer ts.Close()

	client := &http.Client{}
	url := ts.URL + "/v1/menu"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(url)
		if err != nil {
			b.Fatalf("GET failed: %v", err)
		}
		resp.Body.Close()
	}
}
