package db

import (
	"context"

	"github.com/jmoiron/sqlx"
)

func GetOrder(ctx context.Context, orderId string, db sqlx.QueryerContext) (Order, error) {
	var order Order
	err := sqlx.GetContext(ctx, db, &order, "SELECT * FROM orders WHERE id = $1", orderId)

	return order, err
}
