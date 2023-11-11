package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

const (
	TokenLength         = 15
	TokenValidityPeriod = 24 // In hours
)

type Login struct {
	ID       string `pg:"id"`
	Password string `pg:"password"`
}

type LoginToken struct {
	ID         string    `pg:"id"`
	Token      string    `pg:"token"`
	ValidUntil time.Time `pg:"valid_until"`
}

func LoginUser(ctx context.Context, login *Login) (*LoginToken, error) {
	if err := CheckUserPassword(ctx, login.ID, login.Password); err != nil {
		return nil, err
	}
	loginToken, err := login.getLoginToken(ctx)
	if err != nil && err != ErrTokenNotFound && err != ErrTokenExpired {
		return nil, err
	}
	if err == ErrTokenNotFound || err == ErrTokenExpired {
		// generate token
		log.Println("Generating new token for user ", login.ID)
		token := generateSecureToken(TokenLength)
		newLoginToken, err := HandleInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) (interface{}, error) {
			loginToken, err := login.dbAddLoginToken(ctx, tx, token)
			if err != nil {
				return nil, err
			}

			return loginToken, nil
		})
		if err != nil {
			return nil, err
		}
		return newLoginToken.(*LoginToken), nil
	}
	return loginToken, nil
}

func (login *Login) dbAddLoginToken(ctx context.Context, tx pgx.Tx, token string) (*LoginToken, error) {
	valid_until := time.Now().Add(time.Hour * time.Duration(TokenValidityPeriod))
	_, err := tx.Exec(ctx,
		`INSERT INTO user_tokens (id, token, valid_until) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET token = EXCLUDED.token, valid_until = EXCLUDED.valid_until`,
		login.ID, token, valid_until)

	return &LoginToken{login.ID, token, valid_until}, err
}

func (login *Login) dbReadToken(ctx context.Context, tx pgx.Tx, token string) (*LoginToken, error) {
	res := []*LoginToken{}
	err := pgxscan.Select(context.Background(), db, &res, `SELECT * FROM user_tokens WHERE token = $1`, token)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, ErrTokenNotFound
	}
	return res[0], nil
}

func (login *Login) dbReadTokenByID(ctx context.Context, tx pgx.Tx, userID string) (*LoginToken, error) {
	res := []*LoginToken{}
	err := pgxscan.Select(context.Background(), db, &res, `SELECT * FROM user_tokens WHERE id = $1`, userID)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, ErrTokenNotFound
	}
	return res[0], nil
}

func generateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func (login *Login) getLoginToken(ctx context.Context) (*LoginToken, error) {
	loginToken, err := HandleInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) (interface{}, error) {
		loginToken, err := login.dbReadTokenByID(ctx, tx, login.ID)
		if err != nil {
			return "", err
		}
		return loginToken, nil
	})
	if err != nil {
		return nil, err
	}
	loginTokenStruct := loginToken.(*LoginToken)
	if loginTokenStruct.ValidUntil.Before(time.Now()) {
		return nil, ErrTokenExpired
	}
	return loginToken.(*LoginToken), nil
}

/* Get token and make sure it's not expired */
func ValidateLoginToken(ctx context.Context, token string) (string, error) {
	loginToken, err := HandleInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) (interface{}, error) {
		login := &Login{}
		loginToken, err := login.dbReadToken(ctx, tx, token)
		if err != nil {
			return "", err
		}
		return loginToken, nil
	})
	if err != nil {
		return "", err
	}
	loginTokenStruct := loginToken.(*LoginToken)
	if loginTokenStruct.ValidUntil.Before(time.Now()) {
		return "", ErrTokenExpired
	}
	return loginTokenStruct.ID, nil
}
