package storage

import (
	"context"
	"log"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type FriendRequest struct {
	ID       string `pg:"id"`
	FriendID string `pg:"friend_id"`
}

func (req *FriendRequest) dbAddFriend(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO friends (id, friend_id) VALUES ($1, $2) ON CONFLICT (id, friend_id) DO NOTHING`,
		req.ID, req.FriendID)

	return err
}

func (req *FriendRequest) dbDeleteFriend(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx,
		`DELETE from friends WHERE id = $1 AND friend_id = $2`,
		req.ID, req.FriendID)

	return err
}

func dbLoadFriends(ctx context.Context) ([]FriendRequest, error) {
	res := []FriendRequest{}

	rows, err := db.Query(ctx, `SELECT id, friend_id FROM friends;`)

	defer rows.Close()
	if err != nil {
		return nil, err
	}

	if err := pgxscan.ScanAll(&res, rows); err != nil {
		return nil, err
	}

	return res, err
}

func dbLoadFriendsByUser(ctx context.Context, userID string) ([]FriendRequest, error) {
	res := []FriendRequest{}

	rows, err := db.Query(ctx, `SELECT id, friend_id FROM friends WHERE id = $1;`, userID)

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
	req := &FriendRequest{ID: userID, FriendID: friendID}
	_, err := HandleInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) (interface{}, error) {
		err := req.dbAddFriend(ctx, tx)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		log.Println("Cannot add friend")
	}
	friends, err := GetFriendsByUser(ctx, userID)
	if err != nil {
		log.Panicln("Cannot get friends")
	}
	if len(friends) > CELEBRITY_THRESHOLD {
		SetCelebrity(ctx, userID)
	}
	return err
}

func DeleteFriend(ctx context.Context, userID string, friendID string) error {
	req := &FriendRequest{ID: userID, FriendID: friendID}
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

func GetFriendsByUser(ctx context.Context, userID string) ([]FriendRequest, error) {
	friends, err := dbLoadFriendsByUser(ctx, userID)
	if err != nil {
		return nil, nil
	}
	return friends, err
}
