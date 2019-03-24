package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Matsushin/qiitan-api/config"
	"github.com/Matsushin/qiitan-api/logger"
	"github.com/Matsushin/qiitan-api/model"
	"github.com/Matsushin/qiitan-api/mysql"
	"github.com/Matsushin/qiitan-api/response"
	as "github.com/aerospike/aerospike-client-go"
	"github.com/aerospike/aerospike-client-go/types"
	"github.com/ugorji/go/codec"
)

type cacheTicker struct {
	t        *time.Ticker
	deadline time.Duration
}

// InitCacheUpdateSchedule ... キャッシュ自動更新スケジュール登録処理
func InitCacheUpdateSchedule(ctx context.Context) {
	cfg := config.MustFromContext(ctx)
	updateScheduleSec := cfg.UpdCache.LikeRankingSec * time.Second
	ticker := &cacheTicker{
		t:        time.NewTicker(updateScheduleSec),
		deadline: updateScheduleSec / 2,
	}

	defer func(ticker *cacheTicker) {
		if err := recover(); err != nil {
			logger.WithoutContext().Error("キャッシュ更新に異常が発生したため定期更新を停止します。修正後再デプロイしてください。 ")
			ticker.t.Stop()
		}
		logger.WithoutContext().Info("キャッシュの更新を停止しました - " + time.Now().String())
	}(ticker)

	for tt := range ticker.t.C {
		logger.WithoutContext().Info("キャッシュの更新を開始しました - " + tt.String())
		updateCache(ctx, ticker)
	}
}

func updateCache(ctx context.Context, ticker *cacheTicker) {
	done := NewChannelErrorResult()

	go updateCacheFunc(ctx, done)

	select {
	case cacheChannel := <-done: // キャッシュ更新成功
		if cacheChannel.Err == nil {
			logger.WithoutContext().Info("キャッシュの更新に成功しました - " + time.Now().String())
		} else {
			if cacheChannel.Panic {
				close(done)
				panic(cacheChannel.Err) // サブgoroutineから受け取ったpanicをメインgoroutineに伝播させる
			}
			logger.WithoutContext().Error("キャッシュの更新に失敗しました ")
		}
	case <-time.After(ticker.deadline): // deadlineまでにdoneチャネルが終了しない
		logger.WithoutContext().Error("キャッシュの更新が時間内に終了しませんでした")
		<-done
	}

	close(done)
}

func updateCacheFunc(ctx context.Context, done chan<- ChannelErrorResult) {
	err := updateCacheLikeRanking(ctx)
	done <- ChannelErrorResult{Err: err, Panic: false}
}

func updateCacheLikeRanking(ctx context.Context) error {
	cfg := config.MustFromContext(ctx)
	db, _ := mysql.GetConnection(cfg)
	rows, err := model.GetLikeRanking(db)
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
	return PutLikeRanking(ctx, likeRankingReports)

}

// PutLikeRanking ... いいねランキングのキャッシュ保管
func PutLikeRanking(ctx context.Context, likeRankingReports response.LikeRankingReports) error {
	cfg := config.MustFromContext(ctx)
	var buf []byte
	mh := &codec.MsgpackHandle{RawToString: true}
	codec.NewEncoderBytes(&buf, mh).Encode(likeRankingReports)

	client, _ := as.NewClient(cfg.Aerospike.Server.Host, cfg.Aerospike.Server.Port)
	key, err := as.NewKey(cfg.Aerospike.LikeRankingDB.Namespace, cfg.Aerospike.LikeRankingDB.Set, cfg.Aerospike.LikeRankingDB.Key)
	if err != nil {
		return err
	}
	bins := as.BinMap{"selialized": buf}

	err = client.Put(nil, key, bins)
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

// GetLikeRanking ... いいねランキングキャッシュ取得
func GetLikeRanking(ctx *gin.Context) (response.LikeRankingReports, error) {
	cfg, _ := config.FromContextByGin(ctx)
	var likeRankingReports response.LikeRankingReports
	client, _ := as.NewClient(cfg.Aerospike.Server.Host, cfg.Aerospike.Server.Port)
	key, err := as.NewKey(cfg.Aerospike.LikeRankingDB.Namespace, cfg.Aerospike.LikeRankingDB.Set, cfg.Aerospike.LikeRankingDB.Key)
	if err != nil {
		return likeRankingReports, err
	}
	res, err := client.Get(nil, key)
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
