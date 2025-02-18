package models

import "time"

type LogModel struct {
	Id         int    `json:"id"`
	Code       string `json:"code"`
	Text       string `json:"text"`
	CreateTime string `json:"create_time"`
}

func AddLog(model LogModel) bool {
	date := time.Now().Format("200601")
	var tableNames = "cm_log" + date
	if !db.HasTable(tableNames) {
		createLogTable(tableNames)
	}
	res := db.Table(tableNames).Create(&model).Error == nil
	return res
}

// 创建表
func createLogTable(tableName string) {
	createTable := `CREATE TABLE ` + tableName + `(
		id int NOT NULL AUTO_INCREMENT COMMENT 'ID',
  		code varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT '' COMMENT '标识',
  		text text CHARACTER SET utf8 COLLATE utf8_general_ci,
  		create_time varchar(60) DEFAULT '' COMMENT '时间',
  		PRIMARY KEY (id) USING BTREE,
  		KEY code (code) USING BTREE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='日志'`
	db.Exec(createTable)
}
