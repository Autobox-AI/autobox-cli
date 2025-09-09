package models

import (
	"testing"
	"time"
)

func TestSimulationStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   SimulationStatus
		expected string
	}{
		{"Pending", StatusPending, "pending"},
		{"Running", StatusRunning, "running"},
		{"Completed", StatusCompleted, "completed"},
		{"Failed", StatusFailed, "failed"},
		{"Stopped", StatusStopped, "stopped"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("got %s, want %s", tt.status, tt.expected)
			}
		})
	}
}

func TestSimulationStruct(t *testing.T) {
	now := time.Now()
	sim := &Simulation{
		ID:          "test-123",
		Name:        "test-simulation",
		ContainerID: "abc123def456",
		Status:      StatusRunning,
		CreatedAt:   now,
		StartedAt:   &now,
		Config: SimulationConfig{
			ConfigPath:  "/app/config.json",
			MetricsPath: "/app/metrics.json",
			Image:       "autobox-engine:latest",
		},
	}

	if sim.ID != "test-123" {
		t.Errorf("ID: got %s, want test-123", sim.ID)
	}

	if sim.Status != StatusRunning {
		t.Errorf("Status: got %s, want %s", sim.Status, StatusRunning)
	}

	if sim.Config.Image != "autobox-engine:latest" {
		t.Errorf("Image: got %s, want autobox-engine:latest", sim.Config.Image)
	}
}

func TestMetricsStruct(t *testing.T) {
	metrics := &Metrics{
		CPUUsage:    45.5,
		MemoryUsage: 62.3,
		NetworkIO: NetworkStats{
			BytesReceived:    1024,
			BytesTransmitted: 2048,
		},
		DiskIO: DiskStats{
			BytesRead:    4096,
			BytesWritten: 8192,
		},
		Timestamp: time.Now(),
	}

	if metrics.CPUUsage != 45.5 {
		t.Errorf("CPUUsage: got %f, want 45.5", metrics.CPUUsage)
	}

	if metrics.NetworkIO.BytesReceived != 1024 {
		t.Errorf("BytesReceived: got %d, want 1024", metrics.NetworkIO.BytesReceived)
	}
}