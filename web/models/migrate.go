package models

import (
	"api-360proxy/web/db/clickhousedb"
	"fmt"
)

type FlowDayUnlimited struct {
	UID       int   `gorm:"default:0;index;comment:'用户id'" json:"uid"`
	AccountID int   `gorm:"default:0;index;comment:'用户代理表 id'" json:"account_id"`
	Today     int64 `gorm:"default:0;comment:'使用时间'" json:"today"`
	Flows     int64 `gorm:"default:0;comment:'使用时间'" json:"flows"`
	Hour      int   `gorm:"default:0;comment:'小时'" json:"hour"`
}

func GetOldDataList(tableName string) (list []FlowDayUnlimited, err error) {
	err = db.Table(tableName).Select("uid,account_id,today,sum(flows) as flows").Group("uid,account_id,today").Find(&list).Error
	return list, err
}

// 批量写入数据
func BatchInsertFlowDayLog(data []FlowDayUnlimited, month string) (err error) {
	tableName := "st_user_flow_day_used_"
	ym := month
	table := tableName + ym
	if !clickhousedb.ClickhouseDb.Migrator().HasTable(table) {
		createFlowDayTable(table)
	}

	err = clickhousedb.ClickhouseDb.Table(table).CreateInBatches(&data, len(data)).Error
	return err
}

// 创建表
func createFlowDayTable(tableName string) {
	sql := `CREATE TABLE IF NOT EXISTS ` + tableName + ` (
	  uid UInt32 DEFAULT 0 COMMENT '用户id',
	  account_id UInt32 DEFAULT 0 COMMENT '用户子账户ID',
	  today UInt32 DEFAULT 0 COMMENT '零点时间戳',
	  hour UInt32 DEFAULT 0 COMMENT '小时',
	  flows UInt64 DEFAULT 0 COMMENT '使用流量（字节）',
	) 
	ENGINE = MergeTree
	PARTITION BY toYYYYMMDD(toDateTime(today))
	ORDER BY (uid,today,account_id) 
	COMMENT '统计不限量使用记录';`
	if err := clickhousedb.ClickhouseDb.Exec(sql).Error; err != nil {
		fmt.Println("failed to create table:", tableName, "error:", err.Error())
	}
}
