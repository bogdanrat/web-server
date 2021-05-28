package config

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"path"
	"path/filepath"
	"strings"
)

type ServerConfig struct {
	ListenAddress string
	GinMode       string
}

type DbConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
	SslMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type AuthenticationConfig struct {
	AccessTokenDuration  int // minutes
	RefreshTokenDuration int // minutes
	MFA                  bool
	Channel              string
}

type GRPCConfig struct {
	Deadline       int64 // milliseconds
	UseCompression bool
}

type Config struct {
	Server         ServerConfig
	DB             DbConfig
	Redis          RedisConfig
	Authentication AuthenticationConfig
	GRPC           GRPCConfig
}

var (
	AppConfig  Config
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
	// SetEnvPrefix defines a prefix that ENVIRONMENT variables will use
	// e.g., WEB_SERVER_SOME-VAR
	viper.SetEnvPrefix("WEB_SERVER")
}