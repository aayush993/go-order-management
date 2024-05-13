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
	receiveRoutingKey := os.Getenv(receiveRoutingKeyStr)

	// Initialize rabbitMQ Client Service
	rabbitmqService, err := common.NewRabbitMQService(amqpServerURL)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ service: %v", err)
	}
	defer rabbitmqService.Close()

	log.Printf("Checking orders in queue to process payments. To exit, press CTRL+C")
	err = rabbitmqService.Consume(exchangeName, receiveRoutingKey, func(msgs <-chan amqp.Delivery) {
		for d := range msgs {
			requesId := d.CorrelationId

			var req common.PaymentRequest

			err := json.Unmarshal(d.Body, &req)
			if err != nil {
				log.Printf("Failed to decode message.  error: %v", err)
				continue
			}

			// Simulate payment processing
			var res common.PaymentResponse
			res.OrderID = req.OrderID

			var message string
			if req.TotalPrice <= 1000 {
				message = "Payment successful"
				res.PaymentStatus = common.PaymentSuccessfull
			} else {
				res.PaymentStatus = common.PaymentFailed
				message = "Payment failed: Insufficient funds"
			}

			log.Printf("[%s] %s for order id: %v", requesId, message, req.OrderID)

			// Publish payment response
			body, err := json.Marshal(res)
			if err != nil {
				log.Printf("[%s] failed to marshal json: %v", requesId, err)
			}

			err = rabbitmqService.Publish(exchangeName, d.ReplyTo, body, "", requesId)
			if err != nil {
				log.Printf("[%s] Failed to publish payment response: %v", requesId, err)
			}

			d.Ack(false)
			log.Printf("[%s] Payment response published", requesId)
		}
	})
	if err != nil {
		log.Fatal(err)
	}

}
