package cache

import (
	"errors"
	"fmt"
	"time"

	"github.com/Matsushin/qiitan-api/logger"
	"github.com/Matsushin/qiitan-api/model"
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
func InitCacheUpdateSchedule() {
	updateScheduleSec := 60 * time.Second
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
		updateCache(ticker)
	}
}

func updateCache(ticker *cacheTicker) {
	done := NewChannelErrorResult()

	go updateCacheFunc(done)

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

func updateCacheFunc(done chan<- ChannelErrorResult) {
	err := updateCacheLikeRanking()
	done <- ChannelErrorResult{Err: err, Panic: false}
}

func updateCacheLikeRanking() error {
	rows, err := model.GetLikeRanking()
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
	return PutLikeRanking(likeRankingReports)

}

// PutLikeRanking ... いいねランキングのキャッシュ保管
func PutLikeRanking(likeRankingReports response.LikeRankingReports) error {
	var buf []byte
	mh := &codec.MsgpackHandle{RawToString: true}
	codec.NewEncoderBytes(&buf, mh).Encode(likeRankingReports)

	client, _ := as.NewClient("127.0.0.1", 3000)
	key, err := as.NewKey("test", "testset", "likeRanking")
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
func GetLikeRanking() (response.LikeRankingReports, error) {
	var likeRankingReports response.LikeRankingReports
	client, _ := as.NewClient("127.0.0.1", 3000)
	key, err := as.NewKey("test", "testset", "likeRanking")
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
