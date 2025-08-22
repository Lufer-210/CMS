package post

import (
	"CMS/internal/logger"
	"CMS/internal/middleware"
	"CMS/internal/services"
	"CMS/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LikePostData struct {
	PostID uint `json:"post_id" binding:"required"`
}

func LikePost(c *gin.Context) {
	var data LikePostData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		logger.GetLogger().Errorf("点赞参数错误: %v", err)
		utils.JsonErrorWithCode(c, 1001, "参数错误")
		return
	}

	// 从JWT token中获取用户ID，而不是从前端传递
	userID := middleware.GetUserIDFromContext(c)
	if userID == 0 {
		logger.GetLogger().Error("点赞失败: 无法获取用户ID")
		utils.JsonErrorWithCode(c, 1002, "用户认证失败")
		return
	}

	logger.GetLogger().Infof("用户尝试点赞: user_id=%d, post_id=%d", userID, data.PostID)

	// 调用服务层切换点赞状态
	result, serviceErr := services.ToggleLike(
		strconv.FormatUint(uint64(data.PostID), 10),
		strconv.FormatUint(uint64(userID), 10),
	)
	if serviceErr != nil {
		logger.GetLogger().Errorf("点赞操作失败: user_id=%d, post_id=%d, error=%v", userID, data.PostID, serviceErr)
		utils.JsonErrorWithCode(c, serviceErr.Code, serviceErr.Message)
		return
	}

	_, err = services.IsUserLikedPost(data.PostID, userID)
	if err != nil {
		logger.GetLogger().Errorf("检查用户点赞状态失败: user_id=%d, post_id=%d, error=%v", userID, data.PostID, err)
	}
	logger.GetLogger().Infof("用户点赞操作成功: user_id=%d, post_id=%d, likes=%d", userID, data.PostID, result["likes"])

	utils.JsonSuccessWithCode(c, 200, nil)
}
