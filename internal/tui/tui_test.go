package tui

import (
	"context"
	"testing"

	"github.com/briheet/sen/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNewTuiRejectsUnknownTheme(t *testing.T) {
	configuration := &config.Config{
		Project: config.Project{Name: "project", Theme: "missing"},
	}

	_, err := NewTui(context.Background(), configuration)
	require.ErrorContains(t, err, `unknown theme "missing"`)
}
