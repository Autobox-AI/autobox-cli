package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Autobox-AI/autobox-cli/internal/docker"
	"github.com/Autobox-AI/autobox-cli/pkg/models"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	runImage       string
	runConfig      string
	runMetricsPath string
	runServer      string
	runVolumes     []string
	runEnv         []string
	runName        string
	runDetach      bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a new simulation",
	Long: `Run a new Autobox simulation container with the specified configuration.

Examples:
  autobox run --config simulation.json --metrics metrics.json
  autobox run --image autobox-engine:v1.0 --name "test-simulation"
  autobox run --env OPENAI_API_KEY=sk-... --volume ./configs:/app/configs`,
	RunE: runSimulation,
}

func init() {
	home, _ := os.UserHomeDir()
	defaultVolume := fmt.Sprintf("%s/.autobox/configs:/app/configs", home)

	runCmd.Flags().StringVarP(&runImage, "image", "i", "autobox-engine:latest", "Docker image to use")
	runCmd.Flags().StringVarP(&runConfig, "config", "c", "/app/configs/simulation.json", "Path to simulation config file")
	runCmd.Flags().StringVarP(&runMetricsPath, "metrics", "m", "/app/configs/metrics.json", "Path to metrics config file")
	runCmd.Flags().StringVarP(&runServer, "server", "s", "/app/configs/server.json", "Path to server config file")
	runCmd.Flags().StringSliceVarP(&runVolumes, "volume", "V", []string{defaultVolume}, "Volume mounts (format: host:container)")
	runCmd.Flags().StringSliceVarP(&runEnv, "env", "e", []string{}, "Environment variables (format: KEY=VALUE)")
	runCmd.Flags().StringVarP(&runName, "name", "n", "", "Simulation name")
	runCmd.Flags().BoolVarP(&runDetach, "detach", "d", false, "Run in detached mode")
}

func runSimulation(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	envMap := make(map[string]string)
	for _, env := range runEnv {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// runName will be set later after reading config

	// If using default volume (not explicitly overridden), ensure directories exist
	home, _ := os.UserHomeDir()
	defaultVolume := fmt.Sprintf("%s/.autobox/configs:/app/configs", home)

	if len(runVolumes) == 1 && runVolumes[0] == defaultVolume {
		// Using default volume, ensure directories exist
		configDirs := []string{
			filepath.Join(home, ".autobox", "configs"),
			filepath.Join(home, ".autobox", "configs", "simulations"),
			filepath.Join(home, ".autobox", "configs", "metrics"),
			filepath.Join(home, ".autobox", "configs", "server"),
			filepath.Join(home, ".autobox", "logs"),
		}

		for _, dir := range configDirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}

		// Create default config files if they don't exist
		simulationFile := filepath.Join(home, ".autobox", "configs", "simulation.json")
		if _, err := os.Stat(simulationFile); os.IsNotExist(err) {
			defaultSimConfig := `{
  "name": "default-simulation",
  "agents": [],
  "duration": 3600,
  "output": "/app/logs/results.json"
}`
			if err := os.WriteFile(simulationFile, []byte(defaultSimConfig), 0644); err != nil {
				return fmt.Errorf("failed to create default simulation config: %w", err)
			}
		}

		metricsFile := filepath.Join(home, ".autobox", "configs", "metrics.json")
		if _, err := os.Stat(metricsFile); os.IsNotExist(err) {
			defaultMetricsConfig := `{
  "enabled": true,
  "interval": 60,
  "collectors": ["cpu", "memory", "network", "disk"]
}`
			if err := os.WriteFile(metricsFile, []byte(defaultMetricsConfig), 0644); err != nil {
				return fmt.Errorf("failed to create default metrics config: %w", err)
			}
		}
	}

	// Handle empty volume flag (user wants no volumes)
	volumes := runVolumes
	if len(volumes) == 1 && volumes[0] == "" {
		volumes = []string{}
	}

	// Read simulation name from config file
	simName := ""
	if runConfig != "" {
		// Check if config file is in the container or on host
		configPath := runConfig
		if strings.HasPrefix(runConfig, "/app/configs/") {
			// Config is in container, map to host path
			configPath = filepath.Join(home, ".autobox", "configs", filepath.Base(runConfig))
		}

		// Try to read the config file
		if configData, err := os.ReadFile(configPath); err == nil {
			var config map[string]interface{}
			if err := json.Unmarshal(configData, &config); err == nil {
				if name, ok := config["name"].(string); ok {
					simName = name
				}
			}
		}
	}

	// Use provided name (from --name flag) or extracted name from config or default
	if runName != "" {
		simName = runName // User explicitly provided a name via --name flag
	} else if simName == "" {
		simName = fmt.Sprintf("simulation-%d", os.Getpid()) // fallback to default if no name found
	}

	simConfig := models.SimulationConfig{
		Name:        simName,
		ConfigPath:  runConfig,
		MetricsPath: runMetricsPath,
		ServerPath:  runServer,
		Image:       runImage,
		Environment: envMap,
		Volumes:     volumes,
	}

	fmt.Printf("%s Running simulation...\n", color.YellowString("→"))
	if verbose {
		fmt.Printf("  Image: %s\n", runImage)
		fmt.Printf("  Config: %s\n", runConfig)
		fmt.Printf("  Metrics: %s\n", runMetricsPath)
		fmt.Printf("  Server: %s\n", runServer)
		if len(runVolumes) > 0 {
			fmt.Printf("  Volumes: %s\n", strings.Join(runVolumes, ", "))
		}
	}

	simulation, err := client.LaunchSimulation(ctx, simConfig)
	if err != nil {
		return fmt.Errorf("failed to run simulation: %w", err)
	}

	fmt.Printf("%s Simulation running successfully!\n", color.GreenString("✓"))
	fmt.Printf("  ID: %s\n", color.CyanString(simulation.ID))
	fmt.Printf("  Container: %s\n", simulation.ContainerID[:12])
	fmt.Printf("  Status: %s\n", colorizeStatus(simulation.Status))

	if !runDetach {
		fmt.Printf("\n%s Following logs (press Ctrl+C to detach)...\n\n", color.YellowString("→"))
		return followLogs(ctx, client, simulation.ContainerID)
	}

	return nil
}

func followLogs(ctx context.Context, client *docker.Client, containerID string) error {
	logs, err := client.GetSimulationLogs(ctx, containerID, 100)
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}
	fmt.Print(logs)
	return nil
}

func colorizeStatus(status models.SimulationStatus) string {
	switch status {
	case models.StatusRunning:
		return color.GreenString(string(status))
	case models.StatusCompleted:
		return color.BlueString(string(status))
	case models.StatusFailed:
		return color.RedString(string(status))
	case models.StatusStopped:
		return color.YellowString(string(status))
	default:
		return string(status)
	}
}
