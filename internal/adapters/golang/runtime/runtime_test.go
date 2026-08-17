package runtime

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCollectRuntimeData(t *testing.T) {
	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "go.mod"), []byte("module example\n\ngo 1.25\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte("package main\n\nfunc main() { select {} }\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	observed, err := NewRuntime(context.Background(), sourceDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = observed.Process.Cleanup() })
	done := make(chan error, 1)
	go func() { done <- observed.Process.Run() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		err = observed.CollectMetrics(ctx)
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal(err)
		case <-time.After(10 * time.Millisecond):
		}
	}

	results := make(chan error, 2)
	go func() { results <- observed.CollectProfile(ctx, "cpu", time.Second) }()
	go func() { results <- observed.CollectTrace(ctx, time.Second) }()
	for range 2 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	if observed.Profiles["cpu"] == nil || observed.Trace == nil {
		t.Fatal("collector returned incomplete runtime data")
	}
	if err := observed.Process.Stop(); err != nil {
		t.Fatal(err)
	}
	<-done
}
