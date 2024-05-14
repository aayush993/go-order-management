package main

import (
	"fmt"
	"strconv"

	"github.com/aayush993/go-order-management/common"
)

type Service interface {
	CreateOrder(string, string, int64) (*Order, error)
	GetOrder(int) (*Order, error)
	UpdateOrderStatus(string, string) error
}

type OrderManagementService struct {
	repo Storage
}

func NewOrderManagementService(repo Storage) Service {
	return &OrderManagementService{
		repo: repo,
	}
}

func (s *OrderManagementService) CreateOrder(customerId, productId string, quantity int64) (*Order, error) {

	// Validate customer Id
	err := validateCustomerInfo(s.repo, customerId)
	if err != nil {
		return nil, err
	}
	// Get product price
	product, err := getProductInformation(s.repo, productId)
	if err != nil {
		return nil, err
	}

	order := NewOrder(customerId, productId, quantity, product.Price)

	if err := s.repo.CreateOrder(order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderManagementService) GetOrder(id int) (*Order, error) {

	order, err := s.repo.GetOrderByID(id)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderManagementService) UpdateOrderStatus(orderId, paymentStatus string) error {

	// Get order Status
	var orderStatus string
	switch paymentStatus {
	case common.PaymentSuccessfull:
		orderStatus = OrderConfirmed
	case common.PaymentFailed:
		orderStatus = OrderCanceled
	default:
		return fmt.Errorf("invalid payment status: %v", paymentStatus)

	}

	err := s.repo.UpdateOrderStatus(orderId, orderStatus)
	if err != nil {
		return err
	}

	return nil

}

func validateCustomerInfo(repo Storage, custId string) error {
	customerId, err := strconv.Atoi(custId)
	if err != nil {
		return fmt.Errorf("invalid customer id %s", custId)
	}

	customer, err := repo.GetCustomerByID(customerId)
	if err != nil || customer == nil {
		return fmt.Errorf("invalid customer id %s", custId)
	}

	return nil
}

func getProductInformation(repo Storage, prodId string) (*Product, error) {
	productId, err := strconv.Atoi(prodId)
	if err != nil {
		return nil, fmt.Errorf("invalid product id %s", prodId)
	}

	return repo.GetProductByID(productId)

}
