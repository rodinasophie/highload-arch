package storage

import (
	"context"
	"highload-arch/pkg/common"
	"log"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type UnreadMessageCount struct {
	Count       int    `pg:"count"`
	AuthorID    string `pg:"author_id"`
	RecepientID string `pg:"recepient_id"`
}

func (req *UnreadMessageCount) dbIncMessageCount(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO unread_messages (author_id, recepient_id) VALUES ($1, $2) ON CONFLICT (author_id, recepient_id) DO UPDATE SET count = EXCLUDED.count + 1`,
		req.AuthorID, req.RecepientID)

	return err
}

func (req *UnreadMessageCount) dbDecMessageCount(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO unread_messages (author_id, recepient_id) VALUES ($1, $2) ON CONFLICT (author_id, recepient_id) DO UPDATE SET count = EXCLUDED.count - 1`,
		req.AuthorID, req.RecepientID)

	return err
}

func dbGetMessageCount(ctx context.Context, userID, to string) ([]UnreadMessageCount, error) {
	res := []UnreadMessageCount{}
	rows, err := db.Query(ctx,
		`SELECT count FROM unread_messages WHERE author_id = $1 and recepient_id = $2`, userID, to)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	if err := pgxscan.ScanAll(&res, rows); err != nil {
		return nil, err
	}

	return res, err
}

func GetMessageCount(ctx context.Context, userID, to string) (*UnreadMessageCount, error) {
	count, err := dbGetMessageCount(ctx, userID, to)
	if err != nil {
		return nil, err
	}
	if len(count) == 0 {
		return nil, common.ErrNoMessagesFound
	}
	return &count[0], nil
}

func UpdateMessageCount(ctx context.Context, msg *common.MessageCountRequest) error {
	req := UnreadMessageCount{AuthorID: msg.AuthorID, RecepientID: msg.RecepientID}
	var err error
	if msg.Action == common.INCREMENT_MESSAGE_COUNT_ACTION {
		_, err = HandleInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) (interface{}, error) {
			err := req.dbIncMessageCount(ctx, tx)
			if err != nil {
				return nil, err
			}
			return nil, nil
		})
	} else if msg.Action == common.DECREMENT_MESSAGE_COUNT_ACTION {
		_, err = HandleInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) (interface{}, error) {
			err := req.dbDecMessageCount(ctx, tx)
			if err != nil {
				return nil, err
			}
			return nil, nil
		})
	} else {
		log.Printf("Unknown action: %s", msg.Action)
	}
	return err
}
