package handler

import (
	"runtime"
	"time"

	"github.com/NattX28/payment-concurrency-api/internal/service"
	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	paymentService *service.PaymentService
	startTime      time.Time
}

func NewHealthHandler(paymentService *service.PaymentService) *HealthHandler {
	return &HealthHandler{
		paymentService: paymentService,
		startTime:      time.Now(),
	}
}

func (h *HealthHandler) Check(c *fiber.Ctx) error {
	uptime := time.Since(h.startTime)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	paymentStats := h.paymentService.GetStats()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":         "healthy",
		"timestamp":      time.Now().Format(time.RFC3339),
		"uptime_seconds": int(uptime.Seconds()),
		"system": fiber.Map{
			"goroutines": runtime.NumGoroutine(),
			"memory_mb":  memStats.Alloc / 1024 / 1024,
			"num_cpu":    runtime.NumCPU(),
		},
		"payments": fiber.Map{
			"total":        paymentStats.TotalPayments,
			"pending":      paymentStats.PendingPayments,
			"processing":   paymentStats.ProcessingPayments,
			"completed":    paymentStats.CompletedPayments,
			"failed":       paymentStats.FailedPayments,
			"total_amount": paymentStats.TotalAmount,
		},
	})
}
