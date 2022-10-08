package db

import (
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
	db.AutoMigrate(&model.Subtitle{}, &model.Project{}, &model.SubtitleOrder{}, &model.User{})
	return db, nil
}

var Mdb, _ = connectionDB()
var Rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

func TestCreate() {
	subtitle := model.Subtitle{
		InputTime:    "93:67:88",
		ProjectId:    1,
		TranslatedBy: "SanYue",
		Subtitle:     "test insert!@#$%^&*#@(!)",
	}
	Mdb.Create(&subtitle)
}

func GetRoomSubtitles(roomid string) ([]model.Subtitle, error) {
	var project model.Project
	pidResult := Mdb.Where("project_name = ?", roomid).First(&project)
	if pidResult.Error != nil {
		return nil, pidResult.Error
	}
	var subtitles []model.Subtitle
	subResult := Mdb.Where("project_id = ?", project.ID).Find(&subtitles)
	if subResult.Error != nil {
		return nil, subResult.Error
	}
	return subtitles, nil
}
