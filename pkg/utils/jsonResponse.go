package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 统一响应格式
type APIResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

// JsonResponse 通用响应函数
func JsonResponse(c *gin.Context, httpCode, code int, msg string, data interface{}) {
	c.JSON(httpCode, APIResponse{
		Code: code,
		Data: data,
		Msg:  msg,
	})
}

func JsonSuccessWithCode(c *gin.Context, code int, data interface{}) {
	JsonResponse(c, http.StatusOK, code, "success", data)
}

func JsonErrorWithCode(c *gin.Context, code int, message string) {
	JsonResponse(c, http.StatusOK, code, message, nil)
}

func JsonValidationError(c *gin.Context, message string) {
	JsonResponse(c, http.StatusBadRequest, 200, message, nil)
}
