package middleware

import (
	"CMS/internal/logger"
	"CMS/internal/models"
	"CMS/pkg/utils"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
)

// GlobalErrorHandler 全局错误处理中间件（优化版）
func GlobalErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行后续中间件和业务逻辑
		c.Next()

		// 处理主动返回的错误（通过 c.Error 传递的错误）
		if len(c.Errors) > 0 {
			handleBusinessErrors(c)
			return
		}

		// 处理 panic 错误
		defer func() {
			if err := recover(); err != nil {
				handlePanicError(c, err)
			}
		}()
	}
}

// 处理业务逻辑中主动返回的错误
func handleBusinessErrors(c *gin.Context) {
	// 获取最后一个错误（通常是最关键的错误）
	err := c.Errors.Last()
	if err == nil {
		return
	}

	// 提取请求上下文信息（用于日志）
	reqInfo := getRequestInfo(c)
	logger.GetLogger().Errorf("[请求错误] %s, 错误详情: %v", reqInfo, err.Err)

	// 区分错误类型，返回对应响应
	switch e := err.Err.(type) {
	case *models.ServiceError:
		// 业务层自定义错误（已知错误）
		utils.JsonResponse(c, http.StatusOK, e.Code, e.Message, nil)
	case *gin.Error:
		// Gin 框架错误（如参数绑定失败）
		msg := "参数验证失败: " + e.Error()
		// 简化错误信息（去除内部堆栈）
		if idx := strings.Index(msg, "\n"); idx != -1 {
			msg = msg[:idx]
		}
		utils.JsonResponse(c, http.StatusBadRequest, 400, msg, nil)
	default:
		// 其他未知错误（避免暴露敏感信息）
		utils.JsonResponse(c, http.StatusInternalServerError, 500, "服务器内部错误", nil)
	}
	c.Abort() // 终止后续响应处理
}

// 处理 panic 错误
func handlePanicError(c *gin.Context, err interface{}) {
	// 提取请求上下文和堆栈信息
	reqInfo := getRequestInfo(c)
	stack := string(debug.Stack())
	logger.GetLogger().Errorf("[Panic错误] %s, 错误详情: %v\n堆栈信息: %s", reqInfo, err, stack)

	// 返回统一的内部错误响应
	utils.JsonResponse(c, http.StatusInternalServerError, 500, "服务器内部错误", nil)
	c.Abort()
}

// 获取请求上下文信息（方便日志排查）
func getRequestInfo(c *gin.Context) string {
	userID := GetUserIDFromContext(c) // 复用现有获取用户ID的方法
	return strings.Join([]string{
		"method=" + c.Request.Method,
		"path=" + c.Request.URL.Path,
		"user_id=" + string(rune(userID)),
		"ip=" + c.ClientIP(),
	}, ", ")
}
