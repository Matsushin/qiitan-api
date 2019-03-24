package config

import (
	"context"
	"io/ioutil"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/Matsushin/qiitan-api/logger"
	"github.com/gin-gonic/gin"
)

// Config APIのConfigを管理する構造体
type Config struct {
	MySQL     MySQLConfig
	Aerospike AerospikeConfig
	UpdCache  UpdCache
}

// ctxKeyConfig ... コンフィグキャッシュキー
type ctxKeyConfig struct{}

const (
	// ConfigEnvKey ...コンフィグ環境変数
	ConfigEnvKey = "QIITAN_API_CONFIG"
	ctxConfigKey = "qiitan_api_config"
)

var (
	instance *Config
	once     sync.Once
	err      error
)

// NewConfig ...コンフィグ作成
func NewConfig(fpath string) (*Config, error) {
	cfg := &Config{}

	contents, err := replaceEnvFile(fpath)
	if err != nil {
		logger.WithoutContext().Fatalf("Loading MySQL Config FAILED!!: %+v", err)
		return nil, err
	}
	_, err = toml.Decode(contents, cfg)
	if err != nil {
		logger.WithoutContext().Fatalf("Loading MySQL Config FAILED!!: %+v", err)
		return nil, err
	}
	return cfg, err
}

// NewContext ...コンテキスト作成
func NewContext(ctx context.Context) context.Context {
	cfg, err := NewConfig(os.Getenv(ConfigEnvKey))
	if err != nil {
		panic(err)
	}
	return ToContext(ctx, cfg)
}

// NewContextByGin Ginで使うコンテキストに設定情報を紐付ける
func NewContextByGin(ctx *gin.Context) {
	cfg, err := NewConfig(os.Getenv(ConfigEnvKey))
	if err != nil {
		panic(err)
	}
	ctx.Set(ctxConfigKey, cfg)
}

// ToContext ...コンテキストにキャッシュを紐付ける
func ToContext(ctx context.Context, conf *Config) context.Context {
	return context.WithValue(ctx, ctxKeyConfig{}, conf)
}

// FromContext ... コンテキスト取得 Option
func FromContext(ctx context.Context) (*Config, bool) {
	conf, ok := ctx.Value(ctxKeyConfig{}).(*Config)
	return conf, ok
}

// FromContextByGin ... Ginのコンテキストから設定情報取得
func FromContextByGin(ctx *gin.Context) (*Config, bool) {
	raw, ok := ctx.Get(ctxConfigKey)
	if !ok || raw == nil {
		logger.WithoutContext().Errorf("Contextから設定情報を取得できません（不在）。")
		return nil, false
	}
	conf, ok := raw.(*Config)
	if !ok {
		logger.WithoutContext().Errorf("Contextから設定情報を取得できません（型の不一致）。")
		return nil, false
	}
	return conf, ok
}

// MustFromContext ... コンテキスト取得 取得できなければエラー
func MustFromContext(ctx context.Context) *Config {
	conf, ok := FromContext(ctx)
	if !ok {
		panic("ContextからConfigが取得できない。") // そもそもプログラムが正常なら起動時に設定するのでここは通らないはず。(起動時Contextに入れる際に失敗したらpanicにしている)
	}
	return conf
}

// 環境変数が見つからないとき空の文字列 "" が入る。(その後のRDS接続で補足されきちんとpanicしてくれるはず)
func replaceEnvFile(path string) (string, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return os.ExpandEnv(string(contents)), nil
}

// Get Configを取得する
func Get() *Config {
	once.Do(func() {
		instance, err = NewConfig(os.Getenv(ConfigEnvKey))
	})
	return instance
}

// Err Configのerrorを取得する
func Err() error {
	Get()
	return err
}
