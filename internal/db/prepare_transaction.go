package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Order struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Label     string    `db:"label"`
	CreatedAt time.Time `db:"created_at"`
}

const InsertOrderQuery = "INSERT INTO orders ( id,  user_id, label, created_at ) VALUES (:id, :user_id, :label, :created_at)" // nolint: lll // sql statement

type NamedExecutorContext interface {
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

func InsertUser(ctx context.Context, dbConn NamedExecutorContext, order *Order) error {
	_, err := dbConn.NamedExecContext(ctx, InsertOrderQuery, order)

	return err //nolint:wrapcheck // should be wrapped in service layer
}

func PrepareTransaction(ctx context.Context, dbConn sqlx.ExecerContext, txID string) error {
	_, err := dbConn.ExecContext(ctx, fmt.Sprintf("PREPARE TRANSACTION '%s'", txID))

	return err //nolint:wrapcheck // should be wrapped in service layer
}

func CommitTransaction(ctx context.Context, dbConn sqlx.ExecerContext, txID string) error {
	_, err := dbConn.ExecContext(ctx, fmt.Sprintf("COMMIT PREPARED '%s'", txID))

	return err //nolint:wrapcheck // should be wrapped in service layer
}

func RollBackTransaction(ctx context.Context, dbConn sqlx.ExecerContext, txID string) error {
	_, err := dbConn.ExecContext(ctx, fmt.Sprintf("ROLLBACK PREPARED '%s'", txID))

	return err //nolint:wrapcheck // should be wrapped in service layer
}
