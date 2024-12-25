package models

import (
	"api-360proxy/web/pkg/util"
	"math"
)

type StUsedToday struct {
	Id    int    `json:"id"`    // ID
	Uid   int    `json:"uid"`   // uid
	Today int    `json:"today"` // 当日凌晨时间戳
	Cate  string `json:"cate"`  // 类型：isp：住宅代理，all：全球静态，us：美国静态，custom：普通流量
	Num   int64  `json:"num"`   // 数量
}

// 查询用户使用记录
func GetUserUsed(uid int, cate string, start, end int) (list []StUsedToday) {
	dbs := db.Table("st_used_today").Where("uid = ?", uid).Where("today >= ? and today <= ?", start, end)
	if cate == "isp" {
		dbs = dbs.Where("cate = ?", cate)
	} else if cate == "static" {
		dbs = dbs.Where("cate = ? or cate = ?", "all", "us")
	} else if cate == "flow" {
		dbs = dbs.Where("cate = ?", "custom")
	}
	dbs.Order("today asc", true).Find(&list)
	return
}

//// 流量使用
//type StUrlToday struct {
//	Today int   `json:"today"` // 当日凌晨时间戳
//	Flows int64 `json:"flows"` // 数量
//}

// 流量使用
type StUrlToday struct {
	Today int    `json:"today"` // 当日凌晨时间戳
	Flows int64  `json:"flows"` // 数量
	Url   string `json:"url"`   // 地址
}

// 查询用户使用记录
//func GetUrlUsed(uid int, url string, start, end int) (list []StUrlToday) {
//	dbs := db.Table("st_url_today").Select("today,sum(flows) as flows").Where("uid = ?", uid).Where("today >= ? and today <= ?", start, end)
//	if url != "" {
//		dbs = dbs.Where("url = ?", url)
//	}
//	dbs.Order("today asc", true).Group("today").Find(&list)
//	return
//}
// 查询用户使用记录

func GetUrlUsed(uid, accountId int, start, end int) (list []StUrlToday) {
	tableNum := int(math.Ceil(float64(uid) / 30000))
	tableStr := "st_top_url" + util.ItoS(tableNum)
	dbs := db.Table(tableStr).Select("today,flows,url").Where("uid = ?", uid).Where("today >= ? and today <= ?", start, end)
	if accountId > 0 {
		dbs = dbs.Where("account_id = ?", accountId)
	}
	dbs.Order("today asc", true).Find(&list)
	return
}

// 查询用户使用记录
func GetUrlWhiteUsed(uid, accountId int, start, end int) (list []StUrlToday) {
	dbs := db.Table("st_white_url_today").Select("today,flows,url").Where("uid = ?", uid).Where("today >= ? and today <= ?", start, end)
	if accountId > 0 {
		dbs = dbs.Where("account_id = ?", accountId)
	}
	dbs.Order("today asc", true).Find(&list)
	return
}

// 流量使用
type StUrlLists struct {
	Url string `json:"url"`
}

// 查询用户使用记录
func GetUrlList(uid int, start int, url string) (list []StUrlLists) {
	tableNum := int(math.Ceil(float64(uid) / 30000))
	tableStr := "st_top_url" + util.ItoS(tableNum)
	dbs := db.Table(tableStr).Select("url").Where("uid = ?", uid).Where("today >= ?", start)
	if url != "" {
		dbs = dbs.Where("url like ?", "%"+url+"%")
	}
	dbs.Group("url").Find(&list)
	return
}
