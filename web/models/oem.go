package models

// 渠道表
type OemModel struct {
	Id      int    `json:"id"`
	Oem     string `json:"oem"`
	Name    string `json:"name"`
	Client  string `json:"client"`
	DownUrl string `json:"down_url"`
	Version string `json:"version"`
	Html    string `json:"html"`
}

var OemTabName = "cm_oem"

// 获取渠道信息
func GetOemInfo(oem string, client int) (infos OemModel, err error) {
	err = db.Table(OemTabName).Where("oem = ?", oem).Where("client = ?", client).Where("status =?", 1).Order("id desc", true).First(&infos).Error
	return
}

// 获取渠道信息
func GetOemList() (infos []OemModel, err error) {
	err = db.Table(OemTabName).Where("status =?", 1).Order("id desc", true).Find(&infos).Error
	return
}
