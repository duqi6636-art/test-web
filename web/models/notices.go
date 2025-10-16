package models

import "api-360proxy/web/pkg/util"

type ResNoticeV1 struct {
	Id         int    `json:"id"`
	Title      string `json:"title"`
	Brief      string `json:"brief"`       //简介
	Content    string `json:"content"`     //内容
	CreateTime string `json:"create_time"` //公告时间
	IsRead     int    `json:"is_read"`     //是否已读
	Sort       int    `json:"sort"`        //排序值
	Timestamp  int    `json:"timestamp"`   //公告时间
}

// CmNoticeMsg 新版站内信
type CmNoticeMsg struct {
	Id         int    `json:"id"`
	Title      string `json:"title"`
	Brief      string `json:"brief"`       //简介
	Content    string `json:"content"`     //内容
	TitleZh    string `json:"title_zh"`    //繁体中文
	BriefZh    string `json:"brief_zh"`    //简介 繁体中文
	ContentZh  string `json:"content_zh"`  //内容 繁体中文
	ShowType   int    `json:"show_type"`   //展示类型
	CreateTime int    `json:"create_time"` //公告时间
	Cate       string `json:"cate"`        //类型 login_warning work_order_pre auto_renew_insufficient等
	Uid        int    `json:"uid"`         //用户信息
	ReadTime   int    `json:"read_time"`   //已读时间  0 未读
	PushTime   int    `json:"push_time"`   //类型是其他的时候 给哪些用户展示的
	Sort       int    `json:"sort"`        //排序值
	Admin      string `json:"admin"`       //操作人
}

var cMNoticeMsgTable = "cm_notice_msg"

// GetNoticeMsgList 获取列表
func GetNoticeMsgList(uid, limit int) (data []CmNoticeMsg, err error) {
	nowTime := util.GetNowInt()
	var noticeMsgField = "id,title,brief,content,create_time,show_type,title_zh,brief_zh,content_zh,cate,uid,read_time,push_time,sort"
	dbs := db.Table(cMNoticeMsgTable).Select(noticeMsgField).
		Where("uid =?", uid).
		Where("push_time < ?", nowTime).
		Where("status = ?", 1).
		Order("sort desc,id desc")
	if limit > 0 { //数量
		dbs = dbs.Limit(limit)
	}
	err = dbs.Find(&data).Error
	return
}

// EditNoticeMsg 更新信息
func EditNoticeMsg(id int, data map[string]interface{}) bool {
	err := db.Table(cMNoticeMsgTable).Where("id = ?", id).Updates(data).Error
	if err != nil {
		return false
	}
	return true
}

// GetNoticeMsgInfo 获取详情
func GetNoticeMsgInfo(id, uid int) (data CmNotice, err error) {
	dbs := db.Table(cMNoticeMsgTable).Select("id,title,brief,content,create_time,show_type")
	err = dbs.Where("id = ?", id).Where("uid = ?", uid).First(&data).Error
	return
}

// BatchAddNoticeMsgLog 批量写入站内信记录
func BatchAddNoticeMsgLog(list []CmNoticeMsg) (err error) {
	err = MasterWriteDb.Table(cMNoticeMsgTable).CreateInBatches(&list, len(list)).Error
	return
}
