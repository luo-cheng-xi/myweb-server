package dao

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"myweb/app/user/internal/conf"
	"myweb/app/user/internal/model/po"
)

// NewData 初始化gorm数据库连接.
func NewData(conf *conf.Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(conf.MysqlConf.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("gorm.Open error %#v", err)
	}
	//完成建表工作
	migrate(db)
	return db
}

// 完成建表工作的方法
func migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&po.User{},
	)
	if err != nil {
		log.Fatal("autoMigrate error")
	}
}
