package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"tempo-otlp-trace-demo/models"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Simulate handles custom simulation requests
func Simulate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "GET /api/simulate",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/simulate"),
	)

	// Parse query parameters with defaults
	depth := getIntParam(r, "depth", 3)
	breadth := getIntParam(r, "breadth", 2)
	duration := getIntParam(r, "duration", 100)
	variance := getFloatParam(r, "variance", 0.5)

	// Limit parameters to reasonable values
	if depth > 10 {
		depth = 10
	}
	if breadth > 5 {
		breadth = 5
	}
	if duration > 1000 {
		duration = 1000
	}
	if variance > 1.0 {
		variance = 1.0
	}

	span.SetAttributes(
		attribute.Int("simulate.depth", depth),
		attribute.Int("simulate.breadth", breadth),
		attribute.Int("simulate.duration", duration),
		attribute.Float64("simulate.variance", variance),
	)

	startTime := time.Now()
	spanCount := 0

	// Generate trace tree recursively
	spanCount = generateTraceTree(ctx, 1, depth, breadth, duration, variance, &spanCount)

	totalDuration := time.Since(startTime)

	// Get trace ID from span context
	traceID := span.SpanContext().TraceID().String()

	response := models.SimulateResponse{
		TraceID:   traceID,
		SpanCount: spanCount,
		Duration:  totalDuration.String(),
		Message:   fmt.Sprintf("Generated %d spans with depth=%d, breadth=%d", spanCount, depth, breadth),
	}

	span.SetAttributes(
		attribute.String("trace.id", traceID),
		attribute.Int("trace.span_count", spanCount),
		attribute.Int64("trace.duration_ms", totalDuration.Milliseconds()),
	)
	span.SetStatus(codes.Ok, "simulation completed")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func generateTraceTree(ctx context.Context, currentDepth, maxDepth, breadth, baseDuration int, variance float64, spanCount *int) int {
	if currentDepth > maxDepth {
		return *spanCount
	}

	// Create spans at current level
	for i := 0; i < breadth; i++ {
		spanName := fmt.Sprintf("level-%d-span-%d", currentDepth, i+1)
		ctx, span := tracer.Start(ctx, spanName)

		span.SetAttributes(
			attribute.Int("span.depth", currentDepth),
			attribute.Int("span.breadth_index", i+1),
			attribute.String("span.type", "simulated"),
		)

		*spanCount++

		// Calculate duration with variance
		minDuration := int(float64(baseDuration) * (1.0 - variance))
		maxDuration := int(float64(baseDuration) * (1.0 + variance))
		if minDuration < 1 {
			minDuration = 1
		}
		actualDuration := minDuration + rand.Intn(maxDuration-minDuration+1)

		// Simulate work
		time.Sleep(time.Duration(actualDuration) * time.Millisecond)

		// Recursively create child spans
		if currentDepth < maxDepth {
			generateTraceTree(ctx, currentDepth+1, maxDepth, breadth, baseDuration, variance, spanCount)
		}

		span.SetAttributes(attribute.Int("span.duration_ms", actualDuration))
		span.SetStatus(codes.Ok, "span completed")
		span.End()
	}

	return *spanCount
}

func getIntParam(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}

func getFloatParam(r *http.Request, key string, defaultValue float64) float64 {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}
	return defaultValue
}
