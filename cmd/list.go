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

var (
	listAll bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all simulations",
	Long: `List all Autobox simulations with their current status.

Examples:
  autobox list
  autobox list --all
  autobox list --output json`,
	RunE: runList,
}

func init() {
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "Show all simulations (including stopped)")
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	simulations, err := client.ListSimulations(ctx)
	if err != nil {
		return fmt.Errorf("failed to list simulations: %w", err)
	}

	if !listAll {
		simulations = filterRunningSimulations(simulations)
	}

	switch output {
	case "json":
		return outputJSON(simulations)
	case "yaml":
		return outputYAML(simulations)
	default:
		return outputListTable(simulations)
	}
}

func filterRunningSimulations(simulations []*models.Simulation) []*models.Simulation {
	var running []*models.Simulation
	for _, sim := range simulations {
		if sim.Status == models.StatusRunning {
			running = append(running, sim)
		}
	}
	return running
}

func outputListTable(simulations []*models.Simulation) error {
	if len(simulations) == 0 {
		fmt.Println(color.YellowString("No simulations found"))
		return nil
	}

	fmt.Printf("\n%s Found %d simulation(s)\n\n", color.CyanString("▶"), len(simulations))

	fmt.Printf("%-12s  %-30s  %-12s  %-16s  %-12s\n", "ID", "NAME", "STATUS", "CREATED", "RUNNING FOR")
	fmt.Println(strings.Repeat("-", 90))

	for _, sim := range simulations {
		runningFor := "-"
		if sim.StartedAt != nil && sim.Status == models.StatusRunning {
			duration := time.Since(*sim.StartedAt)
			runningFor = formatDuration(duration)
		}

		statusStr := colorizeStatus(sim.Status)
		idStr := color.CyanString(sim.ID)

		fmt.Printf("%-12s  %-30s  %-12s  %-16s  %-12s\n",
			idStr,
			truncate(sim.Name, 30),
			statusStr,
			sim.CreatedAt.Format("2006-01-02 15:04"),
			runningFor,
		)
	}

	running := countByStatus(simulations, models.StatusRunning)
	completed := countByStatus(simulations, models.StatusCompleted)
	failed := countByStatus(simulations, models.StatusFailed)

	fmt.Printf("\nSummary: ")
	if running > 0 {
		fmt.Printf("%s ", color.GreenString("%d running", running))
	}
	if completed > 0 {
		fmt.Printf("%s ", color.BlueString("%d completed", completed))
	}
	if failed > 0 {
		fmt.Printf("%s ", color.RedString("%d failed", failed))
	}
	fmt.Println()

	return nil
}

func countByStatus(simulations []*models.Simulation, status models.SimulationStatus) int {
	count := 0
	for _, sim := range simulations {
		if sim.Status == status {
			count++
		}
	}
	return count
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
