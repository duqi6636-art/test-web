package models

import "time"

type CmNotice struct {
	Id          int    `json:"id"`
	Type        int    `json:"type"`
	Cate        string `json:"cate"`
	Title       string `json:"title"`
	Brief       string `json:"brief"`   //简介
	Content     string `json:"content"` //内容
	TitleZh     string `json:"title_zh"`
	BriefZh     string `json:"brief_zh"`     //简介
	ContentZh   string `json:"content_zh"`   //内容
	Users       string `json:"users"`        //用户信息
	CreateTime  int    `json:"create_time"`  //公告时间
	ReleaseTime int    `json:"release_time"` //发布时间
	Code        string `json:"code"`
}

type ResNotice struct {
	Id          int    `json:"id"`
	Type        int    `json:"type"`
	Title       string `json:"title"`
	Brief       string `json:"brief"`        //简介
	Content     string `json:"content"`      //内容
	CreateTime  string `json:"create_time"`  //公告时间
	ReleaseTime string `json:"release_time"` //发布时间
	IsRead      int    `json:"is_read"`      //是否已读
}

// 获取列表
func GetNoticeList() (data []CmNotice, err error) {
	dbs := db.Table("cm_notice")
	err = dbs.Where("release_time <= ?", int(time.Now().Unix())).
		Where("platform = ? or platform = ?", "all", "web").
		Where("status = ?", 1).
		Order("release_time desc").
		Find(&data).Error
	return
}

// 获取列表
func GetNoticeList_bak(lang string) (data []CmNotice, err error) {
	dbs := db.Table("cm_notice")
	if lang != "" { //语言
		dbs = dbs.Where("lang =?", lang)
	}
	err = dbs.Where("release_time <= ?", int(time.Now().Unix())).
		Where("platform = ? or platform = ?", "all", "web").
		Where("status = ?", 1).
		Order("release_time desc").
		Find(&data).Error
	return
}

// 获取列表
func GetNoticeListLimit(lang string, limit int) (data []CmNotice, err error) {
	dbs := db.Table("cm_notice")
	if lang != "" { //语言
		dbs = dbs.Where("lang =?", lang)
	}
	err = dbs.Where("release_time <= ?", int(time.Now().Unix())).Where("platform = ? or platform = ?", "all", "web").Where("status = ?", 1).Order("release_time desc").Limit(limit).Find(&data).Error
	return
}

// 获取详情
func GetNoticeInfo(id int) (data CmNotice, err error) {
	dbs := db.Table("cm_notice")
	err = dbs.Where("id = ?", id).First(&data).Error
	return
}

// 获取列表
func GetNoticeListByIds(ids []string) (data []CmNotice, err error) {
	err = db.Table("cm_notice").Where("id in (?)", ids).Find(&data).Error
	return
}

func GetLastNotice(lang string) (data CmNotice, err error) {
	dbs := db.Table("cm_notice")
	err = dbs.Where("lang = ?", lang).
		Where("type = 1").
		Order("release_time desc").
		First(&data).Error
	return
}

// 获取列表
func GetOtherNoticeList(lang string, banCodes []string) (data []CmNotice, err error) {
	dbs := db.Table("cm_notice")
	dbs = dbs.Where("lang =?", lang).
		Where("code not in (?)", banCodes)
	err = dbs.Where("release_time <= ?", int(time.Now().Unix())).
		Where("platform = ? or platform = ?", "all", "web").
		Where("status = ?", 1).
		Order("release_time desc").Find(&data).Error
	return
}

type CmNoticeUser struct {
	Id         int    `json:"id"`
	Nid        int    `json:"nid"`
	Uid        int    `json:"uid"`
	Ip         string `json:"ip"`
	Lang       string `json:"lang"`
	Version    string `json:"version"`
	IsDel      int    `json:"is_del"`
	CreateTime int    `json:"create_time"`
	Code       string `json:"code"`
}

func AddUserNotice(data CmNoticeUser) (err error) {
	err = db.Table("cm_notice_user").Create(&data).Error
	return
}

func GetUserNoticeList(uid int) (data []CmNoticeUser, err error) {
	err = db.Table("cm_notice_user").Where("uid =?", uid).Order("id desc").Find(&data).Error
	return
}

func GetUserNoticeInfo(uid, nid int) (data CmNoticeUser, err error) {
	err = db.Table("cm_notice_user").Where("uid =?", uid).Where("nid =?", nid).Where("is_del =?", 0).First(&data).Error
	return
}
