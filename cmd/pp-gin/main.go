package main

import (
	"example.com/oligzeev/pp-gin/internal/cache"
	config2 "example.com/oligzeev/pp-gin/internal/config"
	"example.com/oligzeev/pp-gin/internal/database"
	"example.com/oligzeev/pp-gin/internal/domain"
	"example.com/oligzeev/pp-gin/internal/metric"
	"example.com/oligzeev/pp-gin/internal/rest"
	"example.com/oligzeev/pp-gin/internal/service"
	"example.com/oligzeev/pp-gin/internal/tracing"
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/uber/jaeger-client-go"
	jaegerconf "github.com/uber/jaeger-client-go/config"
	"io"
	"strconv"

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

	router := initRouter(cfg.Rest, []domain.RestHandler{
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

func initScheduler(cfg config2.SchedulerConfig, jobRepo *database.JobRepo, orderService domain.OrderService,
	readMappingService domain.ReadMappingService) {
	if cfg.Enabled {
		scheduler := service.NewJobScheduler(cfg, jobRepo, orderService, readMappingService)
		scheduler.Start()
	}
}

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

func initDatabase(cfg config2.DbConfig) *sqlx.DB {
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func initRouter(cfg config2.RestConfig, handlers []domain.RestHandler) *gin.Engine {
	router := gin.Default()

	// Jaeger middleware initialization
	router.Use(tracing.Middleware())

	// Swagger handler initialization
	// From the root directory: swag init --dir ./ --generalInfo ./cmd/pp-gin/main.go --output ./api/swagger
	router.GET(cfg.SwaggerUrl+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Prometheus handler initialization
	// prom := ginprom.NewPrometheus("gin") and prom.Use(router)
	router.GET(cfg.MetricsUrl, metric.PrometheusHandler())

	// PProf handler initialization
	// https://github.com/gin-contrib/pprof
	// go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
	pprof.Register(router)

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

// ***************************
// *** Create repositories ***
// ***************************

func NewReadMappingService(cfg config2.CacheConfig, repo *database.ReadMappingRepo) domain.ReadMappingService {
	service, err := cache.NewCachedReadMappingService(cfg.DefaultEntityCount, service.NewReadMappingService(repo))
	if err != nil {
		log.Fatal(err)
	}
	return tracing.NewSpanReadMappingService(service)
}

func NewProcessService(cfg config2.CacheConfig, db *sqlx.DB, repo *database.ProcessRepo) domain.ProcessService {
	service, err := cache.NewCachedProcessRepo(cfg.DefaultEntityCount, service.NewProcessService(db, repo))
	if err != nil {
		log.Fatal(err)
	}
	return tracing.NewSpanProcessService(service)
}

func NewOrderService(cfg config2.CacheConfig, db *sqlx.DB, processService domain.ProcessService,
	orderRepo *database.OrderRepo, jobRepo *database.JobRepo) domain.OrderService {

	service := service.NewOrderService(db, processService, orderRepo, jobRepo)
	cached, err := cache.NewCachedOrderService(cfg.DefaultEntityCount, service)
	if err != nil {
		log.Fatal(err)
	}
	return tracing.NewSpanOrderService(cached)
}
