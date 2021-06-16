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

type AWSConfig struct {
	Region            string
	DatabaseSecretARN string
}

type Config struct {
	Service ServiceConfig
	AWS     AWSConfig
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

func SetAWSSession(sess *session.Session) {
	AWSSession = sess
}
