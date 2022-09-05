package opentelemetry

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// Middleware returns middleware that will trace incoming requests.
// The service parameter should describe the name of the (virtual)
// server handling the request.
func Tracing(service string, opts ...Option) gin.HandlerFunc {
	tracer := NewTracer(trace.SpanKindServer, opts...)

	return func(c *gin.Context) {
		savedCtx := c.Request.Context()
		defer func() {
			//end to use raw request ctx
			c.Request = c.Request.WithContext(savedCtx)
		}()

		opts := []trace.SpanStartOption{
			trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", c.Request)...),
			trace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(c.Request)...),
			trace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(service, c.FullPath(), c.Request)...),
		}

		spanName := c.FullPath()
		if spanName == "" {
			spanName = fmt.Sprintf("HTTP %s route not found", c.Request.Method)
		}

		ctx, span := tracer.Start(savedCtx, spanName, propagation.HeaderCarrier(c.Request.Header), opts...)

		// pass the span through the request context
		c.Request = c.Request.WithContext(ctx)

		// serve the request to the next middleware
		c.Next()

		status := c.Writer.Status()
		attrs := semconv.HTTPAttributesFromHTTPStatusCode(status)
		spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(status)
		span.SetAttributes(attrs...)
		span.SetStatus(spanStatus, spanMessage)
		var err error
		if len(c.Errors) > 0 {
			//span.SetAttributes(attribute.String("gin.errors", c.Errors.String()))
			err = fmt.Errorf("gin.errors:%s", c.Errors.String())
		}
		tracer.End(ctx, span, "", err)
	}
}

// ContextWithSpanFromGinCtx for Inject span context
func ContextWithSpanFromGinCtx(c *gin.Context) context.Context {
	return c.Request.Context()
}
