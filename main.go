// main.go
package main

import (
	"CMS/config"
	"CMS/internal/pkg/database"
	"CMS/internal/router"
	"CMS/internal/services"
	"CMS/pkg/redis"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var cfg *config.Config // 全局配置变量

func main() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}
	config.LoadedConfig = cfg // 注入全局配置（需在config/config.go中添加LoadedConfig变量）

	database.Init()
	redis.Init() // 初始化Redis

	// 启动定时同步任务
	go startLikeSyncTask()

	r := gin.Default()
	router.Init(r)
	err = r.Run(":" + strconv.Itoa(cfg.Server.Port))
	if err != nil {
		log.Fatal(err)
	}
}

// startLikeSyncTask 启动点赞数据同步任务
func startLikeSyncTask() {
	// 立即执行一次同步
	services.SyncDiffLikesToRedis()

	// 之后每5分钟执行一次
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		services.SyncDiffLikesToRedis()
	}
}
