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
	projectID   = os.Getenv("PROJECT_ID")
	serviceName = os.Getenv("K_SERVICE")  // https://cloud.google.com/run/docs/container-contract?hl=ja#services-env-vars
	revision    = os.Getenv("K_REVISION") // https://cloud.google.com/run/docs/container-contract?hl=ja#services-env-vars
	loglevel    = Default(os.Getenv("LOG_LEVEL"), "0")
	Factory     = Must(NewLoggerFactory(context.Background(), projectID, serviceName, revision))
)

func New(req *http.Request, traceID, spanID string) *Logger {
	return Factory.New(req, traceID, spanID)
}

func Default(v, w string) string {
	if v != "" {
		return v
	}

	return w
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
