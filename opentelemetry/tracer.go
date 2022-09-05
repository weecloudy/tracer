package opentelemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

// Option is tracing option.
type Option func(*options)

type options struct {
	tracerProvider trace.TracerProvider
	propagator     propagation.TextMapPropagator
}

// WithPropagator with tracer propagator.
func WithPropagator(propagator propagation.TextMapPropagator) Option {
	return func(opts *options) {
		opts.propagator = propagator
	}
}

// WithTracerProvider with tracer provider.
// Deprecated: use otel.SetTracerProvider(provider) instead.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return func(opts *options) {
		opts.tracerProvider = provider
	}
}

// Tracer is otel span tracer
type Tracer struct {
	tracer trace.Tracer
	kind   trace.SpanKind
	opt    *options
}

// NewTracer create tracer instance
func NewTracer(kind trace.SpanKind, opts ...Option) *Tracer {
	op := options{
		propagator: propagation.NewCompositeTextMapPropagator(Metadata{}, propagation.Baggage{}, propagation.TraceContext{}),
	}
	for _, o := range opts {
		o(&op)
	}

	// op.tracerProvider is nil, use default global.TracerProvider
	if op.tracerProvider != nil {
		otel.SetTracerProvider(op.tracerProvider)
	}

	switch kind {
	case trace.SpanKindClient:
		return &Tracer{tracer: otel.Tracer("weecloudy-tracer", trace.WithInstrumentationVersion(SemVersion())), kind: kind, opt: &op}
	case trace.SpanKindServer:
		return &Tracer{tracer: otel.Tracer("weecloudy-tracer", trace.WithInstrumentationVersion(SemVersion())), kind: kind, opt: &op}
	default:
		panic(fmt.Sprintf("unsupported span kind: %v", kind))
	}
}

// Start start tracing span
func (t *Tracer) Start(ctx context.Context, spanName string, carrier propagation.TextMapCarrier, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if t.kind == trace.SpanKindServer {
		ctx = t.opt.propagator.Extract(ctx, carrier)
	}
	ctx, span := t.tracer.Start(ctx,
		spanName,
		append(opts, trace.WithSpanKind(t.kind))...,
	)
	if t.kind == trace.SpanKindClient {
		t.opt.propagator.Inject(ctx, carrier)
	}

	return ctx, span
}

// End finish tracing span
func (t *Tracer) End(ctx context.Context, span trace.Span, m interface{}, err error, kv ...attribute.KeyValue) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "OK")
	}

	span.SetAttributes(kv...)

	// for pb
	if p, ok := m.(proto.Message); ok {
		if t.kind == trace.SpanKindServer {
			span.SetAttributes(attribute.Key("send_msg.size").Int(proto.Size(p)))
		} else {
			span.SetAttributes(attribute.Key("recv_msg.size").Int(proto.Size(p)))
		}
	}

	span.End()
}
