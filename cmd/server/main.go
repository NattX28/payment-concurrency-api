package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NattX28/payment-concurrency-api/internal/handler"
	"github.com/NattX28/payment-concurrency-api/internal/middleware"
	"github.com/NattX28/payment-concurrency-api/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	fmt.Println(`
		╔═══════════════════════════════════════╗
		║   Payment Processing API (Go)         ║
		║   Showcasing Concurrency Patterns     ║
		╚═══════════════════════════════════════╝
	`)

	log.Println("Starting Payment API Server...")

	// Initialize
	paymentService := service.NewPaymentService()
	log.Println("Payment service initialized")

	paymentHandler := handler.NewPaymentHandler(paymentService)
	healthHandler := handler.NewHealthHandler(paymentService)
	log.Println("Payment handler initialized")

	rateLimiter := middleware.NewRateLimiter(10, 10)
	log.Println("Rate limiter initialized")

	app := fiber.New(fiber.Config{
		AppName:               "Payment Concurrency API",
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: "15:04:05",
	}))

	app.Use(cors.New())

	app.Get("/health", healthHandler.Check)

	api := app.Group("/api/v1")

	api.Use(rateLimiter.Middleware())

	// Payment routes
	payments := api.Group("/payments")
	payments.Post("/", paymentHandler.CreatePayment)
	payments.Get("/:id", paymentHandler.GetPayment)
	payments.Get("/metrics/stats", paymentHandler.GetStats)

	log.Println("Routes registered")
	log.Println("   POST   /api/v1/payments")
	log.Println("   GET    /api/v1/payments/:id")
	log.Println("   GET    /api/v1/payments/metrics/stats")
	log.Println("   GET    /health")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	go func() {
		addr := fmt.Sprintf(":%s", port)

		if err := app.Listen(addr); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	log.Println("Shutting down server gracefully...")

	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	paymentService.Shutdown()
	log.Println("Server stopped gracefully")
	log.Println("Goodbye!")
}
