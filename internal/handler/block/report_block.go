package block

import (
	"CMS/internal/logger"
	"CMS/internal/middleware"
	"CMS/internal/models"
	"CMS/internal/services"
	"CMS/pkg/utils"

	"github.com/gin-gonic/gin"
)

type ReportPostData struct {
	PostID uint   `json:"post_id" binding:"required"`
	Reason string `json:"reason" binding:"required"`
}

func ReportPost(c *gin.Context) {
	var data ReportPostData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		logger.GetLogger().Errorf("举报帖子参数错误: %v", err)
		utils.JsonErrorWithCode(c, 1001, "参数错误")
		return
	}

	// 从上下文获取当前用户ID
	userID := middleware.GetUserIDFromContext(c)
	if userID == 0 {
		logger.GetLogger().Error("举报帖子失败: 无法获取用户ID")
		utils.JsonErrorWithCode(c, 1002, "用户认证失败")
		return
	}

	_, err = services.GetPostByID(data.PostID)
	if err != nil {
		logger.GetLogger().Errorf("举报帖子失败: 帖子不存在 post_id=%d, error=%v", data.PostID, err)
		utils.JsonErrorWithCode(c, 1003, "帖子不存在")
		return
	}
	block := models.Block{
		UserID:   userID,
		TargetID: data.PostID,
		Reason:   data.Reason,
	}
	err = services.CreateBlock(block)
	if err != nil {
		logger.GetLogger().Errorf("举报帖子失败: 保存举报记录失败 user_id=%d, post_id=%d, error=%v", userID, data.PostID, err)
		utils.JsonErrorWithCode(c, 1004, "举报失败")
		return
	}

	logger.GetLogger().Infof("用户举报帖子成功: user_id=%d, post_id=%d", userID, data.PostID)
	utils.JsonSuccessWithCode(c, 200, nil)
}
func GetReportList(c *gin.Context) {
	// 从上下文获取当前用户ID
	userID := middleware.GetUserIDFromContext(c)
	if userID == 0 {
		logger.GetLogger().Error("获取举报列表失败: 无法获取用户ID")
		utils.JsonErrorWithCode(c, 1002, "用户认证失败")
		return
	}
	// 调用服务层获取举报列表
	reportList, err := services.GetReportListByUserID(userID)
	if err != nil {
		logger.GetLogger().Errorf("获取举报列表失败: user_id=%d, error=%v", userID, err)
		utils.JsonErrorWithCode(c, 1003, "获取举报列表失败")
		return
	}

	utils.JsonSuccessWithCode(c, 200, gin.H{
		"report_list": reportList,
	})
}
