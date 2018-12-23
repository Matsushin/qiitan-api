package handler

import (
	"github.com/Matsushin/qiitan-api/logger"
	"github.com/gin-gonic/gin"
)

// NewHandler engineを生成する
func NewHandler() *gin.Engine {
	r := gin.New()
	r.Use(logger.Logger)

	v1 := r.Group("/v1")

	v1.GET("/articles/", V1GetArticles)

	return r
}
