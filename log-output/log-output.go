package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const logInterval = 5 * time.Second

func logMessage(msg string) {
	currentTime := time.Now().UTC().Format(time.RFC3339)
	fmt.Printf("%s: %s\n", currentTime, msg)
}

func main() {
	msg := uuid.New().String()

	for {
		logMessage(msg)
		time.Sleep(logInterval)
	}
}
