package logger

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ctxKeyRequestID = "requestID"
)

// ReqID ginのミドルウェア用の関数
func ReqID(ctx *gin.Context) {
	id, err := uuid.NewRandom()
	if err != nil {
		WithoutContext().Errorf("Generating RequestID FAILED!!: %+v", err)
		// 失敗しないはずだが、もし失敗したら、ゼロ値で処理継続する
	}
	requestID := id.String()

	ctx.Set(ctxKeyRequestID, requestID)
	ctx.Header("X-Uzou-Request-Id", requestID)

	ctx.Next()
}

// GetReqID ctxからrequestIDを取得
func GetReqID(ctx *gin.Context) (string, bool) {
	raw, ok := ctx.Get(ctxKeyRequestID)
	if !ok || raw == nil {
		WithoutContext().Errorf("ContextからRequestIDを取得できません（不在）。")
		return "", false
	}
	ret, ok := raw.(string)
	if !ok {
		WithoutContext().Errorf("ContextからRequestIDを取得できません（型の不一致）。")
		return "", false
	}

	return ret, true
}
