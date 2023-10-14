package storage

import (
	"context"
	"highload-arch/pkg/config"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	DbUsersFirstName  = "first_name"
	DbUsersSecondName = "second_name"
)

var db *pgxpool.Pool

func CreateConnectionPool() {
	var err error
	db, err = pgxpool.Connect(context.Background(), config.GetString("db.connstr"))
	if err != nil {
		log.Fatal(err)
	}
}
