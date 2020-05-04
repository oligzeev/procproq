package logging

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

// E.g.
// https://github.com/antonfisher/nested-logrus-formatter/blob/master/formatter.go
// https://github.com/x-cray/logrus-prefixed-formatter/blob/master/formatter.go
type TextFormatter struct {
	TimestampFormat string
}

// Logrus formatter
func (f *TextFormatter) Format(entry *log.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	//printField(b, "timestamp", entry.Time.Format(f.TimestampFormat))
	b.WriteString(entry.Time.Format(f.TimestampFormat))
	b.WriteByte(' ')

	//printField(b, "level", entry.Level.String()[:4])
	b.WriteString(strings.ToUpper(entry.Level.String()[:4]))
	b.WriteByte(' ')

	if len(entry.Data) > 0 {
		for key, value := range entry.Data {
			printInterfaceField(b, key, value)
		}
		b.WriteByte(' ')
	}
	if entry.Message != "" {
		//printField(b, "message", entry.Message)
		b.WriteString(entry.Message)
	}
	b.WriteByte('\n')
	return b.Bytes(), nil
}

// Gin log formatter
// Usage router.Use(gin.LoggerWithFormatter(logging.GinLogFormatter))
var GinLogTimestampFormat = "15.04.05 02.01.2006.000000000"

func GinLogFormatter(params gin.LogFormatterParams) string {
	b := &bytes.Buffer{}
	//printField(b, "timestamp", params.TimeStamp.Format(GinLogTimestampFormat))
	b.WriteString(params.TimeStamp.Format(GinLogTimestampFormat))
	b.WriteByte(' ')

	//printField(b, "level", "debu")
	b.WriteString("DEBU")
	b.WriteByte(' ')

	printField(b, "statusCode", strconv.Itoa(params.StatusCode))
	printField(b, "method", params.Method)
	printField(b, "path", params.Path)
	printField(b, "bodySize", strconv.Itoa(params.BodySize))
	printField(b, "clientIP", params.ClientIP)
	//printField(b, "proto", params.Request.Proto)
	printField(b, "latency", params.Latency.String())
	//printField(b, "userAgent", params.Request.UserAgent()[:20])
	printField(b, "errorMessage", params.ErrorMessage)
	b.WriteByte('\n')
	return b.String()
}

// Gin router debug format function
func DebugPrintRouteFunc(httpMethod, absolutePath, handlerName string, nuHandlers int) {
	/*log.WithFields(log.Fields{
		"httpMethod":   httpMethod,
		"absolutePath": absolutePath,
		"handlerName":  handlerName,
		"nuHandlers":   nuHandlers,
	}).Debug()*/
	log.Debugf("%s %s - %s (%d)", httpMethod, absolutePath, handlerName, nuHandlers)
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
