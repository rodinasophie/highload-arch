package storage

import (
	"context"
	"fmt"

	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
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
		user, err := dbGetUserById(ctx, tx, id)
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

func SearchUsers(ctx context.Context, firstName string, secondName string) ([]User, error) {
	regex := make(map[string]string)
	regex[DbUsersFirstName] = firstName
	regex[DbUsersSecondName] = secondName
	users, err := dbGetUsersByRegex(ctx, regex)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func dbGetUserById(ctx context.Context, tx pgx.Tx, userID string) (*User, error) {
	res := []*User{}
	err := pgxscan.Select(context.Background(), Db(), &res, `SELECT * FROM users WHERE id = $1`, userID)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, ErrUserNotFound
	}
	return res[0], nil
}

func dbGetUsersByRegex(ctx context.Context, regexMap map[string]string) ([]User, error) {
	res := []User{}
	regexFilter := ``
	for key, value := range regexMap {
		val := string('\'') + string(value) + string('%') + string('\'')
		newFilter := fmt.Sprintf(`%s LIKE %s`, key, val)
		if regexFilter != `` {
			regexFilter += ` and ` + newFilter
		} else {
			regexFilter = newFilter
		}
	}
	regexFilter = `SELECT * FROM users WHERE ` + regexFilter + ` ORDER BY id`
	rows, err := Db().Query(context.Background(), regexFilter)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	if err := pgxscan.ScanAll(&res, rows); err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, ErrUserNotFound
	}
	return res, nil
}
