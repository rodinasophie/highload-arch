package main

import (
	"context"
	"highload-arch/pkg/backend"
	"highload-arch/pkg/backend/endpoints"
	"highload-arch/pkg/config"
	"highload-arch/pkg/storage"
	"log"
	"net/http"
	"os"
)

func setLoggerFile(filename string) *os.File {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(f)
	return f
}

func main() {

	/*f := setLoggerFile("./logs/service.log")
	defer f.Close()*/

	config.Load("local-config.yaml")
	log.Println("Connecting to Postgres")

	storage.CreateConnectionPool()
	storage.CreateReplicaConnectionPool()
	log.Println("Connecting to Cache")
	storage.ConnectToCache()

	//log.Printf("Connecting to TT")
	//storage.ConnectToTarantool()
	//defer storage.CloseTarantoolConnection()

	log.Printf("Connecting to RabbitMQ")
	storage.ConnectToRabbitMQ()
	defer storage.CloseRabbitMQ()
	go endpoints.ReadPostCreatedMessageFromQueueUpdateCache(context.Background(), storage.CacheUpdatePostsForUser)
	log.Printf("Server started")
	router := backend.NewRouter()

	log.Fatal(http.ListenAndServe(config.GetString("server.port"), router))

}
