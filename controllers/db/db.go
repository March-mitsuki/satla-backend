package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"
	"vvvorld/controllers/password"
	"vvvorld/model"

	"github.com/go-redis/redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var Mdb *gorm.DB

func ConnectionDB() error {
	dns := os.Getenv("DB_DSN")
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                     dns,
		DontSupportRenameIndex:  true,
		DontSupportRenameColumn: true,
	}), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return err
	}
	Mdb = db
	Mdb.AutoMigrate(&model.Subtitle{}, &model.Project{}, &model.SubtitleOrder{}, &model.User{})
	return nil
}

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
		newOrder := model.SubtitleOrder{
			ProjectId: project.ID,
			Order:     ",",
		}
		newOrderResult := Mdb.Create(&newOrder)
		if newOrderResult.Error != nil {
			return nil, "", newOrderResult.Error
		}
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
		sql := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
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
		sqlResult := tx.Exec(sql)
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
		sql := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
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
		sqlResult := tx.Exec(sql)
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
	result := Mdb.Model(&model.Subtitle{}).Where(
		"id = ?",
		arg.ID,
	).Select("checked_by", "subtitle", "origin").Updates(model.Subtitle{
		CheckedBy: arg.CheckedBy,
		Subtitle:  arg.Subtitle,
		Origin:    arg.Origin,
	})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func CreateTranslatedSub(sub model.Subtitle, pname string) (model.Subtitle, error) {
	err := Mdb.Transaction(func(tx *gorm.DB) error {
		var project model.Project
		searchResult := Mdb.Where("project_name = ?", pname).First(&project)
		if searchResult.Error != nil {
			return searchResult.Error
		}
		(&sub).ProjectId = project.ID
		createResult := tx.Create(&sub)
		if createResult.Error != nil {
			return createResult.Error
		}
		sql := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
			orderResults := tx.Model(
				&model.SubtitleOrder{},
			).Where(
				"project_id = ?",
				sub.ProjectId,
			).Update(
				"order",
				gorm.Expr(
					"CONCAT(`order`, '?,')",
					sub.ID,
				),
			)
			if orderResults.Error != nil {
				panic(orderResults.Error)
			}
			return orderResults
		})
		sqlResult := tx.Exec(sql)
		if sqlResult.Error != nil {
			return sqlResult.Error
		}
		return nil
	})
	if err != nil {
		return model.Subtitle{}, err
	}
	return sub, nil
}

func DeleteSubtitle(sub model.Subtitle) error {
	err := Mdb.Transaction(func(tx *gorm.DB) error {
		delResult := tx.Delete(&sub)
		if delResult.Error != nil {
			return delResult.Error
		}
		sql := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
			orderResults := tx.Model(
				&model.SubtitleOrder{},
			).Where(
				"project_id = ?",
				sub.ProjectId,
			).Update(
				"order",
				gorm.Expr(
					"REPLACE(`order`, ',?,', ',')",
					sub.ID,
				),
			)
			if orderResults.Error != nil {
				panic(orderResults.Error)
			}
			return orderResults
		})
		sqlResult := tx.Exec(sql)
		if sqlResult.Error != nil {
			return sqlResult.Error
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func ReorderSubtitle(projectId, dragId, dropId uint) error {
	// 当前不论从前往后拖还是从后往前拖, 拖动元素永远在放置元素的前面
	// 所以db只需要一个逻辑
	err := Mdb.Transaction(func(tx *gorm.DB) error {
		var subtitles []model.Subtitle
		searchResult := Mdb.Find(&subtitles, []uint{dragId, dropId})
		if searchResult.Error != nil {
			fmt.Printf("无法确认是否存在该字幕, 无法交换位置")
			return searchResult.Error
		} else if len(subtitles) < 2 {
			return errors.New("只存在一方的字幕, 无法交换位置")
		}
		sql := Mdb.ToSQL(func(tx *gorm.DB) *gorm.DB {
			orderResults := tx.Model(
				&model.SubtitleOrder{},
			).Where(
				"project_id = ?",
				projectId,
			).Update(
				"order",
				gorm.Expr(
					"REPLACE(?, ',?,', ',?,?,')",
					gorm.Expr(
						"REPLACE(`order`, ',?,', ',')",
						dragId,
					),
					dropId,
					dragId,
					dropId,
				),
			)
			if orderResults.Error != nil {
				panic(orderResults.Error)
			}
			return orderResults
		})
		// fmt.Printf("\n --- reorder sql is: \n - %v ---", sql)
		sqlResult := Mdb.Exec(sql)
		if sqlResult.Error != nil {
			return sqlResult.Error
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func DirectSendSubtitle(sub model.Subtitle, pname string) (model.Subtitle, error) {
	// 直接发送会根据client发过来的sub新建一行已经被软删除了的subtitle (不更新order)
	err := Mdb.Transaction(func(tx *gorm.DB) error {
		var project model.Project
		searchResult := tx.Where("project_name = ?", pname).First(&project)
		if searchResult.Error != nil {
			return searchResult.Error
		}
		(&sub).SendTime = &sql.NullTime{Time: time.Now(), Valid: true}
		(&sub).ProjectId = project.ID
		(&sub).DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
		createResult := tx.Create(&sub)
		if createResult.Error != nil {
			return createResult.Error
		}
		return nil
	})
	if err != nil {
		return model.Subtitle{}, err
	}
	return sub, nil
}

func SendSubtitle(sub model.Subtitle) error {
	// 发送字幕会软删除并且更新send_by行, 然后更新order行
	err := Mdb.Transaction(func(tx *gorm.DB) error {
		updateResult := tx.Model(
			&sub,
		).Select(
			"send_by",
			"deleted_at",
			"send_time",
		).UpdateColumns(model.Subtitle{
			SendBy:   sub.SendBy,
			SendTime: &sql.NullTime{Time: time.Now(), Valid: true},
			CustomeModel: model.CustomeModel{
				DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
			},
		})
		if updateResult.Error != nil {
			return updateResult.Error
		}
		sql := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
			orderResults := tx.Model(
				&model.SubtitleOrder{},
			).Where(
				"project_id = ?",
				sub.ProjectId,
			).Update(
				"order",
				gorm.Expr(
					"REPLACE(`order`, ',?,', ',')",
					sub.ID,
				),
			)
			if orderResults.Error != nil {
				panic(orderResults.Error)
			}
			return orderResults
		})
		sqlResult := tx.Exec(sql)
		if sqlResult.Error != nil {
			return sqlResult.Error
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func ChangeUserPassword(arg ArgChangeUserPassword) error {
	err := Mdb.Transaction(func(tx *gorm.DB) error {
		var user model.User
		searchResult := tx.Clauses(clause.Locking{
			Strength: "UPDATE",
			Options:  "NOWAIT",
		}).First(&user, arg.ID)
		if searchResult.Error != nil {
			return searchResult.Error
		}
		passErr := password.ComparePassword(user.PasswordHash, arg.OldPass)
		if passErr != nil {
			return passErr
		}
		newPassHash, encryptPassErr := password.EncryptPassword(arg.NewPass)
		if encryptPassErr != nil {
			return encryptPassErr
		}
		updateResult := tx.Model(&user).Update("password_hash", newPassHash)
		if updateResult.Error != nil {
			return updateResult.Error
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
