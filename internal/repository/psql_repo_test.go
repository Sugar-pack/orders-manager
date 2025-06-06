package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func newMock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return sqlx.NewDb(db, "sqlmock"), mock
}

func TestPrepareInsertOrder_OK(t *testing.T) {
	db, mock := newMock(t)
	repo := NewPsqlRepository(db)
	ctx := context.Background()
	order := &Order{ID: uuid.New(), UserID: uuid.New(), Label: "label", CreatedAt: time.Now().UTC()}
	txID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO orders").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("PREPARE TRANSACTION").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectRollback()

	err := repo.PrepareInsertOrder(ctx, order, txID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPrepareInsertOrder_InsertErr(t *testing.T) {
	db, mock := newMock(t)
	repo := NewPsqlRepository(db)
	ctx := context.Background()
	order := &Order{ID: uuid.New(), UserID: uuid.New(), Label: "label", CreatedAt: time.Now().UTC()}
	txID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO orders").WillReturnError(fmt.Errorf("insert err"))
	mock.ExpectRollback()

	err := repo.PrepareInsertOrder(ctx, order, txID)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPrepareInsertOrder_PrepareErr(t *testing.T) {
	db, mock := newMock(t)
	repo := NewPsqlRepository(db)
	ctx := context.Background()
	order := &Order{ID: uuid.New(), UserID: uuid.New(), Label: "label", CreatedAt: time.Now().UTC()}
	txID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO orders").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("PREPARE TRANSACTION").WillReturnError(fmt.Errorf("prep err"))
	mock.ExpectExec("ROLLBACK PREPARED").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectRollback()

	err := repo.PrepareInsertOrder(ctx, order, txID)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCommitInsertTransaction(t *testing.T) {
	db, mock := newMock(t)
	repo := NewPsqlRepository(db)
	ctx := context.Background()
	txID := uuid.New()

	mock.ExpectExec("COMMIT PREPARED").WillReturnResult(sqlmock.NewResult(1, 1))
	err := repo.CommitInsertTransaction(ctx, txID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCommitInsertTransaction_Error(t *testing.T) {
	db, mock := newMock(t)
	repo := NewPsqlRepository(db)
	ctx := context.Background()
	txID := uuid.New()

	mock.ExpectExec("COMMIT PREPARED").WillReturnError(fmt.Errorf("commit err"))
	err := repo.CommitInsertTransaction(ctx, txID)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRollbackInsertTransaction(t *testing.T) {
	db, mock := newMock(t)
	repo := NewPsqlRepository(db)
	ctx := context.Background()
	txID := uuid.New()

	mock.ExpectExec("ROLLBACK PREPARED").WillReturnResult(sqlmock.NewResult(1, 1))
	err := repo.RollbackInsertTransaction(ctx, txID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRollbackInsertTransaction_Error(t *testing.T) {
	db, mock := newMock(t)
	repo := NewPsqlRepository(db)
	ctx := context.Background()
	txID := uuid.New()

	mock.ExpectExec("ROLLBACK PREPARED").WillReturnError(fmt.Errorf("rollback err"))
	err := repo.RollbackInsertTransaction(ctx, txID)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrder_OK(t *testing.T) {
	db, mock := newMock(t)
	repo := NewPsqlRepository(db)
	ctx := context.Background()
	orderID := uuid.New()
	order := &Order{ID: orderID, UserID: uuid.New(), Label: "label", CreatedAt: time.Now().UTC()}

	rows := sqlmock.NewRows([]string{"id", "user_id", "label", "created_at"}).
		AddRow(order.ID, order.UserID, order.Label, order.CreatedAt)
	mock.ExpectQuery("SELECT").WithArgs(orderID.String()).WillReturnRows(rows)

	res, err := repo.GetOrder(ctx, orderID)
	assert.NoError(t, err)
	assert.Equal(t, order.ID, res.ID)
	assert.Equal(t, order.UserID, res.UserID)
	assert.Equal(t, order.Label, res.Label)
	assert.WithinDuration(t, order.CreatedAt, res.CreatedAt, time.Second)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrder_Error(t *testing.T) {
	db, mock := newMock(t)
	repo := NewPsqlRepository(db)
	ctx := context.Background()
	orderID := uuid.New()

	mock.ExpectQuery("SELECT").WithArgs(orderID.String()).WillReturnError(fmt.Errorf("get err"))

	res, err := repo.GetOrder(ctx, orderID)
	assert.Error(t, err)
	assert.NotNil(t, res)
	assert.NoError(t, mock.ExpectationsWereMet())
}
