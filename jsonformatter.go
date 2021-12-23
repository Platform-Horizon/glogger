package glogger

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

// JSONFormatter struct
type JSONFormatter struct{}

// Format function will set how to format entry in JSON
func (formatter *JSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data)+4)

	data["time"] = entry.Time.Unix()
	data["message"] = entry.Message
	data["level"] = entry.Level

	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	encoder := json.NewEncoder(b)

	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON: %v", err)
	}

	return b.Bytes(), nil
}
