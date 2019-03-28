package handler

import (
	"context"
	"net/http"

	"github.com/Matsushin/qiitan-api/config"
	"github.com/Matsushin/qiitan-api/logger"
	stock_ranking "github.com/Matsushin/qiitan-api/model/stock_ranking"
	"github.com/Matsushin/qiitan-api/mysql"
	"github.com/Matsushin/qiitan-api/response"
	"github.com/gin-gonic/gin"
)

// V1GetStockRanking ストック数の記事ランキングレポートを返却する
func V1GetStockRanking(ctx *gin.Context) {

	cacheStockRankingReports, err := stock_ranking.GetStockRankingCache(ctx)
	if err != nil {
		logger.Error(ctx, err)
		response.UnexpectedError.Respond(ctx)
		return
	}

	if len(cacheStockRankingReports.StockRankingList) > 0 {
		ctx.JSON(http.StatusOK, cacheStockRankingReports)
		return
	}

	cfg, _ := config.FromContextByGin(ctx)
	db, _ := mysql.GetConnection(cfg)
	rows, err := stock_ranking.GetStockRanking(db)
	if err != nil {
		logger.Info(ctx, err)
		response.UnexpectedError.Respond(ctx)
		return
	}
	defer rows.Close()

	var stockRankingList response.StockRankingList
	for rows.Next() {
		var id int
		var title string
		var stockCount int

		err := rows.Scan(&id, &title, &stockCount)
		if err != nil {
			logger.Error(ctx, err)
			response.UnexpectedError.Respond(ctx)
			return
		}

		if !stockRankingList.Contains(id) {
			l := response.StockRanking{ID: id, Title: title, StockCount: stockCount}
			stockRankingList.AddStockRanking(l)
		}
	}

	stockRankingReports := response.StockRankingReports{StockRankingList: stockRankingList}

	context := context.Background()
	context = config.NewContext(context)
	stock_ranking.PutStockRankingCache(context, stockRankingReports)

	ctx.JSON(http.StatusOK, stockRankingReports)
}
