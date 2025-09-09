package cmd

import (
	"context"
	"fmt"

	"github.com/Autobox-AI/autobox-cli/internal/docker"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [SIMULATION_ID]",
	Short: "Stop a running simulation",
	Long: `Stop a running Autobox simulation container.
	
Examples:
  autobox stop abc123def456`,
	Args: cobra.ExactArgs(1),
	RunE: runStop,
}

func runStop(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	simulationID := args[0]

	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	fmt.Printf("%s Stopping simulation %s...\n", color.YellowString("→"), simulationID)

	if err := client.StopSimulation(ctx, simulationID); err != nil {
		return fmt.Errorf("failed to stop simulation: %w", err)
	}

	fmt.Printf("%s Simulation stopped successfully\n", color.GreenString("✓"))
	return nil
}