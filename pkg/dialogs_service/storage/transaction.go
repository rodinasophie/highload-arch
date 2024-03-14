package storage

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type Callback func(context.Context, pgx.Tx) (interface{}, error)

func HandleInTransaction(ctx context.Context, callback Callback) (interface{}, error) {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()
	var val interface{}
	val, err = callback(ctx, tx)
	if err != nil {
		return nil, err
	}

	return val, nil
}
