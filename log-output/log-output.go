package main

import (
	"fmt"
	"io"
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

func createIndexHandler (logFile, pingsApi, file, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fileData, _ := os.ReadFile(file)
		logData, _ := os.ReadFile(logFile)
		res, err := http.Get(pingsApi)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		fmt.Fprintf(w, "file content: %s", string(fileData))
		fmt.Fprintf(w, "env variable: MESSAGE=%s\n", message)
		fmt.Fprint(w, string(logData))
		fmt.Fprint(w, "Ping / Pongs: ")
		io.Copy(w, res.Body)
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
	pingsApi := readEnv("PINGS_API", "localhost")
	file := readEnv("FILE", "/dev/null")
	message := readEnv("MESSAGE", "/dev/null")

	switch role {
	case "writer":
		fmt.Println("Started writing to", logFile)
		doLog(logFile, msg)
	case "reader":
		fmt.Println("Started reading")
		fmt.Println("Log file:", logFile)
		fmt.Println("File:", file)
		fmt.Println("Message:", message)
		fmt.Println("Pings API:", pingsApi)
		http.HandleFunc("/", createIndexHandler(logFile, pingsApi, file, message))
		fmt.Println("Server started in port", port)
		http.ListenAndServe(":" + port, nil)
	default:
		fmt.Printf("Invalid env var ROLE=%s\n", role)
	}
}
