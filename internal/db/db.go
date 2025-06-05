package db

import (
	"context"
	"fmt"

	"github.com/Sugar-pack/users-manager/pkg/logging"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/Sugar-pack/orders-manager/internal/config"
)

// Connect creates new db connection.
func Connect(ctx context.Context, conf *config.DB) (*sqlx.DB, error) {
	logger := logging.FromContext(ctx)
	logger.WithField("conn_string", conf.ConnString).Trace("connecting to db")

	conn, err := otelsqlx.ConnectContext(ctx, "pgx", conf.ConnString,
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
	)
	if err != nil {
		logger.WithError(err).Error("unable to connect to database")

		return nil, fmt.Errorf("unable to connect to database %w", err)
	}
	conn.SetMaxOpenConns(conf.MaxOpenCons)
	conn.SetConnMaxLifetime(conf.ConnMaxLifetime)

	return conn, nil
}

// Disconnect drops db connection.
func Disconnect(ctx context.Context, dbConn *sqlx.DB) error {
	logger := logging.FromContext(ctx)
	logger.Trace("disconnecting db")

	return dbConn.Close()
}
