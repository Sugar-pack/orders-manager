package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PsqlRepository struct {
	db *sqlx.DB
}

func NewPsqlRepository(db *sqlx.DB) *PsqlRepository {
	return &PsqlRepository{db: db}
}

func (p *PsqlRepository) PrepareInsertOrder(ctx context.Context, order *Order, txID uuid.UUID) (err error) {
	transaction, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func(tx *sqlx.Tx) {
		// prepared transaction is independent of commit or rollback. here we just release tx
		errRollback := tx.Rollback()
		if errRollback != nil {
			if !errors.Is(errRollback, sql.ErrTxDone) {
				err = errRollback
			}
		}
	}(transaction)

	_, err = transaction.NamedExecContext(ctx,
		"INSERT INTO orders ( id,  user_id, label, created_at ) VALUES (:id, :user_id, :label, :created_at)", order)
	if err != nil {
		return err
	}

	_, err = transaction.ExecContext(ctx, fmt.Sprintf("PREPARE TRANSACTION '%s'", txID.String()))
	if err != nil {
		defer func(ctx context.Context, dbConn sqlx.ExecerContext, txID uuid.UUID) {
			errRollBack := p.RollbackInsertTransaction(ctx, txID)
			if errRollBack != nil {
				err = errRollBack
			}
		}(ctx, transaction, txID)
	}

	return err
}

func (p *PsqlRepository) CommitInsertTransaction(ctx context.Context, txID uuid.UUID) error {
	_, err := p.db.ExecContext(ctx, fmt.Sprintf("COMMIT PREPARED '%s'", txID))

	return err
}

func (p *PsqlRepository) RollbackInsertTransaction(ctx context.Context, txID uuid.UUID) error {
	_, err := p.db.ExecContext(ctx, fmt.Sprintf("ROLLBACK PREPARED '%s'", txID))

	return err
}

func (p *PsqlRepository) GetOrder(ctx context.Context, id uuid.UUID) (*Order, error) {
	var order Order
	err := sqlx.GetContext(ctx, p.db, &order, "SELECT * FROM orders WHERE id = $1", id.String())

	return &order, err
}
