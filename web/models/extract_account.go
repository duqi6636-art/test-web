package models

// 代理国家
type ExtractCountry struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
	Img     string `json:"img"`
	Keyword string `json:"keyword"`
	Sort    int    `json:"sort"`
	Num     int    `json:"num"`
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

// 获取所有洲省数据
func GetStateByCountry(country, state string) (list []ExtractProvince) {
	dbt := db.Table("cm_ex_state")
	if country != "" {
		dbt = dbt.Where("country = ?", country)
	}
	if state != "" {
		dbt = dbt.Where("state = ?", state)
	}
	dbt.Where("status = ?", 1).Order("sort desc,id desc").Find(&list)
	return
}

// 代理城市
type ExtractCity struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	Country string `json:"country"`
	Status  int    `json:"status"`
	State   string `json:"state"`
	Num     int    `json:"num"`
}

// 返回city 信息
type ResExtractCity struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	//Country string `json:"country"`
	//State   string `json:"state"`
}

// 获取所有城市数据
func GetAllCity() (list []ExtractCity) {

	dbs := db.Table("cm_ex_city").Where("status = ?", 1)
	dbs.Order("sort desc").Find(&list)
	return
}

// 获取所有城市数据
func GetCityByCountry(country, state, city string) (list []ExtractCity) {
	dbs := db.Table("cm_ex_city").Where("country = ?", country)
	if state != "" {
		dbs = dbs.Where("state = ?", state)
	}
	if city != "" {
		dbs = dbs.Where("code like ?", "%"+city+"%")
	}
	dbs.Where("status = ?", 1).Order("sort desc").Find(&list)
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

// 代理 ISP - 运营商
type ExtractIsp struct {
	Id      int    `json:"id"`
	Isp     string `json:"isp"`
	IspName string `json:"isp_name"`
	Country string `json:"country"`
}

// 获取所有运营商数据 根据国家
func GetIspCountry(country string) (list []ExtractIsp) {
	dbt := db.Table("cm_ex_isp")
	if country != "" {
		dbt = dbt.Where("country = ?", country)
	}
	dbt.Where("status = ?", 1).Find(&list)
	return
}

// ------------------------ISP 地区  查询 cm_area 表数据
type ExtractIspCountry struct {
	Country string `json:"country"`
	State   string `json:"state"`
	City    string `json:"city"`
}

// 获取国家地区数据 根据国家
func GetIspCountryList() (list []ExtractIspCountry) {
	tables := "ncm_area"
	dbArea.Table(tables).
		Select("country").
		Group("country").
		Find(&list)

	return
}

// 获取州/省数据
func GetIspStateBy(country, state string) (list []ExtractIspCountry) {
	tables := "ncm_area"
	dbs := dbArea.Table(tables).
		Select("state").
		Where("country = ?", country)

	if state != "" {
		dbs = dbs.Where("state = ?", state)
	}
	dbs.Group("state").
		Find(&list)
	return
}

// 获取城市数据
func GetIspCityBy(country, state, city string) (list []ExtractIspCountry) {
	tables := "ncm_area"
	dbs := dbArea.Table(tables).
		Select("state,city").
		Where("country = ?", country)
	if state != "" {
		dbs = dbs.Where("state = ?", state)
	}
	if city != "" {
		dbs = dbs.Where("city like ?", "%"+city+"%")
	}
	dbs.Group("state,city").
		Find(&list)
	return
}

// ------------------------ISP 地区  查询 ncm_isp 表数据
type ExtractIspAsn struct {
	Id      int    `json:"id"`
	Country string `json:"country"`
	Isp     string `json:"isp"`
	Asn     string `json:"asn"`
}

// 获取isp数据
func GetIspAsnBy(country, asn, isp string) (list []ExtractIspAsn) {
	tables := "ncm_isp"
	dbs := dbArea.Table(tables).
		Select("id,country,isp,asn").
		Where("country = ?", country)
	if asn != "" {
		dbs = dbs.Where("asn = ?", asn)
	}
	if isp != "" {
		dbs = dbs.Where("isp like ?", "%"+isp+"%")
	}
	dbs.Find(&list)
	return
}
