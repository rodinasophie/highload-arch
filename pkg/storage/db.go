package storage

import (
	"context"
	"highload-arch/pkg/config"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/tarantool/go-tarantool/v2"
)

const (
	DbUsersFirstName    = "first_name"
	DbUsersSecondName   = "second_name"
	CELEBRITY_THRESHOLD = 1000000
)

var db *pgxpool.Pool
var replicaDb *pgxpool.Pool
var cache *redis.Client

var tt *tarantool.Connection
var rbmq *amqp.Connection

const DB_USE_REPLICA = false
const DB_CITUS_ENABLED = false

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

type Callback func(context.Context, pgx.Tx) (interface{}, error)

func ConnectToRabbitMQ() {
	url := config.GetString("rabbitmq.url")
	var err error
	rbmq, err = amqp.Dial(url)
	if err != nil {
		panic(err)
	}
}

func CloseRabbitMQ() {
	rbmq.Close()
}

func CloseTarantoolConnection() {
	if tt != nil {
		tt.Close()
	}
}

func HandleInTransaction(ctx context.Context, callback Callback) (interface{}, error) {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()
	var val interface{}
	val, err = callback(ctx, tx)
	if err != nil {
		return nil, err
	}

	return val, nil
}
