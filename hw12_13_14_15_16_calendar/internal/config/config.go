package config

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Logger    LoggerConf
	Storage   StorageConf
	HTTP      HTTPConf
	GRPC      GRPCConf
	Rabbit    RabbitConf
	Scheduler SchedulerConf
}

type LoggerConf struct {
	Level string
}

type StorageConf struct {
	Type string
	SQL  SQLConf
}

type SQLConf struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

type HTTPConf struct {
	Host string
	Port int
}

type GRPCConf struct {
	Host string
	Port int
}

type RabbitConf struct {
	Host     string
	Port     int
	Username string
	Password string
	Queue    string
}

type SchedulerConf struct {
	Interval time.Duration
}

func New(path string) (Config, error) {
	var config Config

	if _, err := toml.DecodeFile(path, &config); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	return config, nil
}
