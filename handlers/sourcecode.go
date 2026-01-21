package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"tempo-otlp-trace-demo/models"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	mappings     map[string]models.SourceCodeMapping
	mappingsLock sync.RWMutex
	mappingsFile = "source_code_mappings.json"
)

// MappingFile represents the structure of the mapping JSON file
type MappingFile struct {
	Mappings []models.SourceCodeMapping `json:"mappings"`
}

// init loads the source code mappings on startup
func init() {
	if err := LoadMappings(); err != nil {
		fmt.Printf("Warning: Failed to load source code mappings: %v\n", err)
		mappings = make(map[string]models.SourceCodeMapping)
	}
}

// LoadMappings loads source code mappings from the JSON file
func LoadMappings() error {
	mappingsLock.Lock()
	defer mappingsLock.Unlock()

	file, err := os.Open(mappingsFile)
	if err != nil {
		return fmt.Errorf("failed to open mappings file: %w", err)
	}
	defer file.Close()

	var mappingFile MappingFile
	if err := json.NewDecoder(file).Decode(&mappingFile); err != nil {
		return fmt.Errorf("failed to decode mappings file: %w", err)
	}

	// Convert array to map for faster lookup
	mappings = make(map[string]models.SourceCodeMapping)
	for _, mapping := range mappingFile.Mappings {
		mappings[mapping.SpanName] = mapping
	}

	return nil
}

// SaveMappings saves source code mappings to the JSON file
func SaveMappings() error {
	mappingsLock.RLock()
	defer mappingsLock.RUnlock()

	// Convert map to array
	mappingArray := make([]models.SourceCodeMapping, 0, len(mappings))
	for _, mapping := range mappings {
		mappingArray = append(mappingArray, mapping)
	}

	mappingFile := MappingFile{
		Mappings: mappingArray,
	}

	file, err := os.Create(mappingsFile)
	if err != nil {
		return fmt.Errorf("failed to create mappings file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(mappingFile); err != nil {
		return fmt.Errorf("failed to encode mappings: %w", err)
	}

	return nil
}

// SourceCodeRequest represents the request body for source code query
type SourceCodeRequest struct {
	SpanName string `json:"spanName" example:"POST /api/order/create"`
}

// GetSourceCode handles requests to retrieve source code for a span
// @Summary Get source code for a span
// @Description Retrieves the source code associated with a specific span name
// @Tags Source Code
// @Accept json
// @Produce json
// @Param request body SourceCodeRequest true "Span name to query"
// @Success 200 {object} models.SourceCodeResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Mapping not found"
// @Router /api/source-code [post]
func GetSourceCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "POST /api/source-code",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/source-code"),
	)

	// Parse request body
	var req SourceCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request")
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.SpanName == "" {
		span.RecordError(fmt.Errorf("missing required parameter"))
		span.SetStatus(codes.Error, "missing parameter")
		http.Error(w, "Missing required parameter: spanName", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("span.name", req.SpanName),
	)

	// Look up source code mapping
	mappingsLock.RLock()
	mapping, found := mappings[req.SpanName]
	mappingsLock.RUnlock()

	if !found {
		span.SetStatus(codes.Error, "mapping not found")
		http.Error(w, fmt.Sprintf("No source code mapping found for span: %s", req.SpanName), http.StatusNotFound)
		return
	}

	// Read source code from file
	sourceCode, err := readSourceCode(mapping.FilePath, mapping.StartLine, mapping.EndLine)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read source code")
		http.Error(w, fmt.Sprintf("Failed to read source code: %v", err), http.StatusInternalServerError)
		return
	}

	// Build response
	response := models.SourceCodeResponse{
		SpanName:     req.SpanName,
		FilePath:     mapping.FilePath,
		FunctionName: mapping.FunctionName,
		StartLine:    mapping.StartLine,
		EndLine:      mapping.EndLine,
		SourceCode:   sourceCode,
	}

	span.SetAttributes(
		attribute.String("source.file_path", mapping.FilePath),
		attribute.String("source.function_name", mapping.FunctionName),
	)
	span.SetStatus(codes.Ok, "source code retrieved")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// readSourceCode reads specific lines from a source file
func readSourceCode(filePath string, startLine, endLine int) (string, error) {
	// Get the project root directory
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	fullPath := filepath.Join(workDir, filePath)
	file, err := os.Open(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", fullPath, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	lineNum := 1

	for scanner.Scan() {
		if lineNum >= startLine && lineNum <= endLine {
			lines = append(lines, scanner.Text())
		}
		if lineNum > endLine {
			break
		}
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return strings.Join(lines, "\n"), nil
}

// UpdateMappings handles requests to update source code mappings
// @Summary Update source code mappings
// @Description Updates or adds new source code mappings
// @Tags Mappings
// @Accept json
// @Produce json
// @Param request body models.MappingRequest true "Mappings to update"
// @Success 200 {object} models.MappingResponse
// @Failure 400 {string} string "Invalid request"
// @Router /api/mappings [post]
func UpdateMappings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "POST /api/mappings",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/mappings"),
	)

	// Parse request
	var req models.MappingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(req.Mappings) == 0 {
		span.SetStatus(codes.Error, "no mappings provided")
		http.Error(w, "No mappings provided", http.StatusBadRequest)
		return
	}

	// Update mappings in memory
	mappingsLock.Lock()
	for _, mapping := range req.Mappings {
		mappings[mapping.SpanName] = mapping
	}
	mappingsLock.Unlock()

	// Save to file
	if err := SaveMappings(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to save mappings")
		http.Error(w, fmt.Sprintf("Failed to save mappings: %v", err), http.StatusInternalServerError)
		return
	}

	response := models.MappingResponse{
		Status:  "success",
		Message: "Mappings updated successfully",
		Count:   len(req.Mappings),
	}

	span.SetAttributes(attribute.Int("mappings.count", len(req.Mappings)))
	span.SetStatus(codes.Ok, "mappings updated")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMappings handles requests to retrieve all source code mappings
// @Summary Get all source code mappings
// @Description Returns all configured source code mappings
// @Tags Mappings
// @Produce json
// @Success 200 {object} models.MappingRequest
// @Router /api/mappings [get]
func GetMappings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "GET /api/mappings",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/mappings"),
	)

	mappingsLock.RLock()
	mappingArray := make([]models.SourceCodeMapping, 0, len(mappings))
	for _, mapping := range mappings {
		mappingArray = append(mappingArray, mapping)
	}
	mappingsLock.RUnlock()

	response := models.MappingRequest{
		Mappings: mappingArray,
	}

	span.SetAttributes(attribute.Int("mappings.count", len(mappingArray)))
	span.SetStatus(codes.Ok, "mappings retrieved")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteMapping handles requests to delete a source code mapping
// @Summary Delete a source code mapping
// @Description Deletes a specific source code mapping by span name
// @Tags Mappings
// @Produce json
// @Param span_name query string true "Span name to delete"
// @Success 200 {object} models.MappingResponse
// @Failure 400 {string} string "Missing parameter"
// @Failure 404 {string} string "Mapping not found"
// @Router /api/mappings [delete]
func DeleteMapping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "DELETE /api/mappings",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/mappings"),
	)

	spanName := r.URL.Query().Get("span_name")
	if spanName == "" {
		span.SetStatus(codes.Error, "missing span_name parameter")
		http.Error(w, "Missing required parameter: span_name", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.String("span.name", spanName))

	mappingsLock.Lock()
	_, found := mappings[spanName]
	if !found {
		mappingsLock.Unlock()
		span.SetStatus(codes.Error, "mapping not found")
		http.Error(w, "Mapping not found", http.StatusNotFound)
		return
	}

	delete(mappings, spanName)
	mappingsLock.Unlock()

	// Save to file
	if err := SaveMappings(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to save mappings")
		http.Error(w, fmt.Sprintf("Failed to save mappings: %v", err), http.StatusInternalServerError)
		return
	}

	response := models.MappingResponse{
		Status:  "success",
		Message: fmt.Sprintf("Mapping for '%s' deleted successfully", spanName),
		Count:   1,
	}

	span.SetStatus(codes.Ok, "mapping deleted")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ReloadMappings handles requests to reload mappings from file
// @Summary Reload mappings from file
// @Description Reloads source code mappings from the configuration file
// @Tags Mappings
// @Produce json
// @Success 200 {object} models.MappingResponse
// @Failure 500 {string} string "Failed to reload"
// @Router /api/mappings/reload [post]
func ReloadMappings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "POST /api/mappings/reload",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/mappings/reload"),
	)

	if err := LoadMappings(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to reload mappings")
		http.Error(w, fmt.Sprintf("Failed to reload mappings: %v", err), http.StatusInternalServerError)
		return
	}

	mappingsLock.RLock()
	count := len(mappings)
	mappingsLock.RUnlock()

	response := models.MappingResponse{
		Status:  "success",
		Message: "Mappings reloaded successfully",
		Count:   count,
	}

	span.SetAttributes(attribute.Int("mappings.count", count))
	span.SetStatus(codes.Ok, "mappings reloaded")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
