package grpcapi

import (
	"context"

	"go.opentelemetry.io/otel"

	"github.com/Sugar-pack/orders-manager/internal/tracing"

	"github.com/Sugar-pack/users-manager/pkg/logging"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Sugar-pack/orders-manager/internal/db"
	"github.com/Sugar-pack/orders-manager/pkg/pb"
)

type TnxConfirmingService struct {
	pb.TnxConfirmingServiceServer
	dbConn *sqlx.DB
}

func (s *TnxConfirmingService) SendConfirmation(ctx context.Context,
	confirmation *pb.Confirmation,
) (*pb.ConfirmationResponse, error) {
	ctx, span := otel.Tracer(tracing.TracerName).Start(ctx, "SendConfirmation")
	defer span.End()

	logger := logging.FromContext(ctx)
	logger.Info("Confirmation request received")
	TnxID := confirmation.Tnx

	if confirmation.Commit {
		errCommit := db.CommitTransaction(ctx, s.dbConn, TnxID)
		if errCommit != nil {
			logger.WithError(errCommit).Error("commit tx failed")

			return nil, status.Error(codes.Internal, "commit tx failed") //nolint:wrapcheck // should be wrapped as is
		}
	} else {
		errRollback := db.RollBackTransaction(ctx, s.dbConn, TnxID)
		if errRollback != nil {
			logger.WithError(errRollback).Error("rollback tx failed")

			return nil, status.Error(codes.Internal, "rollback tx failed") //nolint:wrapcheck // should be wrapped as is
		}
	}

	return &pb.ConfirmationResponse{}, nil
}
