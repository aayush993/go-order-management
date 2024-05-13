package common

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type MqSvc interface {
	Close()
	Publish(string, string, []byte, string, string) error
	Consume(string, string, func(<-chan amqp.Delivery)) error
}

// RabbitMQService represents the RabbitMQ client service
type RabbitMQService struct {
	conn *amqp.Connection
}

// NewRabbitMQService creates a new instance of RabbitMQService
func NewRabbitMQService(amqpServerURL string) (*RabbitMQService, error) {
	var conn *amqp.Connection
	var err error

	if amqpServerURL == "" {
		amqpServerURL = "amqp://guest:guest@localhost:5672/"
	}

	// Retry loop with exponential backoff
	for attempt := 1; attempt <= 10; attempt++ {
		conn, err = amqp.Dial(amqpServerURL)
		if err == nil {
			log.Printf("Connected to RabbitMQ successfully!")
			break
		}

		// Exponential backoff
		delay := time.Duration(2^attempt) * time.Second
		fmt.Printf("Failed to connect to RabbitMQ (attempt %d): %s Retrying in %v...\n", attempt, err, delay)
		time.Sleep(delay)
	}

	if err != nil {
		return nil, err
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
func (s *RabbitMQService) Publish(exchange, routingKey string, body []byte, replyRoutingKey string, requestId string) error {
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

	var message amqp.Publishing
	if replyRoutingKey != "" {
		message = amqp.Publishing{
			ContentType:   "application/json",
			ReplyTo:       replyRoutingKey,
			CorrelationId: requestId,
			Body:          body,
		}
	} else {
		message = amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: requestId,
			Body:          body,
		}
	}

	err = ch.Publish(
		exchange,   // Exchange
		routingKey, // Routing key
		false,      // Mandatory
		false,      // Immediate
		message,
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

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set qos: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
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
