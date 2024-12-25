package models

type ConfCountryCode struct {
	Id      int    `json:"id"`
	Code    string `json:"code"`    // 国家标识
	Country string `json:"country"` // 国家昵称
}

// 模糊查询国家代码信息
func GetCountryCount(code string) (num int) {
	dbs := db.Table("conf_country_code")
	if code != "" {
		dbs = dbs.Where("keyword like ?", "%"+code+"%")
	}
	dbs.Order("country asc",true).Count(&num)
	return
}

// 获取分页列表
func GetCountry(code string)(data[] ConfCountryCode)  {
	dbs := db.Table("conf_country_code")
	if code != "" {
		dbs = dbs.Where("keyword like ?", "%"+code+"%")
	}
	dbs.Order("country asc",true).Find(&data)
	return
}

// 获取分页列表
func GetCountryPage(offset,limit int,code string)(data[] ConfCountryCode)  {
	dbs := db.Table("conf_country_code")
	if code != "" {
		dbs = dbs.Where("keyword like ?", "%"+code+"%")
	}
	dbs.Offset(offset).Limit(limit).Order("country asc",true).Find(&data)
	return
}