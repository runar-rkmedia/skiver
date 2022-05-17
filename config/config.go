package config

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

type Config struct {
	// Level for logging
	// Enum: [trace debug info warn warning error panic]
	LogLevel string `cfg:"log-level" default:"info" description:"Log-level to use. Can be trace,debug,info,warn(ing),error or panic"`
	// Enum: [human json]
	LogFormat string    `cfg:"log-format" default:"human" description:"Format of the logs. Can be human or json"`
	Api       ApiConfig `cfg:"api" description:"Used with the api-server"`

	// If set, will enable a self-check that monitors the applications resource-usage. Used for debugging, and monitoring outside of any orcestrator like kubernetes
	SelfCheck bool `cfg:"selv-check" default:"true" description:"Enables a self check to check resources."`

	// Global translator-services that should be available
	TranslatorServices []TranslatorService
	// Options for Authentication
	Authentication AuthConfig
	// Set to enable gzip-module
	Gzip bool

	Metrics Metrics
}

type Metrics struct {
	Enabled bool

	// If set, will be exposed on a different port. if not set, it will be exposed on the same port.
	// This can be useful to not expose the metrics publicly.
	Port int
}

type AuthConfig struct {
	// Defines how long a Session should be valid for.
	SessionLifeTime time.Duration
}

// TDB
type TranslatorService struct {
	// Enum: [bind libre]
	Kind     string
	ApiToken string
	Endpoint string
}
type ApiConfig struct {
	Address         string `cfg:"address" default:"0.0.0.0" description:"Address (interface) to listen to)"`
	RedirectPort    int    `cfg:"redirect-port" default:"80" description:"Used normally to redirect from http to https. Will be ignored if zero or same as listening-port"`
	Port            int    `cfg:"port" default:"80" description:"Port to listen to"`
	CertFile        string `cfg:"cert-file" default:"" description:"Number of request to make total"`
	CertKey         string `cfg:"cert-key" default:"" description:"Number of request to make total"`
	DBLocation      string `cfg:"db-path" default:"./storage/db.bbolt" description:"Filepath to where to store the database"`
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	// If set, will register debug-handlers at
	// - /debug/vars
	// - /debug/vars/
	// - /debug/pprof/
	// - /debug/pprof/cmdline
	// - /debug/pprof/profile
	// - /debug/pprof/symbol
	// - /debug/pprof/trace
	Debug bool
}

func GetConfig() *Config {
	var cfg Config
	err := (viper.Unmarshal(&cfg))
	if err != nil {
		panic(err)
	}
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

	viper.SetDefault("Api.ShutdownTimeout", time.Second*20)
	viper.SetDefault("Api.WriteTimeout", time.Second*40)
	viper.SetDefault("Api.IdleTimeout", time.Second*120)
	viper.SetDefault("Api.ReadTimeout", time.Second*5)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			b, err := toml.Marshal(Config{})
			if err != nil {
				panic(err)
			}
			err = os.WriteFile("skiver.toml", b, 0644)
			if err != nil {
				panic(err)
			}
		} else {
			return fmt.Errorf("Fatal error config file: %w \n", err)
		}
	}
	return nil
}
