package grpcapi

import (
	"context"

	pb2 "github.com/Sugar-pack/orders-manager/pkg/pb"
	"github.com/Sugar-pack/users-manager/pkg/logging"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type OrderService struct {
	pb2.OrdersManagerServiceServer
	dbConn *sqlx.DB
}

func (s OrderService) InsertOrder(cxt context.Context, order *pb2.Order) (*pb2.OrderTnxResponse, error) {
	logger := logging.FromContext(cxt)
	logger.Info("ReceiveOrder")
	orderID := uuid.New()
	txID := uuid.New()
	query := "INSERT INTO orders ( id,  user_id, label, created_at ) VALUES ($1, $2, $3, $4)"
	_, err := s.dbConn.Exec(query, orderID.String(), order.UserId, order.Label, order.CreatedAt.AsTime())
	if err != nil {
		logger.Error("Failed to insert order ", err)

		return nil, err
	}

	return &pb2.OrderTnxResponse{
		Id:  orderID.String(),
		Tnx: txID.String(),
	}, nil
}
