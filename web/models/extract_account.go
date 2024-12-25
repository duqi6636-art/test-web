package models

// 代理国家
type ExtractCountry struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
	Img     string `json:"img"`
	Keyword string `json:"keyword"`
	Sort    int    `json:"sort"`
}

// 获取所有国家数据
func GetAllCountry(limit int, country string) (list []ExtractCountry) {
	dbs := db.Table("cm_ex_country").Where("status = ?", 1)
	if country != "" {
		dbs = dbs.Where("keyword like ?", "%"+country+"%")
	}
	if limit > 0 {
		dbs = dbs.Limit(limit)
	}
	dbs.Order("sort desc").Find(&list)
	return
}

// 获取所有国家数据
func GetAllCountryV2(country string) (list []ExtractCountry) {
	dbs := db.Table("cm_ex_country").Where("status = ?", 1)
	if country != "" {
		dbs = dbs.Where("keyword like ?", "%"+country+"%")
	}
	dbs.Order("sort desc").Find(&list)
	return
}

// 获取所有国家数据
func GetByCountry(country string) (list ExtractCountry) {
	db.Table("cm_ex_country").Where("country = ?", country).Where("status = ?", 1).First(&list)
	return
}

// 代理洲省
type ExtractProvince struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Status int    `json:"status"`
}

// 获取所有洲省数据
func GetStateByCid(cid int) (list []ExtractProvince) {
	db.Table("cm_extract_state").Where("cid = ? and status = ?", cid, 1).Order("sort desc").Find(&list)
	return
}

// 代理城市
type ExtractCity struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	Country string `json:"country"`
	Status  int    `json:"status"`
}

// 获取所有城市数据
func GetAllCity() (list []ExtractCity) {

	dbs := db.Table("cm_ex_city").Where("status = ?", 1)
	dbs.Order("sort desc").Find(&list)
	return
}

// 获取所有城市数据
func GetCityByCountry(country, city string) (list []ExtractCity) {
	dbs := db.Table("cm_ex_city").Where("country = ?", country)
	if city != "" {
		dbs = dbs.Where("code like ?", "%"+city+"%")
	}
	dbs.Where("status = ?", 1).Order("sort desc").Find(&list)
	//db.Table("md_ex_city").Where("country = ? and status = ?", country, 1).Order("sort desc").Find(&list)
	return
}

// 获取所有国家数据
func GetAllFlowDayCountry(country string) (list []ExtractCountry) {
	dbs := db.Table("cm_flow_day_country").Where("status = ?", 1)
	if country != "" {
		dbs = dbs.Where("keyword like ?", "%"+country+"%")
	}
	dbs.Order("sort desc").Find(&list)
	return
}

type MdExCountryPort struct {
	ID         uint   `gorm:"column:id;primary_key"`
	Name       string `gorm:"column:name"`
	Country    string `gorm:"column:country"`
	Keyword    string `gorm:"column:keyword"`
	Num        int    `gorm:"column:num"`
	Img        string `gorm:"column:img"`
	Status     int    `gorm:"column:status"`
	Sort       int    `gorm:"column:sort"`
	Admin      string `gorm:"column:admin"`
	UpdateTime int    `gorm:"column:update_time"`
	Port1      string `gorm:"column:port1"`
	Port2      string `gorm:"column:port2"`
	Port3      string `gorm:"column:port3"`
}
type ResExtractCountryCity struct {
	Name     string              `json:"name"`
	Country  string              `json:"country"`
	Img      string              `json:"img"`
	Keyword  string              `json:"keyword"`
	Sort     int                 `json:"sort"`
	Collect  int                 `json:"collect"`
	CityList []ExtractCity       `json:"city_list,omitempty"`
	Value    string              `json:"value"` //前端展示需要用到
	Ports    []map[string]string `json:"ports"`
}

// 获取端口
func GetPortsByCountry() (list []MdExCountryPort) {

	dbs := db.Table("cm_country_port").Where("status = ?", 1)
	dbs.Order("sort desc").Find(&list)

	return
}

// 根据国家获取国家端口
func GetCountryPortByCountry(country string) (countryPort MdExCountryPort) {

	dbs := db.Table("cm_country_port").
		Where("status = ?", 1).
		Where("country = ?", country)
	dbs.Order("sort desc").First(&countryPort)
	return
}

// 获取长效Isp国家域名
func GetLongIspPortsByCountry() (list []MdExCountryPort) {

	dbs := db.Table("cm_country_port_longisp").Where("status = ?", 1)
	dbs.Order("sort desc").Find(&list)
	return
}
