package grpcapi

import (
	"context"
	"database/sql"
	"errors"
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

	"github.com/Sugar-pack/orders-manager/internal/config"
	"github.com/Sugar-pack/orders-manager/internal/db"
	"github.com/Sugar-pack/orders-manager/internal/migration"
	"github.com/Sugar-pack/orders-manager/pkg/pb"
)

//e2e test
func TestTnxConfirmingService_SendConfirmation_True(t *testing.T) {
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
		Cmd: []string{"postgres",
			"-c", "log_statement=all",
			"-c", "log_destination=stderr",
			"--max_prepared_transactions=100",
		},
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

	address := "localhost:8081"
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

	testTx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatalf("begin transaction failed: '%s'", err)
	}
	defer func(testTx *sqlx.Tx) {
		errRollback := testTx.Rollback()
		if errRollback != nil {
			if !errors.Is(errRollback, sql.ErrTxDone) {
				t.Fatalf("rollback failed: '%s'", errRollback)
			}
		}
	}(testTx)

	order := &db.Order{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Label:     "e2e test",
		CreatedAt: time.Time{},
	}
	txID := uuid.New().String()
	_, err = testTx.NamedExecContext(ctx, db.InsertOrderQuery, order)
	if err != nil {
		t.Fatalf("insert order failed: '%s'", err)
	}
	_, err = testTx.ExecContext(ctx, fmt.Sprintf("PREPARE TRANSACTION '%s'", txID))
	if err != nil {
		t.Fatalf("prepare transaction failed: '%s'", err)
	}

	// preparation done

	transactionClient := pb.NewTnxConfirmingServiceClient(grpcConn)
	confirmation := &pb.Confirmation{
		Tnx:    txID,
		Commit: true,
	}
	confirmationResponse, err := transactionClient.SendConfirmation(ctx, confirmation)
	assert.NoError(t, err)
	assert.NotNil(t, confirmationResponse, "confirmationResponse should be empty")

	var orderID, label, userID, createdAt string

	err = dbConn.QueryRowContext(ctx, fmt.Sprintf("SELECT id, label, user_id, created_at FROM orders WHERE id = '%s'", order.ID)).Scan(&orderID, &label, &userID, &createdAt)
	if err != nil {
		t.Fatalf("select order failed: '%s'", err)
	}
	assert.Equal(t, order.ID.String(), orderID, "order id should be equal")
	assert.Equal(t, order.Label, label, "order label should be equal")
	assert.Equal(t, order.UserID.String(), userID, "order user id should be equal")
	assert.Equal(t, order.CreatedAt.Format(time.RFC3339), createdAt, "order created at should be equal")
}
