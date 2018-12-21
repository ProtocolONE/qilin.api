package conf

import (
	"encoding/base64"
	"strings"

	"github.com/spf13/viper"
)

const (
	DefaultJwtSignAlgorithm = "RS256"
)

type ServerConfig struct {
	Port 				int
	AllowOrigins 		[]string
	AllowCredentials 	bool
}

type Database struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

type Jwt struct {
	SignatureSecret       []byte
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
		viper.AddConfigPath("$HOME")
		viper.AddConfigPath("./etc")
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := new(Config)
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	if config.Jwt.Algorithm == "" {
		config.Jwt.Algorithm = DefaultJwtSignAlgorithm
	}

	pemKey, err := base64.StdEncoding.DecodeString(config.Jwt.SignatureSecretBase64)
	if err != nil {
		return nil, err
	}

	config.Jwt.SignatureSecret = pemKey

	return config, nil
}
