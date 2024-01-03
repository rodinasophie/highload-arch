package storage

import (
	"context"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type SendRequest struct {
	AuthorID    string    `pg:"author_id"`
	RecepientID string    `pg:"recepient_id"`
	DialogID    string    `pg:"dialog_id"`
	CreatedAt   time.Time `pg:"created_at"`
	Text        string    `pg:"text"`
}

func GetDialogId(author_id, recepient_id string) string {
	dialogID := ""
	if author_id > recepient_id {
		dialogID = author_id + "_" + recepient_id
	} else {
		dialogID = recepient_id + "_" + author_id
	}
	return dialogID
}

func (req *SendRequest) dbAddDialogMessage(ctx context.Context, tx pgx.Tx) error {
	dialogID := GetDialogId(req.AuthorID, req.RecepientID)
	_, err := tx.Exec(ctx,
		`INSERT INTO dialogs (author_id, recepient_id, dialog_id, text, created_at) VALUES ($1, $2, $3, $4, $5)`,
		req.AuthorID, req.RecepientID, dialogID, req.Text, req.CreatedAt)

	return err
}

func dbGetDialog(ctx context.Context, userID, to string) ([]SendRequest, error) {
	res := []SendRequest{}
	dialogID := GetDialogId(userID, to)

	rows, err := db.Query(ctx,
		`SELECT author_id, recepient_id, created_at, text FROM dialogs WHERE dialog_id = $1`, dialogID)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	if err := pgxscan.ScanAll(&res, rows); err != nil {
		return nil, err
	}

	return res, err
}

func SendMessageDB(ctx context.Context, userID, to, text string) error {
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

func DialogListDB(ctx context.Context, userID, to string) ([]SendRequest, error) {
	dialog, err := dbGetDialog(ctx, userID, to)
	if err != nil {
		return nil, err
	}
	return dialog, nil
}
