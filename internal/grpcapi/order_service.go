package grpcapi

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Sugar-pack/users-manager/pkg/logging"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/Sugar-pack/orders-manager/pkg/pb"
)

type OrderService struct {
	pb.OrdersManagerServiceServer
	dbConn *sqlx.DB
}

func (s OrderService) InsertOrder(ctx context.Context, order *pb.Order) (*pb.OrderTnxResponse, error) {
	logger := logging.FromContext(ctx)
	logger.Info("ReceiveOrder")
	orderID := uuid.New()
	txID := uuid.New()

	query := "INSERT INTO orders ( id,  user_id, label, created_at ) VALUES ($1, $2, $3, $4)"

	transaction, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err)

		return nil, fmt.Errorf("prepare tx failed %w", err)
	}
	defer func(tx *sql.Tx) {
		errRollback := tx.Rollback()
		if errRollback != nil {
			logger.Error(errRollback)
		}
	}(transaction)

	_, err = transaction.ExecContext(ctx, query, orderID.String(), order.UserId, order.Label, order.CreatedAt.AsTime())
	if err != nil {
		logger.Error("init transaction failed ", err)

		return nil, fmt.Errorf("init tx failed %w", err)
	}
	_, err = transaction.ExecContext(ctx, fmt.Sprintf("PREPARE TRANSACTION '%s'", txID.String()))
	if err != nil {
		logger.Error("prepare tx failed ", err)

		return nil, fmt.Errorf("prepare tx failed %w", err)
	}
	err = transaction.Commit()
	if err != nil {
		logger.Error("commit tx failed ", err)

		return nil, fmt.Errorf("commit tx failed %w", err)
	}

	return &pb.OrderTnxResponse{
		Id:  orderID.String(),
		Tnx: txID.String(),
	}, nil
}
