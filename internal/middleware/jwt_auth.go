package middleware

import (
	"CMS/internal/models"
	"CMS/internal/pkg/database"
	"CMS/pkg/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("cms_secret_key")

type Claims struct {
	UserID   uint `json:"user_id"`
	UserType int  `json:"user_type"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成JWT token
func GenerateJWT(user *models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   user.ID,
		UserType: user.UserType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// JWTAuthMiddleware JWT鉴权中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.JsonErrorWithCode(c, 401, "缺少Authorization头部")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			utils.JsonErrorWithCode(c, 401, "Authorization头部格式错误")
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			utils.JsonErrorWithCode(c, 401, "无效的token")
			c.Abort()
			return
		}

		// 检查用户是否存在
		var user models.User
		result := database.DB.First(&user, claims.UserID)
		if result.Error != nil {
			utils.JsonErrorWithCode(c, 401, "用户不存在")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("user_type", claims.UserType)
		c.Set("username", claims.Subject)
		c.Next()
	}
}

// GetUserIDFromContext 从上下文中获取用户ID
func GetUserIDFromContext(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(uint)
}

// GetUserTypeFromContext 从上下文中获取用户类型
func GetUserTypeFromContext(c *gin.Context) int {
	userType, exists := c.Get("user_type")
	if !exists {
		return 0
	}
	return userType.(int)
}
