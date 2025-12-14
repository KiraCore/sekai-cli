package docker

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// execCommand executes a command in a Docker container.
func execCommand(ctx context.Context, container, binary string, args ...string) (*ExecResult, error) {
	// Build docker exec command
	dockerArgs := []string{"exec", container, binary}
	dockerArgs = append(dockerArgs, args...)

	cmd := exec.CommandContext(ctx, "docker", dockerArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &ExecResult{
		Stdout: strings.TrimSpace(stdout.String()),
		Stderr: strings.TrimSpace(stderr.String()),
	}

	if err != nil {
		// Get exit code if available
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}

		// Build error with context
		execErr := &sdk.ExecutionError{
			Command:  binary,
			Args:     args,
			ExitCode: result.ExitCode,
			Stderr:   result.Stderr,
			Err:      err,
		}

		// If there's useful info in stderr, include it
		if result.Stderr != "" {
			return result, execErr
		}

		// If stdout has error info (common with sekaid JSON errors), include it
		if result.Stdout != "" && strings.Contains(result.Stdout, "error") {
			execErr.Stderr = result.Stdout
			return result, execErr
		}

		return result, execErr
	}

	return result, nil
}

// execCommandWithInput executes a command with stdin input.
func execCommandWithInput(ctx context.Context, container, binary string, input string, args ...string) (*ExecResult, error) {
	// Build docker exec command with interactive flag for stdin
	dockerArgs := []string{"exec", "-i", container, binary}
	dockerArgs = append(dockerArgs, args...)

	cmd := exec.CommandContext(ctx, "docker", dockerArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &ExecResult{
		Stdout: strings.TrimSpace(stdout.String()),
		Stderr: strings.TrimSpace(stderr.String()),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}

		return result, &sdk.ExecutionError{
			Command:  binary,
			Args:     args,
			ExitCode: result.ExitCode,
			Stderr:   result.Stderr,
			Err:      err,
		}
	}

	return result, nil
}

// IsDockerAvailable checks if Docker is available and running.
func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	return err == nil
}

// IsContainerRunning checks if a specific container is running.
func IsContainerRunning(container string) bool {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", container)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

// GetContainerID returns the full container ID for a container name.
func GetContainerID(container string) (string, error) {
	cmd := exec.Command("docker", "inspect", "-f", "{{.Id}}", container)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get container ID: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ListContainers returns a list of running container names.
func ListContainers() ([]string, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	names := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" {
			result = append(result, name)
		}
	}

	return result, nil
}

// FindSekaiContainer attempts to find a SEKAI container by common naming patterns.
func FindSekaiContainer() (string, error) {
	containers, err := ListContainers()
	if err != nil {
		return "", err
	}

	// Common patterns for SEKAI containers
	patterns := []string{
		"sekai",
		"kira",
		"sekaid",
		"interx",
	}

	for _, container := range containers {
		lower := strings.ToLower(container)
		for _, pattern := range patterns {
			if strings.Contains(lower, pattern) {
				return container, nil
			}
		}
	}

	return "", fmt.Errorf("no SEKAI container found")
}
