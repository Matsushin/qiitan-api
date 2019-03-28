package handler

import (
	"context"
	"net/http"

	"github.com/Matsushin/qiitan-api/config"
	"github.com/Matsushin/qiitan-api/logger"
	like_ranking "github.com/Matsushin/qiitan-api/model/like_ranking"
	"github.com/Matsushin/qiitan-api/mysql"
	"github.com/Matsushin/qiitan-api/response"
	"github.com/gin-gonic/gin"
)

// V1GetLikeRanking いいね数の記事ランキングレポートを返却する
func V1GetLikeRanking(ctx *gin.Context) {

	cacheLikeRankingReports, err := like_ranking.GetLikeRankingCache(ctx)
	if err != nil {
		logger.Error(ctx, err)
		response.UnexpectedError.Respond(ctx)
		return
	}

	if len(cacheLikeRankingReports.LikeRankingList) > 0 {
		ctx.JSON(http.StatusOK, cacheLikeRankingReports)
		return
	}

	cfg, _ := config.FromContextByGin(ctx)
	db, _ := mysql.GetConnection(cfg)
	rows, err := like_ranking.GetLikeRanking(db)
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

	context := context.Background()
	context = config.NewContext(context)
	like_ranking.PutLikeRankingCache(context, likeRankingReports)

	ctx.JSON(http.StatusOK, likeRankingReports)
}
