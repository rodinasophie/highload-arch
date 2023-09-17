package main

import (
	"highload-arch/pkg/backend"
	"highload-arch/pkg/config"
	"highload-arch/pkg/storage"
	"log"
	"net/http"
)

func main() {
	config.Load("config.yaml")
	storage.Connect()

	log.Printf("Server started")
	router := backend.NewRouter()
	log.Fatal(http.ListenAndServe(config.GetString("server.port"), router))
}
