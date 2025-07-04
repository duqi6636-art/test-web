package models

type UserBindPhone struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	Phone      string `json:"phone"`
	Status     int    `json:"status"`
	CountryId  int    `json:"country_id"`
	UpdateTime int64  `json:"update_time"`
	CreateTime int64  `json:"create_time"`
}

const bindPhoneTableName = "cm_user_bind_phone"

func (u *UserBindPhone) Create() {
	db.Table(bindPhoneTableName).Create(u)
}

func (u *UserBindPhone) Save() {
	db.Table(bindPhoneTableName).Where("uid = ?", u.Uid).Save(u)
}

func (u *UserBindPhone) Update(values map[string]interface{}) {
	db.Table(bindPhoneTableName).Where("uid = ?", u.Uid).Updates(values)
}

func (u *UserBindPhone) Get() {
	db.Table(bindPhoneTableName).Where("uid = ?", u.Uid).First(u)
}
