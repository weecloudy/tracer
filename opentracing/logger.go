package opentracing

import (
	"fmt"
	"github.com/weecloudy/logger"
)

type openTracingLogger struct {
}

func (l *openTracingLogger) Error(msg string) {
	logger.NewZapLogger().Logger.Error(msg, []logger.Field{
		logger.Any("msg", msg),
		logger.Any(logger.LogTypeKey, "opentracing"),
	}...)
}

func (l *openTracingLogger) Infof(msg string, args ...interface{}) {
	logger.NewZapLogger().Logger.Info(msg, []logger.Field{
		logger.Any("msg", fmt.Sprintf(msg, args...)),
		logger.Any(logger.LogTypeKey, "opentracing"),
	}...)
}

func (l *openTracingLogger) Debugf(msg string, args ...interface{}) {
	logger.NewZapLogger().Logger.Debug(msg, []logger.Field{
		logger.Any("msg", fmt.Sprintf(msg, args...)),
		logger.Any(logger.LogTypeKey, "opentracing"),
	}...)
}
