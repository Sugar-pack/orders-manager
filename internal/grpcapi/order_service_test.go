package grpcapi

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/Sugar-pack/users-manager/pkg/logging"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Sugar-pack/orders-manager/internal/config"
	"github.com/Sugar-pack/orders-manager/internal/db"
	"github.com/Sugar-pack/orders-manager/internal/migration"
	"github.com/Sugar-pack/orders-manager/pkg/pb"
)

//e2e test
func TestOrderService_InsertOrder_e2e(t *testing.T) {
	logger := logging.GetLogger()
	ctx := logging.WithContext(context.Background(), logger)

	// Prepare test environment. Look for end-section below
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	dbUser := "user_db"
	dbName := "orders_db"
	sslMode := "disable"
	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14.2",
		Env: []string{
			fmt.Sprintf("POSTGRES_USER=%s", dbUser),
			fmt.Sprintf("POSTGRES_DB=%s", dbName),
			"POSTGRES_HOST_AUTH_METHOD=trust",
			"listen_addresses = '*'",
		},
		Cmd: []string{"postgres", "-c", "log_statement=all", "-c", "log_destination=stderr", "--max_prepared_transactions=100"},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}
	defer func() {
		if purgeErr := pool.Purge(resource); purgeErr != nil {
			t.Fatalf("Could not purge resource: %s", purgeErr)
		}
	}()

	hostAndPort := resource.GetHostPort("5432/tcp")
	dbHost, dbPort, err := net.SplitHostPort(hostAndPort)
	if err != nil {
		t.Fatalf("split host-port '%s' failed: '%s'", hostAndPort, err)
	}
	dbConf := &config.DB{
		ConnString:       fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s", dbHost, dbPort, dbUser, dbName, sslMode),
		MigrationDirPath: "../../sql-migrations",
		MigrationTable:   "migrations",
		MaxOpenCons:      20,
		ConnMaxLifetime:  10 * time.Second,
	}

	var dbConn *sqlx.DB
	pool.MaxWait = 30 * time.Second
	if err = pool.Retry(func() error {
		dbConn, err = db.Connect(ctx, dbConf)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	defer func() {
		if disconnectErr := db.Disconnect(ctx, dbConn); disconnectErr != nil {
			log.Fatalf("disconnect failed: '%s'", err)
		}
	}()

	err = migration.Apply(ctx, dbConf)
	if err != nil {
		log.Fatalf("apply migrations failed: '%s'", err)
	}
	// Test environment prepared

	grpcServer, err := CreateServer(logger, dbConn)
	if err != nil {
		t.Fatalf("create grpc server failed: '%s'", err)
	}

	address := "localhost:8080"
	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatalf("listen failed: '%s'", err)
	}
	go func() {
		if errServe := grpcServer.Serve(listener); errServe != nil {
			log.Fatalf("serve failed: '%s'", errServe)
		}
	}()

	grpcConn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial failed: '%s'", err)
	}

	orderClient := pb.NewOrdersManagerServiceClient(grpcConn)

	userID := uuid.New().String()
	label := "e2e test"
	createdAt := timestamppb.Now()
	orderResponse, err := orderClient.InsertOrder(ctx, &pb.Order{
		UserId:    userID,
		Label:     label,
		CreatedAt: createdAt,
	})
	assert.NoError(t, err)
	assert.NotNil(t, orderResponse)

	txID := orderResponse.Tnx

	var dbTxId string
	err = dbConn.QueryRowxContext(ctx, `SELECT gid FROM pg_prepared_xacts WHERE gid = $1`, txID).Scan(&dbTxId)
	assert.NoError(t, err)
	assert.Equal(t, txID, dbTxId, "unexpected db tx id value")
}
