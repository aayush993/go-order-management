package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	OrderPending   = "Pending"
	OrderConfirmed = "Confirmed"
	OrderCanceled  = "Canceled"
)

type Order struct {
	ID         string    `json:"id"`
	CustomerId string    `json:"customerId"`
	ProductId  string    `json:"productId"`
	Quantity   int64     `json:"quantity"`
	TotalPrice float64   `json:"totalPrice"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type Customer struct {
	CustomerId string `json:"customerId"`
	Name       string `json:"name"`
	Email      string `json:"email"`
}

type Product struct {
	ProductId string  `json:"productId"`
	Name      string  `json:"name"`
	Price     float64 `json:"email"`
}

func NewOrder(customerName, product string, quantity int64, productPrice float64) (*Order, error) {
	return &Order{
		ID:         generateNumber(),
		CustomerId: customerName,
		ProductId:  product,
		Quantity:   quantity,
		TotalPrice: generateOrderAmount(quantity, productPrice),
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		Status:     OrderPending,
	}, nil
}

func NewProduct(productName string, price float64) (*Product, error) {
	return &Product{
		ProductId: "1", // Can generate random ID for more entries in future
		Name:      productName,
		Price:     price,
	}, nil
}

func NewCustomer(customerName, email string) (*Customer, error) {
	return &Customer{
		CustomerId: "1", // Can generate random ID for more entries in future
		Name:       customerName,
		Email:      email,
	}, nil
}
func generateNumber() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func generateOrderAmount(quantity int64, productPrice float64) float64 {
	return float64(quantity) * productPrice
}
