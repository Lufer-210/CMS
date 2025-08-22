package models

import (
	"time"
)

type Post struct {
	ID       uint
	Content  string `gorm:"type:text"`
	UserID   uint
	PostTime time.Time
}

type PostResponse struct {
	ID      uint   `json:"id"`
	Content string `json:"content"`
	UserID  uint   `json:"user_id"`
	Time    string `json:"time"`
	Likes   int    `json:"likes"`
}

func (p Post) ToResponse() PostResponse {
	return PostResponse{
		ID:      p.ID,
		Content: p.Content,
		UserID:  p.UserID,
		Time:    p.PostTime.Format("2006-01-02T15:04:05.000-07:00"),
		Likes:   0,
	}
}
