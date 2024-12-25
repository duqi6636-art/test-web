package models

type IpBlack struct {
	Id          int    `json:"id"`
	Uid  		int	   `json:"uid"`
	Ip          string `json:"ip"`
	Email       string `json:"email"`
	Value 		int	   `json:"value"`
	DisableTime int	   `json:"disable_time"`
	CreateTime  int	   `json:"create_time"`
	Type  		int	   `json:"type"`
}

var ipBlackTable = "cm_ip_black"

func AddIpBlack(er IpBlack) bool {
	return db.Table(ipBlackTable).Create(&er).Error == nil
}

func GetIpBlackInfo(where interface{}) (err error, data IpBlack) {
	err = db.Table(ipBlackTable).Where(where).Order(" id desc").First(&data).Error
	return
}

func GetIpBlackList(where interface{}) (data[] IpBlack) {
	db.Table(ipBlackTable).Where(where).Order(" id asc").Find(&data)
	return
}
