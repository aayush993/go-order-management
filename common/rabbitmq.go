package common

import (
	"fmt"

	"github.com/streadway/amqp"
)

type MqSvc interface {
	Close()
	Publish(exchange, routingKey string, body []byte) error
	Consume(exchange, routingKey string, workerFunc func(<-chan amqp.Delivery)) error
}

// RabbitMQService represents the RabbitMQ client service
type RabbitMQService struct {
	conn *amqp.Connection
}

// NewRabbitMQService creates a new instance of RabbitMQService
func NewRabbitMQService(amqpServerURL string) (*RabbitMQService, error) {
	if amqpServerURL == "" {
		amqpServerURL = "amqp://guest:guest@localhost:5672/"
	}
	conn, err := amqp.Dial(amqpServerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	return &RabbitMQService{conn: conn}, nil
}

// Close closes the RabbitMQ connection
func (s *RabbitMQService) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}

// Publish publishes a message to RabbitMQ
func (s *RabbitMQService) Publish(exchange, routingKey string, body []byte) error {
	ch, err := s.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %v", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchange, // name
		"direct", // type
		false,    // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare an exchange: %v", err)
	}

	err = ch.Publish(
		exchange,   // Exchange
		routingKey, // Routing key
		false,      // Mandatory
		false,      // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %v", err)
	}
	fmt.Println("Published to RabbitMQ")
	return nil
}

// Consume consumes messages from RabbitMQ
func (s *RabbitMQService) Consume(exchange, routingKey string, workerFunc func(<-chan amqp.Delivery)) error {
	ch, err := s.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %v", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchange, // name
		"direct", // type
		false,    // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare an exchange: %v", err)
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %v", err)
	}

	err = ch.QueueBind(
		q.Name,     // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind a queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %v", err)
	}

	// Process incoming messages
	forever := make(chan bool)
	go workerFunc(msgs)

	<-forever

	return nil
}
