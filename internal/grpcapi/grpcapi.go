package grpcapi

import (
	"github.com/Sugar-pack/users-manager/pkg/logging"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"

	"github.com/Sugar-pack/orders-manager/pkg/pb"
)

func CreateServer(logger logging.Logger, dbConn *sqlx.DB) (*grpc.Server, error) {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.WithLogger(logger),
			logging.WithUniqTraceID,
			logging.LogBoundaries,
		),
	)

	orderService := &OrderService{
		dbConn: dbConn,
	}
	pb.RegisterOrdersManagerServiceServer(grpcServer, orderService)

	return grpcServer, nil
}
