package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Autobox-AI/autobox-cli/internal/docker"
	"github.com/Autobox-AI/autobox-cli/pkg/models"
	"github.com/fatih/color"
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
If no simulation ID is provided, shows a list of running simulations to choose from.

Examples:
  autobox logs                        # Select from running simulations
  autobox logs abc123def456
  autobox logs abc123def456 --tail 50
  autobox logs --live
  autobox logs abc123def456 --live --tail 20`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLogs,
}

func init() {
	logsCmd.Flags().IntVarP(&logsTail, "tail", "t", 100, "Number of lines to show from the end of the logs")
	logsCmd.Flags().BoolVarP(&logsLive, "live", "l", false, "Stream logs in real-time")
}

func runLogs(cmd *cobra.Command, args []string) error {
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

		simulationID, err = selectSimulationForLogs(running)
		if err != nil {
			return err
		}
		if simulationID == "" {
			return nil
		}
	} else {
		simulationID = args[0]
	}

	if logsLive {
		fmt.Printf("%s Streaming logs for %s (press Ctrl+C to stop)...\n\n",
			color.YellowString("→"), color.CyanString(simulationID[:12]))

		reader, err := client.GetSimulationLogsStream(ctx, simulationID, logsTail)
		if err != nil {
			return fmt.Errorf("failed to get simulation logs: %w", err)
		}
		defer reader.Close()

		_, err = io.Copy(os.Stdout, reader)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to stream logs: %w", err)
		}
		return nil
	}

	logs, err := client.GetSimulationLogs(ctx, simulationID, logsTail)
	if err != nil {
		return fmt.Errorf("failed to get simulation logs: %w", err)
	}

	fmt.Print(logs)
	return nil
}

func selectSimulationForLogs(simulations []*models.Simulation) (string, error) {
	fmt.Printf("\n%s Select a running simulation:\n\n", color.CyanString("▶"))

	for i, sim := range simulations {
		created := sim.CreatedAt.Format("2006-01-02 15:04")
		fmt.Printf("  %s %s %-30s %s (created: %s)\n",
			color.YellowString("[%d]", i+1),
			color.CyanString(sim.ID[:12]),
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
		color.CyanString(selected.ContainerID[:12]),
		selected.Name,
	)

	return selected.ContainerID, nil
}
