package main

import (
	"example.com/oligzeev/pp-gin/internal/cache"
	"example.com/oligzeev/pp-gin/internal/config"
	"example.com/oligzeev/pp-gin/internal/database"
	"example.com/oligzeev/pp-gin/internal/domain"
	"example.com/oligzeev/pp-gin/internal/metrics"
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

	_ "example.com/oligzeev/pp-gin/docs"
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

	readMappingRepo := NewReadMappingRepo(cfg.Cache, db)
	processRepo := NewProcessRepo(cfg.Cache, db)
	jobRepo := NewJobRepo(db)
	orderRepo := NewOrderRepo(cfg.Cache, db)

	processService := service.NewProcessService(db, processRepo, readMappingRepo)
	orderService := service.NewOrderService(db, processRepo, orderRepo, jobRepo)

	initScheduler(cfg.Scheduler, jobRepo, orderRepo, readMappingRepo)

	router := initRouter(cfg.Rest, []domain.RestHandler{
		rest.NewMappingRestHandler(processService),
		rest.NewProcessRestHandler(processService),
		rest.NewJobRestHandler(orderService),
		rest.NewOrderRestHandler(orderService),
	})
	initServer(cfg.Rest, router)
}

// *****************************
// *** Initialize components ***
// *****************************

func initScheduler(cfg config.SchedulerConfig, jobRepo domain.JobRepo, orderRepo domain.OrderRepo,
	readMappingRepo domain.ReadMappingRepo) {
	if cfg.Enabled {
		scheduler := service.NewJobScheduler(cfg, jobRepo, orderRepo, readMappingRepo)
		scheduler.Start()
	}
}

func initConfig(yamlFileName, envPrefix string) *config.ApplicationConfig {
	appConfig, err := config.ReadConfig(yamlFileName, envPrefix)
	if err != nil {
		log.Fatal(err)
	}
	return appConfig
}

func initLogger(cfg config.LoggingConfig) {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(log.Level(cfg.Level))
}

func initTracing(cfg config.TracingConfig) (opentracing.Tracer, io.Closer) {
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

func initDatabase(cfg config.DbConfig) *sqlx.DB {
	db, err := database.DbConnect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func initRouter(cfg config.RestConfig, handlers []domain.RestHandler) *gin.Engine {
	router := gin.Default()

	// Jaeger middleware initialization
	router.Use(tracing.Middleware())

	// Swagger handler initialization
	// From the root directory: swag init --dir ./ --generalInfo ./cmd/pp-gin/main.go
	router.GET(cfg.SwaggerUrl+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Prometheus handler initialization
	// prom := ginprom.NewPrometheus("gin") and prom.Use(router)
	router.GET(cfg.MetricsUrl, metrics.PrometheusHandler())

	// PProf handler initialization
	// https://github.com/gin-contrib/pprof
	// go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
	pprof.Register(router)

	for _, handler := range handlers {
		handler.Register(router)
	}
	return router
}

func initServer(cfg config.RestConfig, r *gin.Engine) {
	if err := endless.ListenAndServe(cfg.Host+":"+strconv.Itoa(cfg.Port), r); err != nil {
		log.Fatal(err)
	}
}

// ***************************
// *** Create repositories ***
// ***************************

func NewReadMappingRepo(cfg config.CacheConfig, db *sqlx.DB) domain.ReadMappingRepo {
	repo, err := cache.NewCachedReadMappingRepo(cfg.DefaultEntityCount, database.NewDbReadMappingRepo(db))
	if err != nil {
		log.Fatal(err)
	}
	return tracing.NewSpanReadMappingRepo(repo)
}

func NewProcessRepo(cfg config.CacheConfig, db *sqlx.DB) domain.ProcessRepo {
	repo, err := cache.NewCachedProcessRepo(cfg.DefaultEntityCount, database.NewDbProcessRepo(db))
	if err != nil {
		log.Fatal(err)
	}
	return tracing.NewSpanProcessRepo(repo)
}

func NewJobRepo(db *sqlx.DB) domain.JobRepo {
	return tracing.NewSpanJobRepo(database.NewDbJobRepo(db))
}

func NewOrderRepo(cfg config.CacheConfig, db *sqlx.DB) domain.OrderRepo {
	repo, err := cache.NewCachedOrderRepo(cfg.DefaultEntityCount, database.NewDbOrderRepo(db))
	if err != nil {
		log.Fatal(err)
	}
	return tracing.NewSpanOrderRepo(repo)
}
