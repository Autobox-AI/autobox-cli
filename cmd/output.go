package cmd

import (
	"encoding/json"
	"os"

	"github.com/Autobox-AI/autobox-cli/pkg/models"
	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func outputYAML(data interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	return encoder.Encode(data)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
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
