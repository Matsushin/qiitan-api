package response

// ArticlesReports はGET v1/articlesのレスポンス
type ArticlesReports struct {
	Articles Articles `json:"articles"`
}

// Articles はArticleのスライス
type Articles []Article

// Contains articleが含まれているかを検査するメソッド
func (as Articles) Contains(articleID int) bool {
	for _, a := range as {
		if a.ID == articleID {
			return true
		}
	}
	return false
}

// AddArticle articlesにarticleを追加する
func (as *Articles) AddArticle(a Article) {
	*as = append(*as, a)
}

// Articleの構造体
type Article struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	LikeCount  int    `json:"like_count"`
	StockCount int    `json:"stock_count"`
}
