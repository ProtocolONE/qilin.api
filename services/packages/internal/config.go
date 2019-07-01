package internal

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Db *DbConfig
}

type DbConfig struct {
	Uri            string `envconfig:"URI" required:"false" default:"mongodb://localhost:27017"`
	Name           string `envconfig:"NAME" required:"false" default:"packages"`
}

func LoadConfig() (*Config, error) {
	config := &Config{}
	return config, envconfig.Process("PACKAGE_SERVICE", config)
}
