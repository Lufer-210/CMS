package post

import (
	"CMS/internal/logger"
	"CMS/internal/models"
	"CMS/internal/services"
	"CMS/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type CreatePostData struct {
	Content string `json:"content" binding:"required"`
	UserID  uint   `json:"user_id" binding:"required"`
}

func CreatePost(c *gin.Context) {
	var data CreatePostData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		logger.GetLogger().Errorf("发布帖子参数错误: %v", err)
		utils.JsonErrorWithCode(c, 1001, "参数错误")
		return
	}

	logger.GetLogger().Infof("用户尝试发布帖子: user_id=%d", data.UserID)

	err = services.CreatePost(models.Post{
		Content:  data.Content,
		UserID:   data.UserID,
		PostTime: time.Now(),
	})
	if err != nil {
		logger.GetLogger().Errorf("创建帖子失败: user_id=%d, error=%v", data.UserID, err)
		utils.JsonErrorWithCode(c, 1002, "创建失败")
		return
	}

	logger.GetLogger().Infof("用户发布帖子成功: user_id=%d", data.UserID)
	utils.JsonSuccessWithCode(c, 200, nil)
}
