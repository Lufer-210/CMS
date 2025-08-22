package database

import (
	"CMS/config"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() *gorm.DB {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	dsn := cfg.Database.DSN
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = autoMigrate(db)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database connection established successfully")
	DB = db
	return db
}
