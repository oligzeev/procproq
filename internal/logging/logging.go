package logging

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strconv"
)

// E.g.
// https://github.com/antonfisher/nested-logrus-formatter/blob/master/formatter.go
// https://github.com/x-cray/logrus-prefixed-formatter/blob/master/formatter.go
type TextFormatter struct {
	TimestampFormat string
}

func (f *TextFormatter) Format(entry *log.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	printField(b, "timestamp", f.TimestampFormat)
	printField(b, "level", entry.Level.String()[:4])
	if len(entry.Data) > 0 {
		for key, value := range entry.Data {
			printInterfaceField(b, key, value)
		}
	}
	if entry.Message != "" {
		printField(b, "message", entry.Message)
	}
	b.WriteByte('\n')
	return b.Bytes(), nil
}

var GinLogTimestampFormat = "15.04.05 02.01.2006.000000000"

func GinLogFormatter(params gin.LogFormatterParams) string {
	b := &bytes.Buffer{}
	printField(b, "timestamp", params.TimeStamp.Format(GinLogTimestampFormat))
	printField(b, "level", "debu")
	printField(b, "method", params.Method)
	printField(b, "path", params.Path)
	printField(b, "statusCode", strconv.Itoa(params.StatusCode))
	printField(b, "bodySize", strconv.Itoa(params.BodySize))
	printField(b, "clientIP", params.ClientIP)
	//printField(b, "proto", params.Request.Proto)
	printField(b, "latency", params.Latency.String())
	//printField(b, "userAgent", params.Request.UserAgent()[:20])
	printField(b, "errorMessage", params.ErrorMessage)
	b.WriteByte('\n')
	return b.String()
}

func DebugPrintRouteFunc(httpMethod, absolutePath, handlerName string, nuHandlers int) {
	b := &bytes.Buffer{}
	b.WriteByte('\n')
	log.WithFields(log.Fields{
		"httpMethod":   httpMethod,
		"absolutePath": absolutePath,
		"handlerName":  handlerName,
		"nuHandlers":   nuHandlers,
	}).Debug()
}

func printField(b *bytes.Buffer, key, value string) {
	b.WriteByte('[')
	b.WriteString(key)
	b.WriteByte('=')
	b.WriteString(value)
	b.WriteString("]")
}
func printInterfaceField(b *bytes.Buffer, key string, value interface{}) {
	b.WriteByte('[')
	b.WriteString(key)
	b.WriteByte('=')
	fmt.Fprint(b, value)
	b.WriteString("]")
}
