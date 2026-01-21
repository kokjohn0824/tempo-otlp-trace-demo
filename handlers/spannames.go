package handlers

import (
	"encoding/json"
	"net/http"
	"sort"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// SpanNameInfo represents information about a span name
type SpanNameInfo struct {
	SpanName     string `json:"span_name" example:"POST /api/order/create"`
	FilePath     string `json:"file_path" example:"handlers/order.go"`
	FunctionName string `json:"function_name" example:"CreateOrder"`
	Description  string `json:"description" example:"Handles order creation with comprehensive tracing"`
	StartLine    int    `json:"start_line" example:"21"`
	EndLine      int    `json:"end_line" example:"85"`
}

// SpanNamesResponse represents the response for span names query
type SpanNamesResponse struct {
	SpanNames []SpanNameInfo `json:"span_names"`
	Count     int            `json:"count" example:"42"`
}

// GetSpanNames handles requests to retrieve all available span names
// GET /api/span-names
// @Summary Get all available span names
// @Description Returns a list of all span names that have source code mappings
// @Tags Source Code
// @Produce json
// @Success 200 {object} SpanNamesResponse
// @Router /api/span-names [get]
func GetSpanNames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "GET /api/span-names",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/span-names"),
	)

	// Get all mappings
	mappingsLock.RLock()
	spanNames := make([]SpanNameInfo, 0, len(mappings))
	for _, mapping := range mappings {
		spanNames = append(spanNames, SpanNameInfo{
			SpanName:     mapping.SpanName,
			FilePath:     mapping.FilePath,
			FunctionName: mapping.FunctionName,
			Description:  mapping.Description,
			StartLine:    mapping.StartLine,
			EndLine:      mapping.EndLine,
		})
	}
	mappingsLock.RUnlock()

	// Sort by span name for consistent output
	sort.Slice(spanNames, func(i, j int) bool {
		return spanNames[i].SpanName < spanNames[j].SpanName
	})

	response := SpanNamesResponse{
		SpanNames: spanNames,
		Count:     len(spanNames),
	}

	span.SetAttributes(attribute.Int("span_names.count", len(spanNames)))
	span.SetStatus(codes.Ok, "span names retrieved")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
