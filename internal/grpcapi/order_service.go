package grpcapi

import (
	"context"

	"github.com/Sugar-pack/users-manager/pkg/logging"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Sugar-pack/orders-manager/internal/repository"
	"github.com/Sugar-pack/orders-manager/internal/tracing"
	"github.com/Sugar-pack/orders-manager/pkg/pb"
)

type OrderService struct {
	pb.OrdersManagerServiceServer
	Repo repository.OrderRepoWith2PC
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

	dbOrder := &repository.Order{
		ID:        orderID,
		UserID:    parseUserID,
		Label:     order.Label,
		CreatedAt: order.CreatedAt.AsTime(),
	}

	err = s.Repo.PrepareInsertOrder(ctx, dbOrder, txID)
	if err != nil {
		logger.WithError(err).Error("Error preparing insert order")

		return nil, status.Error(codes.Internal, "error preparing insert order") //nolint:wrapcheck // should be wrapped as is
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
	parseOrderID, err := uuid.Parse(orderID)
	if err != nil {
		logger.WithError(err).Error("Error parsing order id")

		return nil, status.Error(codes.Internal, "error parsing order id") //nolint:wrapcheck // should be wrapped as is
	}
	order, err := s.Repo.GetOrder(ctx, parseOrderID)
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
