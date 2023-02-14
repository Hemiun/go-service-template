package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// AppConfig - structure for app configuration
type AppConfig struct {
	// Passport - struct for service identification params
	Passport struct {
		// ServiceName - name of the service
		ServiceName string `env:"SERVICE_NAME" yaml:"serviceName"`

		// ServiceInstance - service instance name. Usually k8s pod name
		ServiceInstance string `env:"SERVICE_INSTANCE"`
	} `yaml:"passport"`
	Server struct {
		// Address - address for service listening
		Address string `env:"RUN_ADDRESS" yaml:"address" validate:"required"`
	} `yaml:"server"`
	// Database  - struct for db connection params
	Database struct {
		// Address - database host name
		Address string `env:"DB_ADDRESS" yaml:"address" validate:"required"`

		// Port - database port
		Port int32 `env:"DB_PORT" yaml:"port" validate:"required"`

		// DB - database name
		DB string `env:"DB_NAME" yaml:"db" validate:"required"`

		// Username  - name of user, that used for db connection
		Username string `env:"DB_USERNAME" yaml:"username" validate:"required"`

		// Password - password for username
		Password string `env:"DB_PASSWORD" yaml:"password" validate:"required"`
	} `yaml:"database"`
	PgPool struct {
		// MaxConnLifetime is the duration since creation after which a connection will be automatically closed
		MaxConnLifetime time.Duration `yaml:"maxConnLifetime"`

		// MaxConnIdleTime is the duration after which an idle connection will be automatically closed by the health check.
		MaxConnIdleTime time.Duration `yaml:"maxConnIdleTime"`

		//MaxConns is the maximum size of the pool.
		MaxConns int32 `yaml:"maxConns"`

		// MinConns is the minimum size of the pool.
		MinConns int32 `yaml:"minConns"`

		// HealthCheckPeriod is the duration between checks of the health of idle connections.
		HealthCheckPeriod time.Duration `yaml:"healthCheckPeriod"`

		// LazyConnect. If set to true, pool doesn't do any I/O operation on initialization. And connects to the server only when the pool starts to be used.
		LazyConnect bool `yaml:"lazyConnect"`

		// LogPGX - enable logging inside pgx
		LogPGX bool `env:"LOG_PGX" yaml:"logPGX"`

		// LogLevel - set log level for pgx. Available values are trace, debug, info, warn, error, none
		LogLevel string `env:"PGX_LOG_LEVEL" yaml:"pgxLogLevel"`
	} `yaml:"pgPool"`
	// Kafka struct contains params for apache kafka connection
	Kafka struct {
		// BrokerList - list of brokers ( {"host:port"}[,"host:port"])
		BrokerList []string `env:"BROKERS" envSeparator:"," yaml:"brokerList" validate:"required,min=1,dive,required"`
		// LogSarama enable logging inside sarama
		LogSarama bool `env:"LOG_SARAMA" yaml:"logSarama"`
	} `yaml:"kafka"`
	HTTPClient struct {
		// RequestTimeout - request timeout for http client
		RequestTimeout time.Duration `yaml:"requestTimeout"`
	} `yaml:"httpClient"`
	Logger struct {
		// LogLevel - define global log level
		LogLevel string `env:"LOG_LEVEL" yaml:"logLevel"`
	} `yaml:"logger"`
}

// Validate - method for config validation
func (config *AppConfig) Validate() error {
	return validator.New().Struct(config)
}

func printHeader(msg string) {
	lineSize := 100
	padSize := (lineSize - len(msg) + len(msg)%2) / 2
	fmt.Printf("+%s%s%s+\n", strings.Repeat("-", padSize), msg, strings.Repeat("-", padSize))
}

// PrintEnv print all envs after service start
func PrintEnv() {
	printHeader("Command line args")
	for _, p := range os.Args {
		fmt.Println(p)
	}
	printHeader("Env list")
	for _, e := range os.Environ() {
		fmt.Println(e)
	}
	printHeader("")
}

func printAppConfig(config AppConfig) {
	out, err := yaml.Marshal(config)
	if err != nil {
		fmt.Println("yaml marshall error", err)
		return
	}
	printHeader("AppConfig")
	fmt.Println(string(out))
	printHeader("")
}

// Init - app config initialisation
func (config *AppConfig) Init() error {
	PrintEnv()
	// 1. Set default values
	config.HTTPClient.RequestTimeout = time.Second * 30
	config.PgPool.MaxConnIdleTime = time.Second * 120
	config.PgPool.MaxConns = 5
	config.PgPool.MinConns = 2
	//

	// 2. Application.yaml read
	yamlFile, err := ioutil.ReadFile("./application.yaml")
	if err != nil {
		fmt.Println("can't read file application.yaml", err)
		return err
	}

	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		fmt.Println("can't load app config", err)
		return err
	}

	// 3. Envs
	err = env.Parse(config)
	if err != nil {
		fmt.Println("can't load service config", err)
		return err
	}

	printAppConfig(*config)

	// 4. Validation

	err = config.Validate()
	if err != nil {
		fmt.Println("one or few required params is empty", err)
		return err
	}
	return nil
}

// NewConfig - make new AppConfig
func NewConfig() *AppConfig {
	var target AppConfig

	return &target
}
