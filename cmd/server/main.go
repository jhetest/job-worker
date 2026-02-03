package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"job-worker/pkg/worker"
)

var mgr = worker.NewManager()
const AuthToken = "secret-token-123"

type StartRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func main() {

    http.HandleFunc("/jobs/start", authMiddleware(handleStart))
	http.HandleFunc("/jobs/stop", authMiddleware(handleStop))
	http.HandleFunc("/jobs/status", authMiddleware(handleStatus))
	http.HandleFunc("/jobs/output", authMiddleware(handleOutput))

	log.Println("🔐 HTTPS Job Server starting on :8443...")
	// Note: Use generate_cert.go or openssl to create certs for local testing
	err := http.ListenAndServeTLS(":8443", "server.crt", "server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err)
	}
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") || strings.TrimPrefix(authHeader, "Bearer ") != AuthToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// handleStart triggers a new job
func handleStart(w http.ResponseWriter, r *http.Request) {
    log.Printf("to start job:")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	job, err := mgr.StartJob(req.Command, req.Args)
	if err != nil {
		log.Printf("❌ Failed to start job: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Job Started: %s", job.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}


// handleStop terminates a running job
func handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if err := mgr.StopJob(id); err != nil {
		log.Printf("❌ Failed to stop job %s: %v", id, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("Job Stopped: %s", id)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Job %s stopped", id)
}

// handleStatus returns the current state of a job
func handleStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	job, ok := mgr.GetJob(id) // Note: Ensure GetJob is implemented in pkg/worker
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":     job.ID,
		"status": string(job.Status),
	})
}

// handleOutput retrieves the stdout/stderr buffer
func handleOutput(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	output, err := mgr.GetOutput(id)
	if err != nil {
		log.Printf("⚠️ Output request failed for %s: %v", id, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("Serving output for job: %s", id)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(output))
}
