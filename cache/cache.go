package cache

import (
	"context"
	"errors"
	"time"

	"github.com/Matsushin/qiitan-api/logger"
)

type cacheTicker struct {
	t        *time.Ticker
	deadline time.Duration
}

// InitCacheUpdateSchedule ... キャッシュ自動更新スケジュール登録処理
func InitCacheUpdateSchedule(
	ctx context.Context,
	updateScheduleSec time.Duration,
	updateCacheFunc func(ctx context.Context, done chan<- ChannelErrorResult)) {
	ticker := &cacheTicker{
		t:        time.NewTicker(updateScheduleSec * time.Second),
		deadline: updateScheduleSec * time.Second / 2,
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
		updateCache(ctx, ticker, updateCacheFunc)
	}
}

func updateCache(
	ctx context.Context,
	ticker *cacheTicker,
	updateCacheFunc func(ctx context.Context, done chan<- ChannelErrorResult)) {

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

func UpdateSchedule(
	ctx context.Context,
	done chan<- ChannelErrorResult,
	updateFunc func(ctx context.Context) error) {

	// panic発生時にキャッシュ更新フラグのチャネルを閉じてTickerチャネルを閉じるrecoverにpanicを伝播させる
	defer func(done chan<- ChannelErrorResult) {
		if err := recover(); err != nil {
			done <- ChannelErrorResult{Err: errors.New("Cache update faild."), Panic: true}
		}
	}(done)

	//err := updateCacheLikeRanking(ctx)
	err := updateFunc(ctx)

	done <- ChannelErrorResult{Err: err, Panic: false}
}
