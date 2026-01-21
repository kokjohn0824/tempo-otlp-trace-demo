package tracing

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// TempoSpan represents a span from Tempo
type TempoSpan struct {
	TraceID       string                 `json:"traceID"`
	SpanID        string                 `json:"spanID"`
	OperationName string                 `json:"operationName"`
	StartTime     int64                  `json:"startTime"`
	Duration      int64                  `json:"duration"`
	Tags          []TempoTag             `json:"tags"`
	References    []TempoReference       `json:"references"`
	Process       TempoProcess           `json:"process"`
}

// TempoTag represents a tag/attribute in a span
type TempoTag struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// TempoReference represents a reference to another span
type TempoReference struct {
	RefType string `json:"refType"`
	TraceID string `json:"traceID"`
	SpanID  string `json:"spanID"`
}

// TempoProcess represents the process information
type TempoProcess struct {
	ServiceName string     `json:"serviceName"`
	Tags        []TempoTag `json:"tags"`
}

// TempoTrace represents a complete trace from Tempo
type TempoTrace struct {
	TraceID   string       `json:"traceID"`
	Spans     []TempoSpan  `json:"spans"`
	Processes []TempoProcess `json:"processes"`
}

// TempoResponse represents the response from Tempo API
type TempoResponse struct {
	Data []TempoTrace `json:"data"`
}

// OTLP format structures
type OTLPBatch struct {
	Resource   OTLPResource    `json:"resource"`
	ScopeSpans []OTLPScopeSpan `json:"scopeSpans"`
}

type OTLPResource struct {
	Attributes []OTLPAttribute `json:"attributes"`
}

type OTLPScopeSpan struct {
	Scope OTLPScope  `json:"scope"`
	Spans []OTLPSpan `json:"spans"`
}

type OTLPScope struct {
	Name string `json:"name"`
}

type OTLPSpan struct {
	TraceID           string          `json:"traceId"`
	SpanID            string          `json:"spanId"`
	ParentSpanID      string          `json:"parentSpanId,omitempty"`
	Name              string          `json:"name"`
	Kind              string          `json:"kind"` // Changed from int to string
	StartTimeUnixNano string          `json:"startTimeUnixNano"`
	EndTimeUnixNano   string          `json:"endTimeUnixNano"`
	Attributes        []OTLPAttribute `json:"attributes"`
	Status            OTLPStatus      `json:"status"`
}

type OTLPAttribute struct {
	Key   string     `json:"key"`
	Value OTLPValue  `json:"value"`
}

type OTLPValue struct {
	StringValue string `json:"stringValue,omitempty"`
	IntValue    string `json:"intValue,omitempty"`
	BoolValue   bool   `json:"boolValue,omitempty"`
}

type OTLPStatus struct {
	Code string `json:"code"` // Changed from int to string
}

type OTLPTrace struct {
	Batches []OTLPBatch `json:"batches"`
}

// GetTempoURL returns the Tempo query URL from environment or default
func GetTempoURL() string {
	url := os.Getenv("TEMPO_URL")
	if url == "" {
		url = "http://localhost:3200"
	}
	return url
}

// QueryTraceByID queries Tempo for a trace by trace ID
func QueryTraceByID(traceID string) (*TempoTrace, error) {
	tempoURL := GetTempoURL()
	url := fmt.Sprintf("%s/api/traces/%s", tempoURL, traceID)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to query Tempo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Tempo returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to parse as OTLP format first
	var otlpTrace OTLPTrace
	if err := json.Unmarshal(body, &otlpTrace); err == nil && len(otlpTrace.Batches) > 0 {
		// Convert OTLP format to Jaeger format
		return convertOTLPToJaeger(&otlpTrace, traceID)
	} else {
	}

	// Fallback to Jaeger format
	var trace TempoTrace
	if err := json.Unmarshal(body, &trace); err != nil {
		return nil, fmt.Errorf("failed to parse Tempo response: %w", err)
	}

	return &trace, nil
}

// convertOTLPToJaeger converts OTLP trace format to Jaeger format
func convertOTLPToJaeger(otlp *OTLPTrace, traceID string) (*TempoTrace, error) {
	
	trace := &TempoTrace{
		TraceID: traceID,
		Spans:   make([]TempoSpan, 0),
	}

	// Extract service name from resource attributes
	serviceName := "unknown-service"
	if len(otlp.Batches) > 0 && len(otlp.Batches[0].Resource.Attributes) > 0 {
		for _, attr := range otlp.Batches[0].Resource.Attributes {
			if attr.Key == "service.name" {
				serviceName = attr.Value.StringValue
				break
			}
		}
	}

	// Convert spans
	spanCount := 0
	for _, batch := range otlp.Batches {
		for _, scopeSpan := range batch.ScopeSpans {
			for _, otlpSpan := range scopeSpan.Spans {
				spanCount++
				// Calculate duration
				var startTime, endTime int64
				fmt.Sscanf(otlpSpan.StartTimeUnixNano, "%d", &startTime)
				fmt.Sscanf(otlpSpan.EndTimeUnixNano, "%d", &endTime)
				durationMicros := (endTime - startTime) / 1000 // Convert nanoseconds to microseconds

				// Convert attributes to tags
				tags := make([]TempoTag, 0, len(otlpSpan.Attributes))
				for _, attr := range otlpSpan.Attributes {
					value := attr.Value.StringValue
					if value == "" && attr.Value.IntValue != "" {
						value = attr.Value.IntValue
					}
					tags = append(tags, TempoTag{
						Key:   attr.Key,
						Type:  "string",
						Value: value,
					})
				}

				// Convert references - keep base64 format for parent span ID
				references := make([]TempoReference, 0)
				if otlpSpan.ParentSpanID != "" {
					references = append(references, TempoReference{
						RefType: "CHILD_OF",
						TraceID: traceID,
						SpanID:  otlpSpan.ParentSpanID, // Keep base64 format
					})
				}

				// Keep the original base64 span ID from OTLP
				span := TempoSpan{
					TraceID:       traceID,
					SpanID:        otlpSpan.SpanID, // Keep base64 format
					OperationName: otlpSpan.Name,
					StartTime:     startTime / 1000, // Convert to microseconds
					Duration:      durationMicros,
					Tags:          tags,
					References:    references,
					Process: TempoProcess{
						ServiceName: serviceName,
					},
				}

				trace.Spans = append(trace.Spans, span)
			}
		}
	}

	return trace, nil
}

// FindSpanByID finds a specific span in a trace by span ID
func FindSpanByID(trace *TempoTrace, spanID string) *TempoSpan {
	for i := range trace.Spans {
		if trace.Spans[i].SpanID == spanID {
			return &trace.Spans[i]
		}
	}
	return nil
}

// FindChildSpans finds all child spans of a given span
func FindChildSpans(trace *TempoTrace, parentSpanID string) []TempoSpan {
	var children []TempoSpan
	for _, span := range trace.Spans {
		for _, ref := range span.References {
			if ref.RefType == "CHILD_OF" && ref.SpanID == parentSpanID {
				children = append(children, span)
				break
			}
		}
	}
	return children
}

// GetSpanAttributes extracts span attributes as a map
func GetSpanAttributes(span *TempoSpan) map[string]string {
	attrs := make(map[string]string)
	for _, tag := range span.Tags {
		attrs[tag.Key] = fmt.Sprintf("%v", tag.Value)
	}
	return attrs
}

// FormatDuration formats duration in microseconds to human-readable format
func FormatDuration(durationMicros int64) string {
	duration := time.Duration(durationMicros) * time.Microsecond
	if duration < time.Millisecond {
		return fmt.Sprintf("%.2fμs", float64(durationMicros))
	} else if duration < time.Second {
		return fmt.Sprintf("%.2fms", float64(durationMicros)/1000)
	}
	return fmt.Sprintf("%.2fs", float64(durationMicros)/1000000)
}
