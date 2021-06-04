package config

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"path"
	"path/filepath"
	"strings"
)

type ServiceConfig struct {
	Address string
}

type DbConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
	SslMode  string
}

type Config struct {
	Service ServiceConfig
	DB      DbConfig
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

	if configFileExtension := filepath.Ext(configFileName); configFileExtension != "" {
		configFileType = configFileExtension[strings.Index(configFileExtension, ".")+1:]
	}

	viper.AddConfigPath(configFilePath)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileType)
	viper.SetEnvPrefix("DATABASE_SERVICE")
}
