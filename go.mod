module tracer

go 1.16

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/gin-contrib/logger v0.2.2
	github.com/gin-gonic/gin v1.8.1
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.20.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/uber/jaeger-client-go v2.30.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/weecloudy/common v0.0.0-20220905093133-336a6d03bc89
	github.com/weecloudy/logger v0.1.1-0.20220905093436-6bf18dc0df88
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.34.0
	go.opentelemetry.io/otel v1.9.0
	go.opentelemetry.io/otel/exporters/jaeger v1.9.0
	go.opentelemetry.io/otel/sdk v1.9.0
	go.opentelemetry.io/otel/trace v1.9.0
	google.golang.org/protobuf v1.28.0
)

replace github.com/weecloudy/common => ../common
