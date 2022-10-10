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

func GetRoomSubtitles(roomid string) ([]model.Subtitle, string, error) {
	var project model.Project
	pidResult := Mdb.Where("project_name = ?", roomid).First(&project)
	if pidResult.Error != nil {
		return nil, "", pidResult.Error
	}
	var subtitles []model.Subtitle
	subResult := Mdb.Where("project_id = ?", project.ID).Find(&subtitles)
	if subResult.Error != nil {
		return nil, "", subResult.Error
	}
	var order model.SubtitleOrder
	orderResult := Mdb.Where("project_id = ?", project.ID).First(&order)
	if orderResult.Error != nil {
		return nil, "", orderResult.Error
	}
	return subtitles, order.Order, nil
}

func CreateSubtitleUp(arg ArgAddSubtitle) (uint, error) {
	subtitle := model.Subtitle{
		InputTime:    "00:00:00",
		ProjectId:    arg.ProjectId,
		TranslatedBy: arg.CheckedBy,
		CheckedBy:    arg.CheckedBy,
	}
	err := Mdb.Transaction(func(tx *gorm.DB) error {
		createResult := tx.Create(&subtitle)
		if createResult.Error != nil {
			return createResult.Error
		}
		// 我不知道为什么创建sql就可以,直接执行就提示我变量不匹配,估计是gorm.Expr和sql哪里不匹配
		// orderResult := tx.Model(
		// 	&model.SubtitleOrder{},
		// ).Where(
		// 	"project_id = ?",
		// 	arg.ProjectId,
		// ).Update(
		// 	"order",
		// 	gorm.Expr(
		// 		"REPLACE(`order`, ',?,', ',?,?,')",
		// 		arg.PreSubtitleId,
		// 		subtitle.ID,
		// 		arg.PreSubtitleId,
		// 	),
		// )
		// if orderResult.Error != nil {
		// 	return orderResult.Error
		// }
		sql := Mdb.ToSQL(func(tx *gorm.DB) *gorm.DB {
			orderResults := tx.Model(
				&model.SubtitleOrder{},
			).Where(
				"project_id = ?",
				arg.ProjectId,
			).Update(
				"order",
				gorm.Expr(
					"REPLACE(`order`, ',?,', ',?,?,')",
					arg.PreSubtitleId,
					subtitle.ID,
					arg.PreSubtitleId,
				),
			)
			if orderResults.Error != nil {
				panic(orderResults.Error)
			}
			return orderResults
		})
		sqlResult := Mdb.Exec(sql)
		if sqlResult.Error != nil {
			return sqlResult.Error
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return subtitle.ID, nil
}

func CreateSubtitleDown(arg ArgAddSubtitle) (uint, error) {
	subtitle := model.Subtitle{
		InputTime:    "00:00:00",
		ProjectId:    arg.ProjectId,
		TranslatedBy: arg.CheckedBy,
		CheckedBy:    arg.CheckedBy,
	}
	err := Mdb.Transaction(func(tx *gorm.DB) error {
		createResult := tx.Create(&subtitle)
		if createResult.Error != nil {
			return createResult.Error
		}
		sql := Mdb.ToSQL(func(tx *gorm.DB) *gorm.DB {
			orderResults := tx.Model(
				&model.SubtitleOrder{},
			).Where(
				"project_id = ?",
				arg.ProjectId,
			).Update(
				"order",
				gorm.Expr(
					"REPLACE(`order`, ',?,', ',?,?,')",
					arg.PreSubtitleId,
					arg.PreSubtitleId,
					subtitle.ID,
				),
			)
			if orderResults.Error != nil {
				panic(orderResults.Error)
			}
			return orderResults
		})
		sqlResult := Mdb.Exec(sql)
		if sqlResult.Error != nil {
			return sqlResult.Error
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return subtitle.ID, nil
}

func ChangeSubtitle(arg ArgChangeSubtitle) error {
	updateMap := map[string]interface{}{
		"translated_by": arg.TranslatedBy,
		"checked_by":    arg.CheckedBy,
		"subtitle":      arg.Subtitle,
		"origin":        arg.Origin,
	}
	result := Mdb.Model(&model.Subtitle{}).Where(
		"id = ?",
		arg.ID,
	).Updates(&updateMap)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
