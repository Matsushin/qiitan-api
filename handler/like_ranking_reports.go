package handler

import (
	"net/http"

	"github.com/Matsushin/qiitan-api/cache"
	"github.com/Matsushin/qiitan-api/logger"
	"github.com/Matsushin/qiitan-api/model"
	"github.com/Matsushin/qiitan-api/response"
	"github.com/gin-gonic/gin"
)

// V1GetLikeRanking いいね数の記事ランキングレポートを返却する
func V1GetLikeRanking(ctx *gin.Context) {

	cacheLikeRankingReports, err := cache.GetLikeRanking()
	if err != nil {
		logger.Error(ctx, err)
		response.UnexpectedError.Respond(ctx)
		return
	}

	if len(cacheLikeRankingReports.LikeRankingList) > 0 {
		ctx.JSON(http.StatusOK, cacheLikeRankingReports)
		return
	}

	rows, err := model.GetLikeRanking()
	if err != nil {
		logger.Info(ctx, err)
		response.UnexpectedError.Respond(ctx)
		return
	}
	defer rows.Close()

	var likeRankingList response.LikeRankingList
	for rows.Next() {
		var id int
		var title string
		var likeCount int

		err := rows.Scan(&id, &title, &likeCount)
		if err != nil {
			logger.Error(ctx, err)
			response.UnexpectedError.Respond(ctx)
			return
		}

		if !likeRankingList.Contains(id) {
			l := response.LikeRanking{ID: id, Title: title, LikeCount: likeCount}
			likeRankingList.AddLikeRanking(l)
		}
	}

	likeRankingReports := response.LikeRankingReports{LikeRankingList: likeRankingList}
	cache.PutLikeRanking(likeRankingReports)

	ctx.JSON(http.StatusOK, likeRankingReports)
}
