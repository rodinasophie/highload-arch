package storage

import (
	"context"
	"github.com/jackc/pgx/v4"
	"highload-arch/pkg/config"
	"log"
)

var db *pgx.Conn

func Connect() {
	var err error

	db, err = pgx.Connect(context.Background(), config.GetString("db.connstr"))
	if err != nil {
		log.Fatal(err)
	}
}
