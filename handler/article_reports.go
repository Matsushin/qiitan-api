package handler

import (
	"net/http"

	"github.com/Matsushin/qiitan-api/logger"
	"github.com/Matsushin/qiitan-api/mysql"
	"github.com/Matsushin/qiitan-api/response"
	"github.com/gin-gonic/gin"
)

// V1GetArticles 記事一覧のレポートを返却する
func V1GetArticles(ctx *gin.Context) {

	db := mysql.GetConnection()

	q := `
SELECT 
  a.id AS id, 
  a.title AS title,
  a.user_id AS user_id,
  u.username AS username
FROM articles AS a
LEFT OUTER JOIN users AS u ON a.user_id = u.id 
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
		var userID int
		var username string

		err := rows.Scan(&id, &title, &userID, &username)
		if err != nil {
			logger.Error(ctx, err)
			response.UnexpectedError.Respond(ctx)
			return
		}

		if !articles.Contains(id) {
			a := response.Article{ID: id, Title: title, UserID: userID, Username: username}
			articles.AddArticle(a)
		}
	}

	articlesReports := response.ArticlesReports{Articles: articles}
	ctx.JSON(http.StatusOK, articlesReports)
}
