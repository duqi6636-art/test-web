package models

//  竞价链接
type MdAdLink struct {
	Id               int     `json:"id"`
	Title            string  `json:"title"`
	Name             string  `json:"name"`
	SecretLink       string  `json:"secret_link"`
	Status           int     `json:"status"`
	Category         int     `json:"category"`
	PromotionKeyword string  `json:"promotion_keyword"`
	SpreadCode       string  `json:"spread_code"`
	Spread           string  `json:"spread"`
	JingjiaDomain    string  `json:"jingjia_domain"`
}

var adLinkTable = "cm_ad_link"

func GetAdLinkInfoByMap(where map[string]interface{}) (data MdAdLink) {
	db.Table(adLinkTable).Where(where).First(&data)
	return
}

func EditLinkInfoByMap(where map[string]interface{}, data map[string]interface{}) (err error) {
	err = db.Table(adLinkTable).Where(where).Updates(data).Error
	return err
}

