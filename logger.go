package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/errorreporting"
	"go.opentelemetry.io/otel/trace"
)

const (
	DEFAULT = iota
	DEBUG
	INFO
	NOTICE
	WARNING
	ERROR
	CRITICAL
	ALERT
	EMERGENCY
)

var (
	loglevel = Default(os.Getenv("LOG_LEVEL"), "0")
	Factory  = &LoggerFactory{}
)

func MustSetup(projectID, serviceName, revision string) func() error {
	return Must(Setup(projectID, serviceName, revision))
}

func Setup(projectID, serviceName, revision string) (func() error, error) {
	f, err := NewLoggerFactory(context.Background(), projectID, serviceName, revision)
	if err != nil {
		return nil, fmt.Errorf("new logger factory: %v", err)
	}

	Factory = f
	return f.Close, nil
}

func New(req *http.Request, traceID, spanID string) *Logger {
	return Factory.New(req, traceID, spanID)
}

type Logger struct {
	level   int
	traceID string
	spanID  string
	errC    *errorreporting.Client
	req     *http.Request
}

type LogEntry struct {
	Severity string    `json:"severity"`
	Message  string    `json:"message"`
	Time     time.Time `json:"time"`
	TraceID  string    `json:"logging.googleapis.com/trace"`
	SpanID   string    `json:"logging.googleapis.com/spanId,omitempty"`
}

func (l *Logger) Log(severity, format string, a ...interface{}) {
	if err := json.NewEncoder(os.Stdout).Encode(&LogEntry{
		Severity: severity,
		Time:     time.Now(),
		Message:  fmt.Sprintf(format, a...),
		TraceID:  l.traceID,
		SpanID:   l.spanID,
	}); err != nil {
		log.Printf("encode log entry: %v", err)
	}
}

func (l *Logger) Report(a ...interface{}) {
	for _, aa := range a {
		switch err := aa.(type) {
		case error:
			l.errC.Report(errorreporting.Entry{
				Error: err,
				Req:   l.req,
			})
		}
	}
}

func (l *Logger) Debug(format string, a ...interface{}) {
	if l.level > DEBUG {
		return
	}

	l.Log("Debug", format, a...)
}

func (l *Logger) Info(format string, a ...interface{}) {
	if l.level > INFO {
		return
	}

	l.Log("Info", format, a...)
}

func (l *Logger) Error(format string, a ...interface{}) {
	if l.level > ERROR {
		return
	}

	l.Log("Error", format, a...)
}

func (l *Logger) ErrorReport(format string, a ...interface{}) {
	if l.level > ERROR {
		return
	}

	l.Error(format, a...)
	l.Report(a...)
}

func (l *Logger) Span(span trace.Span) *Logger {
	return &Logger{
		level:   l.level,
		traceID: l.traceID,
		spanID:  span.SpanContext().SpanID().String(),
		errC:    l.errC,
		req:     l.req,
	}
}

func Must(f func() error, err error) func() error {
	if err != nil {
		panic(err)
	}

	return f
}

func Default(v, w string) string {
	if v != "" {
		return v
	}

	return w
}
