package clickhousedb

import (
	"api-360proxy/web/pkg/setting"
	"fmt"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"
)

var ClickhouseDb *gorm.DB
var ClickhouseCherryLogDb *gorm.DB

func InitClickhouseDb() {
	var err error

	dbName := setting.ClickhouseDbConfig.DbName
	user := setting.ClickhouseDbConfig.User
	password := setting.ClickhouseDbConfig.Password
	host := setting.ClickhouseDbConfig.Host
	var dsn = fmt.Sprintf("clickhousedb://%s:%s@%s/%s?dial_timeout=30s&read_timeout=120s",
		user, password, host, dbName)
	ClickhouseDb, err = gorm.Open(clickhouse.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		var errMsg interface{}
		errMsg = err.Error()
		fmt.Println(errMsg)
	}
	if setting.RunMode == "debug" {
		ClickhouseDb.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second, // 慢查询阈值，超过这个阈值的查询将被认为是慢查询
				Colorful:                  true,        // 彩色输出
				IgnoreRecordNotFoundError: true,        // 忽略记录未找到的错误
				LogLevel:                  logger.Info, // 日志级别
			},
		)
	} else {
		ClickhouseDb.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,  // 慢查询阈值，超过这个阈值的查询将被认为是慢查询
				Colorful:                  true,         // 彩色输出
				IgnoreRecordNotFoundError: true,         // 忽略记录未找到的错误
				LogLevel:                  logger.Error, // 日志级别
			},
		)
	}
	ClickhouseDb = ClickhouseDb.Debug()
	fmt.Println("Connect to Clickhouse success!")
}

func InitClickhouseCherryLogDb() {
	var err error

	dbName := setting.ClickhouseDbConfig.CherryLogDbName
	user := setting.ClickhouseDbConfig.User
	password := setting.ClickhouseDbConfig.Password
	host := setting.ClickhouseDbConfig.Host
	var dsn = fmt.Sprintf("clickhousedb://%s:%s@%s/%s?dial_timeout=30s&read_timeout=120s",
		user, password, host, dbName)
	ClickhouseCherryLogDb, err = gorm.Open(clickhouse.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		var errMsg interface{}
		errMsg = err.Error()
		fmt.Println(errMsg)
	}
	if setting.RunMode == "debug" {
		ClickhouseCherryLogDb.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second, // 慢查询阈值，超过这个阈值的查询将被认为是慢查询
				Colorful:                  true,        // 彩色输出
				IgnoreRecordNotFoundError: true,        // 忽略记录未找到的错误
				LogLevel:                  logger.Info, // 日志级别
			},
		)
	} else {
		ClickhouseCherryLogDb.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,  // 慢查询阈值，超过这个阈值的查询将被认为是慢查询
				Colorful:                  true,         // 彩色输出
				IgnoreRecordNotFoundError: true,         // 忽略记录未找到的错误
				LogLevel:                  logger.Error, // 日志级别
			},
		)
	}
	ClickhouseCherryLogDb = ClickhouseCherryLogDb.Debug()
	fmt.Println("Connect to Clickhouse Cherry Log DB success!")
}
