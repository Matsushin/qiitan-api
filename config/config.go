package config

import (
	"sync"

	"github.com/Matsushin/qiitan-api/env"
	"github.com/Matsushin/qiitan-api/logger"
)

// Config APIのConfigを管理する構造体
type Config struct {
	MySQL *MySQLConfig
}

var (
	instance *Config
	once     sync.Once
	err      error
)

func newConfig() (*Config, error) {
	mysqlCfg, err := newMySQLConfig()
	if err != nil {
		logger.WithoutContext().Fatalf("Loading MySQL Config FAILED!!: %+v", err)
		return nil, err
	}

	logger.WithoutContext().Infof("Config initialized: %s", env.GetString())
	return &Config{
		MySQL: mysqlCfg,
	}, nil
}

// Get Configを取得する
func Get() *Config {
	once.Do(func() {
		instance, err = newConfig()
	})
	return instance
}

// Err Configのerrorを取得する
func Err() error {
	Get()
	return err
}
