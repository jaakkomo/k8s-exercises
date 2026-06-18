package main

import (
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
)

var counter atomic.Uint64

func indexHandler(w http.ResponseWriter, r *http.Request) {
	counterValue := counter.Add(1) - 1
	fmt.Fprintf(w, "pong %d\n", counterValue)
}

func main() {
	port := "8080"
	if value, ok := os.LookupEnv("PORT"); ok {
		port = value
	}

	http.HandleFunc("/", indexHandler)
	fmt.Println("Server started in port", port)
	http.ListenAndServe(":" + port, nil)
}
