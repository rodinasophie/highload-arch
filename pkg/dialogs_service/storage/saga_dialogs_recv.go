package storage

import (
	"context"
	"encoding/json"
	"highload-arch/pkg/common"
	"highload-arch/pkg/config"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ConnectClientToRabbitMQ() (*amqp.Connection, error) {
	url := config.GetString("rabbitmq.url")
	var err error
	rbmqClient, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	return rbmqClient, nil
}

func CloseClientRabbitMQ(rbmqClient *amqp.Connection) {
	rbmqClient.Close()
}

type RBMQUpdateDBCallback func(ctx context.Context, req *common.MessageCountRequest) error

func SagaHandleMessageCountUpdated(ctx context.Context, update_dialog_db RBMQUpdateDBCallback) error {
	rbmqClient, err := ConnectClientToRabbitMQ()
	if err != nil {
		log.Println("Could not connect to rabbitmq on client side")
		return err
	}

	defer CloseClientRabbitMQ(rbmqClient)

	ch, err := rbmqClient.Channel()
	if err != nil {
		log.Println("Could not create rabbitmq channel on client side")
		return err
	}

	defer ch.Close()

	err = ch.ExchangeDeclare(
		"unreadMessagesCounted", // name
		"topic",                 // type
		false,                   // durable
		false,                   // auto-deleted
		false,                   // internal
		false,                   // no-wait
		nil,                     // arguments
	)

	if err != nil {
		log.Println("Cannot create exchange on client side")
		return err
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Println("Could not declare queue on client side")
		return err
	}

	routingKey := "*.*"
	err = ch.QueueBind(
		q.Name,                  // queue name
		routingKey,              // routing key
		"unreadMessagesCounted", // exchange
		false,
		nil)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Dialog service(msg recv): [x] %s", d.Body)
			var req common.MessageCountRequest
			err = json.Unmarshal(d.Body, &req)
			if err != nil {
				log.Println("Cannot unmarshal update message count request to bytes array")
			}
			err = update_dialog_db(ctx, &req)
			if err != nil {
				log.Printf("Error while changing dialog state: %s", err)
			} else {
				log.Printf("Successfully updated message state after changing counter's value")
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	return nil

}
