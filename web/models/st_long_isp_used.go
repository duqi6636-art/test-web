package models

import (
	"api-360proxy/web/pkg/util"
	"math"
)

// 动态URL列表
func GetDynamicUrlUsed(uid, accountId int, start, end int) (list []StUrlToday) {
	tableNum := int(math.Ceil(float64(uid) / 50000))
	tableStr := "st_top_url_dynamic_" + util.ItoS(tableNum)
	dbs := db.Table(tableStr).Select("today,flows,url").Where("uid = ?", uid).Where("today >= ? and today <= ?", start, end)
	if accountId > 0 {
		dbs = dbs.Where("account_id = ?", accountId)
	}
	dbs.Order("today asc", true).Find(&list)
	return
}
// 动态URL列表下载
func GetDynamicUrlUsedDown(uid ,accountId int, start, end int) (list []StUrlToday) {
	tableNum := int(math.Ceil(float64(uid) / 50000))
	tableStr := "st_child_url_dynamic_"+util.ItoS(tableNum)
	dbs := db.Table(tableStr).Select("today,flows,url").Where("uid = ?", uid).Where("today >= ? and today <= ?", start, end)
	if accountId > 0 {
		dbs = dbs.Where("account_id = ?", accountId)
	}
	dbs.Order("today asc", true).Find(&list)
	return
}
// 查询长效Isp用户使用记录
func GetLongIspUrlList(uid int, start int, end int, url string) (list []StUrlLists) {
	tableNum := int(math.Ceil(float64(uid) / 50000))
	tableStr := "st_top_url_dynamic_"+util.ItoS(tableNum)
	dbs := db.Table(tableStr).Select("url").Where("uid = ?", uid).Where("today >= ? and today <= ?", start, end)
	if url != "" {
		dbs = dbs.Where("url like ?", "%"+url+"%")
	}
	dbs.Group("url").Find(&list)
	return
}