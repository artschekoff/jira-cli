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
