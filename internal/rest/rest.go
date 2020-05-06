package rest

import (
	"bytes"
	"context"
	"example.com/oligzeev/pp-gin/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
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

type Server struct {
	cfg        domain.ServerRestConfig
	httpServer *http.Server
	router     *gin.Engine
}

func NewServer(cfg domain.ServerRestConfig, handlers []domain.RestHandler) *Server {
	router := gin.New()
	for _, handler := range handlers {
		handler.Register(router)
	}
	httpServer := &http.Server{
		ReadTimeout:  cfg.ReadTimeoutSec * time.Second,
		WriteTimeout: cfg.WriteTimeoutSec * time.Second,
		Addr:         cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Handler:      router,
	}
	return &Server{cfg: cfg, httpServer: httpServer, router: router}
}

func (s Server) Router() *gin.Engine {
	return s.router
}

func (s Server) Start(ctx context.Context) error {
	const op = "RestServer.Start"

	log.Tracef("%s: %s", op, s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return domain.E(op, err)
	}
	log.Tracef("%s: exit", op)
	return ctx.Err()
}

func (s Server) Stop(ctx context.Context) error {
	const op = "RestServer.Stop"

	<-ctx.Done()
	log.Tracef("%s: in progress", op)

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeoutSec*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return domain.E(op, err)
	}
	log.Tracef("%s: finished", op)
	return ctx.Err()
}
