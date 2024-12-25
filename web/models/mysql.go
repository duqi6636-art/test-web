package models

import (
	"api-360proxy/web/pkg/setting"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"
)

var StatisticsDb *gorm.DB

func StatisticsDbSetup() {
	var (
		err                          error
		dbName, user, password, host string
	)

	dbName = setting.DatabaseStatisticsConfig.DbName
	user = setting.DatabaseStatisticsConfig.User
	password = setting.DatabaseStatisticsConfig.Password
	host = setting.DatabaseStatisticsConfig.Host
	StatisticsDb, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user,
		password,
		host,
		dbName)), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	StatisticsDb.Logger = logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second, // 慢查询阈值，超过这个阈值的查询将被认为是慢查询
			Colorful:                  true,        // 彩色输出
			IgnoreRecordNotFoundError: true,        // 忽略记录未找到的错误
			LogLevel:                  logger.Info, // 日志级别
		},
	)

	if err != nil {
		log.Fatalln(err)
	}
}
