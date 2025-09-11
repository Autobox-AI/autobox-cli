package cmd

import (
	"context"
	"fmt"

	"github.com/Autobox-AI/autobox-cli/internal/docker"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	terminateForce bool
	terminateAll   bool
)

var terminateCmd = &cobra.Command{
	Use:   "terminate [SIMULATION_ID]",
	Short: "Terminate and remove a simulation container",
	Long: `Terminate and remove an Autobox simulation container completely.
This command stops the container and removes it from Docker.

Examples:
  # Terminate a specific simulation
  autobox terminate abc123def456

  # Terminate all simulations
  autobox terminate --all

  # Force terminate without confirmation
  autobox terminate abc123def456 --force`,
	Args: func(cmd *cobra.Command, args []string) error {
		if terminateAll && len(args) > 0 {
			return fmt.Errorf("cannot specify simulation ID when using --all flag")
		}
		if !terminateAll && len(args) != 1 {
			return fmt.Errorf("requires exactly one simulation ID (or use --all flag)")
		}
		return nil
	},
	RunE: runTerminate,
}

func init() {
	terminateCmd.Flags().BoolVarP(&terminateForce, "force", "f", false, "Force terminate without confirmation")
	terminateCmd.Flags().BoolVarP(&terminateAll, "all", "a", false, "Terminate all simulations")
}

func runTerminate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	if terminateAll {
		simulations, err := client.ListSimulations(ctx)
		if err != nil {
			return fmt.Errorf("failed to list simulations: %w", err)
		}

		if len(simulations) == 0 {
			fmt.Println("No simulations found")
			return nil
		}

		if !terminateForce {
			fmt.Printf("%s This will terminate and remove %d simulation(s). Continue? [y/N]: ",
				color.YellowString("⚠"), len(simulations))
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Aborted")
				return nil
			}
		}

		terminated := 0
		failed := 0
		for _, sim := range simulations {
			fmt.Printf("%s Terminating simulation %s (%s)...\n",
				color.YellowString("→"), sim.ID, sim.Name)

			if err := client.RemoveSimulation(ctx, sim.ContainerID, true); err != nil {
				fmt.Printf("%s Failed to terminate %s: %v\n",
					color.RedString("✗"), sim.ID, err)
				failed++
			} else {
				fmt.Printf("%s Terminated %s\n", color.GreenString("✓"), sim.ID)
				terminated++
			}
		}

		fmt.Printf("\n%s Terminated %d simulation(s), %d failed\n",
			color.GreenString("Summary:"), terminated, failed)

		if failed > 0 {
			return fmt.Errorf("%d simulation(s) failed to be terminated", failed)
		}
		return nil
	}

	simulationID := args[0]

	if !terminateForce {
		sim, err := client.GetSimulationStatus(ctx, simulationID)
		if err != nil {
			fmt.Printf("%s Terminate and remove simulation %s? [y/N]: ",
				color.YellowString("⚠"), simulationID)
		} else {
			fmt.Printf("%s Terminate and remove simulation %s (%s)? [y/N]: ",
				color.YellowString("⚠"), sim.ID, sim.Name)
		}

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted")
			return nil
		}
	}

	fmt.Printf("%s Terminating simulation %s...\n", color.YellowString("→"), simulationID)

	if err := client.RemoveSimulation(ctx, simulationID, true); err != nil {
		return fmt.Errorf("failed to terminate simulation: %w", err)
	}

	fmt.Printf("%s Simulation terminated and removed successfully\n", color.GreenString("✓"))
	return nil
}
