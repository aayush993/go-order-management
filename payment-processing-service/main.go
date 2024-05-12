package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// PaymentRequest represents the structure of a payment request
type PaymentRequest struct {
	OrderID string `json:"orderId"`
	Amount  int    `json:"amount"`
}

// PaymentResponse represents the structure of a payment response
type PaymentResponse struct {
	OrderID string `json:"orderId"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type Order struct {
	ID           string    `json:"id"`
	CustomerName string    `json:"customerName"`
	Product      string    `json:"product"`
	Quantity     int64     `json:"quantity"`
	Amount       int64     `json:"amount"`
	CreatedAt    time.Time `json:"createdAt"`
	Status       string    `json:"status"`
}

func main() {

	// Initialize rabbitMQ Client Service
	rabbitmqService, err := NewRabbitMQService()
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ service: %v", err)
	}
	defer rabbitmqService.Close()

	log.Printf("Checking orders to process payments. To exit, press CTRL+C")
	err = rabbitmqService.Consume("orders_exchange", "ProcessingOrders", func(msgs <-chan amqp.Delivery) {
		for d := range msgs {
			var order Order
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

			err = rabbitmqService.Publish("orders_exchange", "ProcessedOrders", body)
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
