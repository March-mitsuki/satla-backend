package controllers

// func connectionDB() (*gorm.DB, error) {
// 	db, err := gorm.Open(mysql.New(mysql.Config{
// 		DSN:                     model.Dsn,
// 		DontSupportRenameIndex:  true,
// 		DontSupportRenameColumn: true,
// 	}), &gorm.Config{
// 		DisableForeignKeyConstraintWhenMigrating: true,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	db.AutoMigrate(&model.Subtitle{}, &model.Project{})
// 	return db, nil
// }

// var db, _ = connectionDB()
// var rdb = redis.NewClient(&redis.Options{
// 	Addr: "localhost:6379",
// })

// func getAllSubtitles() ([]model.Subtitle, error) {
// 	var subtitles []model.Subtitle
// 	result := db.Find(&subtitles)
// 	if result.Error != nil {
// 		fmt.Printf("get all subtitles err: %v \n", result.Error)
// 		return nil, result.Error
// 	}
// 	return subtitles, nil
// }
