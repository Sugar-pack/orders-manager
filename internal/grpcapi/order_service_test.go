package grpcapi

import (
	"context"
	"log"
	"testing"

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
