package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Autobox-AI/autobox-cli/internal/docker"
	"github.com/Autobox-AI/autobox-cli/pkg/models"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [SIMULATION_ID]",
	Short: "Get the status of a simulation",
	Long: `Get detailed status information about an Autobox simulation.
If no simulation ID is provided, shows a list of running simulations to choose from.

Examples:
  autobox status                        # Select from running simulations
  autobox status abc123def456           # Show specific simulation
  autobox status abc123def456 --output json
  autobox status abc123def456 -v`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	var simulationID string

	if len(args) == 0 {
		simulations, err := client.ListSimulations(ctx)
		if err != nil {
			return fmt.Errorf("failed to list simulations: %w", err)
		}

		var running []*models.Simulation
		for _, sim := range simulations {
			if sim.Status == models.StatusRunning {
				running = append(running, sim)
			}
		}

		if len(running) == 0 {
			fmt.Println(color.YellowString("No running simulations found"))
			return nil
		}

		simulationID, err = selectSimulation(running)
		if err != nil {
			return err
		}
		if simulationID == "" {
			return nil
		}
	} else {
		simulationID = args[0]
	}

	simulation, err := client.GetSimulationStatus(ctx, simulationID)
	if err != nil {
		return fmt.Errorf("failed to get simulation status: %w", err)
	}

	switch output {
	case "json":
		return outputJSON(simulation)
	case "yaml":
		return outputYAML(simulation)
	default:
		return outputStatusTable(simulation)
	}
}

func selectSimulation(simulations []*models.Simulation) (string, error) {
	fmt.Printf("\n%s Select a running simulation:\n\n", color.CyanString("▶"))

	for i, sim := range simulations {
		created := sim.CreatedAt.Format("2006-01-02 15:04")
		fmt.Printf("  %s %s %-30s %s (created: %s)\n",
			color.YellowString("[%d]", i+1),
			color.CyanString(sim.ID),
			truncate(sim.Name, 30),
			colorizeStatus(sim.Status),
			created,
		)
	}

	fmt.Printf("\n%s Enter selection (1-%d) or 'q' to quit: ",
		color.GreenString("→"), len(simulations))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)

	if strings.ToLower(input) == "q" {
		fmt.Println(color.YellowString("Selection cancelled"))
		return "", nil
	}

	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > len(simulations) {
		return "", fmt.Errorf("invalid selection: %s", input)
	}

	selected := simulations[selection-1]
	fmt.Printf("\n%s Selected: %s (%s)\n\n",
		color.GreenString("✓"),
		color.CyanString(selected.ID),
		selected.Name,
	)

	return selected.ID, nil
}

func outputStatusTable(simulation *models.Simulation) error {
	fmt.Printf("\n%s Simulation Status\n", color.CyanString("▶"))
	fmt.Println(strings.Repeat("─", 50))

	fmt.Printf("%-15s: %s\n", "ID", color.CyanString(simulation.ID))
	fmt.Printf("%-15s: %s\n", "Name", simulation.Name)
	fmt.Printf("%-15s: %s\n", "Container ID", simulation.ContainerID[:12])
	fmt.Printf("%-15s: %s\n", "Status", colorizeStatus(simulation.Status))
	fmt.Printf("%-15s: %s\n", "Created", simulation.CreatedAt.Format(time.RFC3339))

	if simulation.StartedAt != nil {
		fmt.Printf("%-15s: %s\n", "Started", simulation.StartedAt.Format(time.RFC3339))
	}

	if simulation.FinishedAt != nil {
		fmt.Printf("%-15s: %s\n", "Finished", simulation.FinishedAt.Format(time.RFC3339))
		duration := simulation.FinishedAt.Sub(*simulation.StartedAt)
		fmt.Printf("%-15s: %s\n", "Duration", duration.Round(time.Second))
	} else if simulation.StartedAt != nil {
		duration := time.Since(*simulation.StartedAt)
		fmt.Printf("%-15s: %s\n", "Running For", duration.Round(time.Second))
	}

	if verbose {
		fmt.Printf("\n%s Configuration\n", color.CyanString("▶"))
		fmt.Println(strings.Repeat("─", 50))
		fmt.Printf("%-15s: %s\n", "Image", simulation.Config.Image)
		fmt.Printf("%-15s: %s\n", "Config Path", simulation.Config.ConfigPath)
		fmt.Printf("%-15s: %s\n", "Metrics Path", simulation.Config.MetricsPath)

		if len(simulation.Config.Volumes) > 0 {
			fmt.Printf("%-15s: %s\n", "Volumes", strings.Join(simulation.Config.Volumes, ", "))
		}

		if len(simulation.Config.Environment) > 0 {
			fmt.Printf("%-15s:\n", "Environment")
			for k, v := range simulation.Config.Environment {
				fmt.Printf("  %s=%s\n", k, v)
			}
		}
	}

	fmt.Println()
	return nil
}
