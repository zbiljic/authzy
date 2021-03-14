package config

import (
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/imdario/mergo"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/zbiljic/authzy"
	"github.com/zbiljic/authzy/pkg/database/jsonmutexdb"
	"github.com/zbiljic/authzy/pkg/database/leveldb"
	"github.com/zbiljic/authzy/pkg/hash"
	xhttp "github.com/zbiljic/authzy/pkg/http"
	"github.com/zbiljic/authzy/pkg/logger"
)

type Config struct {
	SiteURL  string          `json:"site_url" split_words:"true"`
	Logger   *logger.Config  `json:"logger"`
	Debug    *DebugConfig    `json:"debug" validate:"dive"`
	HTTP     *HTTPConfig     `json:"http" validate:"dive"`
	Hashers  *HashersConfig  `json:"hashers" validate:"dive"`
	Database *DatabaseConfig `json:"database" validate:"dive"`
	API      *APIConfig      `json:"api" validate:"dive"`
	SMTP     *SMTPConfig     `json:"smtp" validate:"dive"`
}

type DebugConfig struct {
	Enabled bool   `json:"enabled" default:"true"`
	Addr    string `json:"addr" default:":6060"`
}

type HTTPConfig struct {
	Addr string `json:"addr" default:":8080"`
}

type HashersConfig struct {
	Argon2 *hash.HasherArgon2Config `json:"argon2" required:"true"`
}

type DatabaseConfig struct {
	Type        string              `json:"type" validate:"required"`
	JSONMutexDB *jsonmutexdb.Config `json:"jsonmutexdb"`
	LevelDB     *leveldb.Config     `json:"leveldb"`
}

type APIConfig struct {
	Secure            bool          `json:"secure" default:"true"`
	RequestIDHeader   string        `json:"request_id_header" split_words:"true" validate:"required"`
	ExternalURL       string        `json:"external_url" split_words:"true"`
	AllowedLogoutURLs []string      `json:"allowed_logout_urls" split_words:"true"`
	CSRF              *CSRFConfig   `json:"csrf" validate:"dive"`
	JWT               *JWTConfig    `json:"jwt" validate:"dive"`
	Mailer            *MailerConfig `json:"mailer" validate:"dive"`
	Cookie            *CookieConfig `json:"cookie" validate:"dive"`
	DisableSignup     bool          `json:"disable_signup" split_words:"true"`
}

// CSRFConfig holds all the CSRF related configuration.
type CSRFConfig struct {
	Enabled bool   `json:"enabled" default:"true"`
	AuthKey string `json:"auth_key" split_words:"true" validate:"required,gte=32"`
}

// JWTConfig holds all the JWT related configuration.
type JWTConfig struct {
	ClaimsNamespace string        `json:"claims_namespace" split_words:"true" validate:"required,gte=3"`
	Exp             int           `json:"exp" default:"3600"`
	Aud             string        `json:"aud"`
	AcceptableSkew  time.Duration `json:"acceptable_skew" split_words:"true" required:"true" default:"30s"`
	DefaultKey      string        `json:"default_key" split_words:"true" validate:"required"`
	KeysJSON        string        `json:"-" envconfig:"keys" validate:"required"`
}

type MailerConfig struct {
	Autoconfirm  bool               `json:"autoconfirm" default:"false"`
	ValidateHost bool               `json:"validate_host" split_words:"true" default:"false"`
	Subjects     EmailContentConfig `json:"subjects"`
	Templates    EmailContentConfig `json:"templates"`
	URLPaths     EmailContentConfig `json:"url_paths" split_words:"true"`
}

// EmailContentConfig holds the configuration for emails, both subjects
// and template URLs.
type EmailContentConfig struct {
	Confirmation string `json:"confirmation"`
	Recovery     string `json:"recovery"`
	EmailChange  string `json:"email_change" split_words:"true"`
}

type CookieConfig struct {
	Key             string `json:"key" validate:"required,gt=1"`
	DurationSeconds int    `json:"duration" envconfig:"duration" default:"86400"`
}

type SMTPConfig struct {
	MaxFrequency time.Duration `json:"max_frequency" split_words:"true" default:"5m"`
	Host         string        `json:"host"`
	Port         int           `json:"port,omitempty" default:"587"`
	User         string        `json:"user"`
	Pass         string        `json:"-"`
	AdminEmail   string        `json:"admin_email" split_words:"true"`
}

func loadEnvironment(filename string) error {
	var err error
	if filename != "" {
		err = godotenv.Load(filename)
	} else {
		err = godotenv.Load()
		// handle if .env file does not exist, this is OK
		if os.IsNotExist(err) {
			return nil
		}
	}

	return err
}

// LoadConfig loads configuration.
func LoadConfig(filename string) (*Config, error) {
	if err := loadEnvironment(filename); err != nil {
		return nil, err
	}

	config := new(Config)
	if err := envconfig.Process(authzy.AppName, config); err != nil {
		return nil, err
	}

	// custom processing for logger config
	if err := envconfig.Process(authzy.AppName, config.Logger); err != nil {
		return nil, err
	}

	config.ApplyDefaults()

	err := config.Validate()
	if err != nil {
		return config, err
	}

	return config, nil
}

// Merge merges present configuration with another one.
func (config *Config) Merge(c *Config) (*Config, error) {
	err := mergo.Merge(config, c,
		mergo.WithOverride,
	)

	return config, err
}

// ApplyDefaults sets defaults for a configuration.
func (config *Config) ApplyDefaults() {
	if config.API.RequestIDHeader == "" {
		config.API.RequestIDHeader = xhttp.XRequestID
	}

	if config.API.Cookie.Key == "" {
		config.API.Cookie.Key = fmt.Sprintf("%s.session-token", authzy.AppName)
		if config.API.Secure {
			config.API.Cookie.Key = "__Secure-" + config.API.Cookie.Key
		}
	}

	if config.API.Mailer.URLPaths.Confirmation == "" {
		config.API.Mailer.URLPaths.Confirmation = "/verify"
	}
	if config.API.Mailer.URLPaths.Recovery == "" {
		config.API.Mailer.URLPaths.Recovery = "/verify"
	}
	if config.API.Mailer.URLPaths.EmailChange == "" {
		config.API.Mailer.URLPaths.EmailChange = "/verify"
	}
}

// Validate validates configuration.
func (config *Config) Validate() error {
	validate := validator.New()

	if err := validate.Struct(*config); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return validationErrors[0]
	}

	return nil
}
