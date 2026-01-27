package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"tempo-otlp-trace-demo/models"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("trace-demo-service")

// CreateOrder handles order creation with comprehensive tracing
func CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "POST /api/order/create",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/order/create"),
	)

	// Parse request
	var req models.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("user.id", req.UserID),
		attribute.String("product.id", req.ProductID),
		attribute.Int("order.quantity", req.Quantity),
		attribute.Bool("simulate.slow", req.Sleep),
	)

	// Step 1: Validate order
	validateOrder(ctx, req)

	// Step 2: Check inventory
	checkInventory(ctx, req.ProductID, req.Quantity)

	// Step 3: Calculate price
	totalCost := calculatePrice(ctx, req.Price, req.Quantity)

	// Step 4: Process payment (with nested spans)
	// If sleep=true, simulate slow payment processing (5 seconds delay)
	processPayment(ctx, req.UserID, totalCost, req.Sleep)

	// Step 5: Create shipment
	createShipment(ctx, req.UserID, req.ProductID)

	// Step 6: Send notifications (with nested spans)
	sendNotification(ctx, req.UserID, "order_created")

	// Step 7: Save to database
	orderID := saveToDatabase(ctx, "orders", req)

	// Return response
	response := models.OrderResponse{
		OrderID:   orderID,
		Status:    "success",
		TotalCost: totalCost,
		Message:   "Order created successfully",
	}

	span.SetAttributes(
		attribute.String("order.id", orderID),
		attribute.Float64("order.total_cost", totalCost),
	)
	span.SetStatus(codes.Ok, "order created")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func validateOrder(ctx context.Context, req models.OrderRequest) {
	_, span := tracer.Start(ctx, "validateOrder")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "validation"),
	)

	// Simulate validation work
	time.Sleep(time.Duration(50+rand.Intn(50)) * time.Millisecond)
	span.SetStatus(codes.Ok, "validation passed")
}

func checkInventory(ctx context.Context, productID string, quantity int) {
	_, span := tracer.Start(ctx, "checkInventory")
	defer span.End()

	span.SetAttributes(
		attribute.String("product.id", productID),
		attribute.Int("requested.quantity", quantity),
		attribute.String("db.system", "postgresql"),
		attribute.String("db.statement", "SELECT stock FROM inventory WHERE product_id = $1"),
	)

	// Simulate database query
	time.Sleep(time.Duration(100+rand.Intn(100)) * time.Millisecond)
	span.SetAttributes(attribute.Int("available.quantity", 100))
	span.SetStatus(codes.Ok, "inventory available")
}

func calculatePrice(ctx context.Context, price float64, quantity int) float64 {
	_, span := tracer.Start(ctx, "calculatePrice")
	defer span.End()

	span.SetAttributes(
		attribute.Float64("unit.price", price),
		attribute.Int("quantity", quantity),
	)

	// Simulate calculation
	time.Sleep(time.Duration(30+rand.Intn(50)) * time.Millisecond)

	totalCost := price * float64(quantity)
	span.SetAttributes(attribute.Float64("total.cost", totalCost))
	span.SetStatus(codes.Ok, "price calculated")

	return totalCost
}

func processPayment(ctx context.Context, userID string, amount float64, simulateSlow bool) {
	ctx, span := tracer.Start(ctx, "processPayment")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.Float64("payment.amount", amount),
		attribute.String("operation.type", "payment"),
		attribute.Bool("simulate.slow", simulateSlow),
	)

	// If simulateSlow is true, add 5 seconds delay to simulate anomaly
	if simulateSlow {
		span.SetAttributes(attribute.String("slow.reason", "simulated_delay"))
		time.Sleep(5 * time.Second)
	}

	// Nested: Call payment gateway
	callPaymentGateway(ctx, amount)

	// Nested: Record transaction
	recordTransaction(ctx, userID, amount)

	span.SetStatus(codes.Ok, "payment processed")
}

func callPaymentGateway(ctx context.Context, amount float64) {
	_, span := tracer.Start(ctx, "callPaymentGateway")
	defer span.End()

	span.SetAttributes(
		attribute.String("payment.gateway", "stripe"),
		attribute.Float64("payment.amount", amount),
		attribute.String("http.method", "POST"),
		attribute.String("http.url", "https://api.stripe.com/v1/charges"),
	)

	// Simulate external API call
	time.Sleep(time.Duration(150+rand.Intn(250)) * time.Millisecond)
	span.SetAttributes(attribute.String("payment.transaction_id", fmt.Sprintf("txn_%d", rand.Int())))
	span.SetStatus(codes.Ok, "payment gateway success")
}

func recordTransaction(ctx context.Context, userID string, amount float64) {
	_, span := tracer.Start(ctx, "recordTransaction")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.Float64("transaction.amount", amount),
		attribute.String("db.system", "postgresql"),
		attribute.String("db.statement", "INSERT INTO transactions (user_id, amount) VALUES ($1, $2)"),
	)

	// Simulate database write
	time.Sleep(time.Duration(20+rand.Intn(30)) * time.Millisecond)
	span.SetStatus(codes.Ok, "transaction recorded")
}

func createShipment(ctx context.Context, userID, productID string) {
	_, span := tracer.Start(ctx, "createShipment")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("product.id", productID),
		attribute.String("operation.type", "shipment"),
	)

	// Simulate shipment creation
	time.Sleep(time.Duration(80+rand.Intn(70)) * time.Millisecond)
	span.SetAttributes(attribute.String("shipment.id", fmt.Sprintf("ship_%d", rand.Int())))
	span.SetStatus(codes.Ok, "shipment created")
}

func sendNotification(ctx context.Context, userID, notificationType string) {
	ctx, span := tracer.Start(ctx, "sendNotification")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("notification.type", notificationType),
	)

	// Nested: Send email
	sendEmail(ctx, userID)

	// Nested: Send SMS
	sendSMS(ctx, userID)

	span.SetStatus(codes.Ok, "notifications sent")
}

func sendEmail(ctx context.Context, userID string) {
	_, span := tracer.Start(ctx, "sendEmail")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("email.provider", "sendgrid"),
	)

	// Simulate email sending
	time.Sleep(time.Duration(30+rand.Intn(30)) * time.Millisecond)
	span.SetStatus(codes.Ok, "email sent")
}

func sendSMS(ctx context.Context, userID string) {
	_, span := tracer.Start(ctx, "sendSMS")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("sms.provider", "twilio"),
	)

	// Simulate SMS sending
	time.Sleep(time.Duration(20+rand.Intn(20)) * time.Millisecond)
	span.SetStatus(codes.Ok, "sms sent")
}

func saveToDatabase(ctx context.Context, table string, data interface{}) string {
	_, span := tracer.Start(ctx, "saveToDatabase")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.table", table),
		attribute.String("db.operation", "INSERT"),
	)

	// Simulate database write
	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

	id := fmt.Sprintf("%s_%d", table, rand.Int())
	span.SetAttributes(attribute.String("record.id", id))
	span.SetStatus(codes.Ok, "data saved")

	return id
}
