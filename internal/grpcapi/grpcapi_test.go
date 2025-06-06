package grpcapi

import (
	"context"
	"testing"
	"time"

	"github.com/Sugar-pack/orders-manager/internal/config"
	"github.com/Sugar-pack/orders-manager/internal/mock"
	"github.com/Sugar-pack/users-manager/pkg/logging"
)

func TestCreateServer(t *testing.T) {
	logger := logging.GetLogger()
	repo := &mock.OrderRepoWith2PC{}
	srv, err := CreateServer(logger, repo)
	if err != nil {
		t.Fatalf("CreateServer error: %v", err)
	}
	if srv == nil {
		t.Fatal("server is nil")
	}
}

func TestServeWithTrace(t *testing.T) {
	logger := logging.GetLogger()
	ctx := logging.WithContext(context.Background(), logger)
	repo := &mock.OrderRepoWith2PC{}
	srv, err := CreateServer(logger, repo)
	if err != nil {
		t.Fatalf("CreateServer error: %v", err)
	}
	cfg := &config.API{Bind: "localhost:0"}
	done := make(chan error)
	go func() {
		done <- ServeWithTrace(ctx, srv, cfg)
	}()
	time.Sleep(100 * time.Millisecond)
	srv.GracefulStop()
	if err := <-done; err != nil {
		t.Fatalf("ServeWithTrace error: %v", err)
	}
}

func TestServeWithTrace_ListenErr(t *testing.T) {
	logger := logging.GetLogger()
	ctx := logging.WithContext(context.Background(), logger)
	srv, _ := CreateServer(logger, &mock.OrderRepoWith2PC{})
	cfg := &config.API{Bind: "localhost:999999"}
	if err := ServeWithTrace(ctx, srv, cfg); err == nil {
		t.Fatal("expected error")
	}
}
