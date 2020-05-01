package main

import (
	config2 "example.com/oligzeev/pp-gin/internal/config"
	"example.com/oligzeev/pp-gin/internal/domain"
	"example.com/oligzeev/pp-gin/internal/metric"
	"example.com/oligzeev/pp-gin/internal/rest"
	"example.com/oligzeev/pp-gin/internal/tracing"
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"github.com/uber/jaeger-client-go"
	jaegerconf "github.com/uber/jaeger-client-go/config"
	"io"
	"strconv"
)

func main() {
	cfg := initConfig("config/pp-stub.yaml", "pp")
	initLogger(cfg.Logging)

	_, closer := initTracing(cfg.Tracing)
	defer closer.Close()

	router := initRouter(cfg.Rest, []domain.RestHandler{
		rest.NewStubRestHandler(cfg.Stub),
	})
	initServer(cfg.Rest, router)
}

// *****************************
// *** Initialize components ***
// *****************************

func initConfig(yamlFileName, envPrefix string) *config2.ApplicationConfig {
	appConfig, err := config2.ReadConfig(yamlFileName, envPrefix)
	if err != nil {
		log.Fatal(err)
	}
	return appConfig
}

func initLogger(cfg config2.LoggingConfig) {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(log.Level(cfg.Level))
}

func initTracing(cfg config2.TracingConfig) (opentracing.Tracer, io.Closer) {
	tracingCfg := jaegerconf.Configuration{
		ServiceName: cfg.ServiceName,
		Sampler: &jaegerconf.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegerconf.ReporterConfig{
			LogSpans: true,
		},
	}
	tracer, closer, err := tracingCfg.NewTracer()
	if err != nil {
		log.Fatal(err)
	}
	opentracing.SetGlobalTracer(tracer)
	return tracer, closer
}

func initRouter(cfg config2.RestConfig, handlers []domain.RestHandler) *gin.Engine {
	router := gin.Default()

	// Jaeger middleware initialization
	router.Use(tracing.Middleware())

	// Prometheus handler initialization
	router.GET(cfg.MetricsUrl, metric.PrometheusHandler())

	for _, handler := range handlers {
		handler.Register(router)
	}
	return router
}

func initServer(cfg config2.RestConfig, r *gin.Engine) {
	if err := endless.ListenAndServe(cfg.Host+":"+strconv.Itoa(cfg.Port), r); err != nil {
		log.Fatal(err)
	}
}
