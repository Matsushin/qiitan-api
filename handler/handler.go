package handler

import (
	"context"

	"github.com/Matsushin/qiitan-api/cache"
	"github.com/Matsushin/qiitan-api/config"
	"github.com/Matsushin/qiitan-api/logger"
	"github.com/gin-gonic/gin"
)

// NewHandler engineを生成する
func NewHandler() *gin.Engine {
	r := gin.New()

	r.Use(config.NewContextByGin, logger.ReqID, logger.Logger)

	v1 := r.Group("/v1")
	pv := r.Group("/pvt")

	v1.GET("/articles/", V1GetArticles)
	v1.GET("/ranking/like", V1GetLikeRanking)
	pv.GET("/health", Health)

	ctx := context.Background()
	ctx = config.NewContext(ctx)
	go cache.InitCacheUpdateSchedule(ctx)

	return r
}
