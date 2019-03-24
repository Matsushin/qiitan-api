package config

import "time"

// UpdCache ...キャッシュ更新頻度情報
type UpdCache struct {
	LikeRankingSec time.Duration
}
