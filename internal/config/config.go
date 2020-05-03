package config

import (
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type RestConfig struct {
	Host             string `yaml:"host"`
	Port             int    `yaml:"port"`
	SwaggerUrl       string `yaml:"swaggerUrl"`
	MetricsUrl       string `yaml:"metricsUrl"`
	ClientRetriesMax int    `yaml:"clientRetriesMax"`
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
	Level int `yaml:"level"`
}

type BalanceConfig struct {
	RetryMax   int    `yaml:"retryMax"`
	RequestUrl string `yaml:"requestUrl"`
}

type SchedulerConfig struct {
	Enabled           bool `yaml:"enabled"`
	PeriodSec         int  `yaml:"periodSec"`
	SendJobTimeoutSec int  `yaml:"sendJobTimeoutSec"`
	SendJobRetriesMax int  `yaml:"sendJobRetriesMax"`
	JobLimit          int  `yaml:"jobLimit"`
}

type StubConfig struct {
	ResponseUrl       string `yaml:"responseUrl"`
	SendJobTimeoutSec int    `yaml:"sendJobTimeoutSec"`
	SendJobRetriesMax int    `yaml:"sendJobRetriesMax"`
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

func ReadConfig(yamlFileName, envPrefix string) (*ApplicationConfig, error) {
	const op = "ReadConfig"

	file, err := os.Open(yamlFileName)
	if err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't open file: %v", yamlFileName), err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't read file: %v", yamlFileName), err)
	}
	config := ApplicationConfig{}
	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't unmarshal file: %v", yamlFileName), err)
	}
	if err = envconfig.Process(envPrefix, &config); err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't apply envconfig with prefix: %v", yamlFileName), err)
	}
	return &config, nil
}
