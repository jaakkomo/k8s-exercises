package main

import (
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
)

var counter atomic.Uint64

func overwrite(file, content string) {
	tmp := file + ".tmp"
	err := os.WriteFile(tmp, []byte(content), 0644)
	if err != nil {
		panic(err)
	}
	os.Rename(tmp, file)
}

func createIndexHandler(file string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newCounter := counter.Add(1)
		oldCounter := newCounter - 1
		fmt.Fprintf(w, "pong %d\n", oldCounter)
		overwrite(file, fmt.Sprintf("%d\n",newCounter))
	}
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
	file := readEnv("FILE", "/dev/null")

	overwrite(file, "0\n")
	http.HandleFunc("/", createIndexHandler(file))
	fmt.Println("Server started in port", port)
	http.ListenAndServe(":" + port, nil)
}
