package main

import (
    "crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"job-worker/pkg/worker"
)

var mgr = worker.NewManager()
const AuthToken = "secret-token-123"

type StartRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func main() {
    // 1. Load CA certificate to verify clients
	caCert, _ := ioutil.ReadFile("ca.crt")
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

    // 2. Configure TLS settings
	tlsConfig := &tls.Config{
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert, // Force mTLS
		MinVersion:   tls.VersionTLS13,               // Only modern TLS
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
	}

	server := &http.Server{
    		Addr:      ":8443",
    		TLSConfig: tlsConfig,
    	}

    http.HandleFunc("/jobs/start", authMiddleware(handleStart))
	http.HandleFunc("/jobs/stop", authMiddleware(handleStop))
	http.HandleFunc("/jobs/status", authMiddleware(handleStatus))
	http.HandleFunc("/jobs/output", authMiddleware(handleOutput))

	log.Println("🛡️ Hardened mTLS Server starting on :8443...")
	log.Fatal(server.ListenAndServeTLS("server.crt", "server.key"))
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Verify we have peer certificates (mTLS ensured this, but safety first)
		if len(r.TLS.PeerCertificates) == 0 {
			http.Error(w, "No client certificate provided", http.StatusUnauthorized)
			return
		}

		// 2. Extract the Common Name
		clientCN := r.TLS.PeerCertificates[0].Subject.CommonName
		log.Printf("🔍 Authorization check for: %s", clientCN)

		// 3. Simple Authorization Scheme
		// In a real app, this might check a database or config file
		allowedClients := map[string]bool{
			"admin-client": true,
		}

		if !allowedClients[clientCN] {
			log.Printf("🚫 Access Denied: %s is not authorized", clientCN)
			http.Error(w, "Forbidden: Unauthorized client identity", http.StatusForbidden)
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
