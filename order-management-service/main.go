package main

import (
	"log"
	"strings"

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
	serverConfig, dbConfig := InitConfig()

	// Initialize rabbitMQ Client Service
	rabbitmqService, err := common.NewRabbitMQService(serverConfig.AmqpUrl)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ service: %v", err)
	}
	defer rabbitmqService.Close()
	log.Printf("[x] Message broker connected")

	// Initialize Postgres Client Service
	dbStore, err := NewPostgresStore(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Printf("[x] Database connected")

	if err := dbStore.CreateTables(); err != nil {
		log.Fatal(err)
	}

	// seed table with customer and product
	seedTables(dbStore)

	svc := NewOrderManagementService(dbStore)

	//Start API Server
	server := NewAPIServer(serverConfig, rabbitmqService, svc)
	server.Run()
}

func seedTables(dbStore *PostgresStore) {
	product := NewProduct("Iphone", 199)

	if err := dbStore.CreateProduct(product); err != nil && !strings.Contains(err.Error(), "duplicate key value") {
		log.Fatalf("Failed to seed database: %v", err)
	}

	customer := NewCustomer("Luke Skywalker", "mail@naboo.com")

	if err := dbStore.CreateCustomer(customer); err != nil && !strings.Contains(err.Error(), "duplicate key value") {
		log.Fatalf("Failed to seed database: %v", err)
	}
}
