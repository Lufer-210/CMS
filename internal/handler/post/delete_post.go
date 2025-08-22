package post

import (
	"CMS/internal/logger"
	"CMS/internal/middleware"
	"CMS/internal/services"
	"CMS/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func DeletePost(c *gin.Context) {
	// 从查询参数中获取post_id
	postIDStr := c.Query("post_id")
	if postIDStr == "" {
		logger.GetLogger().Errorf("删除帖子参数错误: post_id为空")
		utils.JsonErrorWithCode(c, 1001, "参数错误")
		return
	}

	// 转换参数类型
	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	if err != nil {
		logger.GetLogger().Errorf("删除帖子参数错误: post_id不是有效数字: %s", postIDStr)
		utils.JsonErrorWithCode(c, 1002, "参数错误")
		return
	}

	// 从上下文获取当前用户ID
	userID := middleware.GetUserIDFromContext(c)
	if userID == 0 {
		logger.GetLogger().Error("删除帖子失败: 无法获取用户ID")
		utils.JsonErrorWithCode(c, 1003, "用户认证失败")
		return
	}

	logger.GetLogger().Infof("用户尝试删除帖子: user_id=%d, post_id=%d", userID, postID)

	// 获取帖子信息以验证权限
	post, err := services.GetPostByID(uint(postID))
	if err != nil {
		logger.GetLogger().Errorf("删除帖子失败，获取帖子失败: post_id=%d, error=%v", postID, err)
		utils.JsonErrorWithCode(c, 1004, "获取帖子失败")
		return
	}

	// 检查是否是帖子所有者
	if post.UserID != userID {
		logger.GetLogger().Errorf("删除帖子失败，无权限删除: user_id=%d, post_id=%d, post_owner=%d", userID, postID, post.UserID)
		utils.JsonErrorWithCode(c, 1005, "无权限删除")
		return
	}

	// 执行删除操作
	err = services.DeletePostByID(uint(postID))
	if err != nil {
		logger.GetLogger().Errorf("删除帖子失败: post_id=%d, error=%v", postID, err)
		utils.JsonErrorWithCode(c, 1006, "删除失败")
		return
	}

	logger.GetLogger().Infof("用户删除帖子成功: user_id=%d, post_id=%d", userID, postID)
	utils.JsonSuccessWithCode(c, 200, nil)
}
