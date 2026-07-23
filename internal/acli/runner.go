// Package acli provides a safe wrapper around the Atlassian CLI (acli) binary.
package acli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	defaultBinPath = "acli"
	defaultTimeout = 30 * time.Second
)

// allowedSubcmds is the set of acli subcommand paths that this runner may
// execute. Values are space-joined prefix strings, e.g. "jira workitem search".
// User-supplied data is only ever passed as flag values, never as subcommands.
var allowedSubcmds = map[string]bool{
	"jira workitem search":         true,
	"jira workitem view":           true,
	"jira workitem create":         true,
	"jira workitem edit":           true,
	"jira workitem transition":     true,
	"jira workitem assign":         true,
	"jira workitem comment create": true,
	"jira workitem comment update": true,
	"jira workitem comment list":   true,
	"jira sprint view":             true,
	"jira sprint list-workitems":   true,
	"jira sprint create":           true,
	"jira project list":            true,
	"jira project view":            true,
	"jira board search":            true,
	"jira board list-sprints":      true,
	"jira auth login":              true,
	"jira auth logout":             true,
	"jira auth status":             true,
}

// Runner executes acli subcommands via exec.CommandContext.
type Runner struct {
	binPath string
	timeout time.Duration
	logger  *zap.Logger
}

// Option configures a Runner.
type Option func(*Runner)

// WithBinPath overrides the acli binary path (default: "acli" resolved via PATH).
func WithBinPath(path string) Option {
	return func(r *Runner) { r.binPath = path }
}

// WithTimeout overrides the per-command execution timeout (default 30s).
func WithTimeout(d time.Duration) Option {
	return func(r *Runner) { r.timeout = d }
}

// NewRunner creates a Runner with the given logger and options.
func NewRunner(logger *zap.Logger, opts ...Option) *Runner {
	r := &Runner{
		binPath: defaultBinPath,
		timeout: defaultTimeout,
		logger:  logger,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Run executes an acli subcommand with the provided arguments.
//
// subCmdPath is a slice of path tokens, e.g. ["jira", "workitem", "search"].
// flags are additional flag/value arguments, e.g. ["--jql", "project=TEAM", "--json"].
//
// Returns stdout output on success. On non-zero exit the combined stderr is
// wrapped into the returned error.
func (r *Runner) Run(ctx context.Context, subCmdPath []string, flags []string) (string, error) {
	key := strings.Join(subCmdPath, " ")
	if !allowedSubcmds[key] {
		return "", fmt.Errorf("acli subcommand %q is not in the allowed list", key)
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	args := append(subCmdPath, flags...) //nolint:gocritic
	cmd := exec.CommandContext(ctx, r.binPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	r.logger.Debug("acli exec", zap.String("cmd", key), zap.Strings("flags", flags))

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		r.logger.Error("acli exec failed",
			zap.String("cmd", key),
			zap.String("stderr", errMsg),
			zap.Error(err),
		)
		return "", fmt.Errorf("acli %s: %s", key, errMsg)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// RunInteractive executes an acli subcommand attached to the parent process's
// stdin/stdout/stderr. Use this for commands that prompt the user (e.g.
// "jira auth login"). No timeout is applied — the human sets the pace via ctx.
//
// The allowlist is enforced the same way as Run.
func (r *Runner) RunInteractive(ctx context.Context, subCmdPath []string, flags []string) error {
	key := strings.Join(subCmdPath, " ")
	if !allowedSubcmds[key] {
		return fmt.Errorf("acli subcommand %q is not in the allowed list", key)
	}

	args := append(subCmdPath, flags...) //nolint:gocritic
	cmd := exec.CommandContext(ctx, r.binPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	r.logger.Debug("acli exec (interactive)", zap.String("cmd", key), zap.Strings("flags", flags))

	if err := cmd.Run(); err != nil {
		r.logger.Error("acli exec failed",
			zap.String("cmd", key),
			zap.Error(err),
		)
		return fmt.Errorf("acli %s: %w", key, err)
	}
	return nil
}
