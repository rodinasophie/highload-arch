package storage

import (
	"context"
	"fmt"
	"highload-arch/pkg/config"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	tarantool "github.com/tarantool/go-tarantool/v2"
)

const (
	DbUsersFirstName  = "first_name"
	DbUsersSecondName = "second_name"
)

var db *pgxpool.Pool
var replicaDb *pgxpool.Pool
var cache *redis.Client
var tt *tarantool.Connection
var rbmq *amqp.Connection

const DB_USE_REPLICA = false
const DB_CITUS_ENABLED = false
const DB_USE_TARANTOOL = false

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

func ConnectToTarantool() {
	if !DB_USE_TARANTOOL {
		fmt.Println("Tarantool disabled")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(),
		500*time.Millisecond)
	defer cancel()
	dialer := tarantool.NetDialer{
		Address:  config.GetString("tarantool.url"),
		User:     config.GetString("tarantool.user"),
		Password: config.GetString("tarantool.pass"),
	}
	opts := tarantool.Opts{
		Timeout: time.Second,
	}
	var err error
	tt, err = tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		fmt.Println("Connection refused:", err)
	}
}

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

func SendMessage(ctx context.Context, userID, to, text string) error {
	if DB_USE_TARANTOOL {
		return SendMessageTT(ctx, userID, to, text)
	}
	return SendMessageDB(ctx, userID, to, text)
}

func DialogList(ctx context.Context, userID, to string) ([]SendRequest, error) {
	if DB_USE_TARANTOOL {
		return DialogListTT(ctx, userID, to)
	}
	return DialogListDB(ctx, userID, to)
}
