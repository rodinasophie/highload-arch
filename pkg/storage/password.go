package storage

import (
	"context"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"
)

func CheckUserPassword(ctx context.Context, userID string, password string) error {
	requestLogin := &Login{ID: userID, Password: password}
	dbLogin, err := HandleInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) (interface{}, error) {
		dbLogin, err := requestLogin.dbReadPassword(ctx, tx, userID)
		if err != nil {
			return nil, err
		}
		return dbLogin, nil
	})
	if err != nil {
		return err
	}

	if CheckPasswordHash(requestLogin.Password, dbLogin.(*Login).Password) {
		return nil
	} else {
		return ErrPasswordInvalid
	}
}

func (login *Login) dbReadPassword(ctx context.Context, tx pgx.Tx, userID string) (*Login, error) {
	res := []*Login{}
	err := pgxscan.Select(context.Background(), db, &res, `SELECT * FROM user_credentials WHERE id = $1`, userID)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, ErrUserNotFound
	}
	return res[0], nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 7)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
