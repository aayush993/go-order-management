package main

import (
	"log"
	"os"
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
	log.Printf("RabbitMQ connection established")

	// Initialize Postgres Client Service
	dbStore, err := NewPostgresStore(user, pass, dbName, dbHost)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Printf("DB connection established")

	if err := dbStore.CreateTables(); err != nil {
		log.Fatal(err)
	}

	// seed table with customer and product
	seedTables(dbStore)

	//Start API Server
	server := NewAPIServer(":"+port, dbStore, rabbitmqService)
	server.Run()
}

func seedTables(dbStore *PostgresStore) {
	product, err := NewProduct("Iphone", 199)
	if err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	if err := dbStore.CreateProduct(product); err != nil && !strings.Contains(err.Error(), "duplicate key value") {
		log.Fatalf("Failed to seed database: %v", err)
	}

	customer, err := NewCustomer("Luke Skywalker", "mail@naboo.com")
	if err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	if err := dbStore.CreateCustomer(customer); err != nil && !strings.Contains(err.Error(), "duplicate key value") {
		log.Fatalf("Failed to seed database: %v", err)
	}
}
