package post

import (
	"CMS/internal/logger"
	"CMS/internal/middleware"
	"CMS/internal/models"
	"CMS/internal/services"
	"CMS/pkg/utils"

	"github.com/gin-gonic/gin"
)

type UpdatePostData struct {
	PostID  uint   `json:"post_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func UpdatePost(c *gin.Context) {
	var data UpdatePostData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		logger.GetLogger().Errorf("修改帖子参数错误: %v", err)
		c.Error(err) // 绑定错误交给中间件处理
		c.Abort()
		return
	}

	// 从上下文获取当前用户ID（JWT验证后设置）
	userID := middleware.GetUserIDFromContext(c)
	if userID == 0 {
		logger.GetLogger().Error("修改帖子失败: 无法获取用户ID")
		c.Error(&models.ServiceError{Code: 1002, Message: "用户认证失败"})
		c.Abort()
		return
	}

	logger.GetLogger().Infof("用户尝试修改帖子: user_id=%d, post_id=%d", userID, data.PostID)

	post, err := services.GetPostByID(data.PostID)
	if err != nil {
		logger.GetLogger().Errorf("修改帖子失败，获取帖子失败: post_id=%d, error=%v", data.PostID, err)
		c.Error(&models.ServiceError{Code: 1003, Message: "获取帖子失败"})
		c.Abort()
		return
	}

	if post.UserID != userID {
		logger.GetLogger().Errorf("修改帖子失败，无权限修改: user_id=%d, post_id=%d, post_owner=%d", userID, data.PostID, post.UserID)
		c.Error(&models.ServiceError{Code: 1004, Message: "无权限修改"})
		c.Abort()
		return
	}

	err = services.UpdatePostByID(data.PostID, data.Content)
	if err != nil {
		logger.GetLogger().Errorf("修改帖子失败: post_id=%d, error=%v", data.PostID, err)
		c.Error(&models.ServiceError{Code: 1005, Message: "修改失败"})
		c.Abort()
		return
	}

	logger.GetLogger().Infof("用户修改帖子成功: user_id=%d, post_id=%d", userID, data.PostID)
	utils.JsonSuccessWithCode(c, 200, nil)
}
