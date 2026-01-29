package handler

import (
	"log"

	"github.com/NattX28/payment-concurrency-api/internal/model"
	"github.com/NattX28/payment-concurrency-api/internal/service"
	"github.com/gofiber/fiber/v2"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

func (h *PaymentHandler) CreatePayment(c *fiber.Ctx) error {
	var req model.CreatePaymentRequest

	if err := c.BodyParser(&req); err != nil {
		log.Printf("Invalid request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
	}

	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"message": "user_id is required",
		})
	}

	if req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"message": "amount must be greater than 0",
		})
	}

	if len(req.Currency) != 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"message": "currency must be 3-letter code (e.g., USD, THB)",
		})
	}

	response, err := h.paymentService.CreatePayment(req)
	if err != nil {
		log.Printf("Failed to create payment: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Failed to create payment",
			"message": err.Error(),
		})
	}

	log.Printf("Payment created successfully: %s", response.PaymentID)
	return c.Status(fiber.StatusCreated).JSON(response)
}

func (h *PaymentHandler) GetPayment(c *fiber.Ctx) error {
	paymentID := c.Params("id")

	if paymentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Payment ID is required",
		})
	}

	payment, err := h.paymentService.GetPayment(paymentID)
	if err != nil {
		log.Printf("Payment not found: %s", paymentID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Payment not found",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.PaymentResponse{
		Payment: payment,
	})
}

func (h *PaymentHandler) GetStats(c *fiber.Ctx) error {
	stats := h.paymentService.GetStats()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"stats": stats,
	})
}
