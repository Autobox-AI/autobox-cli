# Autobox CLI

[![Tests](https://github.com/Autobox-AI/autobox-cli/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/Autobox-AI/autobox-cli/actions/workflows/tests.yml)
[![codecov](https://codecov.io/gh/Autobox-AI/autobox-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/Autobox-AI/autobox-cli)
[![Go 1.21+](https://img.shields.io/badge/go-1.21+-blue.svg)](https://go.dev/dl/)

A powerful command-line interface for managing Autobox AI simulation containers. Built with Go, Cobra, and Viper for a robust, scalable, and user-friendly CLI experience.

## Features

- 🚀 **Run** new simulations with custom configurations
- 📊 **Monitor** simulation status and metrics in real-time
- 📋 **List** all running simulations
- 📈 **Collect metrics** including CPU, memory, network, and disk I/O
- 📝 **View logs** from simulation containers
- 🛑 **Stop** running simulations gracefully
- 🎨 **Multiple output formats**: table, JSON, YAML
- ⚙️ **Configuration management** via YAML files and environment variables

## Prerequisites

- Go 1.21 or higher
- Docker installed and running
- Autobox Engine Docker image built (see [Autobox Engine](#autobox-engine) section)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/Autobox-AI/autobox-cli.git
cd autobox-cli

# Install dependencies
go mod download

# Build the binary
make build

# Install to $GOPATH/bin
make install
```

### Using Go Install

```bash
go install github.com/Autobox-AI/autobox-cli@latest
```

### Pre-built Binaries

Download the appropriate binary for your platform from the releases page:

```bash
# macOS (Apple Silicon)
curl -L https://github.com/Autobox-AI/autobox-cli/releases/latest/download/autobox-darwin-arm64 -o autobox
chmod +x autobox
sudo mv autobox /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/Autobox-AI/autobox-cli/releases/latest/download/autobox-darwin-amd64 -o autobox
chmod +x autobox
sudo mv autobox /usr/local/bin/

# Linux (x86_64)
curl -L https://github.com/Autobox-AI/autobox-cli/releases/latest/download/autobox-linux-amd64 -o autobox
chmod +x autobox
sudo mv autobox /usr/local/bin/
```

## Quick Start

```bash
# 1. Build the Autobox Engine Docker image
cd ../autobox-engine
docker build -t autobox-engine:latest .

# 2. Run your first simulation
autobox run --name "my-first-simulation"

# 3. Check the status
autobox status <container-id>

# 4. View metrics
autobox metrics <container-id>

# 5. List all simulations
autobox list
```

## Usage

### Run a Simulation

```bash
# Basic run with defaults
autobox run

# Run with custom configuration files
autobox run --config simulation.json --metrics metrics.json

# Run with custom image and environment variables
autobox run --image autobox-engine:v1.0 \
  --env OPENAI_API_KEY=sk-... \
  --env LOG_LEVEL=debug \
  --name "my-simulation"

# Run with volume mounts for config and logs
autobox run \
  --volume ./configs:/app/configs \
  --volume ./logs:/app/logs \
  --name "production-sim"

# Run in detached mode
autobox run --detach --name "background-sim"
```

**Note**: The simulation name displayed in `list` and `status` commands is now read from the simulation configuration file's `name` field, not the file path.

### List Simulations

```bash
# List only running simulations
autobox list

# List all simulations (including stopped/completed)
autobox list --all

# Output as JSON for scripting
autobox list --output json

# Output as YAML
autobox list --output yaml
```

Output example:
```
▶ Found 3 simulation(s)

ID            NAME                            STATUS        CREATED           RUNNING FOR
------------------------------------------------------------------------------------------
abc123def456  Climate Model v2                running       2024-01-15 14:30  2h 45m
def456ghi789  Market Analysis                 running       2024-01-15 16:15  1h 0m
ghi789jkl012  Gift Choice                     completed     2024-01-15 12:00  -

Summary: 2 running 1 completed
```

**Note**: The NAME column shows the actual simulation name from the config file's `name` field, not the config file path.

### Check Simulation Status

```bash
# Interactive selection - shows list of running simulations to choose from
autobox status

# Get detailed status by ID
autobox status abc123def456

# Get status in JSON format for parsing
autobox status abc123def456 --output json

# Verbose output with full configuration details
autobox status abc123def456 -v
```

When no ID is provided, the status command presents an interactive menu:
```
▶ Select a running simulation:

  [1] 32ca259fb21e Gift choice                    running (created: 2025-09-11 08:28)
  [2] bc74ad805192 Test Simulation 2              running (created: 2025-09-11 08:35)

→ Enter selection (1-2) or 'q' to quit:
```

### View Metrics

```bash
# Get real-time metrics
autobox metrics abc123def456

# Output metrics as JSON for monitoring systems
autobox metrics abc123def456 --output json

# Output metrics as YAML
autobox metrics abc123def456 --output yaml
```

Metrics include:
- CPU usage percentage
- Memory usage percentage
- Network I/O (bytes received/transmitted)
- Disk I/O (bytes read/written)
- Custom application metrics (if configured)

### View Logs

```bash
# Get last 100 lines of logs (default)
autobox logs abc123def456

# Get last 50 lines
autobox logs abc123def456 --tail 50

# Get last 500 lines for debugging
autobox logs abc123def456 --tail 500
```

### Stop a Simulation

```bash
# Gracefully stop a running simulation
autobox stop abc123def456

# Stop multiple simulations
for id in $(autobox list --output json | jq -r '.[].id'); do
  autobox stop $id
done
```

## Configuration

Autobox CLI can be configured using:

1. **Configuration file** (`~/.autobox/autobox.yaml`)
2. **Environment variables** (prefixed with `AUTOBOX_`)
3. **Command-line flags**

### Configuration File Example

```yaml
docker:
  host: unix:///var/run/docker.sock
  api_version: "1.41"
  image: autobox-engine:latest

simulation:
  default_image: autobox-engine:latest
  default_config_path: /app/configs/simulation.json
  default_metrics_path: /app/configs/metrics.json
  default_volumes:
    - ./configs:/app/configs
  default_environment:
    LOG_LEVEL: info
  logs_directory: /tmp/autobox/logs
  configs_directory: /tmp/autobox/configs

output:
  format: table
  verbose: false
  color: true
```

### Environment Variables

```bash
export AUTOBOX_DOCKER_HOST=tcp://localhost:2375
export AUTOBOX_SIMULATION_DEFAULT_IMAGE=autobox-engine:v2.0
export AUTOBOX_OUTPUT_FORMAT=json
```

## Advanced Usage

### Scripting and Automation

```bash
# Run simulation and capture ID
SIM_ID=$(autobox run --detach --output json | jq -r '.id')

# Wait for simulation to complete
while [ "$(autobox status $SIM_ID --output json | jq -r '.status')" == "running" ]; do
  echo "Simulation still running..."
  sleep 10
done

# Get final metrics
autobox metrics $SIM_ID --output json > metrics.json

# Clean up
autobox stop $SIM_ID
```

### Integration with CI/CD

```yaml
# GitHub Actions example
- name: Run Autobox Simulation
  run: |
    autobox run \
      --config ${{ github.workspace }}/simulation.json \
      --env GITHUB_SHA=${{ github.sha }} \
      --name "ci-test-${{ github.run_number }}"
    
- name: Check Results
  run: |
    SIM_ID=$(autobox list --output json | jq -r '.[0].id')
    autobox logs $SIM_ID --tail 100
    autobox metrics $SIM_ID
```

### Docker Compose Integration

```yaml
# docker-compose.yml
version: '3.8'
services:
  autobox-engine:
    image: autobox-engine:latest
    labels:
      com.autobox.simulation: "true"
      com.autobox.name: "docker-compose-sim"
    volumes:
      - ./configs:/app/configs
      - ./logs:/app/logs
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
```

## Development

### Go Version Requirements

This project uses **Go 1.24+** for local development to leverage the latest features and dependencies. However, note that:

- **Local Development**: Requires Go 1.24.0 or higher
- **CI/CD**: Currently runs tests on Go 1.21, 1.22, and 1.23 (GitHub Actions limitation)
- **Linting**: Temporarily disabled in CI due to version mismatch; run `make lint` locally

Some dependencies (like `golang.org/x/sys@v0.36.0`) require Go 1.24+, which is why the project targets this version.

### Building

```bash
# Build the binary
make build

# Build for multiple platforms
make release

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Clean build artifacts
make clean

# Install dependencies
make deps
```

### Project Structure

```
autobox-cli/
├── cmd/                    # Command implementations
│   ├── root.go            # Root command and global flags
│   ├── run.go             # Run simulation command
│   ├── list.go            # List simulations command
│   ├── status.go          # Status command
│   ├── metrics.go         # Metrics command
│   ├── logs.go            # Logs command
│   ├── stop.go            # Stop command
│   ├── version.go         # Version command
│   ├── output.go          # Output formatting utilities
│   └── utils_test.go      # Command utilities tests
├── internal/              # Internal packages (not importable)
│   ├── docker/            # Docker client wrapper
│   │   └── client.go      # Docker operations and container management
│   └── config/            # Configuration management
│       ├── config.go      # Viper configuration setup
│       └── config_test.go # Configuration tests
├── pkg/                   # Public packages (importable)
│   └── models/            # Data models
│       ├── simulation.go  # Simulation structures and types
│       └── simulation_test.go # Model tests
├── main.go                # Entry point
├── go.mod                 # Go module definition
├── go.sum                 # Dependency checksums
├── Makefile               # Build automation and tasks
├── README.md              # This file
├── LICENSE                # Apache 2.0 license
└── .autobox.yaml.example  # Example configuration file
```

### Testing

```bash
# Run all tests using Makefile
make test

# Run tests with coverage report
make test-coverage

# Run all tests with Go directly
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage percentage
go test -cover ./...

# Generate coverage report and open in browser
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./cmd
go test ./pkg/models
go test ./internal/config
go test -v ./internal/docker  # verbose since it has no tests yet

# Run a specific test function
go test -v -run TestTruncate ./cmd
go test -v -run TestInit ./internal/config

# Run tests with race detection
go test -race ./...

# Run tests with timeout
go test -timeout 30s ./...

# Run benchmarks
go test -bench=. ./...

# Clean test cache and run tests
go clean -testcache && go test ./...
```

#### Test Structure

The project includes tests for:
- **cmd package**: Command utilities (truncate, formatDuration, formatBytes, etc.)
- **internal/config**: Configuration management with Viper
- **pkg/models**: Data model validation
- **internal/docker**: Docker client operations (tests to be added)

#### Writing Tests

Tests follow Go conventions:
- Test files are named `*_test.go`
- Test functions start with `Test`
- Use table-driven tests for multiple scenarios
- Mock external dependencies (Docker client, file system, etc.)

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

#### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write tests for new functionality
- Update documentation as needed

## Docker Requirements

The CLI requires Docker to be installed and running on your system. It connects to Docker via:

- **Unix socket**: `/var/run/docker.sock` (default on Linux/macOS)
- **TCP**: Configure via `AUTOBOX_DOCKER_HOST` environment variable
- **Docker Desktop**: Works with Docker Desktop on macOS/Windows

### Docker Connection Examples

```bash
# Default Unix socket
autobox list

# Custom Docker host
export AUTOBOX_DOCKER_HOST=tcp://localhost:2375
autobox list

# Docker in Docker (DinD)
export AUTOBOX_DOCKER_HOST=tcp://docker:2376
export AUTOBOX_DOCKER_TLS_VERIFY=1
export AUTOBOX_DOCKER_CERT_PATH=/certs
autobox list
```

## Autobox Engine

The CLI manages containers running the Autobox Engine image. The engine is a Python-based simulation runtime that executes AI agent workflows.

### Building the Engine

```bash
# Clone and build the autobox-engine
cd ../autobox-engine
docker build -t autobox-engine:latest .

# Verify the image
docker images | grep autobox-engine

# Test run
docker run --rm autobox-engine:latest --help
```

### Engine Configuration

The engine expects two configuration files:
- **simulation.json**: Defines the simulation parameters and agent behaviors
- **metrics.json**: Configures metrics collection and reporting

Example simulation.json:
```json
{
  "name": "Market Analysis",
  "agents": [
    {
      "name": "data-collector",
      "type": "collector",
      "config": {
        "sources": ["api", "database"],
        "interval": 60
      }
    },
    {
      "name": "analyzer",
      "type": "analyzer",
      "config": {
        "algorithms": ["ml", "statistical"],
        "threshold": 0.8
      }
    }
  ],
  "duration": 3600,
  "output": "/app/logs/results.json"
}
```

## Troubleshooting

### Common Issues

#### Docker Connection Failed
```bash
# Error: Cannot connect to Docker daemon
# Solution: Ensure Docker is running
docker version

# On macOS/Windows with Docker Desktop
open -a Docker  # macOS
# or start Docker Desktop from system tray
```

#### Permission Denied
```bash
# Error: permission denied while trying to connect to Docker
# Solution: Add user to docker group (Linux)
sudo usermod -aG docker $USER
newgrp docker
```

#### Container Not Found
```bash
# Error: container not found
# Solution: Use the container ID from 'autobox list'
autobox list
autobox status <correct-id>
```

#### Image Not Found
```bash
# Error: autobox-engine:latest not found
# Solution: Build the engine image first
cd ../autobox-engine
docker build -t autobox-engine:latest .
```

### Debug Mode

```bash
# Enable verbose output
autobox -v run

# Check Docker connection
docker ps

# Verify labels on containers
docker inspect <container-id> | grep -A 10 Labels

# Manual container inspection
docker logs <container-id>
docker stats <container-id>
```

## Performance Considerations

- **Container Limits**: Set resource limits to prevent runaway simulations
  ```bash
  autobox run --env MEMORY_LIMIT=2g --env CPU_LIMIT=2
  ```

- **Log Rotation**: Configure log rotation for long-running simulations
  ```bash
  autobox run --env LOG_MAX_SIZE=100m --env LOG_MAX_FILES=5
  ```

- **Monitoring**: Use the metrics command for real-time monitoring
  ```bash
  watch -n 5 'autobox metrics <container-id>'
  ```

## Security

### Best Practices

1. **Never commit secrets**: Use environment variables for sensitive data
   ```bash
   autobox run --env OPENAI_API_KEY=${OPENAI_API_KEY}
   ```

2. **Use volume mounts carefully**: Only mount necessary directories
   ```bash
   autobox run --volume ./configs:/app/configs:ro  # read-only mount
   ```

3. **Network isolation**: Run simulations in isolated networks
   ```bash
   docker network create autobox-net
   autobox run --env DOCKER_NETWORK=autobox-net
   ```

4. **Regular updates**: Keep the CLI and engine updated
   ```bash
   go install github.com/Autobox-AI/autobox-cli@latest
   docker pull autobox-engine:latest
   ```

## License

Apache License 2.0 - See [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/Autobox-AI/autobox-cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Autobox-AI/autobox-cli/discussions)
- **Documentation**: [Wiki](https://github.com/Autobox-AI/autobox-cli/wiki)

---

Built with ❤️ by the Autobox team
