package model

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type CustomeModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Subtitle struct {
	CustomeModel `gorm:"embedded"`
	InputTime    string        `gorm:"not null;type:varchar(64)" json:"input_time"`
	SendTime     *sql.NullTime `json:"send_time"` // 为null则为未发送
	ProjectId    uint          `gorm:"not null" json:"project_id"`
	TranslatedBy string        `gorm:"not null;type:varchar(128)" json:"translated_by"`
	CheckedBy    string        `gorm:"type:varchar(128)" json:"checked_by"` // 为空字符串则为未校对
	Subtitle     string        `gorm:"type:text" json:"subtitle"`           // 翻译
	Origin       string        `gorm:"type:text" json:"origin"`             // 原文
}

type SubtitleOrder struct {
	CustomeModel `gorm:"embedded"`
	ProjectId    uint   `gorm:"not null"`
	Order        string `gorm:"type:text"`
}

type Project struct {
	CustomeModel `gorm:"embedded"`
	ProjectName  string `gorm:"not null;type:varchar(128);uniqueIndex" json:"project_name"`
	Description  string `gorm:"not null;type:varchar(256)" json:"description"`
	PointMan     string `gorm:"not null;type:varchar(64)" json:"point_man"`
	CreatedBy    string `gorm:"not null;type:varchar(128)" json:"created_by"`
}

type User struct {
	CustomeModel `gorm:"embedded"`
	UserName     string `gorm:"not null;type:varchar(128);unique"`
	Email        string `gorm:"not null;type:varchar(256);uniqueIndex"`
	PasswordHash string `gorm:"not null"`
}
