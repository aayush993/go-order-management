package main

import (
	"log"
	"os"

	"github.com/aayush993/go-order-management/common"
)

// All constants
const (
	amqpUrlStr           = "AMQP_SERVER_URL"
	pgUserStr            = "POSTGRES_USER"
	pgPassStr            = "POSTGRES_PASSWORD"
	pgDbStr              = "POSTGRES_DB"
	dbHostStr            = "POSTGRES_HOST"
	portStr              = "PORT"
	exchangeNameStr      = "EXCHANGE_NAME"
	sendRoutingKeyStr    = "SEND_ROUTING_KEY"
	receiveRoutingKeyStr = "RECEIVE_ROUTING_KEY"
)

func main() {

	// Get config from environment
	amqpServerURL := os.Getenv(amqpUrlStr)
	user := os.Getenv(pgUserStr)
	pass := os.Getenv(pgPassStr)
	dbName := os.Getenv(pgDbStr)
	dbHost := os.Getenv(dbHostStr)
	port := os.Getenv(portStr)

	// Initialize rabbitMQ Client Service
	rabbitmqService, err := common.NewRabbitMQService(amqpServerURL)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ service: %v", err)
	}
	defer rabbitmqService.Close()

	// Initialize Postgres Client Service
	dbStore, err := NewPostgresStore(user, pass, dbName, dbHost)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := dbStore.Init(); err != nil {
		log.Fatal(err)
	}

	//Start API Server
	server := NewAPIServer(":"+port, dbStore, rabbitmqService)
	server.Run()
}
