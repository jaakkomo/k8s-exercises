package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

type Todo struct {
	ID   int64
	Text string `binding:"required,max=140"`
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

func fetchTodosHandler(conn *Connection) gin.HandlerFunc {
	return func(c *gin.Context) {
		todos, err := conn.GetTodos(c.Request.Context())
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, todos)
	}
}

func createTodoHandler(conn *Connection, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var todo Todo

		if err := c.ShouldBindJSON(&todo); err != nil {
			logger.Warn(
				"validation-error",
				"error", err.Error(),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		err := conn.CreateTodo(c.Request.Context(), todo)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		logger.Info(
			"created_todo",
			"text", todo.Text,
		)

		c.JSON(http.StatusCreated, todo)
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

	r.GET("/todos", fetchTodosHandler(conn))
	r.POST("/todos", createTodoHandler(conn, logger))

	fmt.Println("Server started in port", port)
	r.Run(":" + port)
}
