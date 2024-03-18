package storage

import (
	"context"
	"highload-arch/pkg/config"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

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
	db, err = pgxpool.Connect(context.Background(), config.GetString("counters.db"))
	if err != nil {
		log.Fatal(err)
	}
}
