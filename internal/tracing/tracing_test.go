package tracing

import (
	"context"
	"testing"

	"github.com/Sugar-pack/users-manager/pkg/logging"
)

func TestInitTracing(t *testing.T) {
	logger := logging.GetLogger()
	ctx := logging.WithContext(context.Background(), logger)
	provider, err := InitTracing(ctx, logger)
	if err != nil {
		t.Fatalf("InitTracing error: %v", err)
	}
	if provider == nil {
		t.Fatal("provider is nil")
	}
	if err := provider.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
}
