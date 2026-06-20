package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var templatesFS embed.FS

func main() {
	r := gin.Default()

	tmpl := template.Must(
		template.ParseFS(templatesFS, "templates/*"),
	)
	r.SetHTMLTemplate(tmpl)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	port := "8080"
	if value, ok := os.LookupEnv("PORT"); ok {
		port = value
	}

	fmt.Println("Server started in port", port)
	r.Run(":" + port)
}
