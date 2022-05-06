package grpcapi

import (
	"context"

	"github.com/google/uuid"

	"github.com/Sugar-pack/orders-manager/internal/repository"

	"github.com/Sugar-pack/users-manager/pkg/logging"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Sugar-pack/orders-manager/internal/tracing"
	"github.com/Sugar-pack/orders-manager/pkg/pb"
)

type TnxConfirmingService struct {
	pb.TnxConfirmingServiceServer
	Repo repository.OrderRepoWith2PC
}

func (s *TnxConfirmingService) SendConfirmation(ctx context.Context,
	confirmation *pb.Confirmation,
) (*pb.ConfirmationResponse, error) {
	ctx, span := otel.Tracer(tracing.TracerName).Start(ctx, "SendConfirmation")
	defer span.End()

	logger := logging.FromContext(ctx)
	logger.Info("Confirmation request received")
	TnxID := confirmation.Tnx
	TnxIdParsed, err := uuid.Parse(TnxID)
	if err != nil {
		logger.WithError(err).Error("Failed to parse TnxID as UUID")

		return nil,
			status.Error(codes.InvalidArgument, "Failed to parse TnxID as UUID") //nolint:wrapcheck //should be wrapped as is
	}

	if confirmation.Commit {
		errCommit := s.Repo.CommitInsertTransaction(ctx, TnxIdParsed)
		if errCommit != nil {
			logger.WithError(errCommit).Error("commit tx failed")

			return nil, status.Error(codes.Internal, "commit tx failed") //nolint:wrapcheck // should be wrapped as is
		}
	} else {
		errRollback := s.Repo.RollbackInsertTransaction(ctx, TnxIdParsed)
		if errRollback != nil {
			logger.WithError(errRollback).Error("rollback tx failed")

			return nil, status.Error(codes.Internal, "rollback tx failed") //nolint:wrapcheck // should be wrapped as is
		}
	}

	return &pb.ConfirmationResponse{}, nil
}
