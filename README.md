![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Fiber](https://img.shields.io/badge/Fiber-v2-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![Fly.io](https://img.shields.io/badge/Fly.io-8B5CF6?style=for-the-badge&logo=flydotio&logoColor=white)

# 💳 Payment Processing API

Payment API showcasing Go concurrency patterns for Fintech applications.

**Live Demo:** `https://your-app.fly.dev`

---

## Features

- **Worker Pool** - 10 goroutines process 1000+ concurrent payments
- **Rate Limiting** - Token bucket (10 req/sec per user, concurrent-safe)
- **Async Webhooks** - Non-blocking via channels
- **Graceful Shutdown** - Zero data loss


---

## Quick Start
```bash
go run cmd/server/main.go
curl http://localhost:3000/health
```

---

## API

**Create Payment**
```bash
curl -X POST http://localhost:3000/api/v1/payments \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user123" \
  -d '{"user_id":"user123","amount":100.50,"currency":"THB"}'

# → {"payment_id":"...", "status":"pending"}
```

**Get Payment**
```bash
curl http://localhost:3000/api/v1/payments/{id}
# → Status: pending → processing → completed/failed
```

**Health Check**
```bash
curl http://localhost:3000/health
# → goroutines, memory, payment stats
```

---

## Architecture
```
Request → Rate Limiter → Handler → Service → Worker Pool (10 workers)
                                              ↓
                                         sync.Map Storage
                                              ↓
                                         Webhook Dispatcher
```

**Key Patterns:**
- `sync.Map` - thread-safe storage
- Buffered channels - task queue
- `sync.WaitGroup` - graceful shutdown

---

## Deploy
```bash
fly launch && fly deploy
# → https://your-app.fly.dev
```

---

## Tech Stack

Go 1.21+ • Fiber v2 • Goroutines & Channels • sync.Map • golang.org/x/time/rate • Docker • Fly.io

---

## Testing
```bash
# Create & check payment
PAYMENT_ID=$(curl -s -X POST http://localhost:3000/api/v1/payments \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user123" \
  -d '{"user_id":"user123","amount":100,"currency":"THB"}' | jq -r '.payment_id')

sleep 3
curl http://localhost:3000/api/v1/payments/$PAYMENT_ID | jq '.payment.status'
```

---

Made with ❤️ | [GitHub](https://github.com/NattX28)
