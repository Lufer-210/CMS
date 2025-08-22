package admin

import (
	"CMS/internal/logger"
	"CMS/internal/middleware"
	"CMS/internal/models"
	"CMS/internal/services"
	"CMS/pkg/utils"

	"github.com/gin-gonic/gin"
)

// GetPendingReports 管理员获取所有未审批的被举报帖子
func GetPendingReports(c *gin.Context) {
	// 从JWT context获取用户ID
	userID := middleware.GetUserIDFromContext(c)
	if userID == 0 {
		c.Error(&models.ServiceError{Code: 401, Message: "用户未认证"})
		c.Abort()
		return
	}

	logger.GetLogger().Infof("管理员尝试获取待审批举报列表: admin_user_id=%d", userID)

	// 检查用户是否为管理员
	isAdmin, serviceErr := services.CheckUserIsAdmin(userID)
	if serviceErr != nil {
		logger.GetLogger().Errorf("获取待审批举报列表失败，权限检查错误: user_id=%d, error=%v", userID, serviceErr)
		c.Error(serviceErr)
		c.Abort()
		return
	}

	if !isAdmin {
		logger.GetLogger().Errorf("获取待审批举报列表失败，权限不足: user_id=%d", userID)
		c.Error(&models.ServiceError{Code: 1003, Message: "权限不足，只有管理员可以查看举报列表"})
		c.Abort()
		return
	}

	// 获取所有未审批的举报列表
	reportList, serviceErr := services.GetPendingReportsForAdmin()
	if serviceErr != nil {
		logger.GetLogger().Errorf("获取待审批举报列表失败: error=%v", serviceErr)
		c.Error(serviceErr)
		c.Abort()
		return
	}

	logger.GetLogger().Infof("获取待审批举报列表成功: admin_user_id=%d, count=%d", userID, len(reportList))
	utils.JsonSuccessWithCode(c, 200, gin.H{
		"report_list": reportList,
	})
}
