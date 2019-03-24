package response

// LikeRankingList はLikeRankingのスライス
type LikeRankingList []LikeRanking

// LikeRankingReports はGET v1/ranking/like
type LikeRankingReports struct {
	LikeRankingList LikeRankingList `json:"articles"`
}

// Contains likeRankingが含まれているかを検査するメソッド
func (lr LikeRankingList) Contains(articleID int) bool {
	for _, a := range lr {
		if a.ID == articleID {
			return true
		}
	}
	return false
}

// AddLikeRanking like_ranking_listにlike_rankingを追加する
func (lr *LikeRankingList) AddLikeRanking(l LikeRanking) {
	*lr = append(*lr, l)
}

// LikeRanking の構造体
type LikeRanking struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	LikeCount int    `json:"like_count"`
}
