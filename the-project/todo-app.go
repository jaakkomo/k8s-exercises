package main

import (
	"embed"
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

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
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

	r.GET("/", indexHandler)
	r.GET("/picture", createPictureHandler(picture, pictureApi))

	fmt.Println("Server started in port", port)
	r.Run(":" + port)
}
