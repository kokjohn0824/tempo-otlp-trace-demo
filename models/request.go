package models

// OrderRequest represents an order creation request
type OrderRequest struct {
	UserID    string  `json:"user_id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Sleep     bool    `json:"sleep,omitempty"` // If true, simulate slow operation by adding 5s delay to processPayment
}

// OrderResponse represents an order creation response
type OrderResponse struct {
	OrderID   string  `json:"order_id"`
	Status    string  `json:"status"`
	TotalCost float64 `json:"total_cost"`
	Message   string  `json:"message"`
}

// UserProfileResponse represents user profile data
type UserProfileResponse struct {
	UserID      string            `json:"user_id"`
	Name        string            `json:"name"`
	Email       string            `json:"email"`
	Preferences map[string]string `json:"preferences"`
}

// ReportRequest represents a report generation request
type ReportRequest struct {
	ReportType string   `json:"report_type"`
	StartDate  string   `json:"start_date"`
	EndDate    string   `json:"end_date"`
	Filters    []string `json:"filters"`
}

// ReportResponse represents a report generation response
type ReportResponse struct {
	ReportID string `json:"report_id"`
	Status   string `json:"status"`
	URL      string `json:"url"`
	Message  string `json:"message"`
	Duration string `json:"duration"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query   string   `json:"query"`
	Filters []string `json:"filters"`
	Page    int      `json:"page"`
	Limit   int      `json:"limit"`
}

// SearchResponse represents search results
type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
	Page    int            `json:"page"`
}

// SearchResult represents a single search result
type SearchResult struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Score       float64 `json:"score"`
}

// BatchRequest represents a batch processing request
type BatchRequest struct {
	Items []string `json:"items"`
}

// BatchResponse represents batch processing response
type BatchResponse struct {
	BatchID        string   `json:"batch_id"`
	ProcessedCount int      `json:"processed_count"`
	FailedCount    int      `json:"failed_count"`
	Status         string   `json:"status"`
	Results        []string `json:"results"`
}

// SimulateRequest represents a custom simulation request
type SimulateRequest struct {
	Depth    int     `json:"depth"`
	Breadth  int     `json:"breadth"`
	Duration int     `json:"duration"`
	Variance float64 `json:"variance"`
}

// SimulateResponse represents simulation response
type SimulateResponse struct {
	TraceID   string `json:"trace_id"`
	SpanCount int    `json:"span_count"`
	Duration  string `json:"duration"`
	Message   string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SourceCodeMapping represents the mapping between span operation name and source code location
type SourceCodeMapping struct {
	SpanName     string `json:"span_name" example:"POST /api/order/create"`     // e.g., "POST /api/order/create"
	FilePath     string `json:"file_path" example:"handlers/order.go"`          // e.g., "handlers/order.go"
	FunctionName string `json:"function_name" example:"CreateOrder"`            // e.g., "CreateOrder"
	StartLine    int    `json:"start_line" example:"21"`                        // Starting line number
	EndLine      int    `json:"end_line" example:"85"`                          // Ending line number
	Description  string `json:"description" example:"Handles order creation"`   // Optional description
}

// SourceCodeResponse represents the response containing source code and metadata
type SourceCodeResponse struct {
	SpanName     string `json:"span_name" example:"POST /api/order/create"`
	FilePath     string `json:"file_path" example:"handlers/order.go"`
	FunctionName string `json:"function_name" example:"CreateOrder"`
	StartLine    int    `json:"start_line" example:"21"`
	EndLine      int    `json:"end_line" example:"85"`
	SourceCode   string `json:"source_code" example:"func CreateOrder(w http.ResponseWriter, r *http.Request) {...}"`
}

// MappingRequest represents a request to add/update source code mapping
type MappingRequest struct {
	Mappings []SourceCodeMapping `json:"mappings"`
}

// MappingResponse represents the response for mapping operations
type MappingResponse struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"Mappings updated successfully"`
	Count   int    `json:"count" example:"5"`
}
