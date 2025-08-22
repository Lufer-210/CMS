package user

import (
	"CMS/internal/logger"
	"CMS/internal/services"
	"CMS/pkg/utils"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LoginData struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var loginData LoginData
	err := c.ShouldBindJSON(&loginData)
	if err != nil {
		logger.GetLogger().Errorf("登录参数错误: %v", err)
		utils.JsonErrorWithCode(c, 1001, "参数错误")
		return
	}
	logger.GetLogger().Infof("用户尝试登录: %s", loginData.Username)

	user, err := services.CheckLogin(loginData.Username, loginData.Password)
	if err != nil {
		// 区分用户不存在和密码错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.GetLogger().Errorf("用户不存在 username=%s", loginData.Username)
			utils.JsonErrorWithCode(c, 200506, "用户不存在")
			return
		}

		// 检查是否是密码错误
		// 先获取用户信息确认用户存在
		user, userErr := services.GetUserByUsername(loginData.Username)
		if userErr == nil && user != nil {
			// 用户存在但密码错误
			logger.GetLogger().Errorf("用户密码错误 username=%s", loginData.Username)
			utils.JsonErrorWithCode(c, 200507, "密码错误")
			return
		}

		// 其他错误
		logger.GetLogger().Errorf("用户登录失败 username=%s, error=%v", loginData.Username, err)
		utils.JsonErrorWithCode(c, 1002, "登录失败")
		return
	}
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		logger.GetLogger().Errorf("生成token失败: username=%s, error=%v", loginData.Username, err)
		utils.JsonErrorWithCode(c, 1003, "生成token失败")
		return
	}
	logger.GetLogger().Infof("用户登录成功 user_id=%d, username=%s, user_type=%d", user.ID, user.Username, user.UserType)

	// 按照API规范返回数据
	utils.JsonSuccessWithCode(c, 200, gin.H{
		"user_id":   user.ID,
		"user_type": user.UserType,
		"token":     token,
	})
}
