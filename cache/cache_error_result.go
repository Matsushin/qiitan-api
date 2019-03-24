package cache

// ChannelErrorResult ... goroutineチャネル結果構造体
type ChannelErrorResult struct {
	Err     error
	Panic   bool
	Message string
}

// NewChannelResult ... CacheChannelチャネル生成
func NewChannelErrorResult() chan ChannelErrorResult {
	return make(chan ChannelErrorResult)
}
