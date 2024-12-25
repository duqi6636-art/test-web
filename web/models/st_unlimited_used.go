package models

import (
	"time"
)

func GetUnlimitedUrlUsed(uid ,accountId int, start, end int) (list []StUrlToday) {
	tableNum := time.Now().Format("200601")
	tableStr := "st_url_unlimited"+tableNum
	dbs := db.Table(tableStr).Select("today,flows,url").Where("uid = ?", uid).Where("today >= ? and today <= ?", start, end)
	if accountId > 0 {
		dbs = dbs.Where("account_id = ?", accountId)
	}
	dbs.Order("today asc", true).Find(&list)
	return
}

func GetUnlimitedUrlUsedDown(uid ,accountId int, start, end int) (list []StUrlToday) {
	tableNum := time.Now().Format("200601")
	tableStr := "st_url_unlimited"+tableNum
	dbs := db.Table(tableStr).Select("today,flows,url").Where("uid = ?", uid).Where("today >= ? and today <= ?", start, end)
	if accountId > 0 {
		dbs = dbs.Where("account_id = ?", accountId)
	}
	dbs.Order("today asc", true).Find(&list)
	return
}

