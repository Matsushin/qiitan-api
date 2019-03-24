package handler

import (
	"net/http"

	"github.com/Matsushin/qiitan-api/config"
	"github.com/Matsushin/qiitan-api/mysql"
	"github.com/gin-gonic/gin"
)

// Health ヘルスチェック用のハンドラ
func Health(ctx *gin.Context) {
	if err := config.Err(); err != nil {
		ctx.String(http.StatusServiceUnavailable, "config: error: %s", err)
		return
	}
	if err := mysql.ConnectionErr(ctx); err != nil {
		ctx.String(http.StatusServiceUnavailable, "mysql connection error: %s", err)
		return
	}

	ctx.String(http.StatusOK, "OK")
}
