package handler

import (
	"net/http"

	"github.com/Matsushin/qiitan-api/logger"
	"github.com/Matsushin/qiitan-api/mysql"
	"github.com/Matsushin/qiitan-api/response"
	"github.com/gin-gonic/gin"
)

// V1GetLikeRanking いいね数の記事ランキングレポートを返却する
func V1GetLikeRanking(ctx *gin.Context) {

	db := mysql.GetConnection()

	q := `
SELECT 
  a.id AS id, 
  a.title AS title,
  count(l.id) AS like_count,
  count(s.id) as stock_count
FROM articles AS a
LEFT OUTER JOIN 
  likes AS l
  ON a.id = l.article_id
LEFT OUTER JOIN 
  stocks AS s
  ON a.id = s.article_id
GROUP BY a.id
ORDER BY like_count desc
LIMIT 20
`
	rows, err := db.Query(q)
	if err != nil {
		logger.Info(ctx, err)
		response.UnexpectedError.Respond(ctx)
		return
	}
	defer rows.Close()

	var articles response.Articles
	for rows.Next() {
		var id int
		var title string
		var likeCount int
		var stockCount int

		err := rows.Scan(&id, &title, &likeCount, &stockCount)
		if err != nil {
			logger.Error(ctx, err)
			response.UnexpectedError.Respond(ctx)
			return
		}

		if !articles.Contains(id) {
			a := response.Article{ID: id, Title: title, LikeCount: likeCount, StockCount: stockCount}
			articles.AddArticle(a)
		}
	}

	articlesReports := response.ArticlesReports{Articles: articles}
	ctx.JSON(http.StatusOK, articlesReports)
}
