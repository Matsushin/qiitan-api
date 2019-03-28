package response

// StockRankingList はStockRankingのスライス
type StockRankingList []StockRanking

// StockRankingReports はGET v1/ranking/stock
type StockRankingReports struct {
	StockRankingList StockRankingList `json:"articles"`
}

// Contains stockRankingが含まれているかを検査するメソッド
func (sr StockRankingList) Contains(articleID int) bool {
	for _, a := range sr {
		if a.ID == articleID {
			return true
		}
	}
	return false
}

// AddStockRanking stock_ranking_listにstock_rankingを追加する
func (sr *StockRankingList) AddStockRanking(s StockRanking) {
	*sr = append(*sr, s)
}

// StockRanking の構造体
type StockRanking struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	StockCount int    `json:"stock_count"`
}
