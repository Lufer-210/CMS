package post

import (
	"CMS/internal/logger"
	"CMS/internal/middleware"
	"CMS/internal/services"
	"CMS/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPostLikes(c *gin.Context) {
	// 从查询参数中获取post_id
	postIDStr := c.Query("post_id")

	// 检查post_id参数是否存在
	if postIDStr == "" {
		logger.GetLogger().Error("获取帖子点赞数失败: 缺少post_id参数")
		utils.JsonErrorWithCode(c, 1001, "缺少post_id参数")
		return
	}

	// 解析post_id参数
	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	if err != nil || postID == 0 {
		logger.GetLogger().Errorf("获取帖子点赞数失败: 无效的post_id参数: %s", postIDStr)
		utils.JsonErrorWithCode(c, 1002, "无效的post_id参数")
		return
	}

	// 获取当前用户ID（如果已登录）
	userID := middleware.GetUserIDFromContext(c)

	logger.GetLogger().Infof("用户尝试获取帖子点赞数: post_id=%d, user_id=%d", postID, userID)

	// 调用服务层获取点赞数和用户点赞状态
	likes, err := services.GetLikesByPostID(uint(postID))
	if err != nil {
		logger.GetLogger().Errorf("获取帖子点赞数失败: post_id=%d, error=%v", postID, err)
		utils.JsonErrorWithCode(c, 1003, "获取点赞数失败")
		return
	}

	// 构造返回数据
	responseData := gin.H{
		"likes": likes,
	}

	// 如果用户已登录，检查用户是否已点赞该帖子
	if userID != 0 {
		userLiked, err := services.IsUserLikedPost(uint(postID), userID)
		if err == nil {
			responseData["user_liked"] = userLiked
		}
	}

	logger.GetLogger().Infof("获取帖子点赞数成功: post_id=%d, likes=%d", postID, likes)

	// 按照API规范返回数据
	utils.JsonSuccessWithCode(c, 200, responseData)
}
