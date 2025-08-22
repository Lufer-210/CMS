package services

import (
	"CMS/internal/logger"
	"CMS/internal/models"
	"CMS/internal/pkg/database"
	"CMS/pkg/redis"
	"context"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

func CreateBlock(block models.Block) error {
	block.CreatedAt = time.Now()
	block.Status = 0 // 初始状态为待审批
	result := database.DB.Create(&block)
	return result.Error
}

// GetReportListByUserID 获取用户的举报列表及审批状态
func GetReportListByUserID(userID uint) ([]map[string]interface{}, error) {
	var blocks []models.Block
	result := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&blocks)
	if result.Error != nil {
		return nil, result.Error
	}

	var reportList []map[string]interface{}
	for _, block := range blocks {
		// 获取被举报的帖子内容
		post, err := GetPostByID(block.TargetID)

		item := map[string]interface{}{
			"post_id": block.TargetID,
			"reason":  block.Reason,
			"status":  block.Status,
		}

		if err != nil {
			item["content"] = "帖子已被删除"
		} else {
			item["content"] = post.Content
		}

		reportList = append(reportList, item)
	}

	return reportList, nil
}

// ProcessReportApproval 处理举报审批
func ProcessReportApproval(postID uint, approval int, adminID uint) *models.ServiceError {
	// 开始事务
	tx := database.DB.Begin()
	if tx.Error != nil {
		return &models.ServiceError{
			Code:    1001,
			Message: "数据库事务启动失败: " + tx.Error.Error(),
		}
	}
	var block models.Block
	if err := tx.Where("target_id = ? AND status = 0", postID).First(&block).Error; err != nil {
		tx.Rollback()
		return &models.ServiceError{
			Code:    1009, // 新增错误码：举报记录不存在或已处理
			Message: "未找到待审批的举报记录: " + err.Error(),
		}
	}
	// 如果审批通过（同意删除），则删除被举报的帖子
	if approval == 1 {
		// 先查询帖子，确认存在
		var post models.Post
		if err := tx.First(&post, postID).Error; err != nil {
			tx.Rollback()
			return &models.ServiceError{
				Code:    1004,
				Message: "帖子不存在: " + err.Error(),
			}
		}

		// 删除帖子
		if err := tx.Delete(&models.Post{}, postID).Error; err != nil {
			tx.Rollback()
			return &models.ServiceError{
				Code:    1003,
				Message: "删除帖子失败: " + err.Error(),
			}
		}

		// 同时删除相关的点赞记录
		if err := tx.Where("post_id = ?", postID).Delete(&models.Like{}).Error; err != nil {
			tx.Rollback()
			return &models.ServiceError{
				Code:    1007,
				Message: "删除帖子点赞记录失败: " + err.Error(),
			}
		}
		// 清理 Redis 缓存
		ctx := context.Background()
		postKey := postLikesKey + strconv.Itoa(int(postID))
		rankKey := likesRankKey
		// 批量删除缓存
		_, err := redis.RedisClient.Pipelined(ctx, func(pipe goredis.Pipeliner) error {
			pipe.Del(ctx, postKey)                             // 删除帖子点赞数
			pipe.ZRem(ctx, rankKey, strconv.Itoa(int(postID))) // 从排行榜移除
			return nil
		})
		if err != nil {
			logger.GetLogger().Errorf("删除帖子后清理Redis缓存失败: postID=%d, err=%v", postID, err)

		}
	}

	block.Status = approval // 直接赋值审批结果（1或2）
	auditLog := models.AuditLog{
		AdminID:  adminID,
		Action:   "approve_report",
		TargetID: postID,
		Detail:   fmt.Sprintf(`{"approval": %d}`, approval),
	}
	if err := tx.Create(&auditLog).Error; err != nil {
		tx.Rollback()
		return &models.ServiceError{Code: 1011, Message: "记录审计日志失败"}
	}
	if err := tx.Save(&block).Error; err != nil {
		tx.Rollback()
		return &models.ServiceError{
			Code:    1010, // 新增错误码：更新举报状态失败
			Message: "更新举报状态失败: " + err.Error(),
		}
	}
	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return &models.ServiceError{
			Code:    1005,
			Message: "事务提交失败: " + err.Error(),
		}
	}

	return nil
}
