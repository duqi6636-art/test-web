package models

import (
	"api-360proxy/web/db/clickhousedb"
	"api-360proxy/web/pkg/util"
)

// 查询流量使用记录
// uid 用户ID
// accountId 在账户ID
// country 使用地区
// address 访问地址
// flowType 流量使用类型 1帐密  2 白名单 3 api
func GetFlowUsedStat(uid, accountId int, start, end int, country, address, flowType string) (list []StUrlToday) {
	ym := util.GetTimeStr(start, "Ym")
	tableStr := "st_user_flow_used_" + ym
	dbs := clickhousedb.ClickhouseDb.
		Table(tableStr).
		Select("today,flows").
		Where("uid = ?", uid).
		Where("today >= ? and today <= ?", start, end)
	if accountId > 0 {
		dbs = dbs.Where("account_id = ?", accountId)
	}
	if country != "" {
		dbs = dbs.Where("country = ?", country)
	}
	if address != "" {
		dbs = dbs.Where("address = ?", address)
	}
	if flowType != "" {
		dbs = dbs.Where("flow_type = ?", flowType)
	}
	dbs.Order("today asc").Find(&list)
	return
}

// 查询不限量使用记录
// uid 用户ID
// accountId 在账户ID
func GetFlowDayUsedStat(uid, accountId int, start, end int) (list []StUrlToday) {
	ym := util.GetTimeStr(start, "Ym")
	tableStr := "st_user_flow_day_used_" + ym
	dbs := clickhousedb.ClickhouseDb.
		Table(tableStr).
		Select("today,flows").
		Where("uid = ?", uid).
		Where("today >= ? and today <= ?", start, end)
	if accountId > 0 {
		dbs = dbs.Where("account_id = ?", accountId)
	}
	dbs.Order("today asc").Find(&list)
	return
}

// 查询轮转ISP流量记录
// uid 用户ID
// accountId 在账户ID
// country 使用地区
// address 访问地址
func GetIspFlowUsedStat(uid, accountId int, start, end int, country, address string) (list []StUrlToday) {
	ym := util.GetTimeStr(start, "Ym")
	tableStr := "st_user_isp_flow_used_" + ym
	dbs := clickhousedb.ClickhouseDb.
		Table(tableStr).
		Select("today,flows").
		Where("uid = ?", uid).
		Where("today >= ? and today <= ?", start, end)
	if accountId > 0 {
		dbs = dbs.Where("account_id = ?", accountId)
	}
	if country != "" {
		dbs = dbs.Where("country = ?", country)
	}
	if address != "" {
		dbs = dbs.Where("address = ?", address)
	}
	dbs.Order("today asc").Find(&list)
	return
}

// 获取 流量使用URL
func GetUrlListStats(uid int, start int, url string) (list []StUrlLists) {
	ym := util.GetTimeStr(start, "Ym")
	tableStr := "st_user_flow_used_" + ym
	dbs := clickhousedb.ClickhouseDb.
		Table(tableStr).
		Select("today,flows").
		Where("uid = ?", uid).
		Where("today >= ?", start)
	if url != "" {
		dbs = dbs.Where("url like ?", "%"+url+"%")
	}
	dbs.Group("url").Find(&list)
	return
}

// 获取 轮转ISP流量URL
func GetIspUrlListStats(uid int, start, end int, url string) (list []StUrlLists) {
	ym := util.GetTimeStr(start, "Ym")
	tableStr := "st_user_isp_flow_used_" + ym
	dbs := clickhousedb.ClickhouseDb.
		Table(tableStr).
		Select("today,flows").
		Where("uid = ?", uid).
		Where("today >= ? and today <= ?", start, end)
	if url != "" {
		dbs = dbs.Where("url like ?", "%"+url+"%")
	}
	dbs.Group("url").Find(&list)
	return
}
