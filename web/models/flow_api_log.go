package models

import (
	"api-360proxy/web/pkg/util"
	"time"
)

// 流量
type LogApiUseInfo struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`         // 用户id
	Num        int    `json:"num"`         // 数量
	Ip         string `json:"ip"`          // 用户IP
	Country    string `json:"country"`     // 国家地区
	Protocol   string `json:"protocol"`    // 协议
	Cate       string `json:"cate"`        // 类型
	Lt         string `json:"lt"`          // 符号
	St         string `json:"st"`          // 符号-其他
	CreateTime int    `json:"create_time"` // 创建时间
}

// 用户流量使用记录
func AddLogApiUseInfo(uid, num int, ip, country, protocol, cate, lt, st string) (err error) {
	log := LogApiUseInfo{
		Uid:        uid,
		Num:        num,
		Ip:         ip,
		Country:    country,
		Protocol:   protocol,
		Cate:       cate,
		Lt:         lt,
		St:         st,
		CreateTime: util.GetNowInt(),
	}
	date := time.Now().Format("20060102")
	var tableName = "log_api_use" + date
	if !StatisticsDb.Migrator().HasTable(tableName) {
		createApiUseLogTable(tableName)
	}
	err = StatisticsDb.Table(tableName).Create(&log).Error
	return
}

// 创建表
func createApiUseLogTable(tableName string) {
	createTables := `CREATE TABLE ` + tableName + `(
		id int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
		uid int(11) NOT NULL COMMENT '用户ID',
		num int(11) NOT NULL COMMENT '数量',
		ip varchar(30) DEFAULT '' COMMENT '用户IP',
		country varchar(30) DEFAULT '' COMMENT '国家地区',
		protocol varchar(30) DEFAULT '0' COMMENT '协议',
		cate varchar(30) DEFAULT '' COMMENT '类型',
		lt varchar(30) DEFAULT '' COMMENT '分隔符',
		st varchar(30) DEFAULT '' COMMENT '分隔符-其他',
		create_time int(11) DEFAULT '0' COMMENT '操作时间',
		PRIMARY KEY (id),
		KEY uid (uid) USING BTREE,
		KEY create_time (create_time) USING BTREE,
		KEY ip (ip) USING BTREE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Api 提取记录';`
	StatisticsDb.Exec(createTables)
}
