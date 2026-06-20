package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

const logInterval = 5 * time.Second

func getStatus(msg string) string {
	currentTime := time.Now().UTC().Format(time.RFC3339)
	return fmt.Sprintf("%s: %s\n", currentTime, msg)
}

func createIndexHandler (msg string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, getStatus(msg))
	}
}

func doLog(msg string) {
	for {
		fmt.Print(getStatus(msg))
		time.Sleep(logInterval)
	}
}

func main() {
	msg := uuid.New().String()

	port := "8080"
	if value, ok := os.LookupEnv("PORT"); ok {
		port = value
	}
	http.HandleFunc("/", createIndexHandler(msg))

	go doLog(msg)
	fmt.Println("Server started in port", port)
	http.ListenAndServe(":" + port, nil)
}
