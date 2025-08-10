package utils

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// CommandResult represents the result of a command execution
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

// ExecuteCommand executes a command with the given arguments
func ExecuteCommand(ctx context.Context, name string, args ...string) *CommandResult {
	cmd := exec.CommandContext(ctx, name, args...)
	
	stdout, err := cmd.Output()
	result := &CommandResult{
		Stdout: strings.TrimSpace(string(stdout)),
	}
	
	if err != nil {
		result.Error = err
		if exitError, ok := err.(*exec.ExitError); ok {
			result.Stderr = strings.TrimSpace(string(exitError.Stderr))
			result.ExitCode = exitError.ExitCode()
		}
	}
	
	return result
}

// ExecuteCommandWithTimeout executes a command with a timeout
func ExecuteCommandWithTimeout(timeout time.Duration, name string, args ...string) *CommandResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	return ExecuteCommand(ctx, name, args...)
}

// IsCommandAvailable checks if a command is available in PATH
func IsCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// GetCommandPath returns the full path of a command
func GetCommandPath(command string) (string, error) {
	return exec.LookPath(command)
}

// GetCommandVersion tries to get the version of a command
func GetCommandVersion(ctx context.Context, command string, versionArgs ...string) (string, error) {
	if len(versionArgs) == 0 {
		versionArgs = []string{"--version"}
	}
	
	result := ExecuteCommand(ctx, command, versionArgs...)
	if result.Error != nil {
		return "", result.Error
	}
	
	// Try to extract version from output
	output := result.Stdout
	if output == "" {
		output = result.Stderr
	}
	
	return strings.TrimSpace(output), nil
}

// WhichCommand finds the path of a command (cross-platform which)
func WhichCommand(command string) (string, bool) {
	path, err := exec.LookPath(command)
	return path, err == nil
}

// GetShell returns the appropriate shell for the current OS
func GetShell() string {
	if runtime.GOOS == "windows" {
		return "cmd"
	}
	return "sh"
}

// GetShellArgs returns the appropriate shell arguments for command execution
func GetShellArgs(command string) []string {
	if runtime.GOOS == "windows" {
		return []string{"/c", command}
	}
	return []string{"-c", command}
}

// ExecuteShellCommand executes a command through the system shell
func ExecuteShellCommand(ctx context.Context, command string) *CommandResult {
	shell := GetShell()
	args := GetShellArgs(command)
	return ExecuteCommand(ctx, shell, args...)
}

// SanitizeCommand sanitizes a command string to prevent injection
func SanitizeCommand(command string) string {
	// Remove potentially dangerous characters
	dangerous := []string{";", "&", "|", "`", "$", "(", ")", "<", ">"}
	for _, char := range dangerous {
		command = strings.ReplaceAll(command, char, "")
	}
	return strings.TrimSpace(command)
}
