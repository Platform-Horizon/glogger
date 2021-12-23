package glogger

import (
	"context"

	"github.com/sirupsen/logrus"
)

type loggerKey struct{}

var defaultLogger *logrus.Entry = logrus.NewEntry(logrus.StandardLogger())

// WithLogger returns a new context with the provided logger
func WithLogger(ctx context.Context, logger *logrus.Entry) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// Get retrivies the current logger from the context. If no logger is availabe, the default logger is returned.
func Get(ctx context.Context) *logrus.Entry {
	logger := ctx.Value(loggerKey{})

	if logger == nil {
		return defaultLogger
	}

	entry, ok := logger.(*logrus.Entry)

	if !ok {
		return defaultLogger
	}

	return entry
}
