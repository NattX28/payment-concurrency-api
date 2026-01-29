package service

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/NattX28/payment-concurrency-api/internal/model"
	"github.com/NattX28/payment-concurrency-api/pkg/pool"
)

type PaymentService struct {
	payments    sync.Map
	workerPool  *pool.WorkerPool
	webhookChan chan model.WebhookPayload
	stats       sync.Map
}

func NewPaymentService() *PaymentService {
	workerPool := pool.NewWorkerPool(10, 100)

	service := &PaymentService{
		workerPool:  workerPool,
		webhookChan: make(chan model.WebhookPayload, 50),
	}

	service.stats.Store("total", int64(0))
	service.stats.Store("pending", int64(0))
	service.stats.Store("processing", int64(0))
	service.stats.Store("completed", int64(0))
	service.stats.Store("failed", int64(0))
	service.stats.Store("total_amount", float64(0))

	workerPool.Start()

	go service.webhookDispatcher()

	return service
}

func (s *PaymentService) CreatePayment(req model.CreatePaymentRequest) (*model.CreatePaymentResponse, error) {
	payment := model.NewPayment(req)
	s.payments.Store(payment.ID, payment)

	// Update stats
	s.incrementStat("total")
	s.incrementStat("pending")
	s.addAmount(payment.Amount)

	log.Printf("Payment created: ID=%s, Amount=%.2f %s", payment.ID, payment.Amount, payment.Currency)

	// Submit to worker pool
	task := pool.Task{
		ID: payment.ID,
		Process: func() error {
			return s.processPayment(payment.ID)
		},
	}

	if err := s.workerPool.SubmitAsync(task); err != nil {
		log.Printf("Failed to submit payment %s to worker pool: %v", payment.ID, err)
		return nil, fmt.Errorf("failed to queue payment for processing: %w", err)
	}

	return &model.CreatePaymentResponse{
		PaymentID: payment.ID,
		Status:    model.StatusPending,
		Message:   "Payment queue for processing",
	}, nil
}

func (s *PaymentService) processPayment(paymentID string) error {
	value, ok := s.payments.Load(paymentID)
	if !ok {
		return fmt.Errorf("payment not found: %s", paymentID)
	}

	payment := value.(*model.Payment)

	// Update status to processing
	payment.Status = model.StatusProcessing
	payment.UpdatedAt = time.Now()
	s.payments.Store(paymentID, payment)

	s.decrementStat("pending")
	s.incrementStat("processing")

	log.Printf("Processing payment: %s", paymentID)

	// Simulate payment gateway API call
	processingTime := time.Duration(2+rand.Intn(3)) * time.Second
	time.Sleep(processingTime)

	// Simulate success/failure (90% success rate)
	success := rand.Float32() < 0.9

	now := time.Now()
	if success {
		payment.Status = model.StatusCompleted
		payment.ProcessedAt = &now
		log.Printf("Payment completed: %s (took %v)", paymentID, processingTime)

		s.decrementStat("processing")
		s.incrementStat("completed")
	} else {
		payment.Status = model.StatusFailed
		log.Printf("Payment failed: %s", paymentID)

		s.decrementStat("processing")
		s.incrementStat("failed")
		s.subtractAmount(payment.Amount)
	}

	payment.UpdatedAt = time.Now()
	s.payments.Store(paymentID, payment)

	s.sendWebhook(payment)

	return nil
}

func (s *PaymentService) GetPayment(paymentID string) (*model.Payment, error) {
	value, ok := s.payments.Load(paymentID)
	if !ok {
		return nil, fmt.Errorf("payment not found")
	}
	payment := value.(*model.Payment)
	return payment, nil
}

func (s *PaymentService) GetStats() model.PaymentStats {
	return model.PaymentStats{
		TotalPayment:       s.getStat("total"),
		PendingPayments:    s.getStat("pending"),
		ProcessingPayments: s.getStat("processing"),
		CompletedPayments:  s.getStat("completed"),
		FailedPayments:     s.getStat("failed"),
		TotalAmount:        s.getAmount(),
		ActiveGoroutines:   runtime.NumGoroutine(),
	}
}

// Send a webhook noti asynchronously
func (s *PaymentService) sendWebhook(payment *model.Payment) {
	webhook := model.WebhookPayload{
		PaymentID: payment.ID,
		Status:    payment.Status,
		Amount:    payment.Amount,
		TimeStamp: time.Now(),
	}

	// Non-blocking send
	select {
	case s.webhookChan <- webhook:
		log.Printf("Webhook queued for payment: %s", payment.ID)
	default:
		log.Printf("Webhook queue full, dropping notification for: %s", payment.ID)
	}
}

func (s *PaymentService) webhookDispatcher() {
	log.Println("Webhook dispatcher started")

	for webhook := range s.webhookChan {
		go func(wh model.WebhookPayload) {
			log.Printf("Sending webhook for payment %s (status: %s)", wh.PaymentID, wh.Status)

			// Simulate HTTP request to webhook endpoint
			time.Sleep(500 * time.Millisecond)

			log.Printf("Webhook sent successfully for payment: %s", wh.PaymentID)
		}(webhook)
	}
}

func (s *PaymentService) Shutdown() {
	log.Println("Shutting down payment service...")

	s.workerPool.Shutdown()
	close(s.webhookChan)

	log.Println("Payment service shutdown complete")
}

// Helper methods for stats
func (s *PaymentService) incrementStat(key string) {
	value, _ := s.stats.LoadOrStore(key, int64(0))
	current := value.(int64)
	s.stats.Store(key, current+1)
}

func (s *PaymentService) decrementStat(key string) {
	value, _ := s.stats.LoadOrStore(key, int64(0))
	current := value.(int64)
	if current > 0 {
		s.stats.Store(key, current-1)
	}
}

func (s *PaymentService) getStat(key string) int64 {
	value, ok := s.stats.Load(key)
	if !ok {
		return 0
	}
	return value.(int64)
}

func (s *PaymentService) addAmount(amount float64) {
	value, _ := s.stats.LoadOrStore("total_amount", float64(0))
	current := value.(float64)
	s.stats.Store("total_amount", current)
}

func (s *PaymentService) subtractAmount(amount float64) {
	value, _ := s.stats.LoadOrStore("total_amount", float64(0))
	current := value.(float64)
	s.stats.Store("total_amount", current-amount)
}

func (s *PaymentService) getAmount() float64 {
	value, ok := s.stats.Load("total_amount")
	if !ok {
		return 0
	}
	return value.(float64)
}
