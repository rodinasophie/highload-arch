package endpoints

import (
	"context"
	"highload-arch/pkg/config"
	"highload-arch/pkg/storage"
	"log"
	"strings"

	"github.com/gorilla/websocket"
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

type RBMQCallback func(conn *websocket.Conn, post []byte)
type RBMQCallbackUpdateCache func(ctx context.Context, userID string, postID string, postText string)

func ReadPostCreatedMessageFromQueue(ctx context.Context, userID string, callback RBMQCallback, ws *websocket.Conn) error {
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
		"createdPosts", // name
		"topic",        // type
		false,          // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
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

	friends, err := storage.GetFriendsByUser(ctx, userID)
	if err != nil {
		log.Println("Cannot get the list of user friends on client side")
		return err
	}
	routingKey := ""
	for _, friend := range friends {
		routingKey = friend.FriendID + ".*"
		err = ch.QueueBind(
			q.Name,         // queue name
			routingKey,     // routing key
			"createdPosts", // exchange
			false,
			nil)
	}

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
			log.Printf(" [x] %s", d.Body)
			callback(ws, d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	return nil

}

func ReadPostCreatedMessageFromQueueUpdateCache(ctx context.Context, callback RBMQCallbackUpdateCache) error {
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
		"createdPosts", // name
		"topic",        // type
		false,          // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
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
		q.Name,         // queue name
		routingKey,     // routing key
		"createdPosts", // exchange
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
			log.Printf(" [x] %s", d.Body)
			res := strings.Split(d.RoutingKey, ".")
			userID := res[0]
			postID := res[1]
			callback(ctx, userID, postID, string(d.Body[:])) // update cache for all users of userID
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	return nil

}
