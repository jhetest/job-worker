package worker

import (
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
)

type Job struct {
	ID        string    `json:"id"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	Status    JobStatus `json:"status"`
	StartTime time.Time `json:"start_time"`
	cmd       *exec.Cmd
	mu        sync.RWMutex
}

type Manager struct {
	jobs map[string]*Job
	mu   sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{jobs: make(map[string]*Job)}
}

func (m *Manager) CreateJob(command string, args []string) (*Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.New().String()
	cmd := exec.Command(command, args...)

	job := &Job{
		ID:      id,
		Command: command,
		Args:    args,
		Status:  StatusRunning,
		StartTime: time.Now(),
		cmd:     cmd,
	}

	m.jobs[id] = job

	// Execute in background
	go func() {
		err := cmd.Run()
		job.mu.Lock()
		defer job.mu.Unlock()
		if err != nil {
			job.Status = StatusFailed
		} else {
			job.Status = StatusCompleted
		}
	}()

	return job, nil
}

func (m *Manager) GetJob(id string) (*Job, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	job, ok := m.jobs[id]
	return job, ok
}