package mock

import (
	"context"
	"testing"

	"github.com/Sugar-pack/orders-manager/internal/repository"
	"github.com/google/uuid"
)

func TestOrderRepoWith2PC_Methods(t *testing.T) {
	m := NewOrderRepoWith2PC(t)
	ctx := context.Background()
	order := &repository.Order{}
	id := uuid.New()

	m.On("PrepareInsertOrder", ctx, order, id).Return(nil)
	m.On("CommitInsertTransaction", ctx, id).Return(nil)
	m.On("RollbackInsertTransaction", ctx, id).Return(nil)
	m.On("GetOrder", ctx, id).Return(&repository.Order{}, nil)

	_ = m.PrepareInsertOrder(ctx, order, id)
	_ = m.CommitInsertTransaction(ctx, id)
	_ = m.RollbackInsertTransaction(ctx, id)
	_, _ = m.GetOrder(ctx, id)
}
