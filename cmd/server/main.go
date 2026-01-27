package main

import (
	"encoding/json"
	"log"
	"net/http"
	"job-worker/pkg/worker"
)

var mgr = worker.NewManager()

type StartRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func main() {
	http.HandleFunc("/jobs", handleJobs)
    log.Println("🚀 Job Worker Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleJobs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req StartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		    log.Printf("❌ Error decoding request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		job, err := mgr.CreateJob(req.Command, req.Args)
        if err != nil {
			log.Printf("❌ Failed to create job for %s: %v", req.Command, err)
			http.Error(w, "Failed to create job", http.StatusInternalServerError)
			return
		}
		log.Printf("✅ Job Created: %s [Cmd: %s]", job.ID, job.Command)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)

	case http.MethodGet:
		id := r.URL.Query().Get("id")
		job, ok := mgr.GetJob(id)
		if !ok {
		    log.Printf("⚠️  Job not found: %s", id)
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(job)
	}
}
