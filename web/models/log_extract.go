package models

import "time"

type LogExtract struct {
	ID         int    `json:"id"`
	Uid        int    `json:"uid"`
	UserName   string `json:"user_name"`
	CreateTime int    `json:"create_time"`
}
type ResLogExtract struct {
	UserName   string `json:"user_name"`
	CreateTime string `json:"create_time"`
}

// 提取记录
var logExtractTable = "cm_log_extract" + time.Now().Format("200601")

func GetIpDeductionCount(uid int) (num int) {
	dbRead.Table(logExtractTable).Where("uid =? ", uid).Count(&num)
	return
}

// 获取分页列表
func GetIpDeductionPage(uid, offset, limit int) (data []LogExtract) {
	dbRead.Table(logExtractTable).Where("uid =? ", uid).Offset(offset).Limit(limit).Order("id desc", true).Find(&data)
	return
}

// 获取分组列表
type NumIpInfo struct {
	Num   int     `json:"num"`
	Today string  `json:"today"`
	Cate  string  `json:"cate"`
	Unit  float64 `json:"unit"`
}

// 提取IP明细
type ExtractIpInfo struct {

	ExtractTime string  `json:"extractTime" gorm:"column:extractTime"`
	IP  string  `json:"ip"`

}

func GetIpDeductionByGroup(uid, create_time int) (data []NumIpInfo) {
	dbRead.Table(logExtractTable).Select("count(id) as num,sum(unit) as unit,FROM_UNIXTIME(create_time,'%Y-%m-%d') as today,cate").Where("uid =? ", uid).Where("create_time >=? ", create_time).Group("today,cate").Find(&data)
	return
}

func GetIpCountByUidAndTime(uid, start, end int, monthArr []string) (data []NumIpInfo) {
	for _, month := range monthArr {
		monthData := []NumIpInfo{}
		monthTableName := "cm_log_extract" + month
		dbRead.Table(monthTableName).
			Select("count(id) as num,sum(unit) as unit,FROM_UNIXTIME(create_time,'%Y-%m-%d') as today,cate").
			Where("uid =? ", uid).
			Where("create_time >=? and create_time <=? ", start, end).
			Group("today,cate").Find(&monthData)
		data = append(data, monthData...)
	}
	return
}
// 获取提取时间提取ip
func GetTimeAndIpByUidAndDate(uid, start, end int, monthArr []string) (data []ExtractIpInfo) {
	for _, month := range monthArr {
		monthData := []ExtractIpInfo{}
		monthTableName := "cm_log_extract" + month

		dbRead.Table(monthTableName).
			Select("FROM_UNIXTIME(create_time,'%Y-%m-%d %H:%i:%s') as extractTime, ip").
			Where("uid =? ", uid).
			Where("create_time >=? and create_time <=? ", start, end).
			Find(&monthData)
		data = append(data, monthData...)
	}
	return
}