package db

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Sugar-pack/orders-manager/internal/config"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestConnect_InvalidConnString(t *testing.T) {
	ctx := context.Background()
	conf := &config.DB{ConnString: "invalid"}
	conn, err := Connect(ctx, conf)
	assert.Error(t, err)
	assert.Nil(t, conn)
}

func TestDisconnect(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	db := sqlx.NewDb(sqlDB, "sqlmock")
	ctx := context.Background()
	mock.ExpectClose()
	assert.NoError(t, Disconnect(ctx, db))
	assert.NoError(t, mock.ExpectationsWereMet())
}
