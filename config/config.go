package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/runar-rkmedia/go-common/logger"
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

	UploadSnapShots map[string]Uploader
}

type Uploader struct {
	S3 *S3UploaderConfig
}
type S3UploaderConfig struct {
	// Endpoint for the s3-compatible service
	Endpoint,
	// The region for the service.
	Region,
	// Bucket to upload into
	BucketID,
	// AccessKey for the bucket / application
	AccessKey,
	// Pretty name, will be displayed to users
	ProviderName,
	// PrivateKey for the bucket / application
	PrivateKey string
	ForcePathStyle bool
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
func getDefaultConfigPaths(name string) ([]string, error) {
	viper.SetConfigName(name)
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	paths := []string{
		path.Join(home, "."+name),
		path.Join(home, ".config", name),
		".",
	}
	for _, p := range paths {
		viper.AddConfigPath(p)
	}

	return paths, nil

}

// Gets a base64-encoded config-file from environment-variables, decodes it and
// loads it into viper This is useful in environments where it is not practical
// to mount a file on disk for configuration, but a structured configuration is
// still wanted.
func getEnvConfig() (bool, error) {
	cfg_b64 := os.Getenv("SKIVER_CONFIG_B64")
	if cfg_b64 == "" {
		return false, nil
	}
	cfg_b, err := base64.StdEncoding.DecodeString(cfg_b64)
	if err != nil {
		_pre_init_fatal_logger(err, "failed to decode SKIVER_CONFIG_B64", nil)
	}
	configType := os.Getenv("SKIVER_CONFIG_B64_TYPE")
	if configType == "" {
		configType = "toml"
	}

	viper.SetConfigType(configType)

	r := bytes.NewReader(cfg_b)
	err = viper.ReadConfig(r)
	if err != nil {
		_pre_init_fatal_logger(err, "viper failed in ReadConfig", nil)
	}
	return true, nil
}

func InitConfig() error {
	viper.SetEnvPrefix("skiver")
	viper.AutomaticEnv()
	viper.SetDefault("Api.ShutdownTimeout", time.Second*20)
	viper.SetDefault("Api.WriteTimeout", time.Second*40)
	viper.SetDefault("Api.IdleTimeout", time.Second*120)
	viper.SetDefault("Api.ReadTimeout", time.Second*5)

	if fromEnv, err := getEnvConfig(); err != nil {
		panic(err)
	} else if fromEnv {
		return nil
	}

	paths := []string{cfgFile}
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		ps, err := getDefaultConfigPaths("skiver")
		if err != nil {
			return err
		}
		paths = ps
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			l := logger.GetLogger("config")
			l.Warn().Interface("default-config-paths", paths).Msg("Config-file was not found in any of the default paths")
			// Config file not found; ignore error if desired
			b, err := toml.Marshal(CreateSampleComfig())
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

func CreateSampleComfig() Config {
	return Config{
		TranslatorServices: []TranslatorService{
			{},
		},
		UploadSnapShots: map[string]Uploader{
			"example": {
				S3: &S3UploaderConfig{
					BucketID:     "my-bucket",
					Endpoint:     "https://s3.us-west-001.backblazeb2.com",
					AccessKey:    "0012345678901234567890123",
					ProviderName: "BackBlaze B2 Public",
					PrivateKey:   "secret",
					Region:       "us-west-001",
				},
			},
		},
	}
}

func WriteToml(w io.Writer, j interface{}) error {
	tomler := toml.NewEncoder(w)
	tomler.SetTagName("json")
	tomler.CompactComments(true)
	tomler.ArraysWithOneElementPerLine(true)
	tomler.SetTagComment("help")
	return tomler.Encode(j)
}

// Should only be used before the logger has been initialized, for instance when parsing config etc.
func _pre_init_fatal_logger(err error, msg string, details map[string]interface{}) {
	l := logger.InitLogger(logger.LogConfig{
		Level:      "debug",
		Format:     "human",
		WithCaller: false,
	})
	lerr := l.Fatal().Err(err)
	for k, v := range details {
		lerr = lerr.Interface(k, v)
	}
	lerr.Msg(msg)
	panic("fatal")
}
