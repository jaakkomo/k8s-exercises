package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Todo struct {
    Text string
}

var todos = []Todo{
	{"Learn Kubernetes"},
	{"Learn Go"},
	{"Learn parallel computing"},
}

func readEnv(env, fallback string) string {
	if value, ok := os.LookupEnv(env); ok {
		return value
	} else {
		return fallback
	}
}

func fetchTodosHandler(c *gin.Context) {
	c.JSON(http.StatusOK, todos)
}

func createTodoHandler(c *gin.Context) {
	var todo Todo

	if err := c.BindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	fmt.Printf("created: %s", todo.Text)

	todos = append(todos, todo)

	c.JSON(http.StatusCreated, todo)
}

func main() {
	r := gin.Default()

	port := readEnv("PORT", "8080")

	r.GET("/todos", fetchTodosHandler)
	r.POST("/todos", createTodoHandler)

	fmt.Println("Server started in port", port)
	r.Run(":" + port)
}
