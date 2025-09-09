# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Autobox CLI is a command-line interface for managing AI simulation containers via Docker. It provides comprehensive tools for launching, monitoring, and managing simulations running in the Autobox Engine.

## Common Development Commands

### Build and Test
```bash
make build           # Build binary to ./bin/autobox
make test            # Run all tests
make test-coverage   # Run tests with coverage report
make fmt             # Format all Go code
make lint            # Run golangci-lint (must be installed)
go test -v ./cmd     # Test specific package with verbose output
```

### Installation and Release
```bash
make install         # Install to $GOPATH/bin
make release         # Build for multiple platforms (linux, darwin, windows)
make clean           # Clean build artifacts
```

### Development Workflow
```bash
make deps            # Download and tidy dependencies
make run             # Build and run the CLI
./bin/autobox        # Run directly after building
```

## High-Level Architecture

### System Design
The CLI acts as a Docker client that manages containers running the Autobox Engine image. It uses container labels (`com.autobox.*`) to track and manage simulations.

### Core Components

**Command Structure (`cmd/`):**
- Each command is a separate file implementing a Cobra command
- `root.go` aggregates all commands and handles global flags
- `output.go` provides unified formatting (table, JSON, YAML)
- Commands communicate with Docker via the internal client wrapper

**Internal Packages (`internal/`):**
- `docker/client.go` - Wraps Docker SDK, handles all container operations
- `config/config.go` - Viper-based configuration management

**Public Models (`pkg/models/`):**
- `simulation.go` - Core data structures for simulations, containers, and metrics

### Key Patterns

**Configuration Hierarchy:**
1. Command-line flags (highest priority)
2. Environment variables (`AUTOBOX_` prefix)
3. Config file (`~/.autobox/autobox.yaml` or `./.autobox.yaml`)
4. Default values in code

**Docker Integration:**
- Connects via Unix socket (`/var/run/docker.sock`) or TCP
- Uses labels for simulation metadata (`com.autobox.simulation`, `com.autobox.name`)
- Collects metrics via Docker stats API
- Manages container lifecycle (create, start, stop, remove)

**Output Formatting:**
- All commands support `--output` flag (table/json/yaml)
- Table format uses tablewriter with optional colors
- Structured data uses standard encoding libraries

### Testing Strategy
- Unit tests for individual functions
- Mock Docker client for testing commands
- Table-driven tests for complex scenarios
- Test files follow `*_test.go` naming convention

## Important Implementation Notes

1. **Docker Client Management**: Always check Docker availability before operations
2. **Error Handling**: Return errors up the stack, handle gracefully at command level
3. **Resource Cleanup**: Ensure containers are properly stopped/removed
4. **Label Convention**: Use `com.autobox.*` labels for all simulation metadata
5. **Configuration**: Support both file and environment variable configuration via Viper
6. **Output Consistency**: Use `cmd/output.go` utilities for all formatted output