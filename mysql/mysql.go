package mysql

import (
	"database/sql"
	"fmt"
	"sync"

	// MySQLのドライバはdefault import
	"github.com/Matsushin/qiitan-api/config"
	"github.com/Matsushin/qiitan-api/logger"
	_ "github.com/go-sql-driver/mysql"
)

var (
	connection     *sql.DB
	connectionOnce sync.Once
	connectionErr  error
)

// GetConnection mysqlのコネクションオブジェクト取得
func GetConnection() *sql.DB {
	connectionOnce.Do(func() {
		cfg := config.Get().MySQL
		url := fmt.Sprintf("%s:%s@(%s:%d)/%s?parseTime=true&loc=UTC", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

		connection, connectionErr = sql.Open("mysql", url)
		if connectionErr != nil {
			logger.WithoutContext().Fatalf("MySQL connection initialization FAILED!!: %+v", connectionErr)
			return
		}

		logger.WithoutContext().Infof("MySQL connection initialized: %s:%d/%s", cfg.Host, cfg.Port, cfg.Database)
	})
	return connection
}

// ConnectionErr コネクションのエラーを取得
func ConnectionErr() error {
	GetConnection()
	return connectionErr
}
