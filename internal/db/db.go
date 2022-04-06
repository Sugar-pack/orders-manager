package db

import (
	"context"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/Sugar-pack/users-manager/pkg/logging"

	"github.com/Sugar-pack/orders-manager/internal/config"
)

// Connect creates new db connection.
func Connect(ctx context.Context, conf *config.DB) (*sqlx.DB, error) {
	logger := logging.FromContext(ctx)
	logger.WithField("conn_string", conf.ConnString).Trace("connecting to db")
	conn, err := sqlx.ConnectContext(ctx, "pgx", conf.ConnString)
	if err != nil {
		logger.WithError(err).Error("unable to connect to database")

		return nil, err
	}
	conn.DB.SetMaxOpenConns(conf.MaxOpenCons)
	conn.DB.SetConnMaxLifetime(conf.ConnMaxLifetime)

	return conn, err
}

// Disconnect drops db connection.
func Disconnect(ctx context.Context, dbConn *sqlx.DB) error {
	logger := logging.FromContext(ctx)
	logger.Trace("disconnecting db")

	return dbConn.Close()
}
