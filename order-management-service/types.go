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

func NewOrder(customerId, productId string, quantity int64, productPrice float64) *Order {
	return &Order{
		ID:         generateNumber(),
		CustomerId: customerId,
		ProductId:  productId,
		Quantity:   quantity,
		TotalPrice: calculateTotalPrice(quantity, productPrice),
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		Status:     OrderPending,
	}
}

func NewProduct(productName string, price float64) *Product {
	return &Product{
		ProductId: "1", // Can generate random ID for more entries in future
		Name:      productName,
		Price:     price,
	}
}

func NewCustomer(customerName, email string) *Customer {
	return &Customer{
		CustomerId: "1", // Can generate random ID for more entries in future
		Name:       customerName,
		Email:      email,
	}
}
func generateNumber() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func calculateTotalPrice(quantity int64, productPrice float64) float64 {
	return float64(quantity) * productPrice
}
