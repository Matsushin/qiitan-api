package logger

import (
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	ctxKeyLogger = "logger"
)

var (
	defaultLogger logrus.FieldLogger
	once          sync.Once
)

// Base ctxがない場合のlogger
type Base struct{}

// GetBase baseLoggerを取得
// Singleton Loggerを取得
// baseパッケージで取得できるLoggerはRequest単位のメタデータを含まないので、各アクションでは原則使わない
func getBase() logrus.FieldLogger {
	once.Do(func() {
		logger := logrus.New()

		logger.Formatter = &logrus.TextFormatter{
			FullTimestamp: true,
		}

		logLevel, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
		if err == nil {
			logger.SetLevel(logLevel)
		} else {
			logger.SetLevel(logrus.DebugLevel)
		}

		// logger.AddHook(&sentryHook{})
		defaultLogger = logger
	})

	return defaultLogger
}

// Get loggerを取得する関数
// Request単位のメタ情報が付与されたLoggerを取得
// もしMiddleware組み込み周りバグがあり、ContextからLoggerを取得できなかった場合は、BaseのLoggerを返す
func get(ctx *gin.Context) logrus.FieldLogger {
	raw, ok := ctx.Get(ctxKeyLogger)
	if !ok || raw == nil {
		base := getBase()
		base.Errorf("ContextからLoggerを取得できません（不在）。")
		return base
	}
	ret, ok := raw.(logrus.FieldLogger)
	if !ok {
		base := getBase()
		base.Errorf("ContextからLoggerを取得できません（型の不一致）。")
		return base
	}

	return ret
}

// WithoutContext ...
func WithoutContext() Base {
	return Base{}
}

// WithFields Fieldsを設定したloggerを取得する
func WithFields(ctx *gin.Context, fields logrus.Fields) *logrus.Entry {
	return get(ctx).WithFields(fields)
}

// Error errorログを出力
func Error(ctx *gin.Context, args ...interface{}) {
	get(ctx).Error(args...)
}

// Info infoログを出力
func Info(ctx *gin.Context, args ...interface{}) {
	get(ctx).Info(args...)
}

// Fatalf fatalfログを出力
func Fatalf(ctx *gin.Context, format string, args ...interface{}) {
	get(ctx).Fatalf(format, args...)
}

// Errorf errorfログを出力
func Errorf(ctx *gin.Context, format string, args ...interface{}) {
	get(ctx).Errorf(format, args...)
}

// Infof infofログを出力
func Infof(ctx *gin.Context, format string, args ...interface{}) {
	get(ctx).Infof(format, args...)
}

// Logger gin用のMiddleware関数
func Logger(ctx *gin.Context) {
	// Start timer
	start := time.Now()

	sharedLogger := WithoutContext().WithFields(logrus.Fields{
		"requestID": "123456789",
	})
	ctx.Set(ctxKeyLogger, sharedLogger)

	// Opening Log
	clientIP := ctx.ClientIP()
	method := ctx.Request.Method
	path := ctx.Request.URL.Path
	raw := ctx.Request.URL.RawQuery
	if raw != "" {
		path = path + "?" + raw
	}

	startStopLogger := sharedLogger.WithFields(logrus.Fields{
		"clientIP": clientIP,
		"method":   method,
		"path":     path,
	})
	startStopLogger.Info("Request Start.")

	// Process request
	ctx.Next()

	// Stop timer
	end := time.Now()
	responseTime := end.Sub(start)
	statusCode := ctx.Writer.Status()

	// Closing Log
	stopLogger := startStopLogger.WithFields(logrus.Fields{
		"status":       statusCode,
		"responseTime": responseTime,
	})

	stopLogger.Info("Request End.")
}

// WithFields Fieldsを設定したBasegerを取得する
func (b Base) WithFields(fields logrus.Fields) *logrus.Entry {
	return getBase().WithFields(fields)
}

// Info infoログを出力
func (b Base) Info(args ...interface{}) {
	getBase().Info(args...)
}

// Fatalf fatalfログを出力
func (b Base) Fatalf(format string, args ...interface{}) {
	getBase().Fatalf(format, args...)
}

// Infof infofログを出力
func (b Base) Infof(format string, args ...interface{}) {
	getBase().Infof(format, args...)
}
