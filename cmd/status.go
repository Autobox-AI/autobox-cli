package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Autobox-AI/autobox-cli/internal/docker"
	"github.com/Autobox-AI/autobox-cli/pkg/models"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [SIMULATION_ID]",
	Short: "Get the status of a specific simulation",
	Long: `Get detailed status information about a specific Autobox simulation.
	
Examples:
  autobox status abc123def456
  autobox status abc123def456 --output json
  autobox status abc123def456 -v`,
	Args: cobra.ExactArgs(1),
	RunE: runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	simulationID := args[0]

	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

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

