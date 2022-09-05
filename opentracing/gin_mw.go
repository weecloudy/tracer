package opentracing

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func Tracing() gin.HandlerFunc {
	if err := InitOpenTracer(); err != nil {
		return func(ctx *gin.Context) {}
	}
	return func(ctx *gin.Context) {

		var startSpan opentracing.Span
		spanName := ctx.Request.URL.Path
		if spanCtx, err := openTracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(ctx.Request.Header)); err == nil {
			startSpan = openTracer.StartSpan(spanName, opentracing.ChildOf(spanCtx))
		} else {
			startSpan = openTracer.StartSpan(spanName)
		}

		ext.HTTPUrl.Set(startSpan, ctx.Request.URL.Path)
		ext.HTTPMethod.Set(startSpan, ctx.Request.Method)
		// pass the span through the request context
		ctx.Request = ctx.Request.WithContext(opentracing.ContextWithSpan(ctx.Request.Context(), startSpan))

		ctx.Next()

		// http response status
		status := ctx.Writer.Status()
		ext.HTTPStatusCode.Set(startSpan, uint16(status))

		startSpan.Finish()
	}
}

// ContextWithSpanFromGinCtx for InjectHTTPRequest span context
func ContextWithSpanFromGinCtx(c *gin.Context) context.Context {
	return c.Request.Context()
}
