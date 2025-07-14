package models

// 接口文档
var user_doc_api_token_table = "cm_user_doc_key"

type UserDocApiTokenModel struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Token      string `json:"token"`
	Key        string `json:"key"`
	Status     int    `json:"status"`
	Ip         string `json:"ip"` // 最后操作IP
	CreateTime int    `json:"create_time"`
	UpdateTime int    `json:"update_time"` // 最后更新时间
}

type ResultUserDocApiToken struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Key      string `json:"key"`
}

// 添加用户serpApiToken
func AddUserDocApiToken(data UserDocApiTokenModel) (err error) {
	err = db.Table(user_doc_api_token_table).Create(&data).Error
	return
}

// 查询用户serpApiToken By Uid
func GetUserDocApiToken(uid int) (err error, info UserDocApiTokenModel) {
	dbs := db.Table(user_doc_api_token_table).Where("uid = ?", uid)
	dbs = dbs.Where("status = ?", 1)
	err = dbs.First(&info).Error
	return
}

// 更新用户serpApiToken
func EditUserDocApiTokenById(id int, user interface{}) (err error) {
	err = db.Table(user_doc_api_token_table).Where("id = ?", id).Updates(user).Error
	return err
}
