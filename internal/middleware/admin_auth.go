package middleware

import (
	"CMS/internal/services"
	"CMS/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware 管理员权限验证中间件
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserIDFromContext(c)
		if userID == 0 {
			utils.JsonErrorWithCode(c, 401, "用户未认证")
			c.Abort()
			return
		}

		isAdmin, err := services.CheckUserIsAdmin(userID)
		if err != nil || !isAdmin {
			utils.JsonErrorWithCode(c, 403, "权限不足，需要管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}
