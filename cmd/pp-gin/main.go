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
	"example.com/oligzeev/pp-gin/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/jmoiron/sqlx"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/uber/jaeger-client-go"
	jaegerconf "github.com/uber/jaeger-client-go/config"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"time"

	"github.com/gin-contrib/pprof"

	_ "example.com/oligzeev/pp-gin/api/swagger"
)

// @title PP Gin
// @version 0.0.1
// @description This is a PP-Gin application.
func main() {
	ctx, done := context.WithCancel(context.Background())
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return util.StartSignalReceiver(groupCtx, done)
	})

	// Initialize configuration
	cfg := initConfig("config/pp-gin.yaml", "pp")

	// Initialize logger
	initLogger(cfg.Logging)

	// Initialize database connection
	db := initDatabase(cfg.DB)
	execTxFunc := func(ctx context.Context, f domain.TxFunc) error {
		return database.ExecTx(ctx, db, f)
	}

	// Initialize open tracing
	_, closer := initTracing(cfg.Tracing)
	defer closer.Close()

	// Initialize repositories
	newUUIDFunc := func() (uuid.UUID, error) {
		return uuid.NewUUID()
	}
	readMappingRepo := database.NewRDBReadMappingRepo(db, newUUIDFunc)
	processRepo := database.NewRDBProcessRepo(db, newUUIDFunc)
	jobRepo := database.NewRDBJobRepo(db)
	orderRepo := database.NewRDBOrderRepo(db, newUUIDFunc)

	// Initialize http clients
	httpClient := retryablehttp.NewClient()
	httpClient.RetryMax = cfg.Rest.Client.RetriesMax
	httpClient.StandardClient().Timeout = cfg.Rest.Client.TimeoutSec * time.Second
	jobStartClient := rest.NewJobStartRestClient(httpClient)

	// Initialize services
	readMappingService := NewReadMappingService(cfg.Cache, readMappingRepo)
	processService := NewProcessService(cfg.Cache, processRepo, execTxFunc)
	orderService := NewOrderService(cfg.Cache, processService, orderRepo, jobRepo, execTxFunc)

	// Initialize scheduler
	if cfg.Scheduler.Enabled {
		group.Go(func() error {
			s := service.NewJobScheduler(cfg.Scheduler, jobRepo, orderService, readMappingService, jobStartClient)
			return s.Start(groupCtx)
		})
	}

	// Initialize rest server
	restServer := rest.NewServer(cfg.Rest.Server, []domain.RestHandler{
		rest.NewMappingRestHandler(readMappingService),
		rest.NewProcessRestHandler(processService),
		rest.NewJobRestHandler(orderService),
		rest.NewOrderRestHandler(orderService),
	})
	initRouter(cfg, restServer.Router())

	group.Go(func() error {
		return restServer.Start(groupCtx)
	})
	group.Go(func() error {
		return restServer.Stop(groupCtx)
	})

	if err := group.Wait(); err != nil && err != context.Canceled {
		log.Error(err)
	}
}

// *****************************
// *** Initialize components ***
// *****************************

func initConfig(yamlFileName, envPrefix string) *domain.ApplicationConfig {
	appConfig, err := appconf.ReadConfig(yamlFileName, envPrefix)
	if err != nil {
		log.Fatal(err)
	}
	return appConfig
}

func initLogger(cfg domain.LoggingConfig) {
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

func initTracing(cfg domain.TracingConfig) (opentracing.Tracer, io.Closer) {
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

func initDatabase(cfg domain.DbConfig) *sqlx.DB {
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func initRouter(cfg *domain.ApplicationConfig, router *gin.Engine) {
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
	router.GET(cfg.Rest.Server.SwaggerUrl+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Prometheus handler initialization
	// prom := ginprom.NewPrometheus("gin") and prom.Use(router)
	router.GET(cfg.Rest.Server.MetricsUrl, metric.PrometheusHandler())

	// PProf handler initialization
	// https://github.com/gin-contrib/pprof
	// go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
	pprof.Register(router)
}

// ***************************
// *** Create repositories ***
// ***************************

func NewReadMappingService(cfg domain.CacheConfig, repo database.ReadMappingRepo) domain.ReadMappingService {
	s := service.NewReadMappingService(repo)
	cached, err := cache.NewCachedReadMappingService(cfg.DefaultEntityCount, s)
	if err != nil {
		log.Fatal(err)
	}
	return tracing.NewSpanReadMappingService(cached)
}

func NewProcessService(cfg domain.CacheConfig, repo database.ProcessRepo, txFunc domain.ExecTxFunc) domain.ProcessService {
	s := service.NewProcessService(repo, txFunc)
	cached, err := cache.NewCachedProcessRepo(cfg.DefaultEntityCount, s)
	if err != nil {
		log.Fatal(err)
	}
	return tracing.NewSpanProcessService(cached)
}

func NewOrderService(cfg domain.CacheConfig, processService domain.ProcessService, orderRepo database.OrderRepo,
	jobRepo database.JobRepo, txFunc domain.ExecTxFunc) domain.OrderService {

	s := service.NewOrderService(processService, orderRepo, jobRepo, txFunc)
	cached, err := cache.NewCachedOrderService(cfg.DefaultEntityCount, s)
	if err != nil {
		log.Fatal(err)
	}
	return tracing.NewSpanOrderService(cached)
}
