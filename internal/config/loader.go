package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SimulationConfigSet struct {
	Name           string                 `json:"name"`
	SimulationPath string                 `json:"simulation_path"`
	MetricsPath    string                 `json:"metrics_path"`
	ServerPath     string                 `json:"server_path"`
	Simulation     map[string]interface{} `json:"simulation"`
	Metrics        interface{}            `json:"metrics"` // Can be map or array
	Server         map[string]interface{} `json:"server"`
}

func LoadSimulationConfig(simulationName string) (*SimulationConfigSet, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configBase := filepath.Join(home, ".autobox", "config")

	fileName := strings.ToLower(strings.ReplaceAll(simulationName, "-", "_"))
	if !strings.HasSuffix(fileName, ".json") {
		fileName = fileName + ".json"
	}

	configSet := &SimulationConfigSet{
		Name: simulationName,
	}

	simPath := filepath.Join(configBase, "simulations", fileName)
	configSet.SimulationPath = simPath
	if simData, err := os.ReadFile(simPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("simulation config not found: %s", fileName)
		}
		return nil, fmt.Errorf("failed to read simulation config: %w", err)
	} else {
		if err := json.Unmarshal(simData, &configSet.Simulation); err != nil {
			return nil, fmt.Errorf("failed to parse simulation config: %w", err)
		}
	}

	metricsPath := filepath.Join(configBase, "metrics", fileName)
	configSet.MetricsPath = metricsPath
	if metricsData, err := os.ReadFile(metricsPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("metrics config not found for simulation '%s': %s", simulationName, fileName)
		}
		return nil, fmt.Errorf("failed to read metrics config: %w", err)
	} else {
		var metricsInterface interface{}
		if err := json.Unmarshal(metricsData, &metricsInterface); err != nil {
			return nil, fmt.Errorf("failed to parse metrics config: %w", err)
		}
		configSet.Metrics = metricsInterface
	}

	serverPath := filepath.Join(configBase, "default.json")
	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		serverPath = filepath.Join(configBase, "server.json")
	}
	configSet.ServerPath = serverPath
	if serverData, err := os.ReadFile(serverPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read server config: %w", err)
		}
	} else {
		if err := json.Unmarshal(serverData, &configSet.Server); err != nil {
			return nil, fmt.Errorf("failed to parse server config: %w", err)
		}
	}

	return configSet, nil
}

func ListAvailableSimulations() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	simDir := filepath.Join(home, ".autobox", "config", "simulations")
	metricsDir := filepath.Join(home, ".autobox", "config", "metrics")

	simFiles, err := os.ReadDir(simDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read simulations directory: %w", err)
	}

	metricsFiles, err := os.ReadDir(metricsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read metrics directory: %w", err)
	}

	metricsMap := make(map[string]bool)
	for _, f := range metricsFiles {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
			metricsMap[f.Name()] = true
		}
	}

	var simulations []string
	for _, f := range simFiles {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
			if metricsMap[f.Name()] {
				name := strings.TrimSuffix(f.Name(), ".json")
				simulations = append(simulations, name)
			}
		}
	}

	return simulations, nil
}

func ValidateSimulationConfig(simulationName string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configBase := filepath.Join(home, ".autobox", "config")
	fileName := strings.ToLower(strings.ReplaceAll(simulationName, "-", "_"))
	if !strings.HasSuffix(fileName, ".json") {
		fileName = fileName + ".json"
	}

	simPath := filepath.Join(configBase, "simulations", fileName)
	if _, err := os.Stat(simPath); os.IsNotExist(err) {
		return fmt.Errorf("simulation config not found: %s", fileName)
	}

	metricsPath := filepath.Join(configBase, "metrics", fileName)
	if _, err := os.Stat(metricsPath); os.IsNotExist(err) {
		return fmt.Errorf("metrics config not found: %s (simulation and metrics configs must have matching names)", fileName)
	}

	return nil
}

func EnsureConfigDirectories() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	dirs := []string{
		filepath.Join(home, ".autobox", "config"),
		filepath.Join(home, ".autobox", "config", "simulations"),
		filepath.Join(home, ".autobox", "config", "metrics"),
		filepath.Join(home, ".autobox", "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
