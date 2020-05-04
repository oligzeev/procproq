package config

import (
	"example.com/oligzeev/pp-gin/internal/domain"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func ReadConfig(yamlFileName, envPrefix string) (*domain.ApplicationConfig, error) {
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
	config := domain.ApplicationConfig{}
	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't unmarshal file: %v", yamlFileName), err)
	}
	if err = envconfig.Process(envPrefix, &config); err != nil {
		return nil, domain.E(op, fmt.Sprintf("can't apply envconfig with prefix: %v", yamlFileName), err)
	}
	return &config, nil
}
