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
	ID   int64
	Text string
	Done bool
}

type TodoClient struct {
	BaseURL string
	Client  *http.Client
}

func (tc *TodoClient) CreateTodo(todo Todo) error {
	body, err := json.Marshal(todo)
	if err != nil {
		return err
	}

	res, err := tc.Client.Post(
		tc.BaseURL+"/todos",
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

func (tc *TodoClient) MarkDone(id string) error {
	url := fmt.Sprintf("%s/todos/%s", tc.BaseURL, id)
	req, _ := http.NewRequest(http.MethodPut, url, nil)
	res, err := tc.Client.Do(req)

	if err != nil {
		return err
	}

	if res.StatusCode >= 300 {
		return fmt.Errorf("mark done failed: %s", res.Status)
	}

	return nil
}

func (tc *TodoClient) FetchTodos() ([]Todo, error) {
	res, err := tc.Client.Get(tc.BaseURL + "/todos")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, errors.New("request status not good")
	}

	var todos []Todo
	err = json.NewDecoder(res.Body).Decode(&todos)
	return todos, err
}

func (tc *TodoClient) Break() {
	res, err := tc.Client.Post(
		tc.BaseURL+"/break",
		"application/json",
		bytes.NewReader([]byte{}),
	)
	if err == nil {
		res.Body.Close()
	}
}

func createIndexHandler(tc *TodoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		todos, err := tc.FetchTodos()
		if err != nil {
			slog.Error("fetching todos failed",
				"error", err,
			)
			c.HTML(http.StatusOK, "unhealthy.html", nil)
			return
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

func markDoneHandler(tc *TodoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := tc.MarkDone(c.Param("id"))

		if err != nil {
			c.Status(http.StatusBadGateway)
			return
		}

		c.Redirect(http.StatusSeeOther, "/")
	}
}

func breakHandler(tc *TodoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		tc.Break()
		c.Redirect(http.StatusSeeOther, "/")
	}
}

func createPictureHandler(picture, pictureApi string, cacheInterval time.Duration) gin.HandlerFunc {
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
	picture := readEnv("PICTURE", "/tmp/todo-app.jpg")
	pictureApi := readEnv("PICTURE_API", "https://picsum.photos/600")
	todosApi := readEnv("TODOS_API", "http://localhost:8084")
	cacheIntervalString := readEnv("CACHE_INTERVAL", "10m")
	cacheInterval, err := time.ParseDuration(cacheIntervalString)
	if err != nil {
		slog.Error("parse cache interval failed",
			"error", err,
		)
		return
	}

	tc := &TodoClient{
		BaseURL: todosApi,
		Client:  &http.Client{},
	}

	r.GET("/", createIndexHandler(tc))
	r.POST("/todos", createTodoHandler(tc))
	r.POST("/todos/:id", markDoneHandler(tc))
	r.GET("/picture", createPictureHandler(picture, pictureApi, cacheInterval))
	r.POST("/break", breakHandler(tc))

	fmt.Println("Server started in port", port)
	r.Run(":" + port)
}
