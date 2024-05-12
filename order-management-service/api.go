package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/aayush993/go-order-management/common"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

var exchangeName, sendRoutingKey, receiveRoutingKey string

type APIServer struct {
	listenAddr  string
	store       Storage
	rabbitmqSvc common.MqSvc
}

func NewAPIServer(listenAddr string, store Storage, rabbitmqSvc common.MqSvc) *APIServer {
	return &APIServer{
		listenAddr:  listenAddr,
		store:       store,
		rabbitmqSvc: rabbitmqSvc,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	exchangeName = os.Getenv(exchangeNameStr)
	sendRoutingKey = os.Getenv(sendRoutingKeyStr)
	receiveRoutingKey = os.Getenv(receiveRoutingKeyStr)

	go s.ProcessPaymentsWorker()

	// Register handlers for HTTP routes
	router.HandleFunc("/orders", makeHTTPHandleFunc(s.HandleOrderCreate)).Methods("POST")
	router.HandleFunc("/orders/{id}", makeHTTPHandleFunc(s.HandleOrderRetrieve)).Methods("GET")

	log.Println("API server now listening on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

// ProcessPaymentsWorker Handles Payment responses from payment processing microservice
func (s *APIServer) ProcessPaymentsWorker() {
	log.Printf("Checking order payment status. To exit, press CTRL+C")
	err := s.rabbitmqSvc.Consume(exchangeName, receiveRoutingKey, func(msgs <-chan amqp.Delivery) {
		for d := range msgs {
			var order common.Order
			err := json.Unmarshal(d.Body, &order)
			if err != nil {
				log.Printf("Failed to decode message: %v", err)
				continue
			}

			if order.Status != "pending" {
				err = s.store.UpdateOrder(&order)
				if err != nil {
					log.Printf("Failed to update orderid %v in the table error: %v", order.ID, err)
					continue
				}
				log.Printf("Order %v is %v", order.ID, order.Status)
			} else {
				log.Printf("Payment for order %v failed", order.ID)
			}

		}
	})
	if err != nil {
		log.Fatal(err)
	}

}

// HandleOrderRetrieve handles the retrieval of an order by ID
func (s *APIServer) HandleOrderRetrieve(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err

	}
	order, err := s.store.GetOrderByID(id)
	if err != nil {
		return err
	}

	return WriteJSONResponse(w, http.StatusOK, order)
}

type CreateOrderRequest struct {
	CustomerName string `json:"customer"`
	Product      string `json:"product"`
	Quantity     int64  `json:"quantity"`
}

// HandleOrderCreate handles the creation of a new order
func (s *APIServer) HandleOrderCreate(w http.ResponseWriter, r *http.Request) error {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	order, err := common.NewOrder(req.CustomerName, req.Product, req.Quantity)
	if err != nil {
		return err
	}

	if err := s.store.CreateOrder(order); err != nil {
		return err
	}

	// Publish message to RabbitMQ
	body, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %v", err)
	}

	err = s.rabbitmqSvc.Publish(exchangeName, sendRoutingKey, body)
	if err != nil {
		return err
	}

	return WriteJSONResponse(w, http.StatusCreated, order)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSONResponse(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func WriteJSONResponse(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}
	return id, nil
}
