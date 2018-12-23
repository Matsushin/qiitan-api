package handler

import (
	"net/http"

	"github.com/Matsushin/qiitan-api/logger"
	"github.com/Matsushin/qiitan-api/mysql"
	"github.com/Matsushin/qiitan-api/response"
	"github.com/gin-gonic/gin"
)

// V1GetArticles プレイスメントごとのレポートを返却する
func V1GetArticles(ctx *gin.Context) {

	db := mysql.GetConnection()

	q := `
SELECT 
  a.id AS id, 
  a.title AS title,
  count(l.id) AS like_count
FROM articles AS a
LEFT OUTER JOIN 
 likes AS l
 ON a.id = l.article_id
GROUP BY a.id
ORDER BY a.id
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

		err := rows.Scan(&id, &title, &likeCount)
		if err != nil {
			logger.Error(ctx, err)
			response.UnexpectedError.Respond(ctx)
			return
		}

		if !articles.Contains(id) {
			a := response.Article{ID: id, Title: title, LikeCount: likeCount}
			articles.AddArticle(a)
		}
	}

	articlesReports := response.ArticlesReports{Articles: articles}
	ctx.JSON(http.StatusOK, articlesReports)
}
