package opentracing

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-client-go/zipkin"
)

//链路追踪实例
var openTracer opentracing.Tracer
var openTracerCloser io.Closer

const (
	envJaegerAgentHost  = "JAEGER_AGENT_HOST"  //环境变量中配置的jaeger host,不配置不开启
	envJaegerAgentPort  = "JAEGER_AGENT_PORT"  //环境变量中配置的jaeger port
	envTraceServiceName = "TRACE_SERVICE_NAME" //环境变量中配置的服务名
)

type Options struct {
	SamplerType      string           //采样类型:const|probabilistic|rateLimiting|remote
	SamplerParam     float64          //采样参数
	JaegerAgentHost  string           //jaeger host
	JaegerAgentPort  string           //jaeger port
	TraceServiceName string           //服务名
	Logger           jaegerlog.Logger //logger
}

type Option func(c *Options)

func SamplerType(samplerType string) Option {
	return func(c *Options) {
		c.SamplerType = samplerType
	}
}
func SamplerParam(samplerParam float64) Option {
	return func(c *Options) {
		c.SamplerParam = samplerParam
	}
}
func JaegerAgentHost(jaegerAgentHost string) Option {
	return func(c *Options) {
		c.JaegerAgentHost = jaegerAgentHost
	}
}
func JaegerAgentPort(jaegerAgentPort string) Option {
	return func(c *Options) {
		c.JaegerAgentPort = jaegerAgentPort
	}
}
func TraceServiceName(traceServiceName string) Option {
	return func(c *Options) {
		c.TraceServiceName = traceServiceName
	}
}
func Logger(logger jaegerlog.DebugLogger) Option {
	return func(c *Options) {
		c.Logger = logger
	}
}

func applyOptions(options ...Option) Options {
	opts := Options{
		SamplerType:      jaeger.SamplerTypeConst,
		SamplerParam:     1,
		JaegerAgentHost:  os.Getenv(envJaegerAgentHost),
		JaegerAgentPort:  os.Getenv(envJaegerAgentPort),
		TraceServiceName: os.Getenv(envTraceServiceName),
	}
	if opts.JaegerAgentPort == "" {
		opts.JaegerAgentPort = "6831"
	}
	if opts.TraceServiceName == "" {
		opts.TraceServiceName = "unknownService"
	}
	if opts.Logger == nil {
		opts.Logger = &openTracingLogger{}
	}
	for _, option := range options {
		option(&opts)
	}
	return opts
}

func InitOpenTracer(op ...Option) (err error) {
	if openTracer != nil {
		return nil
	}
	opts := applyOptions(op...)
	if opts.JaegerAgentHost == "" {
		opts.Logger.Infof("jaeger agent host is empty, opentraceing closed ...")
		return errors.New("jaeger agent host is empty, opentraceing closed")
	}
	agentHost := opts.JaegerAgentHost + ":" + opts.JaegerAgentPort
	var cfg = jaegercfg.Configuration{
		ServiceName: opts.TraceServiceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  opts.SamplerType,
			Param: opts.SamplerParam,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: agentHost,
		},
	}

	// Zipkin shares span ID between client and server spans; it must be enabled via the following option.
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	openTracer, openTracerCloser, err = cfg.NewTracer(
		jaegercfg.Logger(opts.Logger),
		jaegercfg.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.ZipkinSharedRPCSpan(false),
	)
	if err != nil {
		opts.Logger.Infof("openTracer create error %s", err.Error())
		return
	}

	return
}

func CloseOpenTracer() {
	if openTracerCloser != nil {
		openTracerCloser.Close()
		openTracer = nil
		openTracerCloser = nil
	}
}

// GetOpenTracer get open tracer
func GetOpenTracer() opentracing.Tracer {
	if err := InitOpenTracer(); err != nil {
		return nil
	}
	return openTracer
}

//InjectHTTPRequest inject http request for client
func InjectHTTPRequest(ctx context.Context, req *http.Request) {
	if err := InitOpenTracer(); err != nil {
		return
	}
	if span := opentracing.SpanFromContext(ctx); span != nil {
		err := openTracer.Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(req.Header))
		if err != nil {
			new(openTracingLogger).Error("openTracer inject err:" + err.Error())
		}
	}
}
