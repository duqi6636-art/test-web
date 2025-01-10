package models

import (
	"api-360proxy/web/pkg/util"
	"time"
)

type LogLogin struct {
	ID           int32  `json:"id"`
	UID          int    `json:"uid"`
	UserLogin    string `json:"user_login"`
	Ip           string `json:"ip"`
	LoginTime    int64  `json:"login_time"`
	Platform     string `json:"platform"`
	Browser      string `json:"browser"`
	OsInfo       string `json:"os_info"`
	Version      string `json:"version"`
	RegTime      int    `json:"reg_time"`
	Country      string `json:"country"`
	Language     string `json:"language"`
	Lang         string `json:"lang"`
	Today        int    `json:"today"`
	Cate         string `json:"cate"`
	TimeZone     string `json:"time_zone"`
	DeviceNumber string `json:"device_number"`
	UserAgent    string `json:"user_agent"`
	IsPay        int    `json:"is_pay"` //是否是已购买
}

var tableName = "log_login" + time.Now().Format("200601")

func AddLoginLog(info LogLogin) bool {
	if !db.HasTable(tableName) {
		createLoginLogTable(tableName)
	}
	return db.Table(tableName).Create(&info).Error == nil
}

// 创建表
func createLoginLogTable(tableName string) {
	createTable := `CREATE TABLE ` + tableName + `(
		id int NOT NULL AUTO_INCREMENT COMMENT 'ID',
  		uid int DEFAULT '0' COMMENT '用户id',
  		user_login varchar(50) DEFAULT '' COMMENT '登录用户名',
  		ip varchar(60) DEFAULT '' COMMENT '登录ip',
  		login_time int DEFAULT '0' COMMENT '登录时间',
  		platform varchar(30) DEFAULT '' COMMENT '登录终端',
  		cate varchar(50) DEFAULT '' COMMENT '日志类型',
  		browser varchar(255) DEFAULT '' COMMENT '浏览器信息',
		user_agent varchar(255) DEFAULT '' COMMENT '浏览器UA',
  		os_info varchar(255) DEFAULT '' COMMENT '系统信息',
  		version varchar(50) DEFAULT '' COMMENT '版本',
  		channel varchar(50) DEFAULT '' COMMENT '渠道',
  		country varchar(255) DEFAULT '' COMMENT 'IP所在国家',
  		language varchar(255) DEFAULT '' COMMENT '电脑语言',
  		lang varchar(50) DEFAULT '' COMMENT '客户端语言',
  		time_zone varchar(255) DEFAULT '' COMMENT '时区',
  		today int(11) NOT NULL DEFAULT '0' COMMENT '零时时间戳',
  		is_pay int(11) NOT NULL DEFAULT '0' COMMENT '是否付费',
  		reg_time int DEFAULT '0' COMMENT '注册时间',
		device_number varchar(100) DEFAULT NULL COMMENT '设备号',
  		PRIMARY KEY (id) USING BTREE,
  		KEY uid (uid) USING BTREE,
  		KEY country (country) USING BTREE,
		KEY user_login (user_login) USING BTREE,
		KEY login_time (login_time) USING BTREE,
  		KEY today (today)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='登录日志'`
	db.Exec(createTable)
}

// 查询当日登陆数据
func GettodayLogin(uid int) int {
	today := util.GetTodayTime()
	num := 0
	db.Table(tableName).Where("today = ? and uid = ?", today, uid).Count(&num)
	return num
}
