package config

import (
	"fmt"
	"os"
	"path"

	"github.com/pelletier/go-toml"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

type Config struct {
	LogLevel  string    `cfg:"log-level" default:"info" description:"Log-level to use. Can be trace,debug,info,warn(ing),error or panic"`
	LogFormat string    `cfg:"log-format" default:"human" description:"Format of the logs. Can be human or json"`
	Api       ApiConfig `cfg:"api" description:"Used with the api-server"`
	SelfCheck bool      `cfg:"selv-check" default:"true" description:"Enables a self check to check resources."`
}

type ApiConfig struct {
	Address      string `cfg:"address" default:"0.0.0.0" description:"Address (interface) to listen to)"`
	RedirectPort int    `cfg:"redirect-port" default:"80" description:"Used normally to redirect from http to https. Will be ignored if zero or same as listening-port"`
	Port         int    `cfg:"port" default:"80" description:"Port to listen to"`
	CertFile     string `cfg:"cert-file" default:"" description:"Number of request to make total"`
	CertKey      string `cfg:"cert-key" default:"" description:"Number of request to make total"`
	DBLocation   string `cfg:"db-path" default:"./storage/db.bbolt" description:"Filepath to where to store the database"`
}

func GetConfig() *Config {
	var cfg Config
	viper.Unmarshal(&cfg)
	return &cfg
}

func InitConfig() error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		viper.SetConfigName("skiver")
		viper.AddConfigPath(path.Join(home, "skiver"))
		viper.AddConfigPath(path.Join(home, ".config", "skiver"))
		viper.AddConfigPath(".")
	}
	viper.SetEnvPrefix("skiver")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			b, err := toml.Marshal(Config{})
			if err != nil {
				panic(err)
			}
			os.WriteFile("skiver.toml", b, 0644)
		} else {
			return fmt.Errorf("Fatal error config file: %w \n", err)
		}
	}
	return nil
}
