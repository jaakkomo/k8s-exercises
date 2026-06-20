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

func createIndexHandler (file string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		data, _ := os.ReadFile(file)
		fmt.Fprint(w, string(data))
	}
}

func overwrite(file, content string) {
	tmp := file + ".tmp"
	err := os.WriteFile(tmp, []byte(content), 0644)
	if err != nil {
		panic(err)
	}
	os.Rename(tmp, file)
}

func doLog(file, msg string) {
	for {
		overwrite(file, getStatus(msg))
		time.Sleep(logInterval)
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
	msg := uuid.New().String()

	port := readEnv("PORT", "8080")
	role := readEnv("ROLE", "writer")
	file := readEnv("FILE", "/dev/null")

	switch role {
	case "writer":
		fmt.Println("Started writing to", file)
		doLog(file, msg)
	case "reader":
		fmt.Println("Started reading from", file)
		http.HandleFunc("/", createIndexHandler(file))
		fmt.Println("Server started in port", port)
		http.ListenAndServe(":" + port, nil)
	default:
		fmt.Printf("Invalid env var ROLE=%s\n", role)
	}
}
