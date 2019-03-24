package model

import (
	"database/sql"
)

// GetLikeRanking いいねランキングをDBから取得する
func GetLikeRanking(db *sql.DB) (*sql.Rows, error) {
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
