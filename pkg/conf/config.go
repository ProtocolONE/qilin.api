package conf

import (
	"encoding/base64"
	"github.com/pkg/errors"
	"os"
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
	Debug	 			bool
}

type Database struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
	LogMode  bool
}

type Jwt struct {
	SignatureSecret       []byte
	SignatureSecretBase64 string
	Algorithm             string
}

type GeoIP struct {
	DBPath string
}

type Mailer struct {
	ReplyTo				string
	From				string
	Host 				string
	Port 				int
	Username 			string
	Password 			string
	InsecureSkipVerify 	bool
}

// Config the application's configuration
type Config struct {
	Server    ServerConfig
	Database  Database
	Jwt       Jwt
	GeoIP     GeoIP
	LogConfig LoggingConfig
	Mailer  Mailer
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
		return nil, errors.Wrap(err, "Read config file")
	}

	config := new(Config)
	if err := viper.Unmarshal(config); err != nil {
		return nil, errors.Wrap(err, "Unmarshal config file")
	}

	if config.Jwt.Algorithm == "" {
		config.Jwt.Algorithm = DefaultJwtSignAlgorithm
	}

	pemKey, err := base64.StdEncoding.DecodeString(config.Jwt.SignatureSecretBase64)
	if err != nil {
		return nil, errors.Wrap(err, "Decode JWT")
	}

	config.Jwt.SignatureSecret = pemKey

	return config, nil
}

// Config for testing the application jobs
type TestConfig struct {
	Database  Database
}

func LoadTestConfig() (*TestConfig, error) {
	viper.SetEnvPrefix("QILIN")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	configFile := os.Getenv("TEST_CONFIG")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("test.config.yaml")
		viper.AddConfigPath("./")
		viper.AddConfigPath("$HOME")
		viper.AddConfigPath("./etc")
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "Read test config file")
	}

	config := new(TestConfig)
	if err := viper.Unmarshal(config); err != nil {
		return nil, errors.Wrap(err, "Unmarshal test config file")
	}

	return config, nil
}
