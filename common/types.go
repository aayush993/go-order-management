package common

import (
	"fmt"
	"math/rand"
	"time"
)

type Order struct {
	ID           string    `json:"id"`
	CustomerName string    `json:"customerName"`
	Product      string    `json:"product"`
	Quantity     int64     `json:"quantity"`
	Amount       int64     `json:"amount"`
	CreatedAt    time.Time `json:"createdAt"`
	Status       string    `json:"status"`
}

func NewOrder(customerName, product string, quantity int64) (*Order, error) {
	return &Order{
		ID:           generateOrderNumber(),
		CustomerName: customerName,
		Product:      product,
		Quantity:     quantity,
		Amount:       generateOrderAmount(quantity),
		CreatedAt:    time.Now().UTC(),
		Status:       "Pending",
	}, nil
}

func generateOrderNumber() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func generateOrderAmount(quantity int64) int64 {
	return quantity * 100
}
