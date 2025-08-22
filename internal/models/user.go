package models

import "golang.org/x/crypto/bcrypt"

const (
	StudentRole = 1 // 学生用户
	AdminRole   = 2 // 管理员用户
)

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"uniqueIndex;not null;size:20" json:"username"` // 学号作为用户名
	Password string `gorm:"not null" json:"-"`                            // 密码不返回给前端
	Name     string `gorm:"size:50" json:"name"`                          // 用户姓名
	UserType int    `gorm:"default:1" json:"user_type"`                   // 用户类型: 1-学生, 2-管理员
}

func (u *User) CheckPasswordHash(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
