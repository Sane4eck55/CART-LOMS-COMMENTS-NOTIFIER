// Package logger ...
package logger

import (
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Data ...
type Data struct {
	traceID string
	spanID  string
}

var (
	serviceName  = ""
	logger       *zap.SugaredLogger
	defaultLevel = zap.NewAtomicLevelAt(zap.InfoLevel)

	once = new(sync.Once)
)

func initLogger(level zap.AtomicLevel) {
	once.Do(func() {
		config := zap.NewProductionConfig()
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stdout"}
		config.Level.SetLevel(level.Level())
		//в выводе в коллере будет logger/logger.go, если добавить zap.AddCallerSkip(1) то поднимаемся на шаг выше по стеку, и будет, например, service/service.go
		l, err := config.Build(zap.AddCallerSkip(1))
		if err != nil {
			logger.Fatalw(fmt.Sprintf("config.Build logger : %v", err))
		}
		logger = l.Sugar()
	})
}

// GetLogger ...
func GetLogger(srvName string) *zap.SugaredLogger {
	serviceName = srvName
	if logger == nil {
		initLogger(defaultLevel)
	}

	return logger
}

// Fatalw ...
func Fatalw(msg string, keysAndValues ...interface{}) {
	data := extractData(keysAndValues...)
	GetLogger(serviceName).With(
		zap.String("service", serviceName),
		zap.String("traceID", data.traceID),
		zap.String("spanID", data.spanID),
	).Fatalf(msg)
}

// Errorw ...
func Errorw(msg string, keysAndValues ...interface{}) {
	data := extractData(keysAndValues...)
	GetLogger(serviceName).With(
		zap.String("service", serviceName),
		zap.String("traceID", data.traceID),
		zap.String("spanID", data.spanID),
	).Errorf(msg)
}

// Infow ...
func Infow(msg string, keysAndValues ...interface{}) {
	data := extractData(keysAndValues...)
	GetLogger(serviceName).With(
		zap.String("service", serviceName),
		zap.String("traceID", data.traceID),
		zap.String("spanID", data.spanID),
	).Infof(msg)
}

// Sync в случае grasefull shutdown допишутся все логики
func Sync() error {
	return GetLogger(serviceName).Sync()
}

// extractSpan ..
func extractData(keysAndValues ...interface{}) Data {
	data := Data{}
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}

		if key == "span" {
			if span, ok := keysAndValues[i+1].(trace.Span); ok {
				data.traceID = span.SpanContext().TraceID().String()
				data.spanID = span.SpanContext().SpanID().String()
			}
		}
	}

	return data
}
