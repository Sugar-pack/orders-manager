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
	"github.com/Sugar-pack/orders-manager/internal/repository"
	"github.com/Sugar-pack/orders-manager/pkg/pb"
)

const (
	dbUser = "user_db"
	dbName = "orders_db"
)

type NamedExecutorContext interface {
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

func InsertOrder(t *testing.T, ctx context.Context, dbConn NamedExecutorContext, order *repository.Order) error {
	t.Helper()
	_, err := dbConn.NamedExecContext(ctx,
		"INSERT INTO orders ( id,  user_id, label, created_at ) VALUES (:id, :user_id, :label, :created_at)", order)

	return err //nolint:wrapcheck // should be wrapped in service layer
}

func PSQLResource(t *testing.T) (*dockertest.Pool, *dockertest.Resource) {
	t.Helper()
	// Prepare test environment. Look for end-section below
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

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
	return pool, resource
}

func DBConnection(ctx context.Context, t *testing.T, pool *dockertest.Pool, resource *dockertest.Resource) *sqlx.DB {
	t.Helper()

	sslMode := "disable"

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

	err = migration.Apply(ctx, dbConf)
	if err != nil {
		log.Fatalf("apply migrations failed: '%s'", err)
	}
	return dbConn
}

func GRPCConnection(ctx context.Context, t *testing.T, dbConn *sqlx.DB, address string) *grpc.ClientConn {
	t.Helper()

	logger := logging.FromContext(ctx)
	grpcServer, err := CreateServer(logger, dbConn)
	if err != nil {
		t.Fatalf("create grpc server failed: '%s'", err)
	}

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
	return grpcConn
}

//e2e test
func TestTnxConfirmingService_SendConfirmation_True(t *testing.T) {
	logger := logging.GetLogger()
	ctx := logging.WithContext(context.Background(), logger)

	pool, resource := PSQLResource(t)
	defer func() {
		if purgeErr := pool.Purge(resource); purgeErr != nil {
			t.Fatalf("Could not purge resource: %s", purgeErr)
		}
	}()

	dbConn := DBConnection(ctx, t, pool, resource)

	defer func() {
		if disconnectErr := db.Disconnect(ctx, dbConn); disconnectErr != nil {
			log.Fatalf("disconnect failed: '%s'", disconnectErr)
		}
	}()

	// address for each test should be different
	address := "localhost:8081"
	grpcConn := GRPCConnection(ctx, t, dbConn, address)

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

	order := &repository.Order{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Label:     "e2e test",
		CreatedAt: time.Time{},
	}
	txID := uuid.New().String()
	err = InsertOrder(t, ctx, testTx, order)
	if err != nil {
		t.Fatalf("insert order failed: '%s'", err)
	}
	_, err = testTx.ExecContext(ctx, fmt.Sprintf("PREPARE TRANSACTION '%s'", txID))
	if err != nil {
		t.Fatalf("prepare transaction failed: '%s'", err)
	}
	// prepared transaction is commit-independent
	//err = testTx.Commit()
	//if err != nil {
	//	t.Fatalf("commit failed: '%s'", err)
	//}

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

//e2e test
func TestTnxConfirmingService_SendConfirmation_False(t *testing.T) {
	logger := logging.GetLogger()
	ctx := logging.WithContext(context.Background(), logger)

	pool, resource := PSQLResource(t)
	defer func() {
		if purgeErr := pool.Purge(resource); purgeErr != nil {
			t.Fatalf("Could not purge resource: %s", purgeErr)
		}
	}()

	dbConn := DBConnection(ctx, t, pool, resource)

	defer func() {
		if disconnectErr := db.Disconnect(ctx, dbConn); disconnectErr != nil {
			log.Fatalf("disconnect failed: '%s'", disconnectErr)
		}
	}()

	// address for each test should be different
	address := "localhost:8082"
	grpcConn := GRPCConnection(ctx, t, dbConn, address)

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

	order := &repository.Order{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Label:     "e2e test",
		CreatedAt: time.Time{},
	}
	txID := uuid.New().String()
	err = InsertOrder(t, ctx, testTx, order)
	if err != nil {
		t.Fatalf("insert order failed: '%s'", err)
	}
	_, err = testTx.ExecContext(ctx, fmt.Sprintf("PREPARE TRANSACTION '%s'", txID))
	if err != nil {
		t.Fatalf("prepare transaction failed: '%s'", err)
	}
	// prepared transaction is commit-independent
	//err = testTx.Commit()
	//if err != nil {
	//	t.Fatalf("commit failed: '%s'", err)
	//}

	// preparation done

	transactionClient := pb.NewTnxConfirmingServiceClient(grpcConn)
	confirmation := &pb.Confirmation{
		Tnx:    txID,
		Commit: false,
	}
	confirmationResponse, err := transactionClient.SendConfirmation(ctx, confirmation)
	assert.NoError(t, err)
	assert.NotNil(t, confirmationResponse, "confirmationResponse should be empty")

	var orderID, label, userID, createdAt string

	err = dbConn.QueryRowContext(ctx, fmt.Sprintf("SELECT id, label, user_id, created_at FROM orders WHERE id = '%s'", order.ID)).Scan(&orderID, &label, &userID, &createdAt)
	assert.ErrorIs(t, err, sql.ErrNoRows, "order should not exist")

	var dbTxId string
	err = dbConn.QueryRowxContext(ctx, `SELECT gid FROM pg_prepared_xacts WHERE gid = $1`, txID).Scan(&dbTxId)
	assert.ErrorIs(t, err, sql.ErrNoRows, "order should not exist")

}
