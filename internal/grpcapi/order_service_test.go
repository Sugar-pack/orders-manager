package grpcapi

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/Sugar-pack/orders-manager/internal/repository"

	"github.com/Sugar-pack/users-manager/pkg/logging"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Sugar-pack/orders-manager/internal/db"
	"github.com/Sugar-pack/orders-manager/pkg/pb"
)

//e2e test
func TestOrderService_InsertOrder_e2e(t *testing.T) {
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
	address := "localhost:8080"
	grpcConn := GRPCConnection(ctx, t, dbConn, address)

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

//e2e test
func TestOrderService_GetOrder(t *testing.T) {
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
	address := "localhost:8083"
	grpcConn := GRPCConnection(ctx, t, dbConn, address)

	orderClient := pb.NewOrdersManagerServiceClient(grpcConn)

	order := repository.Order{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Label:     "test",
		CreatedAt: time.Now().UTC(),
	}

	_, err := dbConn.NamedExecContext(ctx, `INSERT INTO orders (id, user_id, label, created_at) VALUES (:id, :user_id, :label, :created_at)`, order)
	if err != nil {
		t.Fatalf("could not insert order: %s", err)
	}

	orderResponse, err := orderClient.GetOrder(ctx, &pb.GetOrderRequest{
		Id: order.ID.String(),
	})

	assert.NoError(t, err)
	assert.Equal(t, order.ID.String(), orderResponse.Id)
	assert.Equal(t, order.UserID.String(), orderResponse.UserId)
	assert.Equal(t, order.Label, orderResponse.Label)
}
