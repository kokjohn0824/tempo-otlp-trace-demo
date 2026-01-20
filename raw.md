You want the smallest stack that guarantees real traces land in Tempo with correct parent/child + durations, so your “longest span in a traceID” logic is meaningful.

Minimal demo architecture (recommended)

Service(s) → OTLP (gRPC/HTTP) → OpenTelemetry Collector → Tempo → Grafana (optional UI)

Why: it keeps your app config simple, and the Collector is where you can add batching, retries, filtering, and later fan-out without touching code.

The 3 things people usually miss
	1.	Tempo OTLP bind address in Docker
	•	Newer Tempo defaults OTLP receivers to localhost:4317/4318, which breaks ingestion from other containers unless you set it to 0.0.0.0.  ￼
	2.	Sampling accidentally dropping the “long” trace
	•	For a demo, force 100% sampling (or parent-based 1.0), otherwise your “abnormal but rare” traces disappear.
	3.	Context propagation between services
	•	Each service must extract/inject W3C traceparent (most OTel auto-instrumentation does this). If propagation breaks, you’ll see multiple traceIDs and your “longest span” becomes misleading.

⸻

Docker Compose: Tempo + OTel Collector (+ Grafana)

This is a minimal pattern aligned with Grafana’s “local docker compose” guidance.  ￼

docker-compose.yml (core idea)
	•	otel-collector listens on 4317 (gRPC) and 4318 (HTTP) for apps
	•	collector exports to Tempo via OTLP gRPC
	•	Tempo exposes query API + OTLP receiver (or you can keep ingestion only via collector)

I’m not pasting Grafana’s full example files verbatim (they’re a bit long), but the official repo/docs examples are the baseline reference.  ￼

otel-collector.yaml (minimal)

Key parts:
	•	receivers: otlp with grpc/http
	•	processors: batch (reduces load, improves efficiency)
	•	exporters: otlp to Tempo

Example shape:

receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch: {}

exporters:
  otlp:
    endpoint: tempo:4317
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]

tempo.yaml (minimum to avoid Docker bind pitfall)

Ensure OTLP receivers bind to all interfaces (important in containers).  ￼

distributor:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317
        http:
          endpoint: 0.0.0.0:4318


⸻

Fastest “app emits traces” option: OTel Java Agent (zero/low code)

For Spring Boot, the lowest-effort demo is OpenTelemetry Java agent:
	•	add -javaagent:/path/opentelemetry-javaagent.jar
	•	set env vars to point at the collector

The standard OTLP exporter env var scheme is documented here.  ￼

Practical demo settings:

export OTEL_SERVICE_NAME=my-demo-service
export OTEL_TRACES_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318   # if using OTLP/HTTP
# or for gRPC:
# export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317

export OTEL_METRICS_EXPORTER=none
export OTEL_LOGS_EXPORTER=none

# demo = keep everything
export OTEL_TRACES_SAMPLER=parentbased_traceidratio
export OTEL_TRACES_SAMPLER_ARG=1.0

If your app is running inside docker-compose, use http://otel-collector:4318 (or :4317) instead of localhost.

⸻

Validation checklist (so you don’t chase ghosts)

Ingestion sanity
	•	Collector logs show spans received/exported (turn on a debug/logging exporter temporarily if needed).
	•	Tempo has the traceID (query by traceID via API or Grafana Explore).

Trace quality (for “longest span” correctness)
	•	At least one SERVER span per request.
	•	Nested spans exist (DB/client spans, internal spans).
	•	Durations look realistic (not all ~0ms). Common causes of 0ms:
	•	clock issues in container/VM
	•	incorrect manual instrumentation timestamps (less common with auto-instrumentation)

Tempo query gotchas
	•	If you later rely on searching by attributes (service.name, http.route), make sure they’re actually present (resource attributes vs span attributes).

⸻

Optional but useful (still “minimal”)
	•	Add Grafana with Tempo datasource so you can click around quickly.
	•	Add a second service and an HTTP call between them to prove propagation (you’ll immediately see whether you’re getting a single trace across services).

If you tell me your demo language/runtime (Spring Boot vs Go vs Python) and whether it’s single-service or multi-service, I can tailor the exact minimal compose + env vars for that shape.
If Tempo is already running, the absolute minimal path is:

App → OTLP exporter → Tempo OTLP receiver (4317 gRPC or 4318 HTTP)

Below are two minimal implementations (pick one).

⸻

Option A (recommended for Spring Boot): Zero-code via OpenTelemetry Java Agent

You don’t write tracing code; it auto-instruments Spring MVC, RestTemplate/WebClient, JDBC, etc.

1) Run your app with the agent

java \
  -javaagent:/path/opentelemetry-javaagent.jar \
  -Dotel.service.name=demo-svc \
  -Dotel.traces.exporter=otlp \
  -Dotel.metrics.exporter=none \
  -Dotel.logs.exporter=none \
  -Dotel.exporter.otlp.endpoint=http://TEMPO_HOST:4318 \
  -Dotel.exporter.otlp.protocol=http/protobuf \
  -Dotel.traces.sampler=parentbased_traceidratio \
  -Dotel.traces.sampler.arg=1.0 \
  -jar app.jar

2) Hit any endpoint

You should see traces in Tempo with proper spans/durations.

Use 4318 + http/protobuf first because it’s the least finicky (no gRPC/TLS surprises).

If you prefer gRPC (4317):

-Dotel.exporter.otlp.endpoint=http://TEMPO_HOST:4317 \
-Dotel.exporter.otlp.protocol=grpc


⸻

Option B: Minimal manual tracing (no agent), pure OTLP export

Useful if you’re not on Java/Spring or you want a tiny proof-of-life program.

Java (single file-ish example)

Maven deps (key ones):
	•	io.opentelemetry:opentelemetry-sdk
	•	io.opentelemetry:opentelemetry-exporter-otlp
	•	io.opentelemetry:opentelemetry-semconv (optional)

Core code:

import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.context.Scope;
import io.opentelemetry.exporter.otlp.trace.OtlpGrpcSpanExporter;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.sdk.trace.SdkTracerProvider;
import io.opentelemetry.sdk.trace.export.BatchSpanProcessor;

import static io.opentelemetry.semconv.ResourceAttributes.SERVICE_NAME;

public class DemoTrace {
  static OpenTelemetry init() {
    var exporter = OtlpGrpcSpanExporter.builder()
        .setEndpoint("http://TEMPO_HOST:4317") // or your tempo container DNS
        .build();

    var tracerProvider = SdkTracerProvider.builder()
        .setResource(Resource.getDefault().toBuilder()
            .put(SERVICE_NAME, "demo-svc")
            .build())
        .addSpanProcessor(BatchSpanProcessor.builder(exporter).build())
        .build();

    return OpenTelemetrySdk.builder()
        .setTracerProvider(tracerProvider)
        .build();
  }

  public static void main(String[] args) throws Exception {
    OpenTelemetry otel = init();
    Tracer tracer = otel.getTracer("demo");

    Span parent = tracer.spanBuilder("parent-work").startSpan();
    try (Scope s = parent.makeCurrent()) {
      Span child = tracer.spanBuilder("child-sleep").startSpan();
      try { Thread.sleep(120); } finally { child.end(); }
    } finally {
      parent.end();
    }

    // give batch processor a moment to flush in tiny demos
    Thread.sleep(500);
    System.out.println("done");
  }
}

If you want OTLP/HTTP instead of gRPC, use the HTTP exporter and endpoint http://TEMPO_HOST:4318.

⸻

The “gotchas” that will make it look like it doesn’t work
	1.	Tempo must have OTLP receiver enabled and listening on 0.0.0.0 (inside Docker, localhost binding breaks ingestion from other containers).
	2.	Endpoint must be reachable from the app runtime
	•	If app runs in docker: use http://tempo:4317 (service name), not localhost.
	•	If app runs on host: use http://127.0.0.1:4317 only if ports are published.
	3.	Sampling: set to 1.0 for demos or you’ll “randomly” miss traces.

⸻

If you tell me (1) Spring Boot or not and (2) app runs on host or in Docker, I’ll give you the exact one-liner endpoint you should use (localhost vs tempo DNS) and the simplest working command.
Yes — if you don’t want the Java agent JAR, Go is a good choice for a minimal, explicit tracing demo. The trade-off is you write a bit of code, but it’s predictable and small.

Decision
	•	Fastest + least code changes (Java): keep the agent JAR.
	•	Smallest “clean demo” without any agent binary: Go + OTel SDK (or Python).
	•	Most realistic multi-service trace demo: Go is excellent (simple HTTP + explicit spans).

Your current config: 2 issues to fix (even if you keep Java)
	1.	service.name should not be set via otel.resource.attributes=service.name=... (it works, but better is):

	•	-Dotel.service.name=CHT_aiops

	2.	Port 14317 is nonstandard; make sure Tempo is actually listening there for OTLP gRPC (default is 4317). If your Tempo is behind a gateway/port-mapping, fine; otherwise you’ll “export successfully” but see nothing.

Also remove the trailing space here (it can bite in scripts):
	•	-Dotel.traces.exporter=otlp   ← trailing space

⸻

Minimal Go demo (no agent jar)

This is a single binary that:
	•	starts an HTTP server
	•	creates a trace + child spans per request
	•	exports via OTLP → Tempo (gRPC or HTTP)

1) main.go

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func newTracerProvider(ctx context.Context, tempoEndpoint, serviceName string) (*sdktrace.TracerProvider, error) {
	// gRPC exporter (Tempo OTLP gRPC)
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(tempoEndpoint), // e.g. "192.168.4.138:4317"
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			attribute.String("demo", "go-otlp"),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(1.0))), // 100% for demo
		sdktrace.WithBatcher(exp),
	)
	return tp, nil
}

func main() {
	ctx := context.Background()

	tempo := getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "192.168.4.138:4317") // gRPC host:port
	svc := getenv("OTEL_SERVICE_NAME", "CHT_aiops_go_demo")

	tp, err := newTracerProvider(ctx, tempo, svc)
	if err != nil {
		log.Fatalf("otel init: %v", err)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tr := otel.Tracer("demo")

	http.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tr.Start(r.Context(), "GET /work")
		defer span.End()

		_, child := tr.Start(ctx, "child-sleep")
		time.Sleep(120 * time.Millisecond)
		child.End()

		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok\n"))
	})

	log.Printf("listening on :8080, exporting to %s", tempo)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

2) Run it

export OTEL_EXPORTER_OTLP_ENDPOINT="192.168.4.138:14317"   # use your port if that's correct
export OTEL_SERVICE_NAME="CHT_aiops_go_demo"

go run .   # or go run main.go
curl -s http://localhost:8080/work

If Tempo actually expects default OTLP gRPC, change to:

export OTEL_EXPORTER_OTLP_ENDPOINT="192.168.4.138:4317"


⸻

When Go is “better” than Java-agent

Pros
	•	No attached agent binary, no JVM args games, no classloader issues (WildFly can be painful).
	•	Very small, deterministic demo.
	•	Easier to build multi-service propagation demos.

Cons
	•	You must write spans/instrumentation (unless you add middleware).
	•	Won’t mirror your real Java/WildFly stack behavior (JDBC spans etc.) unless you add equivalents.

⸻

If you want “minimal AND realistic” for your Java stack without agent

You can do Java “manual SDK” too, but it’s usually more code and more error-prone than the agent. For WildFly specifically, the agent is still the most practical way to get server/JDBC spans quickly.

⸻

If you tell me whether your Tempo endpoint 192.168.4.138:14317 is OTLP gRPC or something like collector/gateway, I’ll adjust the Go exporter config (gRPC vs HTTP, path/headers if needed) in one shot.
Use something explicit and disposable — this is a trace-export proof project, not an app.

Recommended

tempo-otlp-trace-demo

Good alternatives (pick by intent)
	•	Language-explicit
	•	tempo-otlp-go-demo
	•	tempo-trace-export-go
	•	Function-explicit
	•	tempo-trace-longest-span-demo
	•	tempo-trace-analysis-input
	•	Infra-oriented
	•	otel-to-tempo-demo
	•	otel-trace-emitter

Suggested minimal directory layout (Go)

tempo-otlp-trace-demo/
├── go.mod
├── main.go
└── README.md   # how to run + endpoint notes

If this project will later be reused as a test trace generator for your “longest span” service, the best name is:

tempo-trace-longest-span-demo

That makes its purpose obvious and future-proof.
