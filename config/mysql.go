package config

import (
	"os"

	"github.com/Matsushin/qiitan-api/env"
)

// MySQLConfig MySQLの設定を管理する構造体
type MySQLConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
}

func newMySQLConfig() (*MySQLConfig, error) {
	if env.IsLocal() {
		return &MySQLConfig{
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Host:     os.Getenv("DB_HOST"),
			Port:     3306,
			Database: "qiitan_development",
		}, nil
	}

	if env.IsTest() {
		return &MySQLConfig{
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Host:     os.Getenv("DB_HOST"),
			Port:     3306,
			Database: "qiitan_test",
		}, nil
	}

	return nil, nil
}
