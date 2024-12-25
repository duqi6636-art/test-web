package models

type CmServerLoad struct {
	Id   int    `json:"id"`   // ID
	Name string `json:"name"` // 名称
	Ip   string `json:"ip"`   // ip
	Area string `json:"area"` // area
}

// 查询负载列表
func GetServerList() (list []CmServerLoad) {
	db.Table("cm_server_load").Where("status = ?", 1).Find(&list)
	return list
}
