package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/March-mitsuki/satla-backend/controllers/password"
	"github.com/March-mitsuki/satla-backend/model"
	"github.com/March-mitsuki/satla-backend/utils/logger"

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
		CreateBatchSize:                          1000,
	})
	if err != nil {
		return err
	}
	Mdb = db
	Mdb.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.Room{},
		&model.Subtitle{},
		&model.SubtitleOrder{},
		&model.AutoList{},
		&model.AutoSub{},
	)
	return nil
}

var Rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

func GetRoomSubtitles(roomId uint) ([]model.Subtitle, string, error) {
	var subtitles []model.Subtitle
	var order model.SubtitleOrder
	err := Mdb.Transaction(func(tx *gorm.DB) error {
		var room model.Room
		pidResult := tx.First(&room, roomId)
		if pidResult.Error != nil {
			return pidResult.Error
		}
		subResult := tx.Where("room_id = ?", room.ID).Find(&subtitles)
		if subResult.Error != nil {
			return subResult.Error
		}
		orderResult := tx.Where("room_id = ?", room.ID).First(&order)
		if orderResult.Error != nil {
			newOrder := model.SubtitleOrder{
				RoomId: room.ID,
				Order:  ",",
			}
			newOrderResult := tx.Create(&newOrder)
			if newOrderResult.Error != nil {
				return newOrderResult.Error
			}
		}
		return nil
	})
	if err != nil {
		return []model.Subtitle{}, "", err
	}
	return subtitles, order.Order, nil
}

func CreateSubtitleUp(arg ArgAddSubtitle) (uint, error) {
	subtitle := model.Subtitle{
		InputTime:    "00:00:00",
		RoomId:       arg.RoomId,
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
				arg.RoomId,
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
		RoomId:       arg.RoomId,
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
				arg.RoomId,
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

func CreateTranslatedSub(sub model.Subtitle) (model.Subtitle, error) {
	err := Mdb.Transaction(func(tx *gorm.DB) error {
		var room model.Room
		searchResult := tx.First(&room, sub.RoomId)
		if searchResult.Error != nil {
			return searchResult.Error
		}
		(&sub).RoomId = room.ID
		createResult := tx.Create(&sub)
		if createResult.Error != nil {
			return createResult.Error
		}
		sql := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
			orderResults := tx.Model(
				&model.SubtitleOrder{},
			).Where(
				"room_id = ?",
				sub.RoomId,
			).Update(
				"order",
				gorm.Expr(
					"CONCAT(`order`, '?,')",
					sub.ID,
				),
			)
			if orderResults.Error != nil {
				panic(fmt.Sprintf("update subtitle order to sql err: %v\n", orderResults.Error))
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
				"room_id = ?",
				sub.RoomId,
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

func ReorderSubtitle(roomId, dragId, dropId uint) error {
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
				"room_id = ?",
				roomId,
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

func DirectSendSubtitle(sub model.Subtitle) (model.Subtitle, error) {
	// 直接发送会根据client发过来的sub新建一行已经被软删除了的subtitle (不更新order)
	(&sub).SendTime = &sql.NullTime{Time: time.Now(), Valid: true}
	(&sub).DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	createResult := Mdb.Create(&sub)
	if createResult.Error != nil {
		return model.Subtitle{}, createResult.Error
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
				"room_id = ?",
				sub.RoomId,
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

func GetRoomAutoLists(roomId uint) ([]model.AutoList, error) {
	var autoLists []model.AutoList
	err := Mdb.Transaction(func(tx *gorm.DB) error {
		var room model.Room
		pidResult := tx.First(&room, roomId)
		if pidResult.Error != nil {
			return pidResult.Error
		}
		subResult := tx.Where("room_id = ?", room.ID).Find(&autoLists)
		if subResult.Error != nil {
			return subResult.Error
		}
		return nil
	})
	if err != nil {
		return autoLists, err
	}
	return autoLists, nil
}

func AddAutoSub(arg ArgAddAutoSub) (model.AutoList, error) {
	var autoList model.AutoList
	err := Mdb.Transaction(func(tx *gorm.DB) error {
		autoList = model.AutoList{
			RoomId:        arg.AutoSubs[0].RoomId,
			FirstSubtitle: arg.AutoSubs[0].Subtitle,
			FirstOrigin:   arg.AutoSubs[0].Origin,
			Memo:          arg.Memo,
		}
		createListResult := tx.Create(&autoList)
		if createListResult.Error != nil {
			return createListResult.Error
		}
		for i := 0; i < len(arg.AutoSubs); i++ {
			elem := &arg.AutoSubs[i]
			elem.ListId = autoList.ID
		}
		createAllAutoSub := tx.Create(&arg.AutoSubs)
		if createAllAutoSub.Error != nil {
			return createAllAutoSub.Error
		}
		return nil
	})
	if err != nil {
		return model.AutoList{}, err
	}
	return autoList, nil
}

func DeleteAutoSub(listId uint) error {
	err := Mdb.Transaction(func(tx *gorm.DB) error {
		delSubResult := tx.Where("list_id = ?", listId).Delete(&model.AutoSub{})
		if delSubResult.Error != nil {
			return delSubResult.Error
		}
		delListResult := tx.Delete(&model.AutoList{}, listId)
		if delListResult.Error != nil {
			return delListResult.Error
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func HandleAutoPlayStart(listId uint) ([]model.AutoSub, error) {
	// 将对应list设置为已播放, 并返回对应所有autoSubs
	var autoSubs []model.AutoSub
	err := Mdb.Transaction(func(tx *gorm.DB) error {
		findResult := tx.Where("list_id = ?", listId).Find(&autoSubs)
		if findResult.Error != nil {
			return findResult.Error
		}
		updateResult := tx.Model(&model.AutoList{}).Where("id = ?", listId).Update("is_sent", true)
		if updateResult.Error != nil {
			return updateResult.Error
		}
		return nil
	})
	if err != nil {
		return autoSubs, err
	}
	return autoSubs, nil
}

func SetRoomListsUnsent(roomId uint) error {
	result := Mdb.Model(&model.AutoList{}).Where("room_id = ?", roomId).Update("is_sent", false)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func ChangeAutoMemo(listId uint, memo string) error {
	result := Mdb.Model(&model.AutoList{}).Where("id = ?", listId).Update("memo", memo)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func BatchAddSubs(subs []model.Subtitle) error {
	roomId := subs[0].RoomId
	err := Mdb.Transaction(func(tx *gorm.DB) error {
		var room model.Room
		searchResult := tx.First(&room, roomId)
		if searchResult.Error != nil {
			return searchResult.Error
		}
		createResult := tx.Create(&subs)
		if createResult.Error != nil {
			return createResult.Error
		}
		var subsIdStr []string
		for _, v := range subs {
			subsIdStr = append(subsIdStr, strconv.FormatUint(uint64(v.ID), 10))
		}
		jsonStr := strings.Join(subsIdStr, ",")
		sql := tx.ToSQL(func(tx *gorm.DB) *gorm.DB {
			orderResults := tx.Model(
				&model.SubtitleOrder{},
			).Where(
				"room_id = ?",
				roomId,
			).Update(
				"order",
				gorm.Expr(
					"CONCAT(`order`, ?)", // 因为gorm会把string自动加单引号,所以这里去掉?两边的单引号
					jsonStr+",",
				),
			)
			if orderResults.Error != nil {
				panic(fmt.Sprintf("update subtitle order to sql err: %v\n", orderResults.Error))
			}
			return orderResults
		})
		logger.Info("db", fmt.Sprintf("tosql: %v", sql))
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

func CheckWsroomType(roomType uint, roomId uint) error {
	if roomType != 1 && roomType != 2 {
		return errors.New("not on room type")
	}
	var room model.Room
	findResult := Mdb.First(&room, roomId)
	if findResult.Error != nil {
		return findResult.Error
	}
	if roomType != room.RoomType {
		return errors.New("not match the current room type")
	}
	return nil
}
