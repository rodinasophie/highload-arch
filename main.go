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
	storage.CreateConnectionPool()

	log.Printf("Server started")
	router := backend.NewRouter()
	log.Fatal(http.ListenAndServe(config.GetString("server.port"), router))

	//server := echo.New()
	//server.HideBanner = true

	// Routes
	//backend.AddRoutes(server)

	// Start REST server
	//err := server.Start(config.GetString("server.port"))
	//server.Logger.Fatal(err)

}
