package models

import "time"

type Block struct {
	ID        uint
	UserID    uint
	TargetID  uint
	Reason    string    `gorm:"type:text"`
	Status    int       `gorm:"default:0"` // 0-待审核, 1-已通过, 2-已拒绝
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type BlockResponse struct {
	PostID  uint   `json:"post_id"`
	Content string `json:"content"`
	Reason  string `json:"reason"`
	Status  int    `json:"status"`
}
