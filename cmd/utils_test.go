package cmd

import (
	"testing"
	"time"

	"github.com/Autobox-AI/autobox-cli/pkg/models"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{"Short string", "hello", 10, "hello"},
		{"Exact length", "hello", 5, "hello"},
		{"Long string", "hello world", 8, "hello..."},
		{"Very long string", "this is a very long string", 10, "this is..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.max)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, result, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"Seconds only", 45 * time.Second, "45s"},
		{"Minutes and seconds", 2*time.Minute + 30*time.Second, "2m 30s"},
		{"Hours and minutes", 3*time.Hour + 15*time.Minute, "3h 15m"},
		{"Full duration", 2*time.Hour + 45*time.Minute + 30*time.Second, "2h 45m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    uint64
		expected string
	}{
		{"Bytes", 512, "512 B"},
		{"Kilobytes", 2048, "2.00 KB"},
		{"Megabytes", 5 * 1024 * 1024, "5.00 MB"},
		{"Gigabytes", 2 * 1024 * 1024 * 1024, "2.00 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestCountByStatus(t *testing.T) {
	simulations := []*models.Simulation{
		{Status: models.StatusRunning},
		{Status: models.StatusRunning},
		{Status: models.StatusCompleted},
		{Status: models.StatusFailed},
		{Status: models.StatusRunning},
	}

	tests := []struct {
		name     string
		status   models.SimulationStatus
		expected int
	}{
		{"Count running", models.StatusRunning, 3},
		{"Count completed", models.StatusCompleted, 1},
		{"Count failed", models.StatusFailed, 1},
		{"Count stopped", models.StatusStopped, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countByStatus(simulations, tt.status)
			if result != tt.expected {
				t.Errorf("countByStatus() for %s = %d, want %d", tt.status, result, tt.expected)
			}
		})
	}
}

func TestFilterRunningSimulations(t *testing.T) {
	simulations := []*models.Simulation{
		{ID: "1", Status: models.StatusRunning},
		{ID: "2", Status: models.StatusCompleted},
		{ID: "3", Status: models.StatusRunning},
		{ID: "4", Status: models.StatusFailed},
		{ID: "5", Status: models.StatusStopped},
	}

	result := filterRunningSimulations(simulations)

	if len(result) != 2 {
		t.Errorf("filterRunningSimulations: got %d simulations, want 2", len(result))
	}

	for _, sim := range result {
		if sim.Status != models.StatusRunning {
			t.Errorf("filterRunningSimulations: got status %s, want %s", sim.Status, models.StatusRunning)
		}
	}
}