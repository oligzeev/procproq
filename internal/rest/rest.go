package rest

import (
	"bytes"
	"context"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"net/http"
)

const (
	ParamId        = "id"
	ParamProcessId = "process_id"
)

type Error struct {
	Ops      []domain.ErrOp `json:"ops"`
	Messages []string       `json:"messages"`
}

func E(err error) *Error {
	return &Error{
		Ops:      domain.EOps(err),
		Messages: domain.EMsgs(err),
	}
}

// opentracing.GlobalTracer() have to be initialized
func Send(ctx context.Context, client *retryablehttp.Client, url, method string, msgBytes []byte) (*http.Response, error) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, method+" "+url)
	defer span.Finish()

	request, err := retryablehttp.NewRequest(method, url, bytes.NewBuffer(msgBytes))
	if err != nil {
		return nil, errors.Wrapf(err, "can't create http request (%s %s)", method, url)
	}
	request.WithContext(spanCtx)

	tracer := opentracing.GlobalTracer()
	err = tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(request.Header))
	if err != nil {
		return nil, errors.Wrapf(err, "can't propagate tracing context (%s %s)", method, url)
	}

	request.Header.Set(domain.HeaderContentType, domain.ContentTypeApplicationJson)
	response, err := client.Do(request)
	if err != nil {
		return nil, errors.Wrapf(err, "can't send http request (%s %s)", method, url)
	}
	return response, nil
}
