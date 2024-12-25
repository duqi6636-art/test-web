package models

import "strings"

// 套餐
type ResSocksIpPackage struct {
	Id          int     `json:"id"`
	Pid         int     `json:"pid"`
	Code        string  `json:"code"`
	PakType     string  `json:"pak_type"`
	Value       int     `json:"value"`
	Name        string  `json:"name"`
	SubName     string  `json:"sub_name"`
	Number      int64   `json:"number"`
	Gift        int64   `json:"gift"`     //专属赠送
	Give        int64   `json:"give"`     //已付费用户赠送
	Discount    float64 `json:"discount"` //用户折扣满减
	Price       float64 `json:"price"`
	Corner      string  `json:"corner"`       //角标
	ActTitle    string  `json:"act_title"`    //活动名称
	ActDesc     string  `json:"act_desc"`     //活动描述
	ActLabel    string  `json:"act_label"`    //活动标签
	CouponLabel string  `json:"coupon_label"` //优惠券标签
	Default     string  `json:"default"`
	IsHot       int     `json:"is_hot"` //是否热门
	ShowPrice   float64 `json:"show_price"`
	//OriginPrice float64 `json:"origin_price"`
	Unit float64 `json:"unit"` //未付费单价
	//OriginUnit  float64 `json:"origin_unit"` //原单价
	AllUnit  float64 `json:"all_unit"` //美国单价
	Currency string  `json:"currency"`
	Day      int     `json:"day"`       //过期天数
	TotalNum int     `json:"total_num"` //总套餐数
	TotalKey int     `json:"total_key"` //套餐Key值 方便模板输出加判断
	//StaticUrl string  `json:"static_url"` //静态资源地址
	Lang string `json:"lang"` //语言

}

// 余额套餐
type ResBalancePackage struct {
	Id      int     `json:"id"`
	Name    string  `json:"name"`
	Cate    string  `json:"cate"`
	Price   float64 `json:"price"`
	Default string  `json:"default"`
}

var packageTable = "cm_package"

func GetSocksPackageInfoById(id int) (data ResSocksIpPackage) {
	dbRead.Table(packageTable).Where("id = ?", id).Where("status=?", 1).First(&data)
	return
}

func GetPackageListByPIds(idArr interface{}) (data []ResSocksIpPackage) {
	dbRead.Table(packageTable).Where("pid in (?)", idArr).Where("status=?", 1).Find(&data)
	return
}

func GetPackageList() (data []ResSocksIpPackage) {
	dbRead.Table(packageTable).Where("status=?", 1).Order("sort desc").Find(&data)
	return
}
func GetStaticPackageList() (data []ResSocksIpPackage) {
	where := map[string]interface{}{
		"pak_type": "static",
		"terminal": "web",
	}
	dbRead.Table(packageTable).Where(where).Where("pid=?", 0).Where("status=?", 1).Order("sort desc").Find(&data)
	return
}

func GetLowPrice(pakType string) (price float64) {
	var data ResSocksIpPackage
	dbRead.Table(packageTable).Where("pak_type = ?", pakType).Where("status=? and unit > 0", 1).Order("unit asc").First(&data)
	return data.Unit
}

func GetStaticLowPrice(day int) (price float64) {
	var data ResSocksIpPackage
	dbRead.Table(packageTable).Where("pak_type = ? and status=? and unit > 0", "static", 1).Where("day = ?", day).Order("unit asc").First(&data)
	return data.Unit
}

// 新版 套餐
type CmPackage struct {
	Id         int     `json:"id"`
	Code       string  `json:"code"`
	PakType    string  `json:"pak_type"`
	Pid        int     `json:"pid"`
	Name       string  `json:"name"`
	SubName    string  `json:"sub_name"` //副标题
	Price      float64 `json:"price"`
	ShowPrice  float64 `json:"show_price"`
	Unit       float64 `json:"unit"`     //IP单价
	UsUnit     float64 `json:"us_unit"`  //美国长效IP单价
	AllUnit    float64 `json:"all_unit"` //全球长效IP单价
	Default    string  `json:"default"`  //是否默认选中
	Value      int64   `json:"value"`
	Day        int     `json:"day"`
	IsHot      int     `json:"is_hot"`    //是否热门推荐
	Gift       int64   `json:"gift"`      //支付方式赠送
	Give       int64   `json:"give"`      //新老用户赠送
	Discount   float64 `json:"discount"`  //折扣满减
	Corner     string  `json:"corner"`    //角标
	ActTitle   string  `json:"act_title"` //活动名称
	ActDesc    string  `json:"act_desc"`  //推荐标题文案
	Currency   string  `json:"currency"`
	ActImg     string  `json:"act_img"`
	VirtualMin float64 `json:"virtual_min"`
	VirtualMax float64 `json:"virtual_max"`
	Status     int     `json:"status"`
	IsAll      int     `json:"is_all"`
	Sort       int     `json:"sort"`        // 排序
	UpdateTime int     `json:"update_time"` // 最后操作时间
	UseType    int     `json:"use_type"`    // 套餐适用 0正常套餐1-新用户套餐 2-老用户套餐
	BuyTimes   int     `json:"buy_times"`   // 可购买次数 0-不限 1-只可购买一次
	Alias      string  `json:"alias"`       //别名
}

// 流量套餐
type ResIpPackageFlow struct {
	Id          int      `json:"id"`
	Pid         int      `json:"pid"`
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	SubName     string   `json:"sub_name"`
	Value       int64    `json:"value"`
	Total       int64    `json:"total"`
	Gift        int64    `json:"gift"` //专属赠送
	Give        int64    `json:"give"` //用户赠送
	Number      int      `json:"number"`
	Price       float64  `json:"price"`
	ShowPrice   float64  `json:"show_price"`
	ActImg      string   `json:"act_img"`      //活动名称
	ActTitle    string   `json:"act_title"`    //活动名称
	ActDesc     string   `json:"act_desc"`     //活动名称
	ActLabel    string   `json:"act_label"`    //活动标签
	CouponLabel string   `json:"coupon_label"` //优惠券标签
	TextArr     []string `json:"text_arr"`     //推荐标题文案
	Unit        float64  `json:"unit"`         //未付费单价
	AllUnit     float64  `json:"all_unit"`     //活动单价
	Corner      string   `json:"corner"`       //角标
	Default     string   `json:"default"`
	IsHot       int      `json:"is_hot"` //是否热门推荐
	Currency    string   `json:"currency"`
	StaticUrl   string   `json:"static_url"` //静态资源地址
	TotalNum    int      `json:"total_num"`  //总套餐数
	GiftUnit    string   `json:"gift_unit"`  // 赠送单位
	Sort        int      `json:"sort"`       // 排序
	UseType     int      `json:"use_type"`   // 套餐适用 0正常套餐1-新用户套餐 2-老用户套餐
	Alias       string   `json:"alias"`      //别名
}

func GetPackageListFlow(pakType string, isOld int) (err error, data []CmPackage) {
	dbs := db.Table(packageTable).Where("pak_type =?", pakType)
	if isOld == 0 {
		dbs = dbs.Where("pid =?", 0)
	} else {
		dbs = dbs.Where("pid >?", 0)
	}
	err = dbs.Where("status=?", 1).Order("sort desc").Find(&data).Error
	return
}

func GetPackageListFlowAgent(day, isOld int) (err error, data []CmPackage) {
	dbs := db.Table(packageTable).Where("pak_type =?", "flow_agent").Where("day = ?", day)
	if isOld == 0 {
		dbs = dbs.Where("pid =?", 0)
	} else {
		dbs = dbs.Where("pid >?", 0)
	}
	err = dbs.Where("status=?", 1).Order("sort desc").Find(&data).Error
	return
}
func GetPackageInfoById(id int) (data CmPackage) {
	dbRead.Table(packageTable).Where("id = ?", id).Where("status=?", 1).First(&data)
	return
}

// / 根据套餐ID获取套餐
func GetPackageListWith(ids []string) (err error, data []CmPackage) {
	dbs := db.Table(packageTable).
		Where("id in (?)", ids)
	err = dbs.Where("status=?", 1).Order("sort desc").Find(&data).Error
	return
}

// 获取新用户5G套餐列表
func GetNewPackageFlowList() (err error, data []CmPackage) {
	dbs := db.Table(packageTable).Where("use_type =?", 1).Where("pak_type =?", "flow")
	err = dbs.Where("status=?", 1).Order("sort desc").Find(&data).Error
	return
}

// 套餐 角标 ，活动文案配置
type CmPackageInfo struct {
	Id        int    `json:"id"`
	PackageId int    `json:"package_id"`
	AreaId    int    `json:"area_id"`
	Lang      string `json:"lang"`      //多语言名称
	Name      string `json:"name"`      //多语言名称
	Corner    string `json:"corner"`    //角标
	ActTitle  string `json:"act_title"` //活动名称
	ActDesc   string `json:"act_desc"`  //推荐标题文案
	ActLabel  string `json:"act_label"` //活动标签
	ActImg    string `json:"act_img"`
	BackImg   string `json:"back_img"`
	Content   string `json:"content"`
}

func GetPackageInfoList() (err error, data []CmPackageInfo) {
	err = db.Table("cm_package_info").Where("status=?", 1).Find(&data).Error
	return
}

// 长效套餐
type ResIpPackageLong struct {
	Id        int     `json:"id"`
	Pid       int     `json:"pid"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Number    int     `json:"number"`
	Corner    string  `json:"corner"` //角标
	Unit      float64 `json:"unit"`
	Price     float64 `json:"price"`
	Default   string  `json:"default"`
	StaticUrl string  `json:"static_url"` //静态资源地址
	IsHot     int     `json:"is_hot"`     //
	//ContinentList []PackageContinentInfo `json:"continent_list"` //大洲数据
}

// 返回的静态大洲信息
type PackageContinentInfo struct {
	Code     string               `json:"code"`      // 国家名称
	Name     string               `json:"name"`      // 国家名称
	Sort     int                  `json:"sort"`      // 国家名称
	AreaList []ResPackageAreaInfo `json:"area_list"` //静态资源地址
}

// 返回的静态大洲信息
type PackageContinentPrice struct {
	PackageId int     `json:"package_id"` // 价格
	Price     float64 `json:"price"`      // 价格
	Unit      float64 `json:"unit"`       // 单价
}

// 返回的静态长效信息
type ResPackageAreaInfo struct {
	Id          int                     `json:"id"`
	PackageId   int                     `json:"package_id"`   // 套餐ID
	Area        string                  `json:"area"`         // 国家名称
	CountryName string                  `json:"country_name"` // 国家名称
	Country     string                  `json:"country"`      // 国家标识
	CountryImg  string                  `json:"country_img"`  // 国家国旗
	Money       float64                 `json:"money"`        // 价格
	Unit        float64                 `json:"unit"`         // 单价
	Default     string                  `json:"default"`
	IsHot       int                     `json:"is_hot"` // 是否热门推荐
	Sort        int                     `json:"sort"`
	Corner      string                  `json:"corner"`     //角标
	StaticUrl   string                  `json:"static_url"` //静态资源地址
	PriceArr    []PackageContinentPrice `json:"price_arr"`
	IpNumber    int                     `json:"ip_number"` // 数量"`
}

// 套餐
type ResStaticPackage struct {
	Id        int      `json:"id"`
	Code      string   `json:"code"`
	Name      string   `json:"name"`
	SubName   string   `json:"sub_name"`
	Price     float64  `json:"price"`
	ShowPrice float64  `json:"show_price"`
	Value     int64    `json:"value"`
	Corner    string   `json:"corner"`  //角标
	Default   string   `json:"default"` //是否默认选中
	IsHot     int      `json:"is_hot"`  //是否热门推荐
	Type      string   `json:"type"`
	ActTitle  string   `json:"act_title"` //活动名称
	ActDesc   string   `json:"act_desc"`  //推荐标题文案
	Content   string   `json:"content"`   //文案
	Unit      float64  `json:"unit"`      //IP单价
	UsUnit    float64  `json:"us_unit"`   //美国长效IP单价
	AllUnit   float64  `json:"all_unit"`  //全球长效IP单价
	Currency  string   `json:"currency"`
	StaticUrl string   `json:"static_url"` //静态资源地址
	NumArr    []string `json:"num_arr"`    //数量配置
	GiftUnit  string   `json:"gift_unit"`  // 赠送单位

}

func AllPackageList() (data []CmPackage, err error) {
	err = db.Table(packageTable).Where("terminal =?", "web").Where("status=?", 1).Order("sort desc").Find(&data).Error
	return
}

type CmPackageAreaInfo struct {
	Id          int     `json:"id"`
	PackageId   int     `json:"package_id"`   // 套餐ID
	Area        string  `json:"area"`         // 国家名称
	CountryName string  `json:"country_name"` // 国家名称
	Country     string  `json:"country"`      // 国家标识
	CountryImg  string  `json:"country_img"`  // 国家国旗
	Money       float64 `json:"money"`        // 价格
	Unit        float64 `json:"unit"`         // 单价
	Default     string  `json:"default"`
	IsHot       int     `json:"is_hot"` // 是否热门推荐
	Sort        int     `json:"sort"`
}

// packageId int
func GetPackageAreaList() (data []CmPackageAreaInfo) {
	dbs := db.Table("cm_package_area").Find(&data)
	//if packageId > 0 {
	//	dbs = dbs.Where("package_id =?", packageId)
	//}
	dbs.Where("status=?", 1).Find(&data)
	return
}

// 返回的大洲信息
type StaticContinent struct {
	Code string `json:"code"` // 国家名称
	Name string `json:"name"` // 国家名称
	Sort int    `json:"sort"` // 排序
}

func GetStaticContinentList(lang string) (data []StaticContinent) {
	name := "name_"
	if lang == "" {
		lang = "en"
	} else if lang == "zh-tw" || lang == "zh" || lang == "tw" || lang == "cn" {
		lang = "tw"
	}
	nameField := name + lang
	db.Table("cm_static_continent").Select("code,"+nameField+" as name,sort").Where("status=?", 1).Find(&data)
	return
}

// 返回的大洲信息
type StaticCountry struct {
	Code     string `json:"code"`      // 国家名称
	Name     string `json:"name"`      // 国家名称
	IpNumber int    `json:"ip_number"` // 数量"`
}

func GetStaticCountryList() (data []StaticCountry) {
	db.Table("cm_static_ip_country").Select("code,name,ip_number").Where("status=?", 1).Find(&data)
	return
}

func GetStaticCountryListByLang(lang string) (data []StaticCountry) {
	name := "name"
	if lang == "" || lang == "en" {
		name = "name"
	} else if lang == "zh-tw" || lang == "zh" || lang == "tw" || lang == "cn" {
		name = "name_tw"
	} else {
		name = "name_" + lang
	}
	db.Table("cm_static_ip_country").Select("code,"+name+" as name,ip_number").Where("status=?", 1).Find(&data)
	return
}

func GetStaticCountryByCountry(country string) (data StaticCountry) {
	country = strings.ToLower(country)
	db.Table("cm_static_ip_country").Select("code,name,ip_number").Where("country=?", country).Where("status=?", 1).First(&data)
	return
}

// 返回的国家信息
type StaticCountryArea struct {
	Code       string `json:"code"`        // 国家名称
	Name       string `json:"name"`        // 国家名称
	CountryImg string `json:"country_img"` // 图标"`
}

func GetStaticCountryByLang(lang string) (data []StaticCountryArea) {
	name := "name"
	if lang == "" || lang == "en" {
		name = "name"
	} else if lang == "zh-tw" || lang == "zh" || lang == "tw" || lang == "cn" {
		name = "name_tw"
	} else {
		name = "name_" + lang
	}
	db.Table("cm_static_ip_country").Select("code,"+name+" as name,country_img").Where("status=?", 1).Find(&data)
	return
}
