package storage

import (
	"context"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type FriendRequest struct {
	UserID   string `pg:"user_id"`
	FriendID string `pg:"friend_id"`
}

func (req *FriendRequest) dbAddFriend(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO friends (user_id, friend_id) VALUES ($1, $2) ON CONFLICT (user_id, friend_id) DO NOTHING`,
		req.UserID, req.FriendID)

	return err
}

func (req *FriendRequest) dbDeleteFriend(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx,
		`DELETE from friends WHERE user_id = $1 AND friend_id = $2`,
		req.UserID, req.FriendID)

	return err
}

func dbLoadFriends(ctx context.Context) ([]FriendRequest, error) {
	res := []FriendRequest{}

	rows, err := db.Query(ctx, `SELECT user_id, friend_id FROM friends;`)

	defer rows.Close()
	if err != nil {
		return nil, err
	}

	if err := pgxscan.ScanAll(&res, rows); err != nil {
		return nil, err
	}

	return res, err
}

func AddFriend(ctx context.Context, userID string, friendID string) error {
	req := &FriendRequest{UserID: userID, FriendID: friendID}
	_, err := HandleInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) (interface{}, error) {
		err := req.dbAddFriend(ctx, tx)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func DeleteFriend(ctx context.Context, userID string, friendID string) error {
	req := &FriendRequest{UserID: userID, FriendID: friendID}
	_, err := HandleInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) (interface{}, error) {
		err := req.dbDeleteFriend(ctx, tx)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func GetFriends(ctx context.Context) ([]FriendRequest, error) {
	friends, err := dbLoadFriends(ctx)
	if err != nil {
		return nil, nil
	}
	return friends, err
}
