package models


type CmVersionModel struct {
	Id                int    `json:"id"`
	Name              string `json:"name"`
	Version           int    `json:"version"`
	ShowVersion       string `json:"show_version"`
	Url               string `json:"url"`
	DownUrl           string `json:"down_url"`
	Desc              string `json:"desc"`
}

var socksVersionTable = "cm_version"

func GetVersionInfo(platform string) (data CmVersionModel) {
	db.Table(socksVersionTable).Where("platform =?",platform).Where("status =?",1).Order("version desc").Order(" id desc").First(&data)
	return
}

