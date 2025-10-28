package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Service struct {
	Name        string `mapstructure:"name"`
	SystemdUnit string `mapstructure:"systemd_unit"`
}

type Config struct {
	Services []Service `mapstructure:"services"`
}

func Get(conf *Config, refreshChan chan<- struct{}) error {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("fatal error config file: %w", err)
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		_ = viper.Unmarshal(conf)
		if refreshChan != nil {
			refreshChan <- struct{}{}
		}
	})
	viper.WatchConfig()
	err = viper.Unmarshal(&conf)

	if err != nil {
		return fmt.Errorf("unable to decode config: %w", err)
	}

	return nil
}
