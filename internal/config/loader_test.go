package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSimulationConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "autobox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configBase := filepath.Join(tmpDir, ".autobox", "config")
	simDir := filepath.Join(configBase, "simulations")
	metricsDir := filepath.Join(configBase, "metrics")

	if err := os.MkdirAll(simDir, 0755); err != nil {
		t.Fatalf("Failed to create simulations dir: %v", err)
	}
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		t.Fatalf("Failed to create metrics dir: %v", err)
	}

	simConfig := map[string]interface{}{
		"name":     "gift_choice",
		"agents":   []string{"agent1", "agent2"},
		"duration": 3600,
	}
	simData, _ := json.Marshal(simConfig)
	if err := os.WriteFile(filepath.Join(simDir, "gift_choice.json"), simData, 0644); err != nil {
		t.Fatalf("Failed to write simulation config: %v", err)
	}

	metricsConfig := map[string]interface{}{
		"enabled":    true,
		"interval":   60,
		"collectors": []string{"cpu", "memory"},
	}
	metricsData, _ := json.Marshal(metricsConfig)
	if err := os.WriteFile(filepath.Join(metricsDir, "gift_choice.json"), metricsData, 0644); err != nil {
		t.Fatalf("Failed to write metrics config: %v", err)
	}

	serverConfig := map[string]interface{}{
		"host": "localhost",
		"port": 8080,
	}
	serverData, _ := json.Marshal(serverConfig)
	if err := os.WriteFile(filepath.Join(configBase, "default.json"), serverData, 0644); err != nil {
		t.Fatalf("Failed to write server config: %v", err)
	}

	configSet, err := LoadSimulationConfig("gift_choice")
	if err != nil {
		t.Fatalf("Failed to load simulation config: %v", err)
	}

	if configSet.Name != "gift_choice" {
		t.Errorf("Expected name 'gift_choice', got '%s'", configSet.Name)
	}

	if configSet.Simulation["name"] != "gift_choice" {
		t.Errorf("Expected simulation name 'gift_choice', got '%v'", configSet.Simulation["name"])
	}

	metricsMap, ok := configSet.Metrics.(map[string]interface{})
	if !ok {
		t.Errorf("Expected metrics to be a map, got %T", configSet.Metrics)
	} else if !metricsMap["enabled"].(bool) {
		t.Errorf("Expected metrics to be enabled")
	}

	if configSet.Server["port"].(float64) != 8080 {
		t.Errorf("Expected server port 8080, got %v", configSet.Server["port"])
	}
}

func TestValidateSimulationConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "autobox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configBase := filepath.Join(tmpDir, ".autobox", "config")
	simDir := filepath.Join(configBase, "simulations")
	metricsDir := filepath.Join(configBase, "metrics")

	if err := os.MkdirAll(simDir, 0755); err != nil {
		t.Fatalf("Failed to create simulations dir: %v", err)
	}
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		t.Fatalf("Failed to create metrics dir: %v", err)
	}

	simData := []byte(`{"name": "test_sim"}`)
	metricsData := []byte(`{"enabled": true}`)

	if err := os.WriteFile(filepath.Join(simDir, "test_sim.json"), simData, 0644); err != nil {
		t.Fatalf("Failed to write simulation config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(metricsDir, "test_sim.json"), metricsData, 0644); err != nil {
		t.Fatalf("Failed to write metrics config: %v", err)
	}

	if err := ValidateSimulationConfig("test_sim"); err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}

	if err := os.WriteFile(filepath.Join(simDir, "no_metrics.json"), simData, 0644); err != nil {
		t.Fatalf("Failed to write simulation config: %v", err)
	}

	if err := ValidateSimulationConfig("no_metrics"); err == nil {
		t.Errorf("Expected validation to fail for missing metrics config")
	}

	if err := os.WriteFile(filepath.Join(metricsDir, "no_sim.json"), metricsData, 0644); err != nil {
		t.Fatalf("Failed to write metrics config: %v", err)
	}

	if err := ValidateSimulationConfig("no_sim"); err == nil {
		t.Errorf("Expected validation to fail for missing simulation config")
	}
}

func TestListAvailableSimulations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "autobox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configBase := filepath.Join(tmpDir, ".autobox", "config")
	simDir := filepath.Join(configBase, "simulations")
	metricsDir := filepath.Join(configBase, "metrics")

	if err := os.MkdirAll(simDir, 0755); err != nil {
		t.Fatalf("Failed to create simulations dir: %v", err)
	}
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		t.Fatalf("Failed to create metrics dir: %v", err)
	}

	testSims := []string{"gift_choice", "holiday_planning", "budget_allocation"}
	for _, sim := range testSims {
		simData := []byte(`{"name": "` + sim + `"}`)
		metricsData := []byte(`{"enabled": true}`)

		if err := os.WriteFile(filepath.Join(simDir, sim+".json"), simData, 0644); err != nil {
			t.Fatalf("Failed to write simulation config: %v", err)
		}
		if err := os.WriteFile(filepath.Join(metricsDir, sim+".json"), metricsData, 0644); err != nil {
			t.Fatalf("Failed to write metrics config: %v", err)
		}
	}

	if err := os.WriteFile(filepath.Join(simDir, "orphan_sim.json"), []byte(`{"name": "orphan"}`), 0644); err != nil {
		t.Fatalf("Failed to write orphan simulation config: %v", err)
	}

	simulations, err := ListAvailableSimulations()
	if err != nil {
		t.Fatalf("Failed to list simulations: %v", err)
	}

	if len(simulations) != 3 {
		t.Errorf("Expected 3 simulations, got %d", len(simulations))
	}

	simMap := make(map[string]bool)
	for _, sim := range simulations {
		simMap[sim] = true
	}

	for _, expected := range testSims {
		if !simMap[expected] {
			t.Errorf("Expected simulation '%s' not found in list", expected)
		}
	}

	if simMap["orphan_sim"] {
		t.Errorf("Orphan simulation should not be listed")
	}
}
