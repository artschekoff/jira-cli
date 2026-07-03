package acli

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRunner_AllowedSubcommand(t *testing.T) {
	logger := zap.NewNop()

	r := NewRunner(logger, WithBinPath("echo"), WithTimeout(5*time.Second))
	out, err := r.Run(context.Background(), []string{"jira", "workitem", "search"}, []string{"--jql", "project=TEST"})
	require.NoError(t, err)
	assert.Contains(t, out, "jira workitem search")
}

func TestRunner_DisallowedSubcommand(t *testing.T) {
	logger := zap.NewNop()
	r := NewRunner(logger)

	_, err := r.Run(context.Background(), []string{"jira", "workitem", "delete"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in the allowed list")
}

func TestRunner_Timeout(t *testing.T) {
	logger := zap.NewNop()
	r := NewRunner(logger, WithBinPath("sleep"), WithTimeout(50*time.Millisecond))

	_, err := r.Run(context.Background(), []string{"jira", "workitem", "search"}, []string{"10"})
	require.Error(t, err)
}

func TestRunner_AllowedAuthSubcommands(t *testing.T) {
	logger := zap.NewNop()
	r := NewRunner(logger, WithBinPath("echo"), WithTimeout(5*time.Second))

	for _, path := range [][]string{
		{"jira", "auth", "login"},
		{"jira", "auth", "logout"},
	} {
		out, err := r.Run(context.Background(), path, nil)
		require.NoError(t, err, "path=%v", path)
		assert.Contains(t, out, "jira auth")
	}
}

func TestRunner_RunInteractive_Allowed(t *testing.T) {
	logger := zap.NewNop()
	// "true" always exits 0 and reads no stdin, safe for a headless test.
	r := NewRunner(logger, WithBinPath("true"))
	err := r.RunInteractive(context.Background(), []string{"jira", "auth", "login"}, nil)
	require.NoError(t, err)
}

func TestRunner_RunInteractive_Disallowed(t *testing.T) {
	logger := zap.NewNop()
	r := NewRunner(logger)
	err := r.RunInteractive(context.Background(), []string{"jira", "workitem", "delete"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in the allowed list")
}
