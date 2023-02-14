package kafka

// KafkaConfig struct contains params for apache kafka connection
type KafkaConfig struct { //nolint:revive
	// BrokerList - list of brokers ( {"host:port"}[,"host:port"])
	BrokerList []string `env:"BROKERS" envSeparator:"," yaml:"brokerList" validate:"required,min=1,dive,required"`
	// LogSarama enable logging inside sarama
	LogSarama bool `env:"LOG_SARAMA" yaml:"logSarama"`
}
