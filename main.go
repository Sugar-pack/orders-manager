package main

import (
	"context"
	"log"
	"net"

	"github.com/Sugar-pack/users-manager/pkg/logging"

	"github.com/Sugar-pack/orders-manager/internal/config"
	"github.com/Sugar-pack/orders-manager/internal/db"
	"github.com/Sugar-pack/orders-manager/internal/grpcapi"
	"github.com/Sugar-pack/orders-manager/internal/migration"
)

func main() {
	ctx := context.Background()
	appConfig, err := config.GetAppConfig()
	if err != nil {
		log.Fatal(err)

		return
	}

	logger := logging.GetLogger()
	ctx = logging.WithContext(ctx, logger)
	err = migration.Apply(ctx, appConfig.Db)

	if err != nil {
		log.Fatal(err)

		return
	}

	dbConn, err := db.Connect(ctx, appConfig.Db)
	if err != nil {
		log.Fatal(err)

		return
	}

	server, err := grpcapi.CreateServer(logger, dbConn)
	if err != nil {
		log.Fatal(err)

		return
	}

	lis, err := net.Listen("tcp", appConfig.API.Bind)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)

		return
	}

	if serveErr := server.Serve(lis); serveErr != nil {
		log.Fatalf("failed to listen: %v", err)

		return
	}
}
