package db

import (
	"context"
	"testing"
	"time"

	"github.com/mmk31585/workout-tracker/internal/config"
)

func TestQueryTimeoutContext_WithLoadedConfig(t *testing.T) {
	config.SetForTest(&config.Config{
		DB: config.DBConfig{QueryTimeout: 30},
	})
	defer config.ResetForTest()

	ctx, cancel := QueryTimeoutContext(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected context to have a deadline")
	}

	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > 31*time.Second {
		t.Fatalf("expected deadline in ~30s, got %v", remaining)
	}
}

func TestQueryTimeoutContext_DefaultWhenConfigNotLoaded(t *testing.T) {
	config.ResetForTest()

	if config.GetConfig() != nil {
		t.Fatal("precondition failed: expected config to be nil")
	}

	ctx, cancel := QueryTimeoutContext(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected context to have a deadline")
	}

	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > 6*time.Second {
		t.Fatalf("expected default deadline of ~5s, got %v", remaining)
	}
}

func TestQueryTimeoutContext_PreservesParentValues(t *testing.T) {
	config.ResetForTest()

	type keyT struct{}
	parent := context.WithValue(context.Background(), keyT{}, "v")

	ctx, cancel := QueryTimeoutContext(parent)
	defer cancel()

	if got, ok := ctx.Value(keyT{}).(string); !ok || got != "v" {
		t.Fatalf("expected parent value to be preserved, got %q (ok=%v)", got, ok)
	}
}
