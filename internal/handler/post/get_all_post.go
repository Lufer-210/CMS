package post

import (
	"CMS/internal/models"
	"CMS/internal/services"
	"CMS/pkg/utils"

	"github.com/gin-gonic/gin"
)

type GetAllPostsResponse struct {
	Posts []models.Post `json:"posts"`
}

func GetAllPosts(c *gin.Context) {
	postlist, err := services.GetAllPostsWithFormat()
	if err != nil {
		utils.JsonErrorWithCode(c, 1001, "获取失败")
		return
	}
	utils.JsonSuccessWithCode(c, 200, gin.H{
		"post_list": postlist,
	})
}
