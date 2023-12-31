package main

import (
	"highload-arch/pkg/backend"
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
	f := setLoggerFile("./logs/service.log")
	defer f.Close()

	config.Load("local-config.yaml")
	log.Println("Connecting to Postgres")
	storage.CreateConnectionPool()
	storage.CreateReplicaConnectionPool()
	log.Println("Connecting to Cache")
	storage.ConnectToCache()
	log.Println("Server started")
	router := backend.NewRouter()

	log.Fatal(http.ListenAndServe(config.GetString("server.port"), router))

}
