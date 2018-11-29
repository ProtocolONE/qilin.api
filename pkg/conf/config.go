package conf

import (
	"crypto/rsa"
	"strings"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port int
}

type Database struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

type Jwt struct {
	SignatureSecret       *rsa.PublicKey
	SignatureSecretBase64 string
	Algorithm             string
}

type GeoIP struct {
	DBPath string
}

// Config the application's configuration
type Config struct {
	Server    ServerConfig
	Database  Database
	Jwt       Jwt
	GeoIP     GeoIP
	LogConfig LoggingConfig
}

// LoadConfig loads the config from a file if specified, otherwise from the environment
func LoadConfig(configFile string) (*Config, error) {
	viper.SetEnvPrefix("QILIN")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("./")
		viper.AddConfigPath("$HOME/.example")
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := new(Config)
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}
	return config, nil
}
