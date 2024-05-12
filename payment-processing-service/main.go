package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/aayush993/go-order-management/common"
	"github.com/streadway/amqp"
)

// All constants
const (
	amqpUrlStr = "AMQP_SERVER_URL"

	exchangeNameStr      = "EXCHANGE_NAME"
	sendRoutingKeyStr    = "SEND_ROUTING_KEY"
	receiveRoutingKeyStr = "RECEIVE_ROUTING_KEY"
)

func main() {

	// Get config from environment
	amqpServerURL := os.Getenv(amqpUrlStr)
	exchangeName := os.Getenv(exchangeNameStr)
	sendRoutingKey := os.Getenv(sendRoutingKeyStr)
	receiveRoutingKey := os.Getenv(receiveRoutingKeyStr)

	// Initialize rabbitMQ Client Service
	rabbitmqService, err := common.NewRabbitMQService(amqpServerURL)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ service: %v", err)
	}
	defer rabbitmqService.Close()

	log.Printf("Checking orders to process payments. To exit, press CTRL+C")
	err = rabbitmqService.Consume(exchangeName, receiveRoutingKey, func(msgs <-chan amqp.Delivery) {
		for d := range msgs {
			var order common.Order
			err := json.Unmarshal(d.Body, &order)
			if err != nil {
				log.Printf("Failed to decode message: %v", err)
				continue
			}

			// Simulate payment processing
			var message string
			if order.Amount <= 1000 {
				message = "Payment successful"
				order.Status = "Confirmed"
			} else {
				message = "Payment failed: Insufficient funds"
			}

			log.Printf("%s for order id: %v", message, order)

			// Publish payment response
			body, err := json.Marshal(order)
			if err != nil {
				log.Printf("failed to marshal order: %v", err)
			}

			err = rabbitmqService.Publish(exchangeName, sendRoutingKey, body)
			if err != nil {
				log.Printf("Failed to publish payment response: %v", err)
			}
			log.Printf("Payment response published")
		}
	})
	if err != nil {
		log.Fatal(err)
	}

}
