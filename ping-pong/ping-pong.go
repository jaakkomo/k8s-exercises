package main

import (
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
)

var counter atomic.Uint64

func indexHandler(w http.ResponseWriter, r *http.Request) {
	newCounter := counter.Add(1)
	oldCounter := newCounter - 1
	fmt.Fprintf(w, "pong %d\n", oldCounter)
}

func pingsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%d", counter.Load())
}

func readEnv(env, fallback string) string {
	if value, ok := os.LookupEnv(env); ok {
		return value
	} else {
		return fallback
	}
}

func main() {
	port := readEnv("PORT", "8080")

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/pings", pingsHandler)
	fmt.Println("Server started in port", port)
	http.ListenAndServe(":" + port, nil)
}
