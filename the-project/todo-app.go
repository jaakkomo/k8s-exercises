package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := "8080"
	if value, ok := os.LookupEnv("PORT"); ok {
		port = value
	}
	fmt.Println("Server started in port", port)
	http.ListenAndServe(":" + port, nil)
}
