package main

import (
	"highload-arch/pkg/backend"
	"highload-arch/pkg/config"
	"highload-arch/pkg/storage"
	"log"
	"net/http"
)

func main() {
	config.Load("local-config.yaml")
	log.Printf("Connecting to Postgres")
	storage.CreateConnectionPool()
	storage.CreateReplicaConnectionPool()
	log.Printf("Connecting to Cache")
	storage.ConnectToCache()
	log.Printf("Connecting to TT")

	storage.ConnectToTarantool()
	defer storage.CloseTarantoolConnection()

	log.Printf("Connecting to RabbitMQ")
	storage.ConnectToRabbitMQ()
	defer storage.CloseRabbitMQ()

	log.Printf("Server started")
	router := backend.NewRouter()

	log.Fatal(http.ListenAndServe(config.GetString("server.port"), router))

}
