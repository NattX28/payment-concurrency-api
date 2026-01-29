package model

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	StatusPending    PaymentStatus = "pending"
	StatusProcessing PaymentStatus = "processing"
	StatusCompleted  PaymentStatus = "complete"
	StatusFailed     PaymentStatus = "failed"
)

// Payment transaction
type Payment struct {
	ID          string        `json:"id"`
	UserID      string        `json:"user_id"`
	Amount      float64       `json:"amount"`
	Currency    string        `json:"currency"`
	Status      PaymentStatus `json:"status"`
	Description string        `json:"description"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	ProcessedAt *time.Time    `json:"processed_at,omitempty"`
}

type CreatePaymentRequest struct {
	UserID      string  `json:"user_id" validate:"required"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Currency    string  `json:"currency" validate:"required,len=3"`
	Description string  `json:"description"`
}

// detailed information
type CreatePaymentResponse struct {
	PaymentID string        `json:"payment_id"`
	Status    PaymentStatus `json:"payment_status"`
	Message   string        `json:"message"`
}

type PaymentResponse struct {
	Payment *Payment `json:"payment"`
}

type PaymentStats struct {
	TotalPayment       int64   `json:"total_payments"`
	PendingPayments    int64   `json:"pending_payments"`
	ProcessingPayments int64   `json:"processing_payments"`
	CompletedPayments  int64   `json:"completed_payments"`
	FailedPayments     int64   `json:"failed_payments"`
	TotalAmount        float64 `json:"total_amount"`
	ActiveGoroutines   int     `json:"active_goroutines"`
}

// Is sent when payment status change
type WebhookPayload struct {
	PaymentID string        `json:"payment_id"`
	Status    PaymentStatus `json:"status"`
	Amount    float64       `json:"amount"`
	TimeStamp time.Time     `json:"timestamp"`
}

func NewPayment(req CreatePaymentRequest) *Payment {
	now := time.Now()
	return &Payment{
		ID:          uuid.New().String(),
		UserID:      req.UserID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: req.Description,
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
