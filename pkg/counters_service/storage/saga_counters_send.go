package storage

import (
	"context"
	"encoding/json"
	"highload-arch/pkg/common"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ReplyToDialogService(ctx context.Context, req *common.MessageCountRequest) error {
	log.Println("Counter service: sending message to RabbitMQ on new dialog message count updated")
	channel, err := rbmq.Channel()
	if err != nil {
		log.Println("RBMQ: Channel creation failed")
		return err
	}
	defer channel.Close()

	err = channel.ExchangeDeclare(
		"unreadMessagesCounted", // name
		"topic",                 // type
		false,                   // durable
		false,                   // auto-deleted
		false,                   // internal
		false,                   // no-wait
		nil,                     // arguments
	)

	if err != nil {
		log.Println("Cannot create exchange")
		return err
	}
	reqBytes, err := json.Marshal(*req)
	if err != nil {
		log.Println("Cannot marshal post request to bytes array")
		return err
	}

	// rounting key: userID.to, we need to increment the counter of messages
	// sent from userID to user 'to'
	// The body also contains the id of the message to return it in the response and update state in dialogs's db
	err = channel.PublishWithContext(ctx,
		"unreadMessagesCounted",          // exchange
		req.AuthorID+"."+req.RecepientID, // routing key
		false,                            // mandatory
		false,                            // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        reqBytes,
		})

	return nil
}
