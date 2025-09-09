package cmd

import (
	"context"
	"fmt"

	"github.com/Autobox-AI/autobox-cli/internal/docker"
	"github.com/spf13/cobra"
)

var (
	logsTail int
)

var logsCmd = &cobra.Command{
	Use:   "logs [SIMULATION_ID]",
	Short: "Get logs from a simulation",
	Long: `Retrieve logs from a specific Autobox simulation container.
	
Examples:
  autobox logs abc123def456
  autobox logs abc123def456 --tail 50`,
	Args: cobra.ExactArgs(1),
	RunE: runLogs,
}

func init() {
	logsCmd.Flags().IntVarP(&logsTail, "tail", "t", 100, "Number of lines to show from the end of the logs")
}

func runLogs(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	simulationID := args[0]

	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	logs, err := client.GetSimulationLogs(ctx, simulationID, logsTail)
	if err != nil {
		return fmt.Errorf("failed to get simulation logs: %w", err)
	}

	fmt.Print(logs)
	return nil
}