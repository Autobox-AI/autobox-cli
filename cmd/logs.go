package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/Autobox-AI/autobox-cli/internal/docker"
	"github.com/spf13/cobra"
)

var (
	logsTail int
	logsLive bool
)

var logsCmd = &cobra.Command{
	Use:   "logs [SIMULATION_ID]",
	Short: "Get logs from a simulation",
	Long: `Retrieve logs from a specific Autobox simulation container.
	
Examples:
  autobox logs abc123def456
  autobox logs abc123def456 --tail 50
  autobox logs abc123def456 --live
  autobox logs abc123def456 --live --tail 20`,
	Args: cobra.ExactArgs(1),
	RunE: runLogs,
}

func init() {
	logsCmd.Flags().IntVarP(&logsTail, "tail", "t", 100, "Number of lines to show from the end of the logs")
	logsCmd.Flags().BoolVarP(&logsLive, "live", "l", false, "Stream logs in real-time")
}

func runLogs(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	simulationID := args[0]

	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	if logsLive {
		// Stream logs in real-time
		reader, err := client.GetSimulationLogsStream(ctx, simulationID, logsTail)
		if err != nil {
			return fmt.Errorf("failed to get simulation logs: %w", err)
		}
		defer reader.Close()

		// Stream logs to stdout
		_, err = io.Copy(os.Stdout, reader)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to stream logs: %w", err)
		}
		return nil
	}

	// Get static logs snapshot
	logs, err := client.GetSimulationLogs(ctx, simulationID, logsTail)
	if err != nil {
		return fmt.Errorf("failed to get simulation logs: %w", err)
	}

	fmt.Print(logs)
	return nil
}