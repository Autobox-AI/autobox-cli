package cmd

import (
	"context"
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
	launchImage      string
	launchConfig     string
	launchMetrics    string
	launchVolumes    []string
	launchEnv        []string
	launchName       string
	launchDetach     bool
)

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch a new simulation",
	Long: `Launch a new Autobox simulation container with the specified configuration.
	
Examples:
  autobox launch --config simulation.json --metrics metrics.json
  autobox launch --image autobox-engine:v1.0 --name "test-simulation"
  autobox launch --env OPENAI_API_KEY=sk-... --volume ./configs:/app/configs`,
	RunE: runLaunch,
}

func init() {
	// Get home directory for default volume
	home, _ := os.UserHomeDir()
	defaultVolume := fmt.Sprintf("%s/.autobox/configs:/app/configs", home)
	
	launchCmd.Flags().StringVarP(&launchImage, "image", "i", "autobox-engine:latest", "Docker image to use")
	launchCmd.Flags().StringVarP(&launchConfig, "config", "c", "/app/configs/simulation.json", "Path to simulation config file")
	launchCmd.Flags().StringVarP(&launchMetrics, "metrics", "m", "/app/configs/metrics.json", "Path to metrics config file")
	launchCmd.Flags().StringSliceVarP(&launchVolumes, "volume", "V", []string{defaultVolume}, "Volume mounts (format: host:container)")
	launchCmd.Flags().StringSliceVarP(&launchEnv, "env", "e", []string{}, "Environment variables (format: KEY=VALUE)")
	launchCmd.Flags().StringVarP(&launchName, "name", "n", "", "Simulation name")
	launchCmd.Flags().BoolVarP(&launchDetach, "detach", "d", false, "Run in detached mode")
}

func runLaunch(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	
	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	envMap := make(map[string]string)
	for _, env := range launchEnv {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	if launchName == "" {
		launchName = fmt.Sprintf("simulation-%d", os.Getpid())
	}

	// If using default volume (not explicitly overridden), ensure directories exist
	home, _ := os.UserHomeDir()
	defaultVolume := fmt.Sprintf("%s/.autobox/configs:/app/configs", home)
	
	if len(launchVolumes) == 1 && launchVolumes[0] == defaultVolume {
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
	volumes := launchVolumes
	if len(volumes) == 1 && volumes[0] == "" {
		volumes = []string{}
	}

	simConfig := models.SimulationConfig{
		ConfigPath:  launchConfig,
		MetricsPath: launchMetrics,
		Image:       launchImage,
		Environment: envMap,
		Volumes:     volumes,
	}

	fmt.Printf("%s Launching simulation...\n", color.YellowString("→"))
	if verbose {
		fmt.Printf("  Image: %s\n", launchImage)
		fmt.Printf("  Config: %s\n", launchConfig)
		fmt.Printf("  Metrics: %s\n", launchMetrics)
		if len(launchVolumes) > 0 {
			fmt.Printf("  Volumes: %s\n", strings.Join(launchVolumes, ", "))
		}
	}

	simulation, err := client.LaunchSimulation(ctx, simConfig)
	if err != nil {
		return fmt.Errorf("failed to launch simulation: %w", err)
	}

	fmt.Printf("%s Simulation launched successfully!\n", color.GreenString("✓"))
	fmt.Printf("  ID: %s\n", color.CyanString(simulation.ID))
	fmt.Printf("  Container: %s\n", simulation.ContainerID[:12])
	fmt.Printf("  Status: %s\n", colorizeStatus(simulation.Status))

	if !launchDetach {
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