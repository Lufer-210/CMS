package models

import "time"

type AuditLog struct {
	ID        uint
	AdminID   uint
	Action    string
	TargetID  uint
	Detail    string    `gorm:"type:json"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
