package user

import (
	"CMS/internal/logger"
	"CMS/internal/models"
	"CMS/internal/services"
	"CMS/pkg/utils"
	"regexp"

	"github.com/gin-gonic/gin"
)

type RegData struct {
	Username string `json:"username" binding:"required"`  // 学号或工号，只能是数字
	Name     string `json:"name" binding:"required"`      // 姓名
	Password string `json:"password" binding:"required"`  // 密码，8-16位
	UserType int    `json:"user_type" binding:"required"` // 1学生，2管理员
}

func Register(c *gin.Context) {
	var req RegData
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("注册参数错误: %v", err)
		utils.JsonErrorWithCode(c, 1001, "参数错误")
		return
	}

	// 校验用户名只由数字组成
	if !regexp.MustCompile(`^\d+$`).MatchString(req.Username) {
		logger.GetLogger().Errorf("注册失败，账号必须为数字: %s", req.Username)
		utils.JsonErrorWithCode(c, 1002, "账号必须为数字")
		return
	}

	// 校验密码长度
	if len(req.Password) < 8 || len(req.Password) > 16 {
		logger.GetLogger().Errorf("注册失败，密码长度不符合要求: %s", req.Username)
		utils.JsonErrorWithCode(c, 1003, "密码长度需为8-16位")
		return
	}

	// 校验用户类型
	if req.UserType != 1 && req.UserType != 2 {
		logger.GetLogger().Errorf("注册失败，用户类型错误: username=%s, user_type=%d", req.Username, req.UserType)
		utils.JsonErrorWithCode(c, 1004, "用户类型错误")
		return
	}

	// 检查用户名是否已存在
	user, err := services.GetUserByUsername(req.Username)
	if err == nil && user != nil {
		logger.GetLogger().Errorf("注册失败，用户名已存在: %s", req.Username)
		utils.JsonErrorWithCode(c, 1005, "注册失败，用户名已存在")
		return
	}

	newUser := &models.User{
		Username: req.Username,
		Name:     req.Name,
		Password: req.Password,
		UserType: req.UserType,
	}

	if err := services.RegisterUser(newUser); err != nil {
		logger.GetLogger().Errorf("注册失败: username=%s, error=%v", req.Username, err)
		utils.JsonErrorWithCode(c, 1006, "注册失败："+err.Error())
		return
	}

	logger.GetLogger().Infof("用户注册成功: username=%s, name=%s, user_type=%d", req.Username, req.Name, req.UserType)
	// 按照API规范返回数据
	utils.JsonSuccessWithCode(c, 200, nil)
}
