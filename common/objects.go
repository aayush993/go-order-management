package common

const (
	PaymentFailed      = "failed"
	PaymentSuccessfull = "successfull"
)

type PaymentRequest struct {
	OrderID    string  `json:"orderId"`
	TotalPrice float64 `json:"totalPrice"`
}

type PaymentResponse struct {
	OrderID       string `json:"orderId"`
	PaymentStatus string `json:"paymentStatus"`
}
