package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"log"
	"time"
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

func (login *Login) dbReadToken(ctx context.Context, tx pgx.Tx, userID string) (*LoginToken, error) {
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
		loginToken, err := login.dbReadToken(ctx, tx, login.ID)
		if err != nil {
			return nil, err
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

func CheckLoginToken(ctx context.Context, token string, userID string) error {
	newLogin := &Login{userID, ""}
	loginToken, err := newLogin.getLoginToken(ctx)
	if err != nil {
		return err
	}
	if loginToken.Token != token {
		return ErrTokenInvalid
	}
	return nil
}
