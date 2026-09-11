// Package execx provides helpers for running external commands.
package execx

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// ExecPath is the path to an executable to run.
type ExecPath string

// ExecOptions configures how a command is executed.
type ExecOptions struct {
	// Stdin, if set, is connected to the command's standard input.
	Stdin io.Reader
	// Env, if set, is appended to the current process's environment for the command.
	Env []string
	// Dir, if set, is the working directory for the command.
	Dir string
}

// String returns p as a plain string.
func (p ExecPath) String() string {
	return string(p)
}

// ExecuteWithOptions runs p with args, applying opts, and returns its combined
// stdout/stderr output. If the command fails, the returned error wraps the underlying
// error and includes the captured output.
func (p ExecPath) ExecuteWithOptions(ctx context.Context, opts ExecOptions, args ...string) (string, error) {
	//nolint:gosec // command execution is the intended purpose of this package
	cmd := exec.CommandContext(ctx, string(p), args...)

	if opts.Stdin != nil {
		cmd.Stdin = opts.Stdin
	}

	if opts.Env != nil {
		cmd.Env = append(os.Environ(), opts.Env...)
	}

	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("execution failed: %w (output: %s)", err, string(output))
	}

	return string(output), nil
}
