package main

import (
	"context"
	"highload-arch/pkg/config"
	"highload-arch/pkg/dialogs_service/routes"
	"highload-arch/pkg/dialogs_service/storage"
	"log"
	"os"

	"net/http"
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
	/*f := setLoggerFile("./logs/dialog-service.log")
	defer f.Close()*/
	config.Load("local-config.yaml")
	log.Printf("Connecting to Postgres")
	storage.CreateConnectionPool()

	log.Printf("Connecting to TT")
	storage.ConnectToTarantool()
	defer storage.CloseTarantoolConnection()

	log.Printf("Connecting to RabbitMQ")
	storage.ConnectToRabbitMQ()
	defer storage.CloseRabbitMQ()

	log.Printf("Running Saga Handler")
	go storage.SagaHandleMessageCountUpdated(context.Background(), storage.MessagedUpdated)

	log.Printf("Server started")
	router := routes.NewRouter()

	log.Fatal(http.ListenAndServe(config.GetString("dialogs.port"), router))
}
