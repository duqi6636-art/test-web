package models

type ConfPayPlat struct {
	Id       int    `json:"id"`
	Code     string `json:"code"`
	ShowName string `json:"show_name"`
}

func GetPayPlatConf(id interface{}) (name string) {
	data := ConfPayPlat{}
	err := db.Table("conf_pay_plat").Where("id =?", id).First(&data).Error
	if err != nil {
		name = ""
	}
	name = data.ShowName
	return
}

func GetPayPlatConfList() (err error, data []ConfPayPlat) {
	err = db.Table("conf_pay_plat").Find(&data).Error
	return
}
