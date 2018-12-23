package response

import (
	"net/http"

	"github.com/Matsushin/qiitan-api/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Error はエラーレスポンスの構造体
type Error struct {
	StatusCode    int    `json:"-"`
	ErrorCode     string `json:"error_code"`
	Message       string `json:"message"`
	IsQiitanError bool   `json:"-"`
}

var (
	ctxKeyError = "error"
	// E3xxx: 認証時の異常
	// E4xxx: リクエスト送信側の異常
	// E5xxx: UZOU側の異常

	// ForbiddenError 認証失敗時のエラーレスポンス
	ForbiddenError = Error{http.StatusForbidden, "E3001", "認証に失敗しました。", false}

	// BadRequestError リクエスト不正時のエラーレスポンス
	BadRequestError = Error{http.StatusBadRequest, "E4001", "リクエストパラメータが不正です。", false}

	// NotFoundError データがない時のエラーレスポンス
	NotFoundError = Error{http.StatusNotFound, "E4002", "データが存在しません。", false}

	// UnexpectedError UZOU側のシステム異常時のエラーレスポンス
	UnexpectedError = Error{http.StatusInternalServerError, "E5999", "想定外のエラーが発生しました。", true}
)

// Respond ctx.JSONを行う
func (e *Error) Respond(ctx *gin.Context) {
	ctx.Set(ctxKeyError, e)
	ctx.JSON(e.StatusCode, e)
}

// ErrorLog エラーをロギングするためのミドルウェア
func ErrorLog(ctx *gin.Context) {
	ctx.Next()

	err, ok := GetError(ctx)
	if !ok {
		return
	}

	if err.IsQiitanError {
		logger.WithFields(ctx, logrus.Fields{
			"errorCode": err.ErrorCode,
			"errorMsg":  err.Message,
		}).Errorf("Error occurred.")
	} else {
		logger.WithFields(ctx, logrus.Fields{
			"errorCode": err.ErrorCode,
			"errorMsg":  err.Message,
		}).Infof("Invalid request.")
	}
}

// GetError ctxからErrorを取得する
func GetError(ctx *gin.Context) (*Error, bool) {
	raw, ok := ctx.Get(ctxKeyError)
	if !ok || raw == nil {
		// エラーが発生していなければnilが返る
		return nil, false
	}
	ret, ok := raw.(*Error)
	if !ok {
		logger.Errorf(ctx, "ContextからErrorを取得できません（型の不一致）。")
		return nil, false
	}

	return ret, true
}
