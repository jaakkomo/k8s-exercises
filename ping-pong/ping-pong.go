package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
)

type App struct {
	conn atomic.Pointer[Connection]
}

type Connection struct {
	conn *pgx.Conn
}

func Connect(ctx context.Context, databaseUrl string) (*Connection, error) {
	conn, err := pgx.Connect(ctx, databaseUrl)
	if err != nil {
		return nil, err
	}

	return &Connection{conn: conn}, nil
}

func (c *Connection) Close(ctx context.Context) error {
	return c.conn.Close(ctx)
}

func (c *Connection) Initialize(ctx context.Context) error {
	_, err := c.conn.Exec(ctx, `
CREATE TABLE IF NOT EXISTS pings (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`,
	)
	return err
}

func (c *Connection) InsertPing(ctx context.Context) error {
	_, err := c.conn.Exec(ctx, `
INSERT INTO pings DEFAULT VALUES
`,
	)
	return err
}

func (c *Connection) GetPingsCount(ctx context.Context) (int, error) {
	var count int

	err := c.conn.QueryRow(ctx, `
SELECT COUNT(*)
FROM pings
`,
	).Scan(&count)

	return count, err
}

func (app *App) HasDbConnection(ctx context.Context) bool {
	conn := app.conn.Load()
	if conn == nil {
		return false
	}

	return conn.conn.Ping(ctx) == nil
}

func (app *App) Connection() *Connection {
	return app.conn.Load()
}

func (app *App) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	oldCount, err := app.Connection().GetPingsCount(ctx)
	app.Connection().InsertPing(ctx)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "pong %d\n", oldCount)
}

func (app *App) Pings(w http.ResponseWriter, r *http.Request) {
	count, err := app.Connection().GetPingsCount(r.Context())
	if err != nil {
		fmt.Println(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%d", count)
}

func (app *App) Health(w http.ResponseWriter, r *http.Request) {
	if !app.HasDbConnection(r.Context()) {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *App) requireReady(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !app.HasDbConnection(r.Context()) {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		next(w, r)
	}
}

func tryConnectUntilConnected(app *App, databaseUrl string) {
	for {
		ctx := context.Background()
		conn, err := Connect(ctx, databaseUrl)
		if err != nil {
			fmt.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}

		conn.Initialize(ctx)
		app.conn.Store(conn)
		fmt.Println("connected to database")
		return
	}
}

func readEnv(env, fallback string) string {
	if value, ok := os.LookupEnv(env); ok {
		return value
	} else {
		return fallback
	}
}

func main() {
	port := readEnv("PORT", "8080")

	databaseUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/postgres",
		readEnv("DB_USER", "postgres"),
		readEnv("DB_PASSWORD", "postgres"),
		readEnv("DB_HOST", "localhost"),
		readEnv("DB_PORT", "5432"),
	)

	app := App{}

	http.HandleFunc("/", app.requireReady(app.Index))
	http.HandleFunc("/pings", app.requireReady(app.Pings))
	http.HandleFunc("/healthz", app.Health)
	fmt.Println("Server started in port", port)
	go tryConnectUntilConnected(&app, databaseUrl)
	http.ListenAndServe(":"+port, nil)
}
