package opentracing

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func Tracer() gin.HandlerFunc {
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
		ctx.Request = ctx.Request.WithContext(opentracing.ContextWithSpan(ctx.Request.Context(), startSpan))
		context.WithValue(ctx, parentSpanCtxKey, startSpan.Context())

		ctx.Next()

		// http response status
		status := ctx.Writer.Status()
		ext.HTTPStatusCode.Set(startSpan, uint16(status))

		startSpan.Finish()
	}
}
