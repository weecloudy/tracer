package opentelemetry

import (
	"bytes"
	"context"
	"go.opentelemetry.io/otel"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/propagation"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestOtel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "opentelemetry Suite")
}

var _ = BeforeSuite(func() {
})

var _ = Describe("NewTracer", func() {
	It("succeed", func() {
		tracer := NewTracer(trace.SpanKindServer)
		Expect(tracer != nil).Should(BeTrue())
	})
})

var _ = Describe("Gin Tracing md", func() {
	It("TracerProviderWithJaegerCollector succeed", func() {
		tp, err := TracerProviderWithJaegerCollector("opentelemetry-app-test")
		if err != nil {
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Cleanly shutdown and flush telemetry when the application exits.
		defer func(ctx context.Context) {
			// Do not make the application hang when it is shutdown.
			ctx, cancel = context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			if err := tp.Shutdown(ctx); err != nil {
				return
			}
		}(ctx)

		engine := gin.New()

		//engine.Use(Tracing("test"))
		engine.Use(otelgin.Middleware("serverTest"))
		engine.GET("/", homeHandler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewReader([]byte("12123")))
		engine.ServeHTTP(w, req)
		data, _ := ioutil.ReadAll(w.Body)
		res := string(data)
		Expect(res != "").Should(BeTrue())
	})

	It("TracerProviderWithJaegerAgent succeed", func() {
		tp, err := TracerProviderWithJaegerAgent("testJaegerAgent")
		if err != nil {
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Cleanly shutdown and flush telemetry when the application exits.
		defer func(ctx context.Context) {
			// Do not make the application hang when it is shutdown.
			ctx, cancel = context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			if err := tp.Shutdown(ctx); err != nil {
				return
			}
		}(ctx)

		engine := gin.New()

		engine.Use(Tracing("test"))
		engine.GET("/", homeHandler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewReader([]byte("12123")))
		engine.ServeHTTP(w, req)
		data, _ := ioutil.ReadAll(w.Body)
		res := string(data)
		Expect(res != "").Should(BeTrue())
	})
})

func homeHandler(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")

	c.String(200, "开始请求...\n")
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)

	syncReq, _ := http.NewRequest("GET", "http://localhost:8080/service", nil)
	tracer := NewTracer(trace.SpanKindClient)
	tracer.Start(ctx, "/service", propagation.HeaderCarrier(syncReq.Header), trace.WithSpanKind(trace.SpanKindClient))
	//otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(syncReq.Header))
	// if 404 not found don't to trace, no span
	_, err := http.DefaultClient.Do(syncReq)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("请求 /service error", err.Error()))
	}
	tracer.End(ctx, span, "", err)

	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

	go func() {
		//ctx, span = otel.Tracer("home").Start(ctx, "async", trace.WithSpanKind(trace.SpanKindClient))
		asyncReq, _ := http.NewRequest("GET", "http://localhost:8080/async", nil)
		tracer.Start(ctx, "/async", propagation.HeaderCarrier(asyncReq.Header), trace.WithSpanKind(trace.SpanKindClient))
		//otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(asyncReq.Header))
		_, err := http.DefaultClient.Do(asyncReq)
		if  err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("请求 /async error", err.Error()))
		}
		tracer.End(ctx, span, "", err)
	}()

	ctx, span = otel.Tracer("home").Start(ctx, "ping-baidu", trace.WithSpanKind(trace.SpanKindClient))
	bdReq, _ := http.NewRequest("GET", "https://www.baidu.com", nil)
	//tracer.Start(ctx, "baidu", propagation.HeaderCarrier(bdReq.Header), trace.WithSpanKind(trace.SpanKindClient))
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(bdReq.Header))
	if _, err := http.DefaultClient.Do(bdReq); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("ping baidu error", err.Error()))
	}
	//tracer.End(ctx, span, "", err)

	c.String(200, "请求结束！")
}
