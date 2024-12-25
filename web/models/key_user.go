package models

// 主要用户配置
type KeyUserWhiteModel struct {
	Id  int `json:"id"`
	Uid int `json:"uid"` //用户ID
}

func GetKeyUserWhiteBy(uid int) (data KeyUserWhiteModel) {
	db.Table("cm_key_user_white").Where("uid = ?", uid).Where("status = ?", 1).First(&data)
	return
}
