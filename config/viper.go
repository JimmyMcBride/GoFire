package config

import (
	"errors"
	"os"

	"github.com/spf13/viper"
)

type EnvVars struct {
	MongodbUri  string `mapstructure:"MONGODB_URI"`
	MongodbName string `mapstructure:"MONGODB_NAME"`
	PORT        string `mapstructure:"PORT"`
}

func LoadConfig() (config EnvVars, err error) {
	env := os.Getenv("GO_ENV")
	if env == "production" {
		return EnvVars{
			MongodbUri:  os.Getenv("MONGODB_URI"),
			MongodbName: os.Getenv("MONGODB_NAME"),
			PORT:        os.Getenv("PORT"),
		}, nil
	}

	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)

	// validate config here
	if config.MongodbUri == "" {
		err = errors.New("MONGODB_URI is required")
		return
	}

	if config.MongodbName == "" {
		err = errors.New("MONGODB_NAME is required")
		return
	}

	return
}
