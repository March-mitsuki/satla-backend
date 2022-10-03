package model

import (
	"time"

	"gorm.io/gorm"
)

type Subtitle struct {
	gorm.Model
	InputTime    string    `gorm:"not null;type:varchar(64)"`
	SendTime     time.Time // 为null则为未发送
	ProjectId    int       `gorm:"not null"`
	ProjectName  string    `gorm:"not null;type:varchar(128)"`
	TranslatedBy string    `gorm:"not null;type:varchar(128)"`
	CheckedBy    string    `gorm:"type:varchar(128)"` // 为null则为未校对
	Subtitle     string    `gorm:"type:text"`         // 翻译
	Origin       string    `gorm:"type:text"`         // 原文
}

type Project struct {
	gorm.Model
	ProjectName string `gorm:"not null;type:varchar(128)"`
	Description string `gorm:"not null;type:varchar(256)"`
	Pointman    string `gorm:"not null;type:varchar(64)"`
	CreatedBy   string `gorm:"not null;type:varchar(128)"`
}

type User struct {
	gorm.Model
	UserName     string `gorm:"not null;type:varchar(128)"`
	Email        string `gorm:"not null;type:varchar(256);uniqueIndex"`
	PasswordHash string `gorm:"not null"`
}
