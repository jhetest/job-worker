package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cli <command> [args...]")
		return
	}

	command := os.Args[1]
	args := os.Args[2:]

	// Prepare request
	payload, _ := json.Marshal(map[string]interface{}{
		"command": command,
		"args":    args,
	})

	fmt.Printf("⏳ Submitting job: %s %v...\n", command, args)

	// Send request
	resp, err := http.Post("http://localhost:8080/jobs", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("❌ Connection Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Handle Server Error
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Server returned error (%d): %s\n", resp.StatusCode, string(body))
		return
	}

	// Success
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("❌ Failed to parse server response: %v\n", err)
		return
	}

	fmt.Printf("✨ Success! Job ID: %s\n", result["id"])
	fmt.Printf("📝 Status: %s\n", result["status"])
}