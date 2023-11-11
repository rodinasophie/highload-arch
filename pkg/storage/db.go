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
var replicaDb *pgxpool.Pool

const DB_USE_REPLICA = false

func Db() *pgxpool.Pool {
	if !DB_USE_REPLICA {
		return db
	}
	return replicaDb
}

func CreateConnectionPool() {
	var err error
	db, err = pgxpool.Connect(context.Background(), config.GetString("db.master"))
	if err != nil {
		log.Fatal(err)
	}
}

func CreateReplicaConnectionPool() {
	if !DB_USE_REPLICA {
		return
	}
	var err error
	replicaDb, err = pgxpool.Connect(context.Background(), config.GetString("db.replica"))
	if err != nil {
		log.Fatal(err)
	}
}
