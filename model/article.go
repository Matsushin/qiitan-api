package model

import (
	"database/sql"

	"github.com/Matsushin/qiitan-api/mysql"
)

// GetLikeRanking いいねランキングをDBから取得する
func GetLikeRanking() (*sql.Rows, error) {
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
ORDER BY like_count desc, id asc
LIMIT 20
`
	rows, err := db.Query(q)
	return rows, err
}
