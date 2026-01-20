package handlers

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"tempo-otlp-trace-demo/models"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// GetUserProfile handles user profile retrieval
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "GET /api/user/profile",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/user/profile"),
	)

	// Get user ID from query params
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "user_12345"
	}

	span.SetAttributes(attribute.String("user.id", userID))

	// Step 1: Authenticate
	authenticate(ctx, userID)

	// Step 2: Query database
	userData := queryDatabase(ctx, "users", userID)

	// Step 3: Load preferences
	preferences := loadPreferences(ctx, userID)

	// Step 4: Format response
	response := formatResponse(ctx, userData, preferences)

	span.SetStatus(codes.Ok, "profile retrieved")
	span.SetAttributes(attribute.Int("http.status_code", 200))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func authenticate(ctx context.Context, userID string) {
	_, span := tracer.Start(ctx, "authenticate")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("auth.method", "jwt"),
	)

	// Simulate authentication
	time.Sleep(time.Duration(20+rand.Intn(30)) * time.Millisecond)
	span.SetStatus(codes.Ok, "authenticated")
}

func queryDatabase(ctx context.Context, table, id string) map[string]interface{} {
	_, span := tracer.Start(ctx, "queryDatabase")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.table", table),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.statement", "SELECT * FROM users WHERE id = $1"),
		attribute.String("record.id", id),
	)

	// Simulate database query
	time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

	data := map[string]interface{}{
		"id":    id,
		"name":  "John Doe",
		"email": "john@example.com",
	}

	span.SetStatus(codes.Ok, "query successful")
	return data
}

func loadPreferences(ctx context.Context, userID string) map[string]string {
	_, span := tracer.Start(ctx, "loadPreferences")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("db.system", "redis"),
		attribute.String("db.operation", "GET"),
	)

	// Simulate cache/database query
	time.Sleep(time.Duration(30+rand.Intn(50)) * time.Millisecond)

	preferences := map[string]string{
		"theme":    "dark",
		"language": "en",
		"timezone": "UTC",
	}

	span.SetAttributes(attribute.Int("preferences.count", len(preferences)))
	span.SetStatus(codes.Ok, "preferences loaded")
	return preferences
}

func formatResponse(ctx context.Context, userData map[string]interface{}, preferences map[string]string) models.UserProfileResponse {
	_, span := tracer.Start(ctx, "formatResponse")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "formatting"),
	)

	// Simulate formatting work
	time.Sleep(time.Duration(10+rand.Intn(20)) * time.Millisecond)

	response := models.UserProfileResponse{
		UserID:      userData["id"].(string),
		Name:        userData["name"].(string),
		Email:       userData["email"].(string),
		Preferences: preferences,
	}

	span.SetStatus(codes.Ok, "response formatted")
	return response
}
