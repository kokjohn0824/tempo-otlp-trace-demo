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

// GenerateReport handles report generation (long-running operation)
// @Summary Generate a report
// @Description Generates a report with comprehensive tracing. LONG TRACE - Generates 10-12 spans with 1500-3500ms duration.
// @Tags Reports
// @Accept json
// @Produce json
// @Param request body models.ReportRequest true "Report generation request"
// @Success 200 {object} models.ReportResponse "Report generated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Router /api/report/generate [post]
func GenerateReport(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "POST /api/report/generate",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/api/report/generate"),
	)

	// Parse request
	var req models.ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("report.type", req.ReportType),
		attribute.String("report.start_date", req.StartDate),
		attribute.String("report.end_date", req.EndDate),
	)

	// Step 1: Validate request
	validateReportRequest(ctx, req)

	// Step 2: Fetch data from multiple sources (with nested spans)
	data := fetchDataFromMultipleSources(ctx, req)

	// Step 3: Process data (with nested spans)
	processedData := processReportData(ctx, data)

	// Step 4: Generate PDF (long operation)
	pdfURL := generatePDF(ctx, processedData, req.ReportType)

	// Step 5: Upload to storage
	storageURL := uploadToStorage(ctx, pdfURL)

	// Step 6: Notify user
	notifyUser(ctx, "report_ready", storageURL)

	reportID := fmt.Sprintf("report_%d", rand.Int())
	duration := time.Since(startTime)

	response := models.ReportResponse{
		ReportID: reportID,
		Status:   "completed",
		URL:      storageURL,
		Message:  "Report generated successfully",
		Duration: duration.String(),
	}

	span.SetAttributes(
		attribute.String("report.id", reportID),
		attribute.String("report.url", storageURL),
		attribute.Int64("report.duration_ms", duration.Milliseconds()),
	)
	span.SetStatus(codes.Ok, "report generated")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func validateReportRequest(ctx context.Context, req models.ReportRequest) {
	_, span := tracer.Start(ctx, "validateRequest")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "validation"),
		attribute.String("report.type", req.ReportType),
	)

	// Simulate validation
	time.Sleep(time.Duration(30+rand.Intn(30)) * time.Millisecond)
	span.SetStatus(codes.Ok, "validation passed")
}

func fetchDataFromMultipleSources(ctx context.Context, req models.ReportRequest) map[string]interface{} {
	ctx, span := tracer.Start(ctx, "fetchDataFromMultipleSources")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "data_fetching"),
	)

	// Nested: Query main database
	mainData := queryMainDB(ctx, req)

	// Nested: Query analytics database
	analyticsData := queryAnalyticsDB(ctx, req)

	// Nested: Fetch from external API
	externalData := fetchExternalAPI(ctx, req)

	data := map[string]interface{}{
		"main":      mainData,
		"analytics": analyticsData,
		"external":  externalData,
	}

	span.SetAttributes(attribute.Int("data.sources", 3))
	span.SetStatus(codes.Ok, "data fetched")
	return data
}

func queryMainDB(ctx context.Context, req models.ReportRequest) map[string]interface{} {
	_, span := tracer.Start(ctx, "queryMainDB")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.name", "main_db"),
		attribute.String("db.statement", "SELECT * FROM transactions WHERE date BETWEEN $1 AND $2"),
	)

	// Simulate long database query
	time.Sleep(time.Duration(200+rand.Intn(200)) * time.Millisecond)

	data := map[string]interface{}{
		"records": 1000,
		"source":  "main_db",
	}

	span.SetAttributes(attribute.Int("query.records", 1000))
	span.SetStatus(codes.Ok, "main db query complete")
	return data
}

func queryAnalyticsDB(ctx context.Context, req models.ReportRequest) map[string]interface{} {
	_, span := tracer.Start(ctx, "queryAnalyticsDB")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "clickhouse"),
		attribute.String("db.name", "analytics_db"),
		attribute.String("db.statement", "SELECT * FROM events WHERE timestamp BETWEEN $1 AND $2"),
	)

	// Simulate analytics query
	time.Sleep(time.Duration(200+rand.Intn(200)) * time.Millisecond)

	data := map[string]interface{}{
		"records": 5000,
		"source":  "analytics_db",
	}

	span.SetAttributes(attribute.Int("query.records", 5000))
	span.SetStatus(codes.Ok, "analytics db query complete")
	return data
}

func fetchExternalAPI(ctx context.Context, req models.ReportRequest) map[string]interface{} {
	_, span := tracer.Start(ctx, "fetchExternalAPI")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", "GET"),
		attribute.String("http.url", "https://api.external.com/data"),
		attribute.String("external.service", "data_provider"),
	)

	// Simulate external API call
	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

	data := map[string]interface{}{
		"records": 500,
		"source":  "external_api",
	}

	span.SetAttributes(attribute.Int("api.records", 500))
	span.SetStatus(codes.Ok, "external api call complete")
	return data
}

func processReportData(ctx context.Context, data map[string]interface{}) map[string]interface{} {
	ctx, span := tracer.Start(ctx, "processData")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "data_processing"),
	)

	// Nested: Aggregate data
	aggregated := aggregateData(ctx, data)

	// Nested: Calculate metrics
	metrics := calculateMetrics(ctx, aggregated)

	processed := map[string]interface{}{
		"aggregated": aggregated,
		"metrics":    metrics,
	}

	span.SetStatus(codes.Ok, "data processed")
	return processed
}

func aggregateData(ctx context.Context, data map[string]interface{}) map[string]interface{} {
	_, span := tracer.Start(ctx, "aggregateData")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "aggregation"),
	)

	// Simulate data aggregation
	time.Sleep(time.Duration(150+rand.Intn(250)) * time.Millisecond)

	aggregated := map[string]interface{}{
		"total_records": 6500,
		"aggregated":    true,
	}

	span.SetAttributes(attribute.Int("aggregated.records", 6500))
	span.SetStatus(codes.Ok, "data aggregated")
	return aggregated
}

func calculateMetrics(ctx context.Context, data map[string]interface{}) map[string]interface{} {
	_, span := tracer.Start(ctx, "calculateMetrics")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "calculation"),
	)

	// Simulate metrics calculation
	time.Sleep(time.Duration(150+rand.Intn(250)) * time.Millisecond)

	metrics := map[string]interface{}{
		"average": 125.5,
		"total":   815750,
		"count":   6500,
	}

	span.SetAttributes(attribute.Int("metrics.count", 3))
	span.SetStatus(codes.Ok, "metrics calculated")
	return metrics
}

func generatePDF(ctx context.Context, data map[string]interface{}, reportType string) string {
	_, span := tracer.Start(ctx, "generatePDF")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "pdf_generation"),
		attribute.String("report.type", reportType),
		attribute.String("pdf.library", "wkhtmltopdf"),
	)

	// Simulate PDF generation (long operation)
	time.Sleep(time.Duration(400+rand.Intn(800)) * time.Millisecond)

	pdfPath := fmt.Sprintf("/tmp/report_%d.pdf", rand.Int())
	span.SetAttributes(
		attribute.String("pdf.path", pdfPath),
		attribute.Int("pdf.pages", 25),
	)
	span.SetStatus(codes.Ok, "pdf generated")

	return pdfPath
}

func uploadToStorage(ctx context.Context, filePath string) string {
	_, span := tracer.Start(ctx, "uploadToStorage")
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.provider", "s3"),
		attribute.String("storage.bucket", "reports"),
		attribute.String("file.path", filePath),
	)

	// Simulate file upload
	time.Sleep(time.Duration(200+rand.Intn(300)) * time.Millisecond)

	url := fmt.Sprintf("https://s3.amazonaws.com/reports/report_%d.pdf", rand.Int())
	span.SetAttributes(attribute.String("storage.url", url))
	span.SetStatus(codes.Ok, "file uploaded")

	return url
}

func notifyUser(ctx context.Context, notificationType, message string) {
	_, span := tracer.Start(ctx, "notifyUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("notification.type", notificationType),
		attribute.String("notification.message", message),
	)

	// Simulate notification
	time.Sleep(time.Duration(50+rand.Intn(50)) * time.Millisecond)
	span.SetStatus(codes.Ok, "user notified")
}
