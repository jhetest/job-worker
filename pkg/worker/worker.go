package worker

import (
    "bytes"
	"fmt"
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
	StatusStopped   JobStatus = "stopped"
)

type Job struct {
	ID        string    `json:"id"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	Status    JobStatus `json:"status"`
	StartTime time.Time `json:"start_time"`
    output    *bytes.Buffer
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
	var buf bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stdout = &buf
    cmd.Stderr = &buf

	job := &Job{
		ID:      id,
		Command: command,
		Args:    args,
		Status:  StatusRunning,
		output:  &buf,
		StartTime: time.Now(),
		cmd:     cmd,
	}

	m.jobs[id] = job

    if err := cmd.Start(); err != nil {
		job.Status = StatusFailed
		return nil, err
	}

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

func (m *Manager) StartJob(command string, args []string) (*Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.New().String()
	var buf bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	job := &Job{
		ID:      id,
		Command: command,
		Status:  StatusRunning,
		output:  &buf,
		cmd:     cmd,
	}

	m.jobs[id] = job

	if err := cmd.Start(); err != nil {
		job.Status = StatusFailed
		return nil, err
	}

	go func() {
		err := cmd.Wait()
		job.mu.Lock()
		defer job.mu.Unlock()
		if job.Status == StatusStopped {
			return
		}
		if err != nil {
			job.Status = StatusFailed
		} else {
			job.Status = StatusCompleted
		}
	}()

	return job, nil
}

func (m *Manager) StopJob(id string) error {
	m.mu.RLock()
	job, ok := m.jobs[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("job not found")
	}

	job.mu.Lock()
	defer job.mu.Unlock()

	if job.Status != StatusRunning {
		return fmt.Errorf("job is not running")
	}

	job.Status = StatusStopped
	return job.cmd.Process.Kill()
}


func (m *Manager) GetOutput(id string) (string, error) {
	m.mu.RLock()
	job, ok := m.jobs[id]
	m.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("job not found")
	}
	return job.output.String(), nil
}