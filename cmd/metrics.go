package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Autobox-AI/autobox-cli/internal/docker"
	"github.com/Autobox-AI/autobox-cli/pkg/models"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics [SIMULATION_ID]",
	Short: "Get metrics for a specific simulation",
	Long: `Get real-time metrics for a specific Autobox simulation.
	
Metrics include CPU usage, memory usage, network I/O, and disk I/O.
	
Examples:
  autobox metrics abc123def456
  autobox metrics abc123def456 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: runMetrics,
}

func runMetrics(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	simulationID := args[0]

	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	metrics, err := client.GetSimulationMetrics(ctx, simulationID)
	if err != nil {
		return fmt.Errorf("failed to get simulation metrics: %w", err)
	}

	switch output {
	case "json":
		return outputJSON(metrics)
	case "yaml":
		return outputYAML(metrics)
	default:
		return outputMetricsTable(metrics)
	}
}

func outputMetricsTable(metrics *models.Metrics) error {
	fmt.Printf("\n%s Simulation Metrics\n", color.CyanString("▶"))
	fmt.Println(strings.Repeat("─", 50))

	fmt.Printf("\n%s Resource Usage\n", color.YellowString("→"))
	fmt.Printf("  %-20s: %s\n", "CPU Usage", formatPercentage(metrics.CPUUsage))
	fmt.Printf("  %-20s: %s\n", "Memory Usage", formatPercentage(metrics.MemoryUsage))

	fmt.Printf("\n%s Network I/O\n", color.YellowString("→"))
	fmt.Printf("  %-20s: %s\n", "Bytes Received", formatBytes(metrics.NetworkIO.BytesReceived))
	fmt.Printf("  %-20s: %s\n", "Bytes Transmitted", formatBytes(metrics.NetworkIO.BytesTransmitted))
	fmt.Printf("  %-20s: %d\n", "Packets Received", metrics.NetworkIO.PacketsReceived)
	fmt.Printf("  %-20s: %d\n", "Packets Transmitted", metrics.NetworkIO.PacketsTransmitted)

	fmt.Printf("\n%s Disk I/O\n", color.YellowString("→"))
	fmt.Printf("  %-20s: %s\n", "Bytes Read", formatBytes(metrics.DiskIO.BytesRead))
	fmt.Printf("  %-20s: %s\n", "Bytes Written", formatBytes(metrics.DiskIO.BytesWritten))

	if len(metrics.Custom) > 0 {
		fmt.Printf("\n%s Custom Metrics\n", color.YellowString("→"))
		for key, value := range metrics.Custom {
			fmt.Printf("  %-20s: %v\n", key, value)
		}
	}

	fmt.Printf("\n%s Timestamp: %s\n", color.WhiteString("•"), metrics.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Println()

	return nil
}

func formatPercentage(value float64) string {
	if value < 50 {
		return color.GreenString("%.2f%%", value)
	} else if value < 80 {
		return color.YellowString("%.2f%%", value)
	}
	return color.RedString("%.2f%%", value)
}

func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}