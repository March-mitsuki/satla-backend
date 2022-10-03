package controller

import (
	"time"
	"vvvorld/model"

	"github.com/go-redis/redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func connectionDB() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                     model.Dsn,
		DontSupportRenameIndex:  true,
		DontSupportRenameColumn: true,
	}), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&model.Subtitle{}, &model.Project{})
	return db, nil
}

var db, _ = connectionDB()
var rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

func TestCreate() (model.Subtitle, error) {
	subtitle := model.Subtitle{
		InputTime:    "26:67:91",
		SendTime:     time.Now(),
		ProjectId:    1,
		ProjectName:  "default",
		TranslatedBy: "三月",
	}
	createResult := db.Create(&subtitle)
	if createResult.Error != nil {
		subtitle.ID = 0
		return subtitle, createResult.Error
	}
	return subtitle, nil
}
