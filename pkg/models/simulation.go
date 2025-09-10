package models

import (
	"time"
)

type SimulationStatus string

const (
	StatusPending   SimulationStatus = "pending"
	StatusRunning   SimulationStatus = "running"
	StatusCompleted SimulationStatus = "completed"
	StatusFailed    SimulationStatus = "failed"
	StatusStopped   SimulationStatus = "stopped"
)

type Simulation struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	ContainerID string           `json:"container_id"`
	Status      SimulationStatus `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	StartedAt   *time.Time       `json:"started_at,omitempty"`
	FinishedAt  *time.Time       `json:"finished_at,omitempty"`
	Config      SimulationConfig `json:"config"`
	Metrics     *Metrics         `json:"metrics,omitempty"`
}

type SimulationConfig struct {
	ConfigPath  string            `json:"config_path"`
	MetricsPath string            `json:"metrics_path"`
	ServerPath  string            `json:"server_path"`
	Image       string            `json:"image"`
	Environment map[string]string `json:"environment"`
	Volumes     []string          `json:"volumes"`
}

type Metrics struct {
	CPUUsage    float64           `json:"cpu_usage"`
	MemoryUsage float64           `json:"memory_usage"`
	NetworkIO   NetworkStats      `json:"network_io"`
	DiskIO      DiskStats         `json:"disk_io"`
	Custom      map[string]interface{} `json:"custom,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
}

type NetworkStats struct {
	BytesReceived    uint64 `json:"bytes_received"`
	BytesTransmitted uint64 `json:"bytes_transmitted"`
	PacketsReceived  uint64 `json:"packets_received"`
	PacketsTransmitted uint64 `json:"packets_transmitted"`
}

type DiskStats struct {
	BytesRead    uint64 `json:"bytes_read"`
	BytesWritten uint64 `json:"bytes_written"`
}