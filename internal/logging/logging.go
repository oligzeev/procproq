package logging

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

// E.g.
// https://github.com/antonfisher/nested-logrus-formatter/blob/master/formatter.go
// https://github.com/x-cray/logrus-prefixed-formatter/blob/master/formatter.go
type TextFormatter struct {
	TimestampFormat string
}

func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	// Timestamp field
	b.WriteString("[timestamp=")
	b.WriteString(entry.Time.Format(f.TimestampFormat))
	b.WriteByte(']')

	// Level field
	b.WriteString("[level=")
	b.WriteString(strings.ToUpper(entry.Level.String()[:4]))
	b.WriteByte(']')

	// Custom fields
	if len(entry.Data) > 0 {
		for key, value := range entry.Data {
			b.WriteByte('[')
			b.WriteString(key)
			b.WriteByte('=')
			fmt.Fprint(b, value)
			b.WriteString("]")
		}
	}

	// Message field
	if entry.Message != "" {
		b.WriteString("[message=")
		b.WriteString(entry.Message)
		b.WriteString("]")
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}
