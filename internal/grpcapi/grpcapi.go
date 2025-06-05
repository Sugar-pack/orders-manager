package grpcapi

import (
	"context"
	"net"

	"github.com/Sugar-pack/users-manager/pkg/logging"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	"github.com/Sugar-pack/orders-manager/internal/config"
	"github.com/Sugar-pack/orders-manager/internal/repository"
	"github.com/Sugar-pack/orders-manager/internal/tracing"
	"github.com/Sugar-pack/orders-manager/pkg/pb"
)

func CreateServer(logger logging.Logger, repo repository.OrderRepoWith2PC) (*grpc.Server, error) {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.WithLogger(logger),
			logging.WithUniqTraceID,
			logging.LogBoundaries,
			otelgrpc.UnaryServerInterceptor(),
		),
	)

	orderService := &OrderService{
		Repo: repo,
	}
	pb.RegisterOrdersManagerServiceServer(grpcServer, orderService)

	transactionService := &TnxConfirmingService{
		Repo: repo,
	}
	pb.RegisterTnxConfirmingServiceServer(grpcServer, transactionService)

	return grpcServer, nil
}

func ServeWithTrace(ctx context.Context, server *grpc.Server, appConfig *config.API) error {
	logger := logging.FromContext(ctx)
	lis, err := net.Listen("tcp", appConfig.Bind)
	if err != nil {
		return err //nolint:wrapcheck //should be wrapped in main
	}
	tracingProvider, err := tracing.InitTracing(ctx, logger)
	if err != nil {
		return err //nolint:wrapcheck //should be wrapped in main
	}
	defer func() {
		if stopErr := tracingProvider.Shutdown(ctx); stopErr != nil {
			logger.WithError(stopErr).Error("shutting down tracer provider failed")
		}
	}()

	if serveErr := server.Serve(lis); serveErr != nil {
		return serveErr //nolint:wrapcheck //should be wrapped in main
	}

	return nil
}
