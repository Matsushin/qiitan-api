package handler

import (
	"github.com/Matsushin/qiitan-api/cache"
	"github.com/Matsushin/qiitan-api/logger"
	"github.com/gin-gonic/gin"
)

// NewHandler engineを生成する
func NewHandler() *gin.Engine {
	r := gin.New()
	r.Use(logger.ReqID, logger.Logger)

	v1 := r.Group("/v1")

	v1.GET("/articles/", V1GetArticles)
	v1.GET("/ranking/like", V1GetLikeRanking)

	go cache.InitCacheUpdateSchedule()
	return r
}
