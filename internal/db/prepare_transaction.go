package db

import (
	"context"
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

func InsertUser(ctx context.Context, dbConn sqlx.ExecerContext, order *Order) error {
	query := "INSERT INTO orders ( id,  user_id, label, created_at ) VALUES ($1, $2, $3, $4)"
	_, err := dbConn.ExecContext(ctx, query, order.ID, order.UserID, order.Label, order.CreatedAt)

	return err //nolint:wrapcheck // should be wrapped in service layer
}

func PrepareTransaction(ctx context.Context, dbConn sqlx.ExecerContext, txID uuid.UUID) error {
	_, err := dbConn.ExecContext(ctx, fmt.Sprintf("PREPARE TRANSACTION '%s'", txID.String()))

	return err //nolint:wrapcheck // should be wrapped in service layer
}
