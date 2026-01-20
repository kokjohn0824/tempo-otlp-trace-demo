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

// Search handles search requests
func Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "GET /api/search",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/search"),
	)

	// Parse query parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		query = "default"
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			page = parsed
		}
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	span.SetAttributes(
		attribute.String("search.query", query),
		attribute.Int("search.page", page),
		attribute.Int("search.limit", limit),
	)

	// Step 1: Parse query
	parsedQuery := parseQuery(ctx, query)

	// Step 2: Search index
	results := searchIndex(ctx, parsedQuery, limit)

	// Step 3: Rank results
	rankedResults := rankResults(ctx, results)

	// Step 4: Fetch details (with nested batch query)
	detailedResults := fetchDetails(ctx, rankedResults)

	// Step 5: Apply filters
	filteredResults := applyFilters(ctx, detailedResults)

	// Build response
	response := models.SearchResponse{
		Results: filteredResults,
		Total:   len(filteredResults),
		Page:    page,
	}

	span.SetAttributes(
		attribute.Int("search.results_count", len(filteredResults)),
	)
	span.SetStatus(codes.Ok, "search completed")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func parseQuery(ctx context.Context, query string) string {
	_, span := tracer.Start(ctx, "parseQuery")
	defer span.End()

	span.SetAttributes(
		attribute.String("search.query", query),
		attribute.String("operation.type", "query_parsing"),
	)

	// Simulate query parsing
	time.Sleep(time.Duration(10+rand.Intn(20)) * time.Millisecond)

	parsedQuery := fmt.Sprintf("parsed:%s", query)
	span.SetAttributes(attribute.String("search.parsed_query", parsedQuery))
	span.SetStatus(codes.Ok, "query parsed")

	return parsedQuery
}

func searchIndex(ctx context.Context, query string, limit int) []map[string]interface{} {
	_, span := tracer.Start(ctx, "searchIndex")
	defer span.End()

	span.SetAttributes(
		attribute.String("search.query", query),
		attribute.Int("search.limit", limit),
		attribute.String("search.engine", "elasticsearch"),
		attribute.String("search.index", "products"),
	)

	// Simulate index search
	time.Sleep(time.Duration(80+rand.Intn(120)) * time.Millisecond)

	// Generate mock results
	results := make([]map[string]interface{}, limit)
	for i := 0; i < limit; i++ {
		results[i] = map[string]interface{}{
			"id":    fmt.Sprintf("item_%d", i+1),
			"score": rand.Float64() * 100,
		}
	}

	span.SetAttributes(attribute.Int("search.hits", len(results)))
	span.SetStatus(codes.Ok, "index searched")

	return results
}

func rankResults(ctx context.Context, results []map[string]interface{}) []map[string]interface{} {
	_, span := tracer.Start(ctx, "rankResults")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "ranking"),
		attribute.Int("results.count", len(results)),
	)

	// Simulate ranking algorithm
	time.Sleep(time.Duration(40+rand.Intn(60)) * time.Millisecond)

	span.SetStatus(codes.Ok, "results ranked")
	return results
}

func fetchDetails(ctx context.Context, results []map[string]interface{}) []models.SearchResult {
	ctx, span := tracer.Start(ctx, "fetchDetails")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "detail_fetching"),
		attribute.Int("items.count", len(results)),
	)

	// Nested: Batch query for details
	detailedResults := batchQuery(ctx, results)

	span.SetStatus(codes.Ok, "details fetched")
	return detailedResults
}

func batchQuery(ctx context.Context, results []map[string]interface{}) []models.SearchResult {
	_, span := tracer.Start(ctx, "batchQuery")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.Int("batch.size", len(results)),
	)

	// Simulate batch database query
	time.Sleep(time.Duration(50+rand.Intn(80)) * time.Millisecond)

	// Convert to SearchResult
	searchResults := make([]models.SearchResult, len(results))
	for i, result := range results {
		searchResults[i] = models.SearchResult{
			ID:          result["id"].(string),
			Title:       fmt.Sprintf("Product %d", i+1),
			Description: fmt.Sprintf("Description for product %d", i+1),
			Score:       result["score"].(float64),
		}
	}

	span.SetAttributes(attribute.Int("query.records", len(searchResults)))
	span.SetStatus(codes.Ok, "batch query complete")

	return searchResults
}

func applyFilters(ctx context.Context, results []models.SearchResult) []models.SearchResult {
	_, span := tracer.Start(ctx, "applyFilters")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "filtering"),
		attribute.Int("results.count", len(results)),
	)

	// Simulate filter application
	time.Sleep(time.Duration(20+rand.Intn(30)) * time.Millisecond)

	span.SetAttributes(attribute.Int("filtered.count", len(results)))
	span.SetStatus(codes.Ok, "filters applied")

	return results
}
