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
	httpSwagger "github.com/swaggo/http-swagger"
)

var sendQueueName, receiveQueueName string

type APIServer struct {
	listenAddr  string
	rabbitmqSvc common.MqSvc
	svc         Service
}

func NewAPIServer(listenAddr string, rabbitmqSvc common.MqSvc, svc Service) *APIServer {
	return &APIServer{
		listenAddr:  listenAddr,
		rabbitmqSvc: rabbitmqSvc,
		svc:         svc,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	sendQueueName = os.Getenv(sendRoutingKeyStr)
	receiveQueueName = os.Getenv(receiveRoutingKeyStr)

	go s.ProcessPaymentsWorker()

	// Register handlers for HTTP routes
	router.HandleFunc("/orders", loggingMiddleware(makeHTTPHandleFunc(s.HandleOrderCreate))).Methods("POST")
	router.HandleFunc("/orders/{id}", loggingMiddleware(makeHTTPHandleFunc(s.HandleOrderRetrieve))).Methods("GET")

	// Serve Swagger UI
	// currently not working
	router.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/docs/swagger.json"), // URL pointing to the generated swagger.json file
	))

	log.Println("Server now listening on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

// HandleOrderRetrieve handles the retrieval of an order by ID
func (s *APIServer) HandleOrderRetrieve(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err

	}

	order, err := s.svc.GetOrder(id)
	if err != nil {
		return err
	}

	return WriteJSONResponse(w, http.StatusOK, order)
}

type CreateOrderRequest struct {
	CustomerId string `json:"customerId"`
	ProductId  string `json:"productId"`
	Quantity   int64  `json:"quantity"`
}

// HandleOrderCreate handles the creation of a new order
// @Summary Create a new order
// @Description Create a new order with customer ID, product ID, and quantity
// @Tags orders
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "Order request"
// @Success 201 {object} Order
// @Router /orders [post]
func (s *APIServer) HandleOrderCreate(w http.ResponseWriter, r *http.Request) error {
	var req CreateOrderRequest
	requestID := r.Header.Get("X-Request-ID")
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	order, err := s.svc.CreateOrder(req.CustomerId, req.ProductId, req.Quantity)
	if err != nil {
		return err
	}

	// Publish message to RabbitMQ
	body, err := json.Marshal(common.PaymentRequest{
		OrderID:    order.ID,
		TotalPrice: order.TotalPrice,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal order: %v", err)
	}

	err = s.rabbitmqSvc.Publish(sendQueueName, body, receiveQueueName, requestID)
	if err != nil {
		return err
	}

	log.Printf("[%s] order %s in queue for processing", requestID, order.ID)
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

// ProcessPaymentsWorker Handles Payment responses from payment processing microservice
func (s *APIServer) ProcessPaymentsWorker() {
	err := s.rabbitmqSvc.Consume(receiveQueueName, func(msgs <-chan amqp.Delivery) {
		for d := range msgs {
			requesId := d.CorrelationId

			var response common.PaymentResponse
			err := json.Unmarshal(d.Body, &response)
			if err != nil {
				log.Printf("[%s] Failed to decode message error: %v", requesId, err)
				d.Ack(false)
				continue
			}

			// Update order status as per business logic
			err = s.svc.UpdateOrderStatus(response.OrderID, response.PaymentStatus)
			if err != nil {
				log.Printf("[%s] Failed to update order status for order id: %s error: %v", requesId, response.OrderID, err)
				d.Ack(false)
				continue
			}

			d.Ack(false)
			log.Printf("[%s] Payment for order id %s is %s", requesId, response.OrderID, response.PaymentStatus)
		}
	})
	if err != nil {
		log.Fatal(err)
	}

}

func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}
	return id, nil
}
