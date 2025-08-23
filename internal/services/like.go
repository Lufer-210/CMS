// internal/services/like.go
package services

import (
	"CMS/internal/logger"
	"CMS/internal/models"
	"CMS/internal/pkg/database"
	"CMS/pkg/redis"
	"context"
	"strconv"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

// Redis 键名定义（原like_key.go内容合并至此）
const (
	postLikesKey    = "post:likes:"     // 帖子点赞数：string类型
	userLikesKey    = "user:likes:"     // 用户点赞记录：hash类型
	likesRankKey    = "post:likes:rank" // 点赞排行榜：zset类型
	updatedPostsKey = "post:updated"    // 记录有更新的帖子ID：set类型（用于diff同步）
	cacheExpire     = 5 * 60            // 缓存过期时间：5分钟（秒）
)

// GetLikesByPostID 从 Redis 获取点赞数，缓存未命中则查数据库并同步到 Redis
func GetLikesByPostID(postID uint) (int, error) {
	key := postLikesKey + strconv.Itoa(int(postID))
	ctx := context.Background()

	// 1. 查Redis（5分钟过期）
	likesStr, err := redis.RedisClient.Get(ctx, key).Result()
	if err == nil {
		likes, _ := strconv.Atoi(likesStr)
		// 滑动过期：每次访问延长5分钟（可选，根据需求决定）
		redis.RedisClient.Expire(ctx, key, cacheExpire*time.Second)
		return likes, nil
	}

	// 2. Redis未命中（过期或不存在），查数据库
	var count int64
	result := database.DB.Model(&models.Like{}).Where("post_id = ?", postID).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	// 3. 同步到Redis（设置5分钟过期）
	if err := redis.RedisClient.Set(ctx, key, count, cacheExpire*time.Second).Err(); err != nil {
		logger.GetLogger().Errorf("同步点赞数到Redis失败: postID=%d, err=%v", postID, err)
	}

	return int(count), nil
}

func IsUserLikedPost(postID, userID uint) (bool, error) {
	key := userLikesKey + strconv.Itoa(int(userID))
	field := strconv.Itoa(int(postID))

	// 1. 先查Redis的hash结构
	exists, err := redis.RedisClient.HExists(context.Background(), key, field).Result()
	if err == nil {
		// 滑动过期：延长缓存时间
		redis.RedisClient.Expire(context.Background(), key, 24*time.Hour)
		return exists, nil
	}

	// 2. Redis查询失败，查数据库
	var count int64
	result := database.DB.Model(&models.Like{}).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}

	// 3. 同步结果到Redis（hash结构）
	userLiked := count > 0
	if userLiked {
		// 记录到hash，并设置过期时间
		if err := redis.RedisClient.HSet(context.Background(), key, field, "1").Err(); err != nil {
			logger.GetLogger().Errorf("同步用户点赞记录到Redis失败: userID=%d, postID=%d, err=%v", userID, postID, err)
		}
		redis.RedisClient.Expire(context.Background(), key, 24*time.Hour)
	}

	return userLiked, nil
}

// ToggleLike 切换用户对帖子的点赞状态
func ToggleLike(postID, userID uint) (map[string]interface{}, *models.ServiceError) {
	// 参数验证
	if postID == 0 {
		return nil, &models.ServiceError{
			Code:    1001,
			Message: "无效的帖子ID",
		}
	}

	if userID == 0 {
		return nil, &models.ServiceError{
			Code:    1002,
			Message: "无效的用户ID",
		}
	}

	ctx := context.Background()
	postKey := postLikesKey + strconv.Itoa(int(postID))
	userKey := userLikesKey + strconv.Itoa(int(userID)) // hash key：user:likes:{userID}
	postIDField := strconv.Itoa(int(postID))            // hash field：postID

	// 1. 先查询用户当前点赞状态（用于判断操作类型）
	isLiked, err := IsUserLikedPost(postID, userID)
	if err != nil {
		return nil, &models.ServiceError{
			Code:    1003,
			Message: "查询点赞状态失败: " + err.Error(),
		}
	}

	// 2. 数据库事务：处理点赞记录的增删
	tx := database.DB.Begin()
	if tx.Error != nil {
		return nil, &models.ServiceError{
			Code:    1005,
			Message: "数据库事务启动失败: " + tx.Error.Error(),
		}
	}

	var dbErr error
	if isLiked {
		// 取消点赞：删除数据库记录
		dbErr = tx.Where("post_id = ? AND user_id = ?", postID, userID).Delete(&models.Like{}).Error
	} else {
		// 点赞：新增数据库记录
		like := models.Like{
			PostID: postID,
			UserID: userID,
		}
		dbErr = tx.Create(&like).Error
	}

	if dbErr != nil {
		tx.Rollback()
		return nil, &models.ServiceError{
			Code:    1006,
			Message: "数据库操作失败: " + dbErr.Error(),
		}
	}

	// 3. 数据库操作成功后，执行Redis操作（确保缓存与数据库一致）
	postIDStr := strconv.Itoa(int(postID))

	_, redisErr := redis.RedisClient.Pipelined(ctx, func(pipe goredis.Pipeliner) error {
		if isLiked {
			// 取消点赞：删除hash中的field
			pipe.HDel(ctx, userKey, postIDField)
			// 减少帖子点赞数
			pipe.Decr(ctx, postKey)
			// 排行榜减1
			pipe.ZIncrBy(ctx, likesRankKey, -1, postIDStr)
		} else {
			// 点赞：添加hash中的field
			pipe.HSet(ctx, userKey, postIDField, "1")
			pipe.Expire(ctx, userKey, 24*time.Hour) // 设置过期时间
			// 增加帖子点赞数
			pipe.Incr(ctx, postKey)
			// 排行榜加1
			pipe.ZIncrBy(ctx, likesRankKey, 1, postIDStr)
		}
		pipe.SAdd(ctx, updatedPostsKey, postIDStr)
		pipe.Expire(ctx, updatedPostsKey, 60*time.Minute)
		return nil
	})

	if redisErr != nil {
		// Redis操作失败：回滚数据库事务（核心！确保一致性）
		tx.Rollback()
		return nil, &models.ServiceError{
			Code:    1007,
			Message: "Redis操作失败: " + redisErr.Error(),
		}
	}

	// 4. 提交数据库事务
	if err := tx.Commit().Error; err != nil {
		return nil, &models.ServiceError{
			Code:    1008,
			Message: "数据库事务提交失败: " + err.Error(),
		}
	}

	// 5. 获取最新点赞数（带滑动过期）
	likes, err := GetLikesByPostID(postID)
	if err != nil {
		return nil, &models.ServiceError{
			Code:    1004,
			Message: "获取点赞数失败: " + err.Error(),
		}
	}

	// 构造响应
	response := map[string]interface{}{
		"likes":    likes,
		"is_liked": !isLiked, // 返回最新的点赞状态
	}
	return response, nil
}

// SyncDiffLikesToRedis 只同步有更新的帖子数据（diff同步）
func SyncDiffLikesToRedis() {
	ctx := context.Background()
	logger.GetLogger().Info("开始diff同步点赞数据到Redis")

	// 1. 获取所有有更新的帖子ID
	updatedPostIDs, err := redis.RedisClient.SMembers(ctx, updatedPostsKey).Result()
	if err != nil {
		logger.GetLogger().Errorf("diff同步失败：获取更新列表错误: %v", err)
		return
	}
	if len(updatedPostIDs) == 0 {
		logger.GetLogger().Info("diff同步：无更新数据，跳过")
		return
	}

	// 2. 批量查询这些帖子的最新点赞数（从数据库）
	var postLikeStats []struct {
		PostID uint
		Count  int64
	}
	if err := database.DB.Model(&models.Like{}).
		Where("post_id IN (?)", updatedPostIDs).
		Select("post_id, count(*) as count").
		Group("post_id").
		Scan(&postLikeStats).Error; err != nil {
		logger.GetLogger().Errorf("diff同步失败：查询更新数据错误: %v", err)
		return
	}

	// 3. 同步到Redis并更新排行榜
	pipe := redis.RedisClient.Pipeline()
	for _, stat := range postLikeStats {
		postIDStr := strconv.Itoa(int(stat.PostID))
		postKey := postLikesKey + postIDStr
		// 更新点赞数缓存（5分钟过期）
		pipe.Set(ctx, postKey, stat.Count, cacheExpire*time.Second)
		// 更新排行榜
		pipe.ZAdd(ctx, likesRankKey, &goredis.Z{Score: float64(stat.Count), Member: postIDStr})
	}
	if _, err := pipe.Exec(ctx); err != nil {
		logger.GetLogger().Errorf("diff同步失败：Redis批量操作错误: %v", err)
		return
	}

	// 4. 清空已同步的更新列表
	redis.RedisClient.Del(ctx, updatedPostsKey)
	logger.GetLogger().Infof("diff同步完成：共同步 %d 个帖子", len(updatedPostIDs))
}
