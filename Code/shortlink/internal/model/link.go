package model

import "time"

// Link 对应 MySQL 表 links。
type Link struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Code       string    `json:"code" gorm:"size:16;uniqueIndex;not null"`
	LongURL    string    `json:"long_url" gorm:"size:2048;not null"`
	ClickCount int64     `json:"click_count" gorm:"not null;default:0"`
	CreatedAt  time.Time `json:"created_at"`
}
