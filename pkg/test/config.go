package qilin_test

import (
	"github.com/kelseyhightower/envconfig"
	"qilin-api/pkg/conf"
)

func LoadTestConfig() (*conf.Config, error) {
	config := &conf.Config{}
	err := envconfig.Process("QAPI", config)

	return config, err
}
