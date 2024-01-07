package storage

import (
	"context"
	"highload-arch/pkg/config"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	tarantool "github.com/tarantool/go-tarantool/v2"
)

var tt *tarantool.Connection

const DB_USE_TARANTOOL = false

var db *pgxpool.Pool

func CreateConnectionPool() {
	var err error
	db, err = pgxpool.Connect(context.Background(), config.GetString("dialogs.db"))
	if err != nil {
		log.Fatal(err)
	}
}

func ConnectToTarantool() {
	if !DB_USE_TARANTOOL {
		log.Println("Tarantool disabled")
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
		log.Println("Connection refused:", err)
	}
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
