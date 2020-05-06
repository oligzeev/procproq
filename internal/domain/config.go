package domain

import "time"

type ServerRestConfig struct {
	Host               string        `yaml:"host"`
	Port               int           `yaml:"port"`
	SwaggerUrl         string        `yaml:"swaggerUrl"`
	MetricsUrl         string        `yaml:"metricsUrl"`
	ReadTimeoutSec     time.Duration `yaml:"readTimeoutSec"`
	WriteTimeoutSec    time.Duration `yaml:"writeTimeoutSec"`
	ShutdownTimeoutSec time.Duration `yaml:"shutdownTimeoutSec"`
}

type ClientRestConfig struct {
	RetriesMax int           `yaml:"retriesMax"`
	TimeoutSec time.Duration `yaml:"timeoutSec"`
}

type RestConfig struct {
	Server ServerRestConfig `yaml:"server"`
	Client ClientRestConfig `yaml:"client"`
}

type DbConfig struct {
	Host               string `yaml:"host"`
	Port               int    `yaml:"port"`
	User               string `yaml:"user"`
	Password           string `yaml:"password"`
	DbName             string `yaml:"dbName"`
	MaxConnections     int    `yaml:"maxConnections"`
	MaxIdleConnections int    `yaml:"maxIdleConnections"`
}

type CacheConfig struct {
	DefaultEntityCount int `yaml:"defaultEntityCount"`
}

type TracingConfig struct {
	ServiceName string `yaml:"serviceName"`
}

type LoggingConfig struct {
	Level           int    `yaml:"level"`
	TimestampFormat string `yaml:"timestampFormat"`
	Default         bool   `yaml:"default"`
}

type BalanceConfig struct {
	RetryMax   int    `yaml:"retryMax"`
	RequestUrl string `yaml:"requestUrl"`
}

type SchedulerConfig struct {
	Enabled   bool          `yaml:"enabled"`
	PeriodSec time.Duration `yaml:"periodSec"`
	JobLimit  int           `yaml:"jobLimit"`
}

type StubConfig struct {
	ResponseUrl string `yaml:"responseUrl"`
}

// Possible tags in https://github.com/kelseyhightower/envconfig
type ApplicationConfig struct {
	Rest      RestConfig      `yaml:"rest"`
	DB        DbConfig        `yaml:"db"`
	Cache     CacheConfig     `yaml:"cache"`
	Tracing   TracingConfig   `yaml:"tracing"`
	Logging   LoggingConfig   `yaml:"logging"`
	Balance   BalanceConfig   `yaml:"balance"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
	Stub      StubConfig      `yaml:"stub"`
}
