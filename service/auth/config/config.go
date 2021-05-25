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

type OpenCensusConfig struct {
	Enabled   bool
	StatsPage string
	Address   string
}

type Config struct {
	Service    ServiceConfig
	OpenCensus OpenCensusConfig
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
	viper.SetEnvPrefix("AUTH_SERVICE")
}
