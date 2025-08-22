package services

import (
	"CMS/internal/models"
	"CMS/internal/pkg/database"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func GetUserByUsername(username string) (user *models.User, err error) {
	result := database.DB.Where("username = ?", username).First(&user)
	err = result.Error
	return
}

func RegisterUser(user *models.User) error {
	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedpassword)

	result := database.DB.Create(&user)
	return result.Error
}

func CheckLogin(username, password string) (*models.User, error) {
	user, err := GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, errors.New("invalid password")
	}
	return user, nil
}

// ... existing code ...
func GetUserByID(id uint) (user *models.User, err error) {
	result := database.DB.First(&user, id)
	err = result.Error
	return
}

// CheckUserIsAdmin 检查用户是否为管理员
func CheckUserIsAdmin(userID uint) (bool, *models.ServiceError) {
	user, err := GetUserByID(userID)
	if err != nil {
		return false, &models.ServiceError{
			Code:    1001,
			Message: "用户不存在",
		}
	}

	return user.UserType == models.AdminRole, nil
}

// AdminReportItem 管理员查看举报列表的响应项
type AdminReportItem struct {
	Username string `json:"username"`
	Content  string `json:"content"`
	Reason   string `json:"reason"`
	PostID   uint   `json:"post_id"`
}

// GetPendingReportsForAdmin 获取管理员待审批的举报列表
func GetPendingReportsForAdmin() ([]AdminReportItem, *models.ServiceError) {
	// 查询所有待审批的举报记录 (Status = 0)
	var blocks []models.Block
	result := database.DB.Where("status = ?", 0).Find(&blocks)
	if result.Error != nil {
		return nil, &models.ServiceError{
			Code:    1001,
			Message: "获取举报列表失败",
		}
	}

	var reportItems []AdminReportItem
	for _, block := range blocks {
		// 获取举报用户信息
		user, err := GetUserByID(block.UserID)
		if err != nil {
			// 如果用户不存在，跳过该记录
			continue
		}

		// 获取被举报的帖子
		post, err := GetPostByID(block.TargetID)
		if err != nil {
			// 如果帖子不存在，使用默认内容
			reportItems = append(reportItems, AdminReportItem{
				Username: user.Username,
				Content:  "帖子已被删除",
				Reason:   block.Reason,
				PostID:   block.TargetID,
			})
		} else {
			reportItems = append(reportItems, AdminReportItem{
				Username: user.Username,
				Content:  post.Content,
				Reason:   block.Reason,
				PostID:   post.ID,
			})
		}
	}

	return reportItems, nil
}
