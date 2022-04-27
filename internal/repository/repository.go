package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Label     string    `db:"label"`
	CreatedAt time.Time `db:"created_at"`
}

type OrderRepoWith2PC interface {
	PrepareInsertOrder(ctx context.Context, order *Order, txId uuid.UUID) error

	CommitInsertTransaction(ctx context.Context, txID uuid.UUID) error

	RollbackInsertTransaction(ctx context.Context, txID uuid.UUID) error

	GetOrder(ctx context.Context, id string) (*Order, error)
}
