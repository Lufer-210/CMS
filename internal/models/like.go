package models

type Like struct {
	ID     uint `gorm:"primaryKey"`
	PostID uint `gorm:"index"`
	UserID uint
}
