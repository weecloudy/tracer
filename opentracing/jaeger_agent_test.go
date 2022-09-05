package opentracing

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/weecloudy/logger"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLopentrace(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "opentracing Suite")
}

var StdLogger *stdLogger
var _ = BeforeSuite(func() {
	StdLogger = &stdLogger{}
})

type stdLogger struct {
}

func (l *stdLogger) Error(msg string) {
	logger.NewZapLogger().Logger.Error(msg, []logger.Field{
		logger.Any("msg", msg),
		logger.Any(logger.LogTypeKey, "opentracing"),
	}...)
}

func (l *stdLogger) Infof(msg string, args ...interface{}) {
	print("===============msg")
	logger.NewZapLogger().Logger.Info(msg, []logger.Field{
		logger.Any("msg", fmt.Sprintf(msg, args...)),
		logger.Any(logger.LogTypeKey, "opentracing"),
	}...)
}

func (l *stdLogger) Debugf(msg string, args ...interface{}) {
	logger.NewZapLogger().Logger.Debug(msg, []logger.Field{
		logger.Any("msg", fmt.Sprintf(msg, args...)),
		logger.Any(logger.LogTypeKey, "opentracing"),
	}...)
}

var _ = Describe("InitOpenTracer succeed", func() {
	It("InitOpenTracer succeed", func() {
		err := InitOpenTracer(
			JaegerAgentHost("127.0.0.1"),
			JaegerAgentPort("6831"),
			SamplerType("const"),
			SamplerParam(1),
			TraceServiceName("testService"),
		)
		Expect(err == nil).Should(BeTrue())
		Expect(openTracer != nil).Should(BeTrue())
		Expect(openTracerCloser != nil).Should(BeTrue())
		CloseOpenTracer()
	})

})

var _ = Describe("InitOpenTracer Fail", func() {
	It("InitOpenTracer Fail", func() {
		v := os.Getenv(envJaegerAgentHost)
		fmt.Printf("v: %v\n", v)
		err := InitOpenTracer(
			JaegerAgentHost(""),
			JaegerAgentPort("6831"),
			SamplerType("const"),
			SamplerParam(1),
			TraceServiceName("testService"),
			Logger(StdLogger),
		)
		Expect(err != nil).Should(BeTrue())
		Expect(openTracer == nil).Should(BeTrue())
		Expect(openTracerCloser == nil).Should(BeTrue())
	})

})

var _ = Describe("InjectHttpRequest Success", func() {

	It("InjectHttpRequest Success", func() {
		err := InitOpenTracer(
			JaegerAgentHost("127.0.0.1"),
			JaegerAgentPort("6831"),
			SamplerType("const"),
			SamplerParam(1),
			TraceServiceName("testService"),
		)
		Expect(err == nil).Should(BeTrue())
		Expect(openTracer != nil).Should(BeTrue())
		Expect(openTracerCloser != nil).Should(BeTrue())

		router := gin.New()
		router.Use(Tracing())
		router.Handle(http.MethodGet, "/", func(ctx *gin.Context) {
			InjectHTTPRequest(ContextWithSpanFromGinCtx(ctx), ctx.Request)
			ctx.JSON(http.StatusOK, map[string]interface{}{
				"code":   0,
				"msg":    "",
				"result": map[string]string{"get": "ok"},
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewReader([]byte("12123")))
		router.ServeHTTP(w, req)
		data, _ := ioutil.ReadAll(w.Body)
		res := string(data)
		Expect(res != "").Should(BeTrue())
		CloseOpenTracer()
	})
})
