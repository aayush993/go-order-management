package main

import (
	"log"

	mbClient "github.com/aayush993/go-order-management/internal/mbroker"
)

func main() {

	// Initialize rabbitMQ Client Service
	rabbitmqService, err := mbClient.NewRabbitMQService()
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ service: %v", err)
	}
	defer rabbitmqService.Close()

	// Initialize Postgres Client Service
	dbStore, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := dbStore.Init(); err != nil {
		log.Fatal(err)
	}

	//Start API Server
	server := NewAPIServer(":3000", dbStore, rabbitmqService)
	server.Run()
}
