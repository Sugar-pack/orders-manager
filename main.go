package main

import (
	"context"
	"log"

	"github.com/Sugar-pack/users-manager/pkg/logging"

	"github.com/Sugar-pack/orders-manager/internal/config"
	"github.com/Sugar-pack/orders-manager/internal/db"
	"github.com/Sugar-pack/orders-manager/internal/grpcapi"
	"github.com/Sugar-pack/orders-manager/internal/migration"
	"github.com/Sugar-pack/orders-manager/internal/repository"
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

	repo := repository.NewPsqlRepository(dbConn)

	server, err := grpcapi.CreateServer(logger, repo)
	if err != nil {
		log.Fatal(err)

		return
	}

	err = grpcapi.ServeWithTrace(ctx, server, appConfig.API)
	if err != nil {
		log.Fatal(err)

		return
	}
}
