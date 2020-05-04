package main

import (
	"context"
	"example.com/oligzeev/pp-gin/internal/cache"
	appconf "example.com/oligzeev/pp-gin/internal/config"
	"example.com/oligzeev/pp-gin/internal/database"
	"example.com/oligzeev/pp-gin/internal/domain"
	"example.com/oligzeev/pp-gin/internal/logging"
	"example.com/oligzeev/pp-gin/internal/metric"
	"example.com/oligzeev/pp-gin/internal/rest"
	"example.com/oligzeev/pp-gin/internal/service"
	"example.com/oligzeev/pp-gin/internal/tracing"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/uber/jaeger-client-go"
	jaegerconf "github.com/uber/jaeger-client-go/config"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/pprof"

	_ "example.com/oligzeev/pp-gin/api/swagger"
)

// @title PP Gin
// @version 0.0.1
// @description This is a PP-Gin application.
func main() {
	cfg := initConfig("config/pp-gin.yaml", "pp")
	initLogger(cfg.Logging)
	db := initDatabase(cfg.DB)

	_, closer := initTracing(cfg.Tracing)
	defer closer.Close()

	readMappingRepo := database.NewReadMappingRepo(db)
	processRepo := database.NewProcessRepo(db)
	jobRepo := database.NewJobRepo(db)
	orderRepo := database.NewOrderRepo(db)

	readMappingService := NewReadMappingService(cfg.Cache, readMappingRepo)
	processService := NewProcessService(cfg.Cache, db, processRepo)
	orderService := NewOrderService(cfg.Cache, db, processService, orderRepo, jobRepo)

	initScheduler(cfg.Scheduler, jobRepo, orderService, readMappingService)

	router := initRouter(*cfg, []domain.RestHandler{
		rest.NewMappingRestHandler(readMappingService),
		rest.NewProcessRestHandler(processService),
		rest.NewJobRestHandler(orderService),
		rest.NewOrderRestHandler(orderService),
	})
	initServer(cfg.Rest, router)
}

// *****************************
// *** Initialize components ***
// *****************************

func initScheduler(cfg appconf.SchedulerConfig, jobRepo *database.JobRepo, orderService domain.OrderService,
	readMappingService domain.ReadMappingService) {
	if cfg.Enabled {
		scheduler := service.NewJobScheduler(cfg, jobRepo, orderService, readMappingService)
		scheduler.Start()
	}
}

func initConfig(yamlFileName, envPrefix string) *appconf.ApplicationConfig {
	appConfig, err := appconf.ReadConfig(yamlFileName, envPrefix)
	if err != nil {
		log.Fatal(err)
	}
	return appConfig
}

func initLogger(cfg appconf.LoggingConfig) {
	if cfg.Default {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
	} else {
		log.SetFormatter(&logging.TextFormatter{
			TimestampFormat: cfg.TimestampFormat,
		})
	}
	log.SetOutput(os.Stdout)
	log.SetLevel(log.Level(cfg.Level))
}

func initTracing(cfg appconf.TracingConfig) (opentracing.Tracer, io.Closer) {
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

func initDatabase(cfg appconf.DbConfig) *sqlx.DB {
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func initRouter(cfg appconf.ApplicationConfig, handlers []domain.RestHandler) *gin.Engine {
	router := gin.New()

	// Logging & Recovery middleware
	if cfg.Logging.Default {
		router.Use(gin.Logger())
	} else if cfg.Logging.Level >= 5 {
		logging.GinLogTimestampFormat = cfg.Logging.TimestampFormat
		gin.DebugPrintRouteFunc = logging.DebugPrintRouteFunc
		router.Use(gin.LoggerWithFormatter(logging.GinLogFormatter))
	}
	router.Use(gin.Recovery())

	// Jaeger middleware initialization
	router.Use(tracing.Middleware())

	// Swagger handler initialization
	// From the root directory: swag init --dir ./ --generalInfo ./cmd/pp-gin/main.go --output ./api/swagger
	router.GET(cfg.Rest.SwaggerUrl+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Prometheus handler initialization
	// prom := ginprom.NewPrometheus("gin") and prom.Use(router)
	router.GET(cfg.Rest.MetricsUrl, metric.PrometheusHandler())

	// PProf handler initialization
	// https://github.com/gin-contrib/pprof
	// go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
	pprof.Register(router)

	for _, handler := range handlers {
		handler.Register(router)
	}
	return router
}

// E.g. https://github.com/gin-gonic/examples/blob/master/graceful-shutdown/graceful-shutdown/server.go
func initServer(cfg appconf.RestConfig, r *gin.Engine) {
	srv := &http.Server{
		Addr:    cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Closing rest server")

	// TBD Close scheduler gracefully

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // TBD Configurable timeout
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Rest server forced to Close: %v", err)
	}
	log.Info("Rest server has been closed")
}

// ***************************
// *** Create repositories ***
// ***************************

func NewReadMappingService(cfg appconf.CacheConfig, repo *database.ReadMappingRepo) domain.ReadMappingService {
	service, err := cache.NewCachedReadMappingService(cfg.DefaultEntityCount, service.NewReadMappingService(repo))
	if err != nil {
		log.Fatal(err)
	}
	return tracing.NewSpanReadMappingService(service)
}

func NewProcessService(cfg appconf.CacheConfig, db *sqlx.DB, repo *database.ProcessRepo) domain.ProcessService {
	service, err := cache.NewCachedProcessRepo(cfg.DefaultEntityCount, service.NewProcessService(db, repo))
	if err != nil {
		log.Fatal(err)
	}
	return tracing.NewSpanProcessService(service)
}

func NewOrderService(cfg appconf.CacheConfig, db *sqlx.DB, processService domain.ProcessService,
	orderRepo *database.OrderRepo, jobRepo *database.JobRepo) domain.OrderService {

	service := service.NewOrderService(db, processService, orderRepo, jobRepo)
	cached, err := cache.NewCachedOrderService(cfg.DefaultEntityCount, service)
	if err != nil {
		log.Fatal(err)
	}
	return tracing.NewSpanOrderService(cached)
}
