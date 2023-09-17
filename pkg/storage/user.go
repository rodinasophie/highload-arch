package storage

import (
	"context"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"time"
)

type User struct {
	ID         string    `pg:"id"`
	FirstName  string    `pg:"first_name"`
	SecondName string    `pg:"second_name"`
	Birthdate  time.Time `pg:"birthdate"`
	Biography  string    `pg:"biography"`
	City       string    `pg:"city"`
}

func AddUser(ctx context.Context, user *User, password string) (string, error) {
	user.ID = uuid.New().String()
	_, err := HandleInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) (interface{}, error) {
		err := user.dbAddUserCredentials(ctx, tx, password)
		if err != nil {
			return nil, err
		}
		err = user.dbAddUser(ctx, tx)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return "", err
	}
	return user.ID, nil
}

func (u *User) dbAddUser(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO users (id, first_name, second_name, birthdate, city, biography) VALUES ($1, $2, $3, $4, $5, $6)`,
		u.ID, u.FirstName, u.SecondName, u.Birthdate, u.City, u.Biography)

	return err
}

func (u *User) dbAddUserCredentials(ctx context.Context, tx pgx.Tx, password string) error {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO user_credentials (id, password) VALUES ($1, $2)`, u.ID, hashedPassword)

	return err
}

func GetUser(ctx context.Context, id string) (*User, error) {
	user, err := HandleInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) (interface{}, error) {
		user, err := dbGetUser(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		return user, nil
	})
	if err != nil {
		return nil, err
	}
	return user.(*User), nil
}

func dbGetUser(ctx context.Context, tx pgx.Tx, userID string) (*User, error) {
	res := []*User{}
	err := pgxscan.Select(context.Background(), db, &res, `SELECT * FROM users WHERE id = $1`, userID)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, ErrUserNotFound
	}
	return res[0], nil
}
