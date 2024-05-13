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
	order, err := s.store.GetOrderByID(id)
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

	// Validate customer Id
	err := validateCustomerInfo(s.store, req.CustomerId)
	if err != nil {
		return err
	}
	// Get product price
	product, err := getProductInformation(s.store, req.ProductId)
	if err != nil {
		return err
	}

	order, err := NewOrder(req.CustomerId, req.ProductId, req.Quantity, product.Price)
	if err != nil {
		return err
	}

	if err := s.store.CreateOrder(order); err != nil {
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
	log.Printf("[%s] Sent order %s for payment processing", requestID, order.ID)

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

			// Get order Status
			var orderStatus string
			switch response.PaymentStatus {
			case common.PaymentSuccessfull:
				orderStatus = OrderConfirmed
			case common.PaymentFailed:
				orderStatus = OrderCanceled
			default:
				log.Printf("[%s] Invalid payment response: %v", requesId, response)
				d.Ack(false)
				continue
			}

			err = s.store.UpdateOrderStatus(response.OrderID, orderStatus)
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

func validateCustomerInfo(store Storage, custId string) error {
	customerId, err := strconv.Atoi(custId)
	if err != nil {
		return fmt.Errorf("invalid customer id %s", custId)
	}

	customer, err := store.GetCustomerByID(customerId)
	if err != nil || customer == nil {
		return fmt.Errorf("invalid customer id %s", custId)
	}

	return nil
}

func getProductInformation(store Storage, prodId string) (*Product, error) {
	productId, err := strconv.Atoi(prodId)
	if err != nil {
		return nil, fmt.Errorf("invalid product id %s", prodId)
	}

	return store.GetProductByID(productId)

}
