package db

import (
	"fmt"
	"vvvorld/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestCreate() {
	subtitle := model.Subtitle{
		InputTime:    "93:67:88",
		ProjectId:    1,
		TranslatedBy: "SanYue",
		Subtitle:     "test insert!@#$%^&*#@(!)",
	}
	Mdb.Create(&subtitle)
}

func TestLockSql() {
	// update会自动上锁, 先读后写才需要上for update的行锁
	lockSql := Mdb.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var subtitle model.Subtitle
		return tx.Clauses(clause.Locking{
			Strength: "UPDATE",
			Options:  "NOWAIT",
		}).First(&subtitle, 2)
	})
	fmt.Printf("\n -------create lock sql:\n %v --------- \n", lockSql)
}

func TestCreateSubtitleUp(arg ArgAddSubtitle) {
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
				6,
				arg.PreSubtitleId,
			),
		)
		if orderResults.Error != nil {
			fmt.Printf("\n ---err here: \n %v -----\n", orderResults.Error)
			return orderResults
		}
		return orderResults
	})
	fmt.Printf("\n -------create sql:\n %v --------- \n", sql)
}

func TestExpr() {
	Mdb.Model(
		&model.SubtitleOrder{},
	).Update(
		"order",
		gorm.Expr(
			"REPLACE(`order`, ',?,', ',?,?,')",
			1,
			6,
			1,
		),
	).Where(
		"project_id = ?",
		1,
	)
}
