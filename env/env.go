package env

import (
	"os"
	"strings"
	"sync"
)

// Env 環境
type Env string

const (
	// PRD 本番環境
	PRD Env = "prd"
	// LOCAL ローカル環境
	LOCAL Env = "local"
	// TEST テスト環境(ci)
	TEST Env = "test"
)

var (
	instance Env
	once     sync.Once
)

// Get 環境の情報を取得
func Get() Env {
	once.Do(func() {
		instance = Env(strings.ToLower(os.Getenv("ENV")))
		if instance == "" {
			instance = LOCAL
		}
	})
	return instance
}

// GetString 環境を文字列で返す
func GetString() string {
	return string(Get())
}

// IsProduction 本番環境の場合trueを返す
func IsProduction() bool {
	return Get() == PRD
}

// IsLocal ローカル環境の場合trueを返す
func IsLocal() bool {
	return Get() == LOCAL
}

// IsTest テスト環境の場合trueを返す
func IsTest() bool {
	return Get() == TEST
}
