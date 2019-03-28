package stock_ranking

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

// GetStockRanking ストックランキングをDBから取得する
func GetStockRanking(db *sql.DB) (*sql.Rows, error) {
	q := `
SELECT 
  a.id AS id, 
  a.title AS title,
  count(s.id) AS stock_count
FROM articles AS a
LEFT OUTER JOIN 
  stocks AS s
  ON a.id = s.article_id
GROUP BY a.id
ORDER BY stock_count desc, id asc
LIMIT 20
`
	rows, err := db.Query(q)
	return rows, err
}

// UpdateCache ...定期的にキャッシュ生成
func UpdateCache(ctx context.Context) {
	cfg := config.MustFromContext(ctx)
	go cache.InitCacheUpdateSchedule(ctx, cfg.UpdCache.StockRankingSec, "StockRanking", updateStockRankingCache)
}

func updateStockRankingCache(ctx context.Context, done chan<- cache.ChannelErrorResult) {
	cache.UpdateSchedule(ctx, done,
		func(ctx context.Context) error {
			cfg := config.MustFromContext(ctx)
			db, _ := mysql.GetConnection(cfg)
			rows, err := GetStockRanking(db)
			if err != nil {
				logger.WithoutContext().Error(err)
				return err
			}
			defer rows.Close()

			var stockRankingList response.StockRankingList
			for rows.Next() {
				var id int
				var title string
				var stockCount int

				err := rows.Scan(&id, &title, &stockCount)
				if err != nil {
					logger.WithoutContext().Error(err)
					return err
				}

				if !stockRankingList.Contains(id) {
					l := response.StockRanking{ID: id, Title: title, StockCount: stockCount}
					stockRankingList.AddStockRanking(l)
				}
			}

			stockRankingReports := response.StockRankingReports{StockRankingList: stockRankingList}
			return PutStockRankingCache(ctx, stockRankingReports)
		},
	)
}

// PutStockRankingCache ... ストックランキングのキャッシュ保管
func PutStockRankingCache(ctx context.Context, stockRankingReports response.StockRankingReports) error {
	cfg := config.MustFromContext(ctx)
	var buf []byte
	mh := &codec.MsgpackHandle{RawToString: true}
	codec.NewEncoderBytes(&buf, mh).Encode(stockRankingReports)

	client, _ := as.NewClient(cfg.Aerospike.Server.Host, cfg.Aerospike.Server.Port)
	asCfg := cfg.Aerospike.StockRankingDB
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
		return fmt.Errorf("[StockRanking] AerospikeError:%v, ResultCode:%v", e.Error(), e.ResultCode())
	default:
		return err
	}
}

// GetStockRankingCache ...ストックランキングをキャッシュ読み取り
func GetStockRankingCache(ctx *gin.Context) (response.StockRankingReports, error) {
	cfg, _ := config.FromContextByGin(ctx)
	var stockRankingReports response.StockRankingReports

	client, _ := as.NewClient(cfg.Aerospike.Server.Host, cfg.Aerospike.Server.Port)
	asCfg := cfg.Aerospike.StockRankingDB
	newKey, err := as.NewKey(asCfg.Namespace, asCfg.Set, asCfg.Key)
	if err != nil {
		return stockRankingReports, err
	}
	res, err := client.Get(nil, newKey)
	if err != nil {
		return stockRankingReports, err
	}
	if res == nil {
		return stockRankingReports, errors.New("key not found")
	}

	mh := codec.MsgpackHandle{RawToString: true}
	codec.NewDecoderBytes(res.Bins["selialized"].([]uint8), &mh).Decode(&stockRankingReports)

	return stockRankingReports, nil
}
