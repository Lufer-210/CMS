package services

import (
	"CMS/internal/models"
	"CMS/internal/pkg/database"
)

func CreatePost(post models.Post) error {
	result := database.DB.Create(&post)
	return result.Error
}

func GetAllPosts() (posts []models.Post, err error) {
	result := database.DB.Order("post_time desc").Find(&posts)
	err = result.Error
	return
}

func GetPostByID(id uint) (post models.Post, err error) {
	result := database.DB.First(&post, id)
	err = result.Error
	return
}

func GetAllPostsWithFormat() ([]models.PostResponse, error) {
	posts, err := GetAllPosts()
	if err != nil {
		return nil, err
	}
	var postResponses []models.PostResponse
	for _, post := range posts {
		postResponse := post.ToResponse()
		likes, err := GetLikesByPostID(post.ID)
		if err == nil {
			postResponse.Likes = likes
		}
		postResponses = append(postResponses, postResponse)
	}
	return postResponses, nil
}
func DeletePostByID(id uint) error {
	result := database.DB.Where("id = ?", id).Delete(&models.Post{})
	return result.Error
}

func UpdatePostByID(id uint, content string) error {
	result := database.DB.Where("id = ?", id).Updates(models.Post{Content: content})
	return result.Error
}
