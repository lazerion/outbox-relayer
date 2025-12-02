package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type RelayerConfig struct {
	Batch       int           `mapstructure:"batch"`
	Timeout     time.Duration `mapstructure:"timeout"`
	MaxAttempts int           `mapstructure:"maxAttempts"`
}

type WebhookConfig struct {
	Url     string        `mapstructure:"url"`
	AuthKey string        `mapstructure:"authKey"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type ScheduleConfig struct {
	Interval time.Duration `mapstructure:"interval"`
}

type Migration struct {
	Path string `mapstructure:"path"`
}

type RedisConfig struct {
	Host     string        `mapstructure:"host"`
	Port     int           `mapstructure:"port"`
	Password string        `mapstructure:"password"`
	DB       int           `mapstructure:"db"`
	TTL      time.Duration `mapstructure:"ttl"`
}

type Config struct {
	Postgres  PostgresConfig `mapstructure:"postgres"`
	Relayer   RelayerConfig  `mapstructure:"relayer"`
	Webhook   WebhookConfig  `mapstructure:"webhook"`
	Schedule  ScheduleConfig `mapstructure:"schedule"`
	Migration Migration      `mapstructure:"migration"`
	Redis     RedisConfig    `mapstructure:"redis"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.AddConfigPath("./internal/config")
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	// Env overrides from docker-compose.yml
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

var Module = fx.Provide(LoadConfig)
