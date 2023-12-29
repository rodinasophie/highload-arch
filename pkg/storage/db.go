package storage

import (
	"context"
	"highload-arch/pkg/config"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	DbUsersFirstName  = "first_name"
	DbUsersSecondName = "second_name"
)

var db *pgxpool.Pool
var replicaDb *pgxpool.Pool
var cache *redis.Client

const DB_USE_REPLICA = false
const DB_CITUS_ENABLED = true

func Db() *pgxpool.Pool {
	if !DB_USE_REPLICA {
		return db
	}
	return replicaDb
}

func CreateConnectionPool() {
	var err error
	default_db := "db.master"
	if DB_CITUS_ENABLED {
		default_db = "citus.master"
	}
	db, err = pgxpool.Connect(context.Background(), config.GetString(default_db))
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

func ConnectToCache() {
	opt, err := redis.ParseURL(config.GetString("cache.url"))
	if err != nil {
		log.Fatal(err)
	}
	cache = redis.NewClient(opt)
	CacheUpdatePosts(context.Background())
}
