package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Service struct {
	Name        string `mapstructure:"name"`
	SystemdUnit string `mapstructure:"systemd_unit"`
}

type Config struct {
	Services []Service `mapstructure:"services"`
}

func Get() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("fatal error config file: %w", err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &config, nil
}
