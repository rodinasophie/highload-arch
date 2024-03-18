package storage

import (
	"context"
	"highload-arch/pkg/common"
	"highload-arch/pkg/config"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/jackc/pgx/v4/pgxpool"
	tarantool "github.com/tarantool/go-tarantool/v2"
)

var tt *tarantool.Connection

const MARK_AS_READ_ON_LISTING = true
const DB_USE_TARANTOOL = false

var db *pgxpool.Pool
var rbmq *amqp.Connection

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
	var err error
	var msg_id string
	if DB_USE_TARANTOOL {
		err = SendMessageTT(ctx, userID, to, text)
	} else {
		msg_id, err = SendMessageDB(ctx, userID, to, text)
	}
	if err != nil {
		return err
	}
	// We surely added message  to the database
	msgReq := &common.MessageCountRequest{AuthorID: userID, RecepientID: to, MessageID: msg_id, Action: common.INCREMENT_MESSAGE_COUNT_ACTION}
	err = SagaUpdateMessageCount(ctx, msgReq)
	if err != nil {
		log.Printf("Unable to inc unread messages")
	}
	// If we were unable to increment count, we just miss it
	return nil
}

func DialogList(ctx context.Context, userID, to string) ([]SendRequest, error) {
	var err error
	var dialogs []SendRequest
	if DB_USE_TARANTOOL {
		dialogs, err = DialogListTT(ctx, userID, to)
	}
	if MARK_AS_READ_ON_LISTING {
		unread_dialogs, err := DialogListUnreadDB(ctx, userID, to)
		if err != nil {
			log.Printf("Cannot list unread dialogs: %s", err)
			return nil, err
		}
		err = DialogMarkAsRead(ctx, unread_dialogs)
		if err != nil {
			log.Printf("Cannot mark as read: %s", err)
			return nil, err
		}
	}
	dialogs, err = DialogListDB(ctx, userID, to)
	if err != nil {
		log.Printf("Cannot list dialogs: %s", err)
		return nil, err
	}
	return dialogs, nil
}

func DialogMarkAsRead(ctx context.Context, dialogs []SendRequest) error {
	var err error
	for _, d := range dialogs {
		msgReq := &common.MessageCountRequest{AuthorID: d.AuthorID, RecepientID: d.RecepientID, MessageID: d.ID, Action: common.DECREMENT_MESSAGE_COUNT_ACTION}
		err = SagaUpdateMessageCount(ctx, msgReq) // decrement mesg count
		if err != nil {
			return err
		}
	}
	return nil
}
