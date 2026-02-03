package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	baseURL   = "https://localhost:8443"
	authToken = "secret-token-123"
)

func getJob(client *http.Client, id string) {
	req, _ := http.NewRequest("GET", baseURL+"/jobs/output?id="+id, nil)
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Connection Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Error (%d): %s\n", resp.StatusCode, string(body))
		return
	}

	fmt.Println("--- Job Output ---")
	io.Copy(os.Stdout, resp.Body)
	fmt.Println("\n--- End of Output ---")
}

func startJob(client *http.Client, command string, args []string) {
	// Prepare the JSON payload
	requestData := map[string]interface{}{
		"command": command,
		"args":    args,
	}
	payload, _ := json.Marshal(requestData)

	// Create request
	req, _ := http.NewRequest("POST", baseURL+"/jobs/start", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Error starting job: %s\n", string(body))
		return
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("🚀 Job started successfully! ID: %s\n", result["id"])
}

func stopJob(client *http.Client, id string) {
	// The stop endpoint expects the ID as a query parameter
	url := fmt.Sprintf("%s/jobs/stop?id=%s", baseURL, id)

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("🛑 Job %s has been terminated.\n", id)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Failed to stop job: %s\n", string(body))
	}
}

func fetchLogs(client *http.Client, id string) {
	req, _ := http.NewRequest("GET", baseURL+"/jobs/output?id="+id, nil)
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Connection Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Error (%d): %s\n", resp.StatusCode, string(body))
		return
	}

	fmt.Println("--- Job Output ---")
	io.Copy(os.Stdout, resp.Body)
	fmt.Println("\n--- End of Output ---")
}



/**
Quick Test commands for your CLI:
Start a long job: go run cmd/cli/main.go start sleep 10

Check Status: go run cmd/cli/main.go status <ID>

Get Output: go run cmd/cli/main.go logs <ID>

Kill Job: go run cmd/cli/main.go stop <ID>

*/

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cli <start|stop|status|logs> [args...]")
		return
	}

	command := os.Args[1]
    fmt.Println("cmd: ", command)
	// Setup custom client to handle self-signed certs for local dev
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
    fmt.Println("cmd: ", command)
	switch command {
    case "start":
		if len(os.Args) < 3 {
			fmt.Println("Usage: cli start <cmd> [args...]")
			return
		}
		fmt.Println("start...")
		startJob(client, os.Args[2], os.Args[3:])

	case "stop":
		if len(os.Args) < 3 {
			fmt.Println("Usage: cli stop <job_id>")
			return
		}
		fmt.Println("stop...")
		stopJob(client, os.Args[2])

    case "status":
        fmt.Println("job status")
		if len(os.Args) < 3 {
			fmt.Println("Usage: cli get <job_id>")
			return
		}
		getJob(client, os.Args[2])
	case "logs":
		if len(os.Args) < 3 {
			fmt.Println("Usage: cli logs <job_id>")
			return
		}
		fetchLogs(client, os.Args[2])
	// ... other cases (start/stop) would follow a similar pattern
	default:
		fmt.Println("Unknown command")
	}
}
