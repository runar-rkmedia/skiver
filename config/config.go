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
	LogLevel string `cfg:"log-level" default:"info" help:"Log-level to use. Can be trace,debug,info,warn(ing),error,fatal or panic" jsonschema:"enum=trace,enum=debug,enum=info,enum=warn,enum=warning,enum=error,enum=fatal,enum=panic"`
	// Enum: [human json]
	LogFormat string    `cfg:"log-format" default:"human" help:"Format of the logs. Can be human or json" jsonschema:"enum=human,enum=json"`
	Api       ApiConfig `cfg:"api" help:"Used with the api-server"`

	// If set, will enable a self-check that monitors the applications resource-usage. Used for debugging, and monitoring outside of any orcestrator like kubernetes
	SelfCheck bool `cfg:"selv-check" default:"true" help:"Enables a self check to check resources."`

	// Global translator-services that should be available
	TranslatorServices []TranslatorService
	// Options for Authentication
	Authentication AuthConfig
	// Set to enable gzip-module for all content served
	Gzip bool

	// Enable Metrics (prometheus-compatible)
	Metrics Metrics

	// Used to upload files to external targets when creating snapshots.
	UploadSnapShots map[string]Uploader

	// Used to upload backups of the database.
	// Can optionally also be used as a source to retreve a backup from on startup, if there is no database.
	DatabaseBackups map[string]BackupConfig `help:"Optional backupendpoints for databases" json:"databaseBackups"`
}

type BackupConfig struct {
	S3 *S3BaseConfig `json:"s3" help:"Use s3 for backup"`
	// If no database is available at startup, this source can be used to fetch the database.
	// Skiver will then use that as a database.
	// This can be useful in environments where there is no readily available persistant storage.
	FetchOnStartup bool
	// The database can be backed up as often as every write, but can be relaxed with this value.
	// Defaults to 10 minutes
	MaxInterval Duration `json:"maxInterval" help:"The database can be backed up as often as every write, but can be relaxed with this value. Defaults to 10 minutes."`
	// Can be used to set a custom objectkey.
	// defaults to "skiver.bbolt"
	FileName string
}

type Uploader struct {
	// S3-compatible target
	S3 *S3UploaderConfig
}
type S3BaseConfig struct {
	// Endpoint for the s3-compatible service
	Endpoint string `json:"endpoint" help:"Endpoint for the s3-compatible service"`
	// The region for the service.
	Region string `json:"region" help:"The region for the service"`

	// Bucket to upload into
	BucketID string `json:"bucketID" help:"Bucket to upload into"`

	// AccessKeyID for the bucket / application
	AccessKey string `json:"accessKey" help:"Accesskey for the bucket / application"`
	// Private key or Secret access key for the bucket / application
	PrivateKey string

	// Name for provider, used for display-puroposes
	ProviderName string `json:"providerName" help:"Pretty name, displayed in logs etc."`
	// If set, will add headers for use with Browser-TTL, CDN-TTL and CloudFlare-TTL
	ForcePathStyle bool `json:"forcePathStyle" help:"If set will add headers for use with Browser-TTL, CDN-TTL and CloudFlare-TTL"`
}
type S3UploaderConfig struct {
	// S3-compatible target
	S3BaseConfig
	// Can be used to override the url that is produced.
	// Golang-templating is available
	// Variables:
	// `.Object`:        The current Object-id (fileName)
	// `.Bucket`:        The current Object-id (fileName)
	// `.EndpointURL`:   net.Url version of the Endpoint
	// `.Endpoint`:      Endpoint as string
	// `.Region`:        Region.
	UrlFormat    string `json:"urlFormat" help:"Can be used to override the url that is produced.\n Golang-templating is available\n Variables:\n '.Object':        The current Object-id (fileName)\n '.Bucket':        The current Object-id (fileName)\n '.EndpointURL':   net.Url version of the Endpoint\n '.Endpoint':      Endpoint as string\n '.Region':        Region. "`
	CacheControl string
}

type Metrics struct {
	Enabled bool

	// If set, will be exposed on a different port. if not set, it will be exposed on the same port.
	// This can be useful to not expose the metrics publicly.
	Port int
}

type AuthConfig struct {
	// Defines how long a Session should be valid for.
	SessionLifeTime Duration `jsonschema="title=Defines how long a session should be valid for"`
}

// TDB
type TranslatorService struct {
	// Enum: [bind libre]
	Kind     string
	ApiToken string
	Endpoint string
}
type ApiConfig struct {
	// Address (interface) to listen to
	Address      string `cfg:"address" default:"0.0.0.0" help:"Address (interface) to listen to)" jsonschema:"default=0.0.0.0"`
	RedirectPort int    `cfg:"redirect-port" default:"80" help:"Used normally to redirect from http to https. Will be ignored if zero or same as listening-port"`
	Port         int    `cfg:"port" default:"80" help:"Port to listen to"`
	CertFile     string `cfg:"cert-file" default:"" help:"Number of request to make total"`
	CertKey      string `cfg:"cert-key" default:"" help:"Number of request to make total"`
	DBLocation   string `cfg:"db-path" default:"./storage/db.bbolt" help:"Filepath to where to store the database"`
	// Timeout used for reads
	ReadTimeout Duration
	// Timeout used for writes
	WriteTimeout Duration
	// Timeout used for idles
	IdleTimeout Duration
	// Timeout used for shutdown
	ShutdownTimeout Duration
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
	err := (viper.Unmarshal(&cfg, viper.DecodeHook(DurationViperHookFunc())))
	if err != nil {
		panic(err)
	}
	if len(cfg.DatabaseBackups) > 0 {
		for k, v := range cfg.DatabaseBackups {
			if v.MaxInterval == 0 {
				v.MaxInterval = Duration(time.Minute * 10)
				cfg.DatabaseBackups[k] = v
			}
		}
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
			filePtr, err := os.OpenFile("skiver.toml", os.O_WRONLY, 0666)
			err = WriteToml(filePtr, CreateSampleComfig())
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
		DatabaseBackups: map[string]BackupConfig{
			"example": {
				MaxInterval: Duration(time.Minute * 10),
				S3: &S3BaseConfig{
					BucketID:     "my-bucket",
					Endpoint:     "https://s3.us-west-001.backblazeb2.com",
					AccessKey:    "0012345678901234567890123",
					ProviderName: "BackBlaze B2 Public",
					Region:       "us-west-001",
					PrivateKey:   "secret",
				},
			},
		},
		UploadSnapShots: map[string]Uploader{
			"example": {
				S3: &S3UploaderConfig{
					S3BaseConfig: S3BaseConfig{
						BucketID:     "my-bucket",
						Endpoint:     "https://s3.us-west-001.backblazeb2.com",
						AccessKey:    "0012345678901234567890123",
						ProviderName: "BackBlaze B2 Public",
						Region:       "us-west-001",
						PrivateKey:   "secret",
					},
				},
			},
		},
	}
}

func WriteToml(w io.Writer, j interface{}) error {
	tomler := toml.NewEncoder(w)
	tomler.SetTagName("json")
	tomler.CompactComments(false)
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
