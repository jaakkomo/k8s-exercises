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

type App struct {
	logFile  string
	pingsApi string
	file     string
	message  string
}

func (app *App) Index(w http.ResponseWriter, req *http.Request) {
	fileData, _ := os.ReadFile(app.file)
	logData, _ := os.ReadFile(app.logFile)
	res, err := http.Get(app.pingsApi)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	fmt.Fprint(w, string(logData))
	fmt.Fprint(w, "Ping / Pongs: ")
	io.Copy(w, res.Body)
	fmt.Fprintf(w, "\nenv variable: MESSAGE=%s\n", app.message)
	fmt.Fprintln(w, "file content:")
	fmt.Fprint(w, string(fileData))
}

func (app *App) Health(w http.ResponseWriter, req *http.Request) {
	res, err := http.Get(app.pingsApi)
	if err != nil || res.StatusCode >= 400 {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
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
	logFile := readEnv("LOG_FILE", "/tmp/log-output")
	pingsApi := readEnv("PINGS_API", "http://localhost:8083/pings")
	file := readEnv("FILE", "/dev/null")
	message := readEnv("MESSAGE", "/dev/null")

	app := App{
		logFile,
		pingsApi,
		file,
		message,
	}

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
		http.HandleFunc("/", app.Index)
		http.HandleFunc("/healthz", app.Health)
		fmt.Println("Server started in port", port)
		http.ListenAndServe(":"+port, nil)
	default:
		fmt.Printf("Invalid env var ROLE=%s\n", role)
	}
}
