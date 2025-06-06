package migration

import (
	"context"
	"testing"

	"github.com/Sugar-pack/orders-manager/internal/config"
)

func TestApply_ConnectError(t *testing.T) {
	ctx := context.Background()
	conf := &config.DB{ConnString: "invalid"}
	err := Apply(ctx, conf)
	if err == nil {
		t.Fatal("expected error")
	}
}
