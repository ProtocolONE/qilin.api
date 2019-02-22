package conf

// Config the application's configuration
type Config struct {
	Server   ServerConfig
	Database Database
	Jwt      Jwt
	Mailer   Mailer
	Notifier Notifier
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

// Jwt specifies all the parameters needed for Jwt middleware
type Jwt struct {
	SignatureSecret string `envconfig:"SECRET" required:"true"`
	Algorithm       string `envconfig:"ALGORITHM" required:"false" default:"HS256"`
}

type Notifier struct {
	Host   string `envconfig:"HOST" required:"false" default:"http://localhost:8000"`
	ApiKey string `envconfig:"API_KEY" required:"true"`
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
