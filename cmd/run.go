package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Autobox-AI/autobox-cli/internal/config"
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
	runListSims    bool
)

var runCmd = &cobra.Command{
	Use:   "run [simulation-name]",
	Short: "Run a new simulation",
	Long: `Run a new Autobox simulation container with the specified configuration.

You can either provide a simulation name to use pre-configured settings from ~/.autobox/config/,
or specify configuration files directly using flags.

Examples:
  # Run a named simulation (loads from ~/.autobox/config/simulations/ and metrics/)
  autobox run gift_choice
  autobox run holiday_planning

  # Run with custom config files
  autobox run --config simulation.json --metrics metrics.json

  # Run with custom image and environment
  autobox run --image autobox-engine:v1.0 --name "test-simulation"
  autobox run --env OPENAI_API_KEY=sk-... --volume ./config:/app/config

  # List available simulations
  autobox run --list`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSimulation,
}

func init() {
	home, _ := os.UserHomeDir()
	defaultVolume := fmt.Sprintf("%s/.autobox/config:/app/config", home)

	runCmd.Flags().StringVarP(&runImage, "image", "i", "autobox-engine:latest", "Docker image to use")
	runCmd.Flags().StringVarP(&runConfig, "config", "c", "", "Path to simulation config file (overrides simulation name)")
	runCmd.Flags().StringVarP(&runMetricsPath, "metrics", "m", "", "Path to metrics config file (overrides simulation name)")
	runCmd.Flags().StringVarP(&runServer, "server", "s", "", "Path to server config file (overrides default)")
	runCmd.Flags().StringSliceVarP(&runVolumes, "volume", "V", []string{defaultVolume}, "Volume mounts (format: host:container)")
	runCmd.Flags().StringSliceVarP(&runEnv, "env", "e", []string{}, "Environment variables (format: KEY=VALUE)")
	runCmd.Flags().StringVarP(&runName, "name", "n", "", "Container name (overrides simulation name)")
	runCmd.Flags().BoolVarP(&runDetach, "detach", "d", false, "Run in detached mode")
	runCmd.Flags().BoolVarP(&runListSims, "list", "l", false, "List available simulations")
}

func runSimulation(cmd *cobra.Command, args []string) error {
	if runListSims {
		simulations, err := config.ListAvailableSimulations()
		if err != nil {
			return fmt.Errorf("failed to list simulations: %w", err)
		}

		if len(simulations) == 0 {
			fmt.Println("No simulations found in ~/.autobox/config/")
			fmt.Println("\nTo create a simulation, add matching JSON files in:")
			fmt.Println("  ~/.autobox/config/simulations/<name>.json")
			fmt.Println("  ~/.autobox/config/metrics/<name>.json")
			return nil
		}

		fmt.Println("Available simulations:")
		for _, sim := range simulations {
			fmt.Printf("  • %s\n", sim)
		}
		fmt.Println("\nRun a simulation with: autobox run <simulation-name>")
		return nil
	}

	ctx := context.Background()

	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	if err := config.EnsureConfigDirectories(); err != nil {
		return fmt.Errorf("failed to create config directories: %w", err)
	}

	var simName string
	var configPath, metricsPath, serverPath string
	home, _ := os.UserHomeDir()

	if len(args) > 0 && runConfig == "" && runMetricsPath == "" {
		simulationName := args[0]

		if err := config.ValidateSimulationConfig(simulationName); err != nil {
			return fmt.Errorf("simulation validation failed: %w", err)
		}

		configSet, err := config.LoadSimulationConfig(simulationName)
		if err != nil {
			return fmt.Errorf("failed to load simulation '%s': %w", simulationName, err)
		}

		simName = simulationName
		configPath = "/app/config/simulations/" + filepath.Base(configSet.SimulationPath)
		metricsPath = "/app/config/metrics/" + filepath.Base(configSet.MetricsPath)

		if configSet.ServerPath != "" {
			serverPath = "/app/config/server.json"
		}

		fmt.Printf("%s Loading simulation '%s'...\n", color.YellowString("→"), simulationName)
		if verbose {
			fmt.Printf("  Simulation: %s\n", configSet.SimulationPath)
			fmt.Printf("  Metrics: %s\n", configSet.MetricsPath)
			if configSet.ServerPath != "" {
				fmt.Printf("  Server: %s\n", configSet.ServerPath)
			}
		}
	} else {
		if runConfig != "" {
			configPath = runConfig
		} else {
			configPath = "/app/config/simulation.json"
			simulationFile := filepath.Join(home, ".autobox", "config", "simulation.json")
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
		}

		if runMetricsPath != "" {
			metricsPath = runMetricsPath
		} else {
			metricsPath = "/app/config/metrics.json"
			metricsFile := filepath.Join(home, ".autobox", "config", "metrics.json")
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

		if runServer != "" {
			serverPath = runServer
		} else {
			serverPath = "/app/config/server.json"
		}

		if configPath != "" && simName == "" {
			localConfigPath := configPath
			if strings.HasPrefix(configPath, "/app/config/") {
				localConfigPath = filepath.Join(home, ".autobox", "config", strings.TrimPrefix(configPath, "/app/config/"))
			}

			if configData, err := os.ReadFile(localConfigPath); err == nil {
				var config map[string]interface{}
				if err := json.Unmarshal(configData, &config); err == nil {
					if name, ok := config["name"].(string); ok {
						simName = name
					}
				}
			}
		}
	}

	if runName != "" {
		simName = runName
	} else if simName == "" {
		simName = fmt.Sprintf("simulation-%d", os.Getpid())
	}

	envMap := make(map[string]string)
	for _, env := range runEnv {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	volumes := runVolumes
	if len(volumes) == 1 && volumes[0] == "" {
		volumes = []string{}
	}

	simConfig := models.SimulationConfig{
		Name:        simName,
		ConfigPath:  configPath,
		MetricsPath: metricsPath,
		ServerPath:  serverPath,
		Image:       runImage,
		Environment: envMap,
		Volumes:     volumes,
	}

	fmt.Printf("%s Running simulation...\n", color.YellowString("→"))
	if verbose {
		fmt.Printf("  Name: %s\n", simName)
		fmt.Printf("  Image: %s\n", runImage)
		fmt.Printf("  Config: %s\n", configPath)
		fmt.Printf("  Metrics: %s\n", metricsPath)
		if serverPath != "" {
			fmt.Printf("  Server: %s\n", serverPath)
		}
		if len(volumes) > 0 {
			fmt.Printf("  Volumes: %s\n", strings.Join(volumes, ", "))
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
