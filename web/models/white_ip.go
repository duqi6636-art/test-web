package models

type IpWhiteModel  struct {
	Id   	int    `json:"id"`
	Ip 		string `json:"ip"`
}

func GetWhiteIpInfoMap(where map[string]interface{}) (data IpWhiteModel) {
	db.Table("cm_ip_white").Where(where).Order(" id desc").First(&data)
	return
}

type PayWhiteModel struct {
	Id   		int    `json:"id"`
	Account 	string `json:"account"`
}
func GetPayWhiteInfoMap(where map[string]interface{}) (data PayWhiteModel) {
	db.Table("pay_white").Where(where).Order(" id desc").First(&data)
	return
}