package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"tempo-otlp-trace-demo/handlers"
	"tempo-otlp-trace-demo/tracing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "tempo-otlp-trace-demo/docs"
)

// @title Tempo OTLP Trace Demo API
// @version 1.0
// @description API for generating traces and retrieving source code mappings for performance analysis
// @host localhost:8080
// @BasePath /

func main() {
	log.Println("Starting Tempo OTLP Trace Demo Service...")

	// Initialize tracer
	ctx := context.Background()
	tp, err := tracing.InitTracer(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Get tracer for middleware
	tracer := otel.Tracer("trace-demo-service")

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/api/order/create", handlers.CreateOrder)
	mux.HandleFunc("/api/user/profile", handlers.GetUserProfile)
	mux.HandleFunc("/api/report/generate", handlers.GenerateReport)
	mux.HandleFunc("/api/search", handlers.Search)
	mux.HandleFunc("/api/batch/process", handlers.ProcessBatch)
	mux.HandleFunc("/api/simulate", handlers.Simulate)

	// Source code analysis endpoints
	mux.HandleFunc("/api/source-code", handlers.GetSourceCode)
	mux.HandleFunc("/api/span-names", handlers.GetSpanNames)
	mux.HandleFunc("/api/mappings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetMappings(w, r)
		case http.MethodPost:
			handlers.UpdateMappings(w, r)
		case http.MethodDelete:
			handlers.DeleteMapping(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/mappings/reload", handlers.ReloadMappings)

	// Swagger UI endpoint
	mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Root endpoint with API documentation
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Tempo OTLP Trace Demo</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        .endpoint { background: #f4f4f4; padding: 10px; margin: 10px 0; border-left: 4px solid #007bff; }
        .method { font-weight: bold; color: #007bff; }
        .path { font-family: monospace; }
        .description { color: #666; margin-top: 5px; }
    </style>
</head>
<body>
    <h1>Tempo OTLP Trace Demo API</h1>
    <p>This service generates traces with various patterns for testing Tempo.</p>
    
    <h2>Available Endpoints:</h2>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/api/order/create</span>
        <div class="description">Create an order (10-12 spans, 600-1500ms)</div>
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/user/profile?user_id=123</span>
        <div class="description">Get user profile (4-5 spans, 110-310ms)</div>
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/api/report/generate</span>
        <div class="description">Generate report - LONG TRACE (10-12 spans, 1500-3500ms)</div>
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/search?q=test&page=1&limit=10</span>
        <div class="description">Search (6-7 spans, 210-530ms)</div>
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/api/batch/process</span>
        <div class="description">Batch processing (6-15 spans, 300-1500ms)</div>
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/simulate?depth=3&breadth=2&duration=100&variance=0.5</span>
        <div class="description">Custom simulation (configurable spans and duration)</div>
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/health</span>
        <div class="description">Health check</div>
    </div>
    
    <h2>Source Code Analysis Endpoints:</h2>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/api/source-code</span>
        <div class="description">Get source code for a specific span (JSON body: {"spanName": "xxx"})</div>
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/span-names</span>
        <div class="description">Get all available span names with mappings</div>
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/mappings</span>
        <div class="description">Get all source code mappings</div>
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/api/mappings</span>
        <div class="description">Update source code mappings</div>
    </div>
    
    <div class="endpoint">
        <span class="method">DELETE</span> <span class="path">/api/mappings?span_name=xxx</span>
        <div class="description">Delete a source code mapping</div>
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/api/mappings/reload</span>
        <div class="description">Reload mappings from file</div>
    </div>
    
    <h2>API Documentation:</h2>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/swagger/</span>
        <div class="description">Interactive Swagger UI for API testing</div>
    </div>
</body>
</html>
`))
	})

	// Wrap mux with tracing middleware
	handler := tracingMiddleware(tracer, mux)

	// Setup HTTP server
	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// tracingMiddleware adds tracing context propagation
func tracingMiddleware(tracer trace.Tracer, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract trace context from incoming request
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		r = r.WithContext(ctx)

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
