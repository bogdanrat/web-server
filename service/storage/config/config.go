package config

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/viper"
	"path"
	"path/filepath"
	"strings"
)

type ServiceConfig struct {
	Address string
}

type ServerConfig struct {
	ListenAddress   string
	GinMode         string
	DevelopmentMode bool
}

type UploadConfig struct {
	MaxFileSize uint32
}

type DiskStorageConfig struct {
	Path string
}

type S3Config struct {
	Domain           string
	Bucket           string
	BucketVersioning bool
	Concurrency      int
	PartSize         int
	MaxAttempts      int
	Timeout          int
}

type AWSConfig struct {
	Region string
	S3     S3Config
}

type PrometheusConfig struct {
	Enabled     bool
	MetricsPath string
}

type Config struct {
	Service       ServiceConfig
	Server        ServerConfig
	Upload        UploadConfig
	StorageEngine string
	DiskStorage   DiskStorageConfig
	AWS           AWSConfig
	Prometheus    PrometheusConfig
}

var (
	AppConfig  Config
	AWSSession *session.Session
	configFile *string
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

func SetAWSSession(sess *session.Session) {
	AWSSession = sess
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
	viper.SetEnvPrefix("STORAGE_SERVICE")
}
