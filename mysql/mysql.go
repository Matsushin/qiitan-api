package mysql

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	// MySQLのドライバはdefault import
	"github.com/Matsushin/qiitan-api/config"
	"github.com/Matsushin/qiitan-api/logger"
)

var (
	connection     *sql.DB
	connectionOnce sync.Once
	connectionErr  error
)

type ctxKeyMysql struct{}

// GetConnection mysqlのコネクションオブジェクト取得
func GetConnection(cfg *config.Config) (*sql.DB, error) {
	connectionOnce.Do(func() {
		url := fmt.Sprintf("%s:%s@(%s:%d)/%s?parseTime=true&loc=UTC", cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)

		connection, connectionErr = sql.Open("mysql", url)
		if connectionErr != nil {
			logger.WithoutContext().Fatalf("MySQL connection initialization FAILED!!: %+v", connectionErr)
			return
		}

		logger.WithoutContext().Infof("MySQL connection initialized: %s:%d/%s", cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)
	})
	return connection, connectionErr
}

// ConnectionErr コネクションのエラーを取得
func ConnectionErr(ctx *gin.Context) error {
	cfg, _ := config.FromContextByGin(ctx)
	_, connectionErr = GetConnection(cfg)
	return connectionErr
}
