package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

type Todo struct {
	ID   int64
	Text string `binding:"required,max=140"`
}

type App struct {
	conn atomic.Pointer[Connection]
	isHealthy atomic.Bool
	logger *slog.Logger
}

type Connection struct {
	conn *pgx.Conn
}

func (app *App) IsHealthy(ctx context.Context) bool {
	conn := app.conn.Load()
	isHealthy := app.isHealthy.Load()
	if conn == nil {
		return false
	}

	return conn.conn.Ping(ctx) == nil && isHealthy
}

func (app *App) Connection() *Connection {
	return app.conn.Load()
}

func tryConnectUntilConnected(app *App, databaseUrl string) {
	for {
		ctx := context.Background()
		conn, err := Connect(ctx, databaseUrl)
		if err != nil {
			app.logger.Error(
				"failed to connect to database",
				"error", err,
			)
			time.Sleep(5 * time.Second)
			continue
		}

		conn.Initialize(ctx)
		app.conn.Store(conn)
		app.logger.Info("connected to database")
		return
	}
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
CREATE TABLE IF NOT EXISTS todos (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    text VARCHAR(140) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`,
	)
	return err
}

func (c *Connection) CreateTodo(ctx context.Context, todo Todo) error {
	_, err := c.conn.Exec(ctx, `
INSERT INTO todos (text)
VALUES ($1)
`,
		todo.Text,
	)
	return err
}

func (c *Connection) GetTodos(ctx context.Context) ([]Todo, error) {
	rows, err := c.conn.Query(ctx, `
SELECT id, text
FROM todos
`,
	)
	if err != nil {
		return nil, err
	}

	todos, err := pgx.CollectRows(rows, pgx.RowToStructByName[Todo])
	if err != nil {
		return nil, err
	}

	return todos, nil
}

func readEnv(env, fallback string) string {
	if value, ok := os.LookupEnv(env); ok {
		return value
	} else {
		return fallback
	}
}

func (app *App) FetchTodos(c *gin.Context) {
	todos, err := app.Connection().GetTodos(c.Request.Context())
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, todos)
}

func (app *App) CreateTodo(c *gin.Context) {
	var todo Todo

	if err := c.ShouldBindJSON(&todo); err != nil {
		app.logger.Warn(
			"validation-error",
			"error", err.Error(),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := app.Connection().CreateTodo(c.Request.Context(), todo)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	app.logger.Info(
		"created_todo",
		"text", todo.Text,
	)

	c.JSON(http.StatusCreated, todo)
}

func (app *App) Health(c *gin.Context) {
	if !app.IsHealthy(c.Request.Context()) {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}

	c.Status(http.StatusOK)
}

func (app *App) Break(c *gin.Context) {
	app.isHealthy.Store(false)
	app.logger.Error("app broke")
	c.Status(http.StatusCreated)
}

func (app *App) requireReady(next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !app.IsHealthy(c.Request.Context()) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "service unavailable",
			})
			return
		}

		next(c)
	}
}

func SlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		logger.Info("request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
		)
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(SlogMiddleware(logger))

	port := readEnv("PORT", "8080")
	databaseUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/postgres",
		readEnv("DB_USER", "postgres"),
		readEnv("DB_PASSWORD", "postgres"),
		readEnv("DB_HOST", "localhost"),
		readEnv("DB_PORT", "5432"),
	)

	app := App{logger: logger}
	app.isHealthy.Store(true)

	r.GET("/todos", app.requireReady(app.FetchTodos))
	r.POST("/todos", app.requireReady(app.CreateTodo))
	r.GET("/healthz", app.Health)
	r.POST("/break", app.Break)

	fmt.Println("Server started in port", port)
	go tryConnectUntilConnected(&app, databaseUrl)
	r.Run(":" + port)
}
