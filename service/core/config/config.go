package config

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/viper"
	"html/template"
	"path"
	"path/filepath"
	"strings"
)

type ServerConfig struct {
	ListenAddress   string
	GinMode         string
	DevelopmentMode bool
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type AuthenticationConfig struct {
	AccessTokenDuration  int64 // minutes
	RefreshTokenDuration int64 // minutes
	MFA                  bool
	Channel              string
}

type SMTPConfig struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	RefreshToken string
}

type RabbitMQConfig struct {
	DefaultUser     string
	DefaultPassword string
	Host            string
	Port            string
	Exchange        string
	Queue           string
}

type SQSConfig struct {
	QueueName                 string
	ContentBasedDeduplication string
	DelaySeconds              string
	MessageRetentionPeriod    string
	MaxNumberOfMessages       int64
	VisibilityTimeout         int64
	WaitTimeSeconds           int64
}

type MessageBrokerConfig struct {
	Broker   string
	RabbitMQ RabbitMQConfig
	SQS      SQSConfig
}

type AWSConfig struct {
	Region            string
	DatabaseSecretARN string
}

type ServicesConfig struct {
	Auth    AuthConfig
	Storage StorageConfig
}

type PrometheusConfig struct {
	Enabled     bool
	MetricsPath string
}

type GRPCConfig struct {
	Deadline       int64 // milliseconds
	UseCompression bool
}
type AuthConfig struct {
	GRPC GRPCConfig
}
type StorageConfig struct {
	GRPC            GRPCConfig
	ImagesPrefix    string
	DocumentsPrefix string
}

type Config struct {
	Server         ServerConfig
	Redis          RedisConfig
	Authentication AuthenticationConfig
	SMTP           SMTPConfig
	MessageBroker  MessageBrokerConfig
	AWS            AWSConfig
	Services       ServicesConfig
	Prometheus     PrometheusConfig
	TemplateCache  map[string]*template.Template
}

var (
	AppConfig  Config
	AWSSession *session.Session
	configFile *string
)

const (
	RabbitMQBroker = "RabbitMQ"
	SQSBroker      = "SQS"
)

func ReadFlags() {
	configFile = flag.String("config", "./config.json", "app config file")
	flag.Parse()
}

func ReadConfiguration() error {
	initViper()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("couldn't read configuration: %v", err)
	}
	if err := viper.Unmarshal(&AppConfig); err != nil {
		return fmt.Errorf("couldn't unmarshal configuration: %v", err)
	}

	return nil
}

func initViper() {
	configFileName := path.Base(*configFile)
	configFilePath := path.Dir(*configFile)
	configFileType := ""

	// Ext(path) returns the file name extension used by path; it is empty if there is no dot.
	if configFileExtension := filepath.Ext(configFileName); configFileExtension != "" {
		configFileType = configFileExtension[strings.Index(configFileExtension, ".")+1:]
	}

	viper.AddConfigPath(configFilePath)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileType)
	viper.SetEnvPrefix("CORE")
}

func SetAWSSession(sess *session.Session) {
	AWSSession = sess
}
