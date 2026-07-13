package main

import (
	"fmt"
	"net/http"
	"os"
)

func readEnv(env, fallback string) string {
	if value, ok := os.LookupEnv(env); ok {
		return value
	} else {
		return fallback
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from version 2")
	})

	port := readEnv("PORT", "8080")

	fmt.Println("Server started in port", port)
	http.ListenAndServe(":"+port, nil)
}
