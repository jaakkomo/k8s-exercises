package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const cacheInterval = 10 * time.Minute

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

func readEnv(env, fallback string) string {
	if value, ok := os.LookupEnv(env); ok {
		return value
	} else {
		return fallback
	}
}

func downloadNewPicture(picture, pictureApi string) {
	res, err := http.Get(pictureApi)
	if err != nil {
		slog.Error("fetching new picture failed",
			"error", err,
		)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		slog.Error("fetching new picture failed",
			"error", res.Status,
		)
		return
	}

	file, err := os.Create(picture)
	if err != nil {
		slog.Error("creating picture file failed",
			"error", err,
		)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		slog.Error("writing to picture file failed",
			"error", err,
		)
		return
	}
}

type Todo struct {
    Text string
}

type TodoClient struct {
	BaseURL string
	Client *http.Client
}

func (tc *TodoClient) CreateTodo(todo Todo) error {
	body, err := json.Marshal(todo)
	if err != nil {
		return err
	}

	res, err := tc.Client.Post(
		tc.BaseURL,
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		return fmt.Errorf("create todo failed: %s", res.Status)
	}

	return nil
}

func (tc *TodoClient) FetchTodos() ([]Todo, error) {
	res, err := tc.Client.Get(tc.BaseURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var todos []Todo
	err = json.NewDecoder(res.Body).Decode(&todos)
	return todos, err
}

func createIndexHandler(tc *TodoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		todos, err := tc.FetchTodos()
		if err != nil {
			slog.Error("fetching todos failed",
				"error", err,
			)
			todos = nil
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"todos": todos,
		})
	}
}

func createTodoHandler(tc *TodoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := tc.CreateTodo(Todo{
			Text: c.PostForm("text"),
		})

		if err != nil {
			c.Status(http.StatusBadGateway)
			return
		}

		c.Redirect(http.StatusSeeOther, "/")
	}
}

func createPictureHandler(picture, pictureApi string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileInfo, err := os.Stat(picture)

		c.Header("Cache-Control", "no-cache")
		if errors.Is(err, os.ErrNotExist) {
			downloadNewPicture(picture, pictureApi)
			c.File(picture)
		} else {
			c.File(picture)
			if time.Since(fileInfo.ModTime()) > cacheInterval {
				go downloadNewPicture(picture, pictureApi)
			}
		}
	}
}

func main() {
	r := gin.Default()

	tmpl := template.Must(
		template.ParseFS(templatesFS, "templates/*"),
	)
	r.SetHTMLTemplate(tmpl)

	staticFolder, _ := fs.Sub(staticFS, "static")
	r.StaticFS("/static", http.FS(staticFolder))

	port := readEnv("PORT", "8080")
	picture := readEnv("PICTURE", "/dev/null")
	pictureApi := readEnv("PICTURE_API", "localhost")
	todosApi := readEnv("TODOS_API", "localhost")

	tc := &TodoClient{
		BaseURL: todosApi,
		Client: &http.Client{},
	}

	r.GET("/", createIndexHandler(tc))
	r.POST("/todos", createTodoHandler(tc))
	r.GET("/picture", createPictureHandler(picture, pictureApi))

	fmt.Println("Server started in port", port)
	r.Run(":" + port)
}
