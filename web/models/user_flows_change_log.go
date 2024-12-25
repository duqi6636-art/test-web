package models

import (
	"api-360proxy/web/pkg/util"
	"time"
)

type UserFlowsChangeLog struct {
	ID          int    `gorm:"primary_key" json:"id"`
	Uid         int    `gorm:"not null"  json:"uid"`
	OldFlows    int64  `gorm:"not null" json:"old_flows"`
	NewFlows    int64  `gorm:"not null" json:"new_flows"`
	ChangeFlows int64  `gorm:"not null" json:"change_flows"`
	ChangeType  string `gorm:"not null" json:"change_type"`
	ChangeTime  int    `gorm:"not null" json:"change_time"`
	Mark        int    `gorm:"not null" json:"mark"`
}

// 用户流量变动记录
func AddUserFlowsChangeLog(info UserFlowsChangeLog) {
	usersChangeLogTable := "user_flows_change_log_" + time.Now().Format("20060102")
	if !StatisticsDb.Migrator().HasTable(usersChangeLogTable) {
		createUserFlowsChangeLogTable(usersChangeLogTable)
	}
	StatisticsDb.Table(usersChangeLogTable).Create(&info)
	return
}

// 创建表
func createUserFlowsChangeLogTable(tableName string) {
	createTable := `CREATE TABLE ` + tableName + `(
		id int(11) unsigned NOT NULL AUTO_INCREMENT,
		uid int(11) DEFAULT NULL COMMENT '用户ID',
		old_flows bigint(32) NOT NULL COMMENT '变动前流量',
		new_flows bigint(32) NOT NULL  COMMENT '变动后流量',
		change_flows bigint(32) NOT NULL  COMMENT '变动流量',
		change_time int(11) NOT NULL DEFAULT '0' COMMENT '变动时间',
		change_type varchar(50) NOT NULL  COMMENT '变动类型',
		mark int(4) DEFAULT NULL COMMENT '增加还是减少 1增加 -1减少',
		PRIMARY KEY (id) USING BTREE
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC COMMENT='用户流量变动记录'`
	StatisticsDb.Exec(createTable)
}

type UserIpChangeLog struct {
	ID         int    `gorm:"primary_key" json:"id"`
	Uid        int    `gorm:"not null"  json:"uid"`
	OldIp      int    `gorm:"not null" json:"old_ip"`
	NewIp      int    `gorm:"not null" json:"new_ip"`
	ChangeIp   int    `gorm:"not null" json:"change_ip"`
	ChangeType string `gorm:"not null" json:"change_type"`
	ChangeTime int    `gorm:"not null" json:"change_time"`
	Mark       int    `gorm:"not null" json:"mark"`
}

// 用户流量变动记录
func AddUserIpChangeLog(info UserIpChangeLog) {
	usersChangeLogTable := "user_ip_change_log_" + time.Now().Format("20060102")
	if !StatisticsDb.Migrator().HasTable(usersChangeLogTable) {
		createUserIpChangeLogTable(usersChangeLogTable)
	}
	StatisticsDb.Table(usersChangeLogTable).Create(&info)
	return
}

// 创建表
func createUserIpChangeLogTable(tableName string) {
	createUserIpChangeLogTableSql := `CREATE TABLE ` + tableName + `(
		id int(11) unsigned NOT NULL AUTO_INCREMENT,
		uid int(11) DEFAULT NULL COMMENT '用户ID',
		old_ip bigint(32) NOT NULL COMMENT '变动前流量',
		new_ip bigint(32) NOT NULL  COMMENT '变动后流量',
		change_ip bigint(32) NOT NULL  COMMENT '变动流量',
		change_time int(11) NOT NULL DEFAULT '0' COMMENT '变动时间',
		change_type varchar(50) NOT NULL  COMMENT '变动类型',
		mark int(4) DEFAULT NULL COMMENT '增加还是减少 1增加 -1减少',
		PRIMARY KEY (id) USING BTREE
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC COMMENT='用户Ip变动记录'`
	StatisticsDb.Exec(createUserIpChangeLogTableSql)
}


// 动态长效
type DynamicIspModel struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	Username   string `json:"username"`
	AccountId  int    `json:"account_id"`
	Value      int64  `json:"value"`
	PreValue   int64  `json:"pre_value"`
	UseValue   int64  `json:"use_value"`
	UserIp     string `json:"user_ip"`
	Mark       int    `json:"mark"` //符号标识 1增加 -1减少
	Cate       string `json:"cate"` //类型 pay购买 cdk 兑换，score积分等
	CreateTime int    `json:"create_time"`
}

// 用户流量使用记录
func AddDynamicIspLog(uid,accountId int, username string, value, preValue, useValue int64, userIp,cate string, mark int) (err error) {
	log := DynamicIspModel{
		Uid:        uid,
		AccountId:  accountId,
		Username:   username,
		Value:      value,
		PreValue:   preValue,
		UseValue:   useValue,
		UserIp:     userIp,
		Cate:       cate,
		Mark:       mark,
		CreateTime: util.GetNowInt(),
	}
	date := time.Now().Format("20060102")
	var tableNames = "log_dynamic_isp" + date
	if !StatisticsDb.Migrator().HasTable(tableNames) {
		createDynamicIspLogTable(tableNames)
	}
	err = StatisticsDb.Table(tableNames).Create(&log).Error
	return
}

//创建表
func createDynamicIspLogTable(tableName string) {
	createTables := `CREATE TABLE ` + tableName + `(
		id int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
		uid int(11) NOT NULL COMMENT '用户ID',
		username varchar(30) DEFAULT '' COMMENT '用户名',
		account_id int(11) NOT NULL COMMENT '子账户ID',
		mark tinyint(4) NOT NULL DEFAULT '0' COMMENT '符号标识 1增加 -1减少',
		cate varchar(30) DEFAULT '' COMMENT '类型 pay购买 cdk 兑换，score积分 flows帐密,white白名单,api接口 等',
		value bigint(20) DEFAULT '0' COMMENT '操作值',
		pre_value bigint(20) DEFAULT '0' COMMENT '操作前数量',
		use_value bigint(20) DEFAULT '0' COMMENT '操作后可用数量',
		user_ip varchar(50) DEFAULT '' COMMENT '用户IP',
		remark varchar(255) DEFAULT '',
		create_time int(11) DEFAULT '0' COMMENT '操作时间',
		PRIMARY KEY (id),
		KEY uid (uid) USING BTREE,
		KEY create_time (create_time) USING BTREE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户IPFlow余额记录';`
	StatisticsDb.Exec(createTables)
}
