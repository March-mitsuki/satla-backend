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

type Project struct {
	CustomeModel `gorm:"embedded"`
	ProjectName  string `gorm:"not null;type:varchar(128);uniqueIndex" json:"project_name"`
	Description  string `gorm:"not null;type:varchar(256)" json:"description"`
	PointMan     string `gorm:"not null;type:varchar(64)" json:"point_man"`
	CreatedBy    string `gorm:"not null;type:varchar(128)" json:"created_by"`
}

type Room struct {
	CustomeModel `gorm:"embedded"`
	ProjectId    uint   `gorm:"not null" json:"project_id"`
	RoomName     string `gorm:"not null;type:varchar(128)" json:"room_name"`
	Description  string `gorm:"not null;type:varchar(256)" json:"description"`
	RoomType     uint   `gorm:"not null;type:tinyint" json:"room_type"` // 1 -> nomal, 2 -> auto
}

type Subtitle struct {
	CustomeModel `gorm:"embedded"`
	InputTime    string        `gorm:"not null;type:varchar(64)" json:"input_time"`
	SendTime     *sql.NullTime `json:"send_time"` // 为null则为未发送
	RoomId       uint          `gorm:"not null" json:"room_id"`
	TranslatedBy string        `gorm:"not null;type:varchar(128)" json:"translated_by"`
	CheckedBy    string        `gorm:"type:varchar(128)" json:"checked_by"` // 为空字符串则为未校对
	SendBy       string        `gorm:"type:varchar(128)" json:"send_by"`
	Subtitle     string        `gorm:"type:text" json:"subtitle"` // 翻译
	Origin       string        `gorm:"type:text" json:"origin"`   // 原文
}

type SubtitleOrder struct {
	CustomeModel `gorm:"embedded"`
	RoomId       uint   `gorm:"not null"`
	Order        string `gorm:"type:text"`
}

type User struct {
	CustomeModel `gorm:"embedded"`
	UserName     string `gorm:"not null;type:varchar(128);unique"`
	Email        string `gorm:"not null;type:varchar(256);uniqueIndex"`
	Permission   *uint  `gorm:"not null;default:1"` // 0 -> 测试用户, 1 -> 普通用户, 2 -> 管理员
	PasswordHash string `gorm:"not null"`
}

type AutoList struct {
	CustomeModel  `gorm:"embedded"`
	RoomId        uint   `gorm:"not null" json:"room_id"`
	FirstSubtitle string `gorm:"not null;type:text" json:"first_subtitle"`
	FirstOrigin   string `gorm:"type:text" json:"first_origin"`
	Memo          string `gorm:"type:varchar(128)" json:"memo"`
	IsSent        bool   `gorm:"not null;default:false" json:"is_sent"`
}

type AutoSub struct {
	CustomeModel `gorm:"embedded"`
	RoomId       uint    `gorm:"not null" json:"room_id"`
	ListId       uint    `gorm:"not null" json:"list_id"`
	Subtitle     string  `gorm:"type:text" json:"subtitle"`
	Origin       string  `gorm:"type:text" json:"origin"`
	Start        float64 `gorm:"type:decimal(10,2)" json:"start"`
	End          float64 `gorm:"type:decimal(10,2)" json:"end"`
	Duration     float64 `gorm:"type:decimal(10,2)" json:"duration"`
}
