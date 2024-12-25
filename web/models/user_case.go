package models

type UserCaseModel struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Case       string `json:"case"`
	Other      string `json:"other"`
	Ip         string `json:"ip"`
	CreateTime int    `json:"create_time"`
}

func AddUserCase(data UserCaseModel) (err error) {
	err = db.Table("cm_log_user_case").Create(&data).Error
	return
}
