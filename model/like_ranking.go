package like_ranking

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Matsushin/qiitan-api/cache"
	"github.com/Matsushin/qiitan-api/config"
	"github.com/Matsushin/qiitan-api/logger"
	"github.com/Matsushin/qiitan-api/mysql"
	"github.com/Matsushin/qiitan-api/response"
	as "github.com/aerospike/aerospike-client-go"
	"github.com/aerospike/aerospike-client-go/types"
	"github.com/gin-gonic/gin"
	"github.com/ugorji/go/codec"
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

// UpdateCache ...定期的にキャッシュ生成
func UpdateCache(ctx context.Context) {
	cfg := config.MustFromContext(ctx)
	go cache.InitCacheUpdateSchedule(ctx, cfg.UpdCache.LikeRankingSec, updateLikeRankingCache)
}

func updateLikeRankingCache(ctx context.Context, done chan<- cache.ChannelErrorResult) {
	cache.UpdateSchedule(ctx, done,
		func(ctx context.Context) error {
			cfg := config.MustFromContext(ctx)
			db, _ := mysql.GetConnection(cfg)
			rows, err := GetLikeRanking(db)
			if err != nil {
				logger.WithoutContext().Error(err)
				return err
			}
			defer rows.Close()

			var likeRankingList response.LikeRankingList
			for rows.Next() {
				var id int
				var title string
				var likeCount int

				err := rows.Scan(&id, &title, &likeCount)
				if err != nil {
					logger.WithoutContext().Error(err)
					return err
				}

				if !likeRankingList.Contains(id) {
					l := response.LikeRanking{ID: id, Title: title, LikeCount: likeCount}
					likeRankingList.AddLikeRanking(l)
				}
			}

			likeRankingReports := response.LikeRankingReports{LikeRankingList: likeRankingList}
			return PutLikeRankingCache(ctx, likeRankingReports)
		},
	)
}

// PutLikeRankingCache ... いいねランキングのキャッシュ保管
func PutLikeRankingCache(ctx context.Context, likeRankingReports response.LikeRankingReports) error {
	cfg := config.MustFromContext(ctx)
	var buf []byte
	mh := &codec.MsgpackHandle{RawToString: true}
	codec.NewEncoderBytes(&buf, mh).Encode(likeRankingReports)

	client, _ := as.NewClient(cfg.Aerospike.Server.Host, cfg.Aerospike.Server.Port)
	asCfg := cfg.Aerospike.LikeRankingDB
	newKey, err := as.NewKey(asCfg.Namespace, asCfg.Set, asCfg.Key)
	if err != nil {
		return err
	}
	bins := as.BinMap{"selialized": buf}

	err = client.Put(nil, newKey, bins)
	if err == nil {
		return nil
	}
	switch e := err.(type) {
	case types.AerospikeError:
		return fmt.Errorf("[LikeRanking] AerospikeError:%v, ResultCode:%v", e.Error(), e.ResultCode())
	default:
		return err
	}
}

// GetLikeRankingCache ...いいねランキングをキャッシュ読み取り
func GetLikeRankingCache(ctx *gin.Context) (response.LikeRankingReports, error) {
	cfg, _ := config.FromContextByGin(ctx)
	var likeRankingReports response.LikeRankingReports

	client, _ := as.NewClient(cfg.Aerospike.Server.Host, cfg.Aerospike.Server.Port)
	asCfg := cfg.Aerospike.LikeRankingDB
	newKey, err := as.NewKey(asCfg.Namespace, asCfg.Set, asCfg.Key)
	if err != nil {
		return likeRankingReports, err
	}
	res, err := client.Get(nil, newKey)
	if err != nil {
		return likeRankingReports, err
	}
	if res == nil {
		return likeRankingReports, errors.New("key not found")
	}

	mh := codec.MsgpackHandle{RawToString: true}
	codec.NewDecoderBytes(res.Bins["selialized"].([]uint8), &mh).Decode(&likeRankingReports)

	return likeRankingReports, nil
}
