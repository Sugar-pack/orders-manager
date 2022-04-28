package grpcapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Sugar-pack/users-manager/pkg/logging"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Sugar-pack/orders-manager/internal/db"
	"github.com/Sugar-pack/orders-manager/internal/tracing"
	"github.com/Sugar-pack/orders-manager/pkg/pb"
)

type OrderService struct {
	pb.OrdersManagerServiceServer
	dbConn *sqlx.DB
}

func (s *OrderService) InsertOrder(ctx context.Context, order *pb.Order) (*pb.OrderTnxResponse, error) {
	ctx, span := otel.Tracer(tracing.TracerName).Start(ctx, "InsertOrder")
	defer span.End()
	logger := logging.FromContext(ctx)
	logger.Info("ReceiveOrder")
	orderID := uuid.New()
	txID := uuid.New()
	parseUserID, err := uuid.Parse(order.UserId)
	if err != nil {
		logger.WithError(err).Error("Error parsing user id")

		return nil, status.Error(codes.Internal, "error parsing user id") //nolint:wrapcheck // should be wrapped as is
	}

	dbOrder := &db.Order{
		ID:        orderID,
		UserID:    parseUserID,
		Label:     order.Label,
		CreatedAt: order.CreatedAt.AsTime(),
	}

	transaction, err := s.dbConn.BeginTxx(ctx, nil)
	if err != nil {
		logger.WithError(err).Error("prepare tx failed")

		return nil, fmt.Errorf("prepare tx failed %w", err)
	}

	defer func(tx *sqlx.Tx) {
		// prepared transaction is independent of commit or rollback. here we just release tx
		errRollback := tx.Rollback()
		if errRollback != nil {
			if !errors.Is(errRollback, sql.ErrTxDone) {
				logger.WithError(err).Error("rollback tx failed")
			}
		}
	}(transaction)

	err = db.InsertOrder(ctx, transaction, dbOrder)
	if err != nil {
		logger.WithError(err).Error("init transaction failed")

		return nil, status.Error(codes.Internal, "init tx failed") //nolint:wrapcheck // should be wrapped as is
	}
	err = db.PrepareTransaction(ctx, transaction, txID.String())
	if err != nil {
		logger.WithError(err).Error("prepare tx failed")
		defer func(ctx context.Context, dbConn sqlx.ExecerContext, txID string) {
			errRollBack := db.RollBackTransaction(ctx, dbConn, txID)
			if errRollBack != nil {
				logger.WithError(errRollBack).Error("rollback prepared tx failed")
			}
		}(ctx, transaction, txID.String())

		return nil, status.Error(codes.Internal, "prepare tx failed") //nolint:wrapcheck // should be wrapped as is
	}

	return &pb.OrderTnxResponse{
		Id:  orderID.String(),
		Tnx: txID.String(),
	}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, request *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	ctx, span := otel.Tracer(tracing.TracerName).Start(ctx, "GetOrder")
	defer span.End()
	logger := logging.FromContext(ctx)
	logger.Info("GetOrder")
	orderID := request.GetId()
	order, err := db.GetOrder(ctx, orderID, s.dbConn)
	if err != nil {
		logger.WithError(err).Error("GetOrder error")

		return nil, status.Error(codes.Internal, "Cant get order by id") //nolint:wrapcheck // should be wrapped as is
	}

	return &pb.OrderResponse{
		Id:        orderID,
		UserId:    order.UserID.String(),
		Label:     order.Label,
		CreatedAt: timestamppb.New(order.CreatedAt),
	}, nil
}
