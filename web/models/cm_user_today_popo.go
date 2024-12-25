package models

type CmUserTodayPopo struct {
	Id    int `json:"id"`    // ID
	Uid   int `json:"uid"`   // 用户id
	Today int `json:"today"` // 当日凌晨时间戳
}

// 创建记录
func CreateUserTodayPopoInfo(info CmUserTodayPopo) {
	db.Table("cm_user_today_notice").Create(&info)
	return
}

// 查询记录
func GetUserTodayPopoInfo(uid, today int) (info CmUserTodayPopo) {
	db.Table("cm_user_today_notice").Where("uid = ? and today = ?", uid, today).First(&info)
	return
}
