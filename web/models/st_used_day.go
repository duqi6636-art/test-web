package models

import (
	"api-360proxy/web/db/clickhousedb"
	"api-360proxy/web/pkg/util"
)

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
