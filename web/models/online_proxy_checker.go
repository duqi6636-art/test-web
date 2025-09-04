package models

import (
	"fmt"
	"time"
)

type LogOnlineChecker struct {
	ID         int    `json:"id"`
	Uid        int    `json:"uid"`         //用户ID
	Username   string `json:"username"`    //检测账户信息
	RspIp      string `json:"rsp_ip"`      //检测IP
	RspRegion  string `json:"rsp_region"`  //检测IP地区
	HttpCode   int    `json:"http_code"`   //检测结果
	Duration   int    `json:"duration"`    //检测延迟时间 -ms毫秒
	UserIp     string `json:"user_ip"`     //用户IP
	UserRegion string `json:"user_region"` //用户IP地区
	Today      int    `json:"today"`
	Cate       string `json:"cate"`        //检测类型 http https  socks5
	CreateTime int    `json:"create_time"` //检测上报时间
}

// BatchLogOnlineChecker 批量添加日志
func BatchLogOnlineChecker(list []LogOnlineChecker) error {
	// 输入验证
	if len(list) == 0 {
		return nil
	}

	date := time.Now().Format("200601")
	var tName = "log_online_checker" + date

	// 检查表是否存在
	if !db.HasTable(tName) {
		if err := createLogOnlineCheckerTable(tName); err != nil {
			return fmt.Errorf("failed to create table %s: %v", tName, err)
		}
	}

	// 使用事务批量插入数据
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, item := range list {
		if err := tx.Table(tName).Create(&item).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// 创建表
func createLogOnlineCheckerTable(tableName string) error {
	createTable := `CREATE TABLE ` + tableName + `(
		id int NOT NULL AUTO_INCREMENT COMMENT 'ID',
  		uid int DEFAULT '0' COMMENT '用户id',
  		username varchar(255) DEFAULT '' COMMENT '检测账户信息',
  		rsp_ip varchar(60) DEFAULT '' COMMENT '检测IP',
  		rsp_region varchar(60) DEFAULT '' COMMENT '检测IP地区',
  		http_code int DEFAULT NULL COMMENT '检测结果状态码',
  		duration int(11) DEFAULT '0' COMMENT '检测延迟时间-ms 毫秒',
  		user_ip varchar(60) DEFAULT '' COMMENT '用户IP',
  		user_region varchar(60) DEFAULT '' COMMENT '用户IP所在地区',
  		today int(11) NOT NULL DEFAULT '0' COMMENT '零时时间戳',
  		cate varchar(30) DEFAULT '' COMMENT '类型 http https socks5',
  		create_time int DEFAULT '0' COMMENT '检测上报时间',
  		PRIMARY KEY (id) USING BTREE,
  		KEY uid (uid) USING BTREE,
  		KEY rsp_region (rsp_region) USING BTREE,
		KEY create_time (create_time) USING BTREE,
  		KEY today (today)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='在线监测日志'`
	return db.Exec(createTable).Error
}
