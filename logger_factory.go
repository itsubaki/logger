package logger

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"cloud.google.com/go/errorreporting"
)

var (
	serviceName = os.Getenv("K_SERVICE")  // https://cloud.google.com/run/docs/container-contract?hl=ja#services-env-vars
	revision    = os.Getenv("K_REVISION") // https://cloud.google.com/run/docs/container-contract?hl=ja#services-env-vars
	loglevel    = LogLevel(os.Getenv("LOG_LEVEL"), "0")
)

type LoggerFactory struct {
	level     int
	projectID string
	errC      *errorreporting.Client
}

func Must(f *LoggerFactory, err error) *LoggerFactory {
	if err != nil {
		panic(err)
	}

	return f
}

func LogLevel(v, w string) string {
	if v == "" {
		return w
	}

	return v
}

func NewLoggerFactory(ctx context.Context, projectID string) (*LoggerFactory, error) {
	c, err := errorreporting.NewClient(ctx, projectID, errorreporting.Config{
		ServiceName:    serviceName,
		ServiceVersion: revision,
	})
	if err != nil {
		return nil, fmt.Errorf("new error reporting client: %v", err)
	}

	l, err := strconv.Atoi(loglevel)
	if err != nil {
		return nil, fmt.Errorf("invalid log level=%v: %v", loglevel, err)
	}

	return &LoggerFactory{
		level:     l,
		projectID: projectID,
		errC:      c,
	}, nil
}

func (f *LoggerFactory) New(req *http.Request, traceID, spanID string) *Logger {
	trace := ""
	if len(traceID) > 0 {
		trace = fmt.Sprintf("projects/%v/traces/%v", f.projectID, traceID)
	}

	return &Logger{
		level:   f.level,
		errC:    f.errC,
		traceID: trace,
		spanID:  spanID,
		req:     req,
	}
}

func (f *LoggerFactory) Close() error {
	f.errC.Flush()
	if err := f.errC.Close(); err != nil {
		return fmt.Errorf("close errorreporing client: %v", err)
	}

	return nil
}
