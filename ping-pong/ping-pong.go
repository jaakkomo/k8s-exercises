package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
)

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

func indexHandler(conn *Connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		oldCount, err := conn.GetPingsCount(ctx)
		conn.InsertPing(ctx)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "pong %d\n", oldCount)
	}
}

func pingsHandler(conn *Connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		count, err := conn.GetPingsCount(r.Context())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%d", count)
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
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
	)

	ctx := context.Background()
	conn, err := Connect(ctx, databaseUrl)
	if err != nil {
		panic(err)
	}
	defer conn.conn.Close(ctx)
	conn.Initialize(ctx)

	http.HandleFunc("/", indexHandler(conn))
	http.HandleFunc("/pings", pingsHandler(conn))
	fmt.Println("Server started in port", port)
	http.ListenAndServe(":"+port, nil)
}
