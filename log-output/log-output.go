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

func createIndexHandler (logFile, pongFile string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logData, _ := os.ReadFile(logFile)
		pongData, _ := os.ReadFile(pongFile)
		fmt.Fprint(w, string(logData))
		fmt.Fprintf(w, "Ping / Pongs: %s", pongData)
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
	logFile := readEnv("LOG_FILE", "/dev/null")
	pongFile := readEnv("PONG_FILE", "/dev/null")

	switch role {
	case "writer":
		fmt.Println("Started writing to", logFile)
		doLog(logFile, msg)
	case "reader":
		fmt.Println("Started reading")
		fmt.Println("Log file:", logFile)
		fmt.Println("Ping pong file:", pongFile)
		http.HandleFunc("/", createIndexHandler(logFile, pongFile))
		fmt.Println("Server started in port", port)
		http.ListenAndServe(":" + port, nil)
	default:
		fmt.Printf("Invalid env var ROLE=%s\n", role)
	}
}
