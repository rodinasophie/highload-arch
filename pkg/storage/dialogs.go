package storage

import (
	"context"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type SendRequest struct {
	AuthorID    string    `pg:"author_id"`
	RecepientID string    `pg:"recepient"`
	CreatedAt   time.Time `pg:"created_at"`
	Text        string    `pg:"text"`
}

func (req *SendRequest) dbAddDialogMessage(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO dialogs (author_id, recepient_id, text, created_at) VALUES ($1, $2, $3, $4)`,
		req.AuthorID, req.RecepientID, req.Text, req.CreatedAt)

	return err
}

func dbGetDialog(ctx context.Context, userID, to string) ([]SendRequest, error) {
	res := []SendRequest{}

	rows, err := db.Query(ctx,
		`SELECT author_id, recepient_id, text FROM dialogs WHERE (author_id = $1 AND recepient_id = $2) OR (author_id = $2 AND recepient_id = $1)`,
		userID, to)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	if err := pgxscan.ScanAll(&res, rows); err != nil {
		return nil, err
	}

	return res, err
}

func SendMessage(ctx context.Context, userID, to, text string) error {
	req := &SendRequest{AuthorID: userID, Text: text, CreatedAt: time.Now(), RecepientID: to}
	_, err := HandleInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) (interface{}, error) {
		err := req.dbAddDialogMessage(ctx, tx)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func DialogList(ctx context.Context, userID, to string) ([]SendRequest, error) {
	dialog, err := dbGetDialog(ctx, userID, to)
	if err != nil {
		return nil, err
	}
	return dialog, nil
}
