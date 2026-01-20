package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"tempo-otlp-trace-demo/models"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ProcessBatch handles batch processing requests
func ProcessBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "POST /api/batch/process",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/batch/process"),
	)

	// Parse request
	var req models.BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Default to 5 items if not specified
	if len(req.Items) == 0 {
		req.Items = []string{"item1", "item2", "item3", "item4", "item5"}
	}

	span.SetAttributes(
		attribute.Int("batch.size", len(req.Items)),
	)

	// Step 1: Validate batch
	validateBatch(ctx, req)

	// Step 2: Process items (with nested spans for each item)
	results := processItems(ctx, req.Items)

	// Step 3: Aggregate results
	aggregated := aggregateResults(ctx, results)

	// Step 4: Save results
	batchID := saveBatchResults(ctx, aggregated)

	// Count successes and failures
	processedCount := 0
	failedCount := 0
	for _, result := range results {
		if result == "success" {
			processedCount++
		} else {
			failedCount++
		}
	}

	response := models.BatchResponse{
		BatchID:        batchID,
		ProcessedCount: processedCount,
		FailedCount:    failedCount,
		Status:         "completed",
		Results:        results,
	}

	span.SetAttributes(
		attribute.String("batch.id", batchID),
		attribute.Int("batch.processed", processedCount),
		attribute.Int("batch.failed", failedCount),
	)
	span.SetStatus(codes.Ok, "batch processed")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func validateBatch(ctx context.Context, req models.BatchRequest) {
	_, span := tracer.Start(ctx, "validateBatch")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "validation"),
		attribute.Int("batch.size", len(req.Items)),
	)

	// Simulate validation
	time.Sleep(time.Duration(30+rand.Intn(30)) * time.Millisecond)
	span.SetStatus(codes.Ok, "batch validated")
}

func processItems(ctx context.Context, items []string) []string {
	ctx, span := tracer.Start(ctx, "processItems")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "batch_processing"),
		attribute.Int("items.count", len(items)),
	)

	results := make([]string, len(items))

	// Process each item with its own span
	for i, item := range items {
		result := processItem(ctx, i+1, item)
		results[i] = result
	}

	span.SetStatus(codes.Ok, "items processed")
	return results
}

func processItem(ctx context.Context, index int, item string) string {
	_, span := tracer.Start(ctx, fmt.Sprintf("processItem-%d", index))
	defer span.End()

	span.SetAttributes(
		attribute.String("item.id", item),
		attribute.Int("item.index", index),
		attribute.String("operation.type", "item_processing"),
	)

	// Simulate item processing
	time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

	// Randomly succeed or fail (90% success rate)
	result := "success"
	if rand.Float64() < 0.1 {
		result = "failed"
		span.SetStatus(codes.Error, "item processing failed")
		span.SetAttributes(attribute.String("error.reason", "random_failure"))
	} else {
		span.SetStatus(codes.Ok, "item processed")
	}

	span.SetAttributes(attribute.String("item.result", result))
	return result
}

func aggregateResults(ctx context.Context, results []string) map[string]interface{} {
	_, span := tracer.Start(ctx, "aggregateResults")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "aggregation"),
		attribute.Int("results.count", len(results)),
	)

	// Simulate aggregation
	time.Sleep(time.Duration(40+rand.Intn(40)) * time.Millisecond)

	successCount := 0
	failedCount := 0
	for _, result := range results {
		if result == "success" {
			successCount++
		} else {
			failedCount++
		}
	}

	aggregated := map[string]interface{}{
		"total":   len(results),
		"success": successCount,
		"failed":  failedCount,
	}

	span.SetAttributes(
		attribute.Int("aggregated.success", successCount),
		attribute.Int("aggregated.failed", failedCount),
	)
	span.SetStatus(codes.Ok, "results aggregated")

	return aggregated
}

func saveBatchResults(ctx context.Context, results map[string]interface{}) string {
	_, span := tracer.Start(ctx, "saveResults")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.table", "batch_results"),
		attribute.String("db.operation", "INSERT"),
	)

	// Simulate database write
	time.Sleep(time.Duration(100+rand.Intn(100)) * time.Millisecond)

	batchID := fmt.Sprintf("batch_%d", rand.Int())
	span.SetAttributes(attribute.String("batch.id", batchID))
	span.SetStatus(codes.Ok, "results saved")

	return batchID
}
