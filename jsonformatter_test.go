package glogger

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

type LogEntry struct {
	Level         string
	Message       string
	Time          int64
	CorrelationId string
	Http          HTTP
	Host          Host
}

func TestJsonFormatter(t *testing.T) {

	t.Run("Base Log Entry must be the same", func(t *testing.T) {
		now := time.Now()
		message := "Incoming Rquest"
		entry := logrus.Entry{
			Level:   logrus.InfoLevel,
			Time:    now,
			Message: message,
		}

		formatter := JSONFormatter{}

		data, err := formatter.Format(&entry)
		actualResult := string(data)

		expected := fmt.Sprintf("{\"level\":\"info\",\"message\":\"%s\",\"time\":%d}\n", message, now.Unix())

		assert.Assert(t, err == nil, "Error is nil")
		assert.Equal(t, actualResult, expected)
	})
}
