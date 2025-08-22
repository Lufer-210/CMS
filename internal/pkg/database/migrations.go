package database

import (
	"CMS/internal/models"

	"gorm.io/gorm"
)

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Block{},
		&models.Like{},
		&models.AuditLog{},
	)
}
