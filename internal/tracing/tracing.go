package tracing

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
)

/*
	https://github.com/jaegertracing/jaeger-client-go
	https://github.com/jaegertracing/jaeger-client-go/blob/master/config/example_test.go
	https://github.com/opentracing/opentracing-go
	https://github.com/opentracing-contrib/go-gin/blob/master/examples/example_test.go
	https://github.com/opentracing-contrib/go-gin/blob/master/ginhttp/server.go
	TBD  r.Use() can use middleware to start spans. Example https://github.com/gin-contrib/opengintracing
	var parentSpanRefFunc = func(sc opentracing.SpanContext) opentracing.StartSpanOption {
		return opentracing.ChildOf(sc)
	}
	func(c *gin.Context) {
		spanContext, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		var span opentracing.Span
		if err != nil {
			span = tracer.StartSpan(c.Request.RequestURI, ext.RPCServerOption(spanContext))
		} else {
			span = tracer.StartSpan(c.Request.RequestURI)
		}
		defer span.Finish()
		c.Request = c.Request.WithContext(opentracing.ContextWithSpan(c.Request.Context(), span))

		c.Next()
	},
	spantracing.NewSpan(tracer, "yyy"),
	spantracing.SpanFromHeaders(tracer, "health", parentSpanRefFunc, true),
	spantracing.InjectToHeaders(tracer, true),
*/
func Middleware() gin.HandlerFunc {
	tracer := opentracing.GlobalTracer()
	return func(c *gin.Context) {
		spanContext, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		var span opentracing.Span
		if err != nil {
			span = tracer.StartSpan(c.Request.RequestURI)
		} else {
			span = tracer.StartSpan(c.Request.Method+" "+c.Request.RequestURI, ext.RPCServerOption(spanContext))
		}
		defer span.Finish()
		c.Request = c.Request.WithContext(opentracing.ContextWithSpan(c.Request.Context(), span))
		c.Next()
	}
}

func StartContextFromSpanStr(ctx context.Context, operationName string, spanStr string) (opentracing.Span, context.Context, error) {
	spanCtx, err := jaeger.ContextFromString(spanStr)
	if err != nil {
		return nil, nil, err
	}
	span := opentracing.GlobalTracer().StartSpan(operationName, ext.RPCServerOption(spanCtx))
	return span, opentracing.ContextWithSpan(ctx, span), nil
}

func SpanStrFromContext(ctx context.Context) (string, error) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		return fmt.Sprint(span.Context()), nil
	}
	return "", errors.New("can't get span from context")
}

// First span := opentracing.SpanFromContext(ctx), then FollowNewSpanFromContext in goroutine
func FollowNewSpanFromContext(span opentracing.Span, operationName string) (opentracing.Span, context.Context) {
	spanCtx := opentracing.ContextWithSpan(context.Background(), span)
	return FollowSpanFromContext(spanCtx, operationName)
}

func FollowSpanFromContext(ctx context.Context, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		opts = append(opts, opentracing.FollowsFrom(parentSpan.Context()))
	}
	span := opentracing.GlobalTracer().StartSpan(operationName, opts...)
	return span, opentracing.ContextWithSpan(ctx, span)
}
