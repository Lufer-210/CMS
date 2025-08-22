package admin

import (
	"CMS/internal/logger"
	"CMS/internal/middleware"
	"CMS/internal/models"
	"CMS/internal/services"
	"CMS/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ApproveReportData 审批举报的数据结构
type ApproveReportData struct {
	UserID   uint `json:"user_id"`
	PostID   uint `json:"post_id" binding:"required"`
	Approval int  `json:"approval" binding:"required"` // 1代表同意，2代表拒绝
}

// ApproveReport 管理员审批被举报的帖子
// POST /api/admin/report
func ApproveReport(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	if userID == 0 {
		logger.GetLogger().Errorf("获取用户ID失败")
		// 改为通过 c.Error 传递错误
		c.Error(&models.ServiceError{Code: 401, Message: "获取用户信息失败"})
		c.Abort()
		return
	}

	var data ApproveReportData
	if err := c.ShouldBindJSON(&data); err != nil {
		logger.GetLogger().Errorf("审批举报参数错误: %v", err)
		// 绑定错误直接传递给 Gin，由中间件识别为 *gin.Error
		c.Error(err)
		c.Abort()
		return
	}

	// 验证approval参数
	if data.Approval != 1 && data.Approval != 2 {
		logger.GetLogger().Errorf("审批举报参数错误: 无效的approval值: %d", data.Approval)
		c.Error(&models.ServiceError{Code: 400, Message: "审批状态参数错误"})
		c.Abort()
		return
	}

	logger.GetLogger().Infof("管理员尝试审批举报: admin_user_id=%d, post_id=%d, approval=%d", userID, data.PostID, data.Approval)

	// 检查用户是否为管理员
	isAdmin, serviceErr := services.CheckUserIsAdmin(userID)
	if serviceErr != nil {
		logger.GetLogger().Errorf("审批举报失败，权限检查错误: user_id=%d, error=%v", userID, serviceErr)
		c.Error(&models.ServiceError{Code: 403, Message: "权限不足，只有管理员可以审批举报"})
		c.Abort()
		return
	}

	if !isAdmin {
		logger.GetLogger().Errorf("审批举报失败，权限不足: user_id=%d", userID)
		c.Error(&models.ServiceError{Code: 403, Message: "权限不足，只有管理员可以审批举报"})
		c.Abort()
		return
	}

	// 处理审批逻辑
	serviceErr = services.ProcessReportApproval(data.PostID, data.Approval, userID)
	if serviceErr != nil {
		logger.GetLogger().Errorf("审批举报失败: post_id=%d, approval=%d, error=%v", data.PostID, data.Approval, serviceErr)
		c.Error(serviceErr) // 直接传递 ServiceError
		c.Abort()
		return
	}

	logger.GetLogger().Infof("管理员审批举报成功: admin_user_id=%d, post_id=%d, approval=%d", userID, data.PostID, data.Approval)
	utils.JsonSuccessWithCode(c, 200, nil)
}
