package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Autobox-AI/autobox-cli/pkg/models"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

const (
	AutoboxLabelPrefix = "com.autobox"
	AutoboxImagePrefix = "autobox-engine"
)

type Client struct {
	cli *client.Client
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &Client{cli: cli}, nil
}

func (c *Client) Close() error {
	return c.cli.Close()
}

func (c *Client) LaunchSimulation(ctx context.Context, config models.SimulationConfig) (*models.Simulation, error) {
	labels := map[string]string{
		fmt.Sprintf("%s.simulation", AutoboxLabelPrefix):  "true",
		fmt.Sprintf("%s.name", AutoboxLabelPrefix):        config.Name,
		fmt.Sprintf("%s.config_path", AutoboxLabelPrefix): config.ConfigPath,
		fmt.Sprintf("%s.created_at", AutoboxLabelPrefix):  time.Now().Format(time.RFC3339),
	}

	containerConfig := &container.Config{
		Image:  config.Image,
		Labels: labels,
		Env:    c.mapToEnvSlice(config.Environment),
		Cmd: []string{
			"--config", config.ConfigPath,
			"--metrics", config.MetricsPath,
			"--server", config.ServerPath,
		},
	}

	hostConfig := &container.HostConfig{
		Binds:      config.Volumes,
		AutoRemove: false,
		RestartPolicy: container.RestartPolicy{
			Name: "no",
		},
	}

	resp, err := c.cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	if err := c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	now := time.Now()
	simulation := &models.Simulation{
		ID:          resp.ID[:12],
		Name:        config.ConfigPath,
		ContainerID: resp.ID,
		Status:      models.StatusRunning,
		CreatedAt:   now,
		StartedAt:   &now,
		Config:      config,
	}

	return simulation, nil
}

func (c *Client) GetSimulationStatus(ctx context.Context, simulationID string) (*models.Simulation, error) {
	containerJSON, err := c.cli.ContainerInspect(ctx, simulationID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	simulation := c.containerToSimulation(containerJSON)
	return simulation, nil
}

func (c *Client) ListSimulations(ctx context.Context) ([]*models.Simulation, error) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("%s.simulation=true", AutoboxLabelPrefix))

	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	simulations := make([]*models.Simulation, 0, len(containers))
	for _, cont := range containers {
		simulation := c.containerListItemToSimulation(cont)
		simulations = append(simulations, simulation)
	}

	return simulations, nil
}

func (c *Client) GetSimulationMetrics(ctx context.Context, simulationID string) (*models.Metrics, error) {
	stats, err := c.cli.ContainerStats(ctx, simulationID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer stats.Body.Close()

	var containerStats container.StatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&containerStats); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	metrics := c.statsToMetrics(containerStats)
	return metrics, nil
}

func (c *Client) StopSimulation(ctx context.Context, simulationID string) error {
	timeout := 30
	stopOptions := container.StopOptions{
		Timeout: &timeout,
	}

	if err := c.cli.ContainerStop(ctx, simulationID, stopOptions); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	return nil
}

func (c *Client) RemoveSimulation(ctx context.Context, simulationID string, force bool) error {
	if force {
		timeout := 10
		stopOptions := container.StopOptions{
			Timeout: &timeout,
		}
		_ = c.cli.ContainerStop(ctx, simulationID, stopOptions)
	}

	removeOptions := container.RemoveOptions{
		Force:         force,
		RemoveVolumes: true,
	}

	if err := c.cli.ContainerRemove(ctx, simulationID, removeOptions); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	return nil
}

func (c *Client) GetSimulationLogs(ctx context.Context, simulationID string, tail int) (string, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       fmt.Sprintf("%d", tail),
	}

	reader, err := c.cli.ContainerLogs(ctx, simulationID, options)
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}
	defer reader.Close()

	logs, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}

	return string(logs), nil
}

func (c *Client) mapToEnvSlice(envMap map[string]string) []string {
	env := make([]string, 0, len(envMap))
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

func (c *Client) containerToSimulation(container types.ContainerJSON) *models.Simulation {
	createdAt, _ := time.Parse(time.RFC3339, container.Created)
	simulation := &models.Simulation{
		ID:          container.ID[:12],
		ContainerID: container.ID,
		Status:      c.containerStateToStatus(container.State),
		CreatedAt:   createdAt,
	}

	if container.State.StartedAt != "" {
		if t, err := time.Parse(time.RFC3339Nano, container.State.StartedAt); err == nil {
			simulation.StartedAt = &t
		}
	}

	if container.State.FinishedAt != "" {
		if t, err := time.Parse(time.RFC3339Nano, container.State.FinishedAt); err == nil {
			simulation.FinishedAt = &t
		}
	}

	if name, ok := container.Config.Labels[fmt.Sprintf("%s.name", AutoboxLabelPrefix)]; ok {
		simulation.Name = name
	}

	return simulation
}

func (c *Client) containerListItemToSimulation(container types.Container) *models.Simulation {
	simulation := &models.Simulation{
		ID:          container.ID[:12],
		ContainerID: container.ID,
		Status:      c.containerStateStringToStatus(container.State),
		CreatedAt:   time.Unix(container.Created, 0),
	}

	if name, ok := container.Labels[fmt.Sprintf("%s.name", AutoboxLabelPrefix)]; ok {
		simulation.Name = name
	}

	return simulation
}

func (c *Client) containerStateToStatus(state *types.ContainerState) models.SimulationStatus {
	switch {
	case state.Running:
		return models.StatusRunning
	case state.Dead:
		return models.StatusFailed
	case state.Paused:
		return models.StatusStopped
	case state.Restarting:
		return models.StatusRunning
	case state.Status == "exited" && state.ExitCode == 0:
		return models.StatusCompleted
	case state.Status == "exited" && state.ExitCode != 0:
		return models.StatusFailed
	default:
		return models.StatusPending
	}
}

func (c *Client) containerStateStringToStatus(state string) models.SimulationStatus {
	switch strings.ToLower(state) {
	case "running":
		return models.StatusRunning
	case "exited":
		return models.StatusCompleted
	case "dead":
		return models.StatusFailed
	case "paused":
		return models.StatusStopped
	default:
		return models.StatusPending
	}
}

func (c *Client) statsToMetrics(stats container.StatsResponse) *models.Metrics {
	var cpuPercent float64
	if stats.PreCPUStats.CPUUsage.TotalUsage > 0 {
		cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
		systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
		if systemDelta > 0 && cpuDelta > 0 {
			cpuPercent = (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
		}
	}

	var memoryPercent float64
	if stats.MemoryStats.Limit > 0 {
		memoryPercent = (float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit)) * 100.0
	}

	return &models.Metrics{
		CPUUsage:    cpuPercent,
		MemoryUsage: memoryPercent,
		NetworkIO: models.NetworkStats{
			BytesReceived:      stats.Networks["eth0"].RxBytes,
			BytesTransmitted:   stats.Networks["eth0"].TxBytes,
			PacketsReceived:    stats.Networks["eth0"].RxPackets,
			PacketsTransmitted: stats.Networks["eth0"].TxPackets,
		},
		DiskIO: models.DiskStats{
			BytesRead:    stats.BlkioStats.IoServiceBytesRecursive[0].Value,
			BytesWritten: stats.BlkioStats.IoServiceBytesRecursive[1].Value,
		},
		Timestamp: time.Now(),
	}
}
