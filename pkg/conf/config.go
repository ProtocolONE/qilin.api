package conf

// Config the application's configuration
type Config struct {
	Server    ServerConfig
	Database  Database
	Auth1     Auth1
	Mailer    Mailer
	Notifier  Notifier
	Enforcer  Enforcer
	Imaginary Imaginary
}

type Enforcer struct {
	Host string `envconfig:"HOST" required:"false" default:"127.0.0.1"`
	Port int    `envconfig:"PORT" required:"false" default:"6379"`
}

// ServerConfig specifies all the parameters needed for http server
type ServerConfig struct {
	Port             int      `envconfig:"PORT" required:"false" default:"8080"`
	AllowOrigins     []string `envconfig:"ALLOW_ORIGINS" required:"false" default:"*"`
	AllowCredentials bool     `envconfig:"ALLOW_CREDENTIALS" required:"false" default:"false"`
	Debug            bool     `envconfig:"DEBUG" required:"false" default:"false"`
}

// Database specifies all the parameters needed for GORM connection
type Database struct {
	DSN     string `envconfig:"DSN" required:"false" default:"postgres://postgres:postgres@localhost:5432/qilin?sslmode=disable"`
	LogMode bool   `envconfig:"DEBUG" required:"false" default:"false"`
}

// Auth1 specifies all the parameters needed for authenticate
type Auth1 struct {
	Issuer       string `envconfig:"ISSUER" required:"true" default:"https://dev-auth1.tst.protocol.one"`
	ClientId     string `envconfig:"CLIENTID" required:"true"`
	ClientSecret string `envconfig:"CLIENTSECRET" required:"true"`
}

type Notifier struct {
	Host   string `envconfig:"HOST" required:"false" default:"http://localhost:8000"`
	ApiKey string `envconfig:"API_KEY" required:"true"`
	Secret string `envconfig:"SECRET" required:"true"`
}

type Imaginary struct {
	Secret string `envconfig:"SECRET" required:"false" default:"123456789"`
}

// Mailer specifies all the parameters needed for dump mail sender
type Mailer struct {
	Host               string `envconfig:"HOST" required:"false" default:"localhost"`
	Port               int    `envconfig:"PORT" required:"false" default:"25"`
	Username           string `envconfig:"USERNAME" required:"false" default:""`
	Password           string `envconfig:"PASSWORD" required:"false" default:""`
	ReplyTo            string `envconfig:"REPLY_TO" required:"false" default:""`
	From               string `envconfig:"FROM" required:"false" default:""`
	InsecureSkipVerify bool   `envconfig:"SKIP_VERIFY" required:"false" default:"true"`
}
