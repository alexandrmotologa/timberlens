package ingest

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/parser"
)

// DemoGenerator produces realistic multi-service microservice log streams.
type DemoGenerator struct {
	consumer LineConsumer
	seqID    uint64
	running  int32
	rateMs   time.Duration
}

// NewDemoGenerator creates a demo stream generator.
func NewDemoGenerator(rateMs time.Duration, consumer LineConsumer) *DemoGenerator {
	if rateMs <= 0 {
		rateMs = 50 * time.Millisecond
	}
	return &DemoGenerator{
		consumer: consumer,
		rateMs:   rateMs,
	}
}

// Start begins emitting synthetic logs until stopped.
func (dg *DemoGenerator) Start() {
	atomic.StoreInt32(&dg.running, 1)
	defer atomic.StoreInt32(&dg.running, 0)

	ticker := time.NewTicker(dg.rateMs)
	defer ticker.Stop()

	for atomic.LoadInt32(&dg.running) == 1 {
		select {
		case <-ticker.C:
			dg.emitBatch()
		}
	}
}

// Stop halts the generator.
func (dg *DemoGenerator) Stop() {
	atomic.StoreInt32(&dg.running, 0)
}

func (dg *DemoGenerator) emitBatch() {
	traceID := "trc-" + randomHex(4)
	r := randomInt(100)

	// Scenario 1: Distributed HTTP Request (API Gateway -> Auth -> Postgres)
	if r < 40 {
		// API Gateway inbound
		dg.emitJSON("api-gateway", "INFO", "Incoming request POST /api/v1/auth/login", traceID, map[string]any{
			"method": "POST",
			"path":   "/api/v1/auth/login",
			"ip":     "192.168.1.105",
		})

		// Auth service processing
		dg.emitJSON("auth-svc", "INFO", "Validating credentials for user alex@example.com", traceID, map[string]any{
			"user_id": "usr-8812",
			"tenant":  "acme-corp",
		})

		// Postgres query
		dg.emitJSON("postgres-db", "INFO", "SELECT id, password_hash FROM users WHERE email = $1 LIMIT 1", traceID, map[string]any{
			"latency_ms": 1.84,
			"rows":       1,
		})

		// API Gateway response
		dg.emitJSON("api-gateway", "INFO", "Request completed 200 OK", traceID, map[string]any{
			"status":     200,
			"latency_ms": 14.2,
		})
		return
	}

	// Scenario 2: Payment Processing Workflow (API Gateway -> Payment Worker -> Stripe)
	if r < 70 {
		dg.emitJSON("api-gateway", "INFO", "Incoming request POST /api/v1/checkout/charge", traceID, map[string]any{
			"method": "POST",
			"path":   "/api/v1/checkout/charge",
			"amount": 49.99,
		})

		dg.emitJSON("payment-worker", "INFO", "Authorizing card transaction via payment gateway", traceID, map[string]any{
			"gateway": "stripe",
			"card":    "visa_4242",
		})

		if r < 65 {
			dg.emitJSON("payment-worker", "INFO", "Charge succeeded ch_3M5qJ6LkdIwHu", traceID, map[string]any{
				"charge_id":  "ch_3M5qJ6LkdIwHu",
				"latency_ms": 145.8,
			})
			dg.emitJSON("api-gateway", "INFO", "Checkout completed 201 Created", traceID, map[string]any{
				"status":     201,
				"latency_ms": 158.4,
			})
		} else {
			dg.emitJSON("payment-worker", "WARN", "Card issuer declined transaction: insufficient_funds", traceID, map[string]any{
				"decline_code": "insufficient_funds",
			})
			dg.emitJSON("api-gateway", "WARN", "Checkout failed 402 Payment Required", traceID, map[string]any{
				"status":     402,
				"latency_ms": 162.1,
			})
		}
		return
	}

	// Scenario 3: Logfmt Background Worker
	if r < 85 {
		dg.emitLogfmt("notification-worker", "INFO", fmt.Sprintf("time=%q level=info service=notification-worker trace_id=%s msg=%q recipient=%q template=%q latency_ms=38.4",
			time.Now().Format(time.RFC3339), traceID, "Email dispatch completed", "alex@example.com", "welcome_v2"))
		return
	}

	// Scenario 4: Error with Stack Trace (Node.js exception)
	if r < 95 {
		stack := `Error: connect ECONNREFUSED 10.0.12.44:5432
    at TCPConnectWrap.afterConnect [as oncomplete] (node:net:1494:16)
    at /app/node_modules/pg-pool/index.js:45:11
    at /app/node_modules/pg/lib/client.js:120:5
    at DatabasePool.acquire (/app/src/db/pool.ts:48:19)
    at OrderRepository.createOrder (/app/src/repositories/orders.ts:88:27)
    at CheckoutHandler.execute (/app/src/handlers/checkout.ts:34:12)
    at /app/node_modules/express/lib/router/layer.js:95:5`

		dg.emitJSON("payment-worker", "ERROR", "Unhandled database connection refusal during transaction commit", traceID, map[string]any{
			"error":      "ECONNREFUSED",
			"stack":      stack,
			"retryCount": 3,
			"latency_ms": 1250.0,
		})
		dg.emitJSON("api-gateway", "ERROR", "Downstream service error 503 Service Unavailable", traceID, map[string]any{
			"status":     503,
			"latency_ms": 1280.5,
		})
		return
	}

	// Scenario 5: Fatal crash or alert
	dg.emitJSON("auth-svc", "FATAL", "RSA private key verification failed: corrupt PEM header", traceID, map[string]any{
		"component": "crypto/jwt",
		"key_id":    "k_2026_q4",
	})
}

func (dg *DemoGenerator) emitJSON(service, level, msg, traceID string, fields map[string]any) {
	payload := map[string]any{
		"time":     time.Now().Format(time.RFC3339Nano),
		"level":    level,
		"service":  service,
		"msg":      msg,
		"trace_id": traceID,
	}
	for k, v := range fields {
		payload[k] = v
	}

	bytesData, _ := json.Marshal(payload)
	seq := atomic.AddUint64(&dg.seqID, 1)
	rec, _ := parser.ParseLine(bytesData, seq)
	if rec != nil && dg.consumer != nil {
		dg.consumer(rec)
	}
}

func (dg *DemoGenerator) emitLogfmt(service, level, raw string) {
	seq := atomic.AddUint64(&dg.seqID, 1)
	rec, _ := parser.ParseLine([]byte(raw), seq)
	if rec != nil && dg.consumer != nil {
		dg.consumer(rec)
	}
}

func randomInt(max int64) int64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(max))
	return n.Int64()
}

func randomHex(bytesLen int) string {
	b := make([]byte, bytesLen)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
