package models

type CmDnsDomainModel struct {
	Id      int    `json:"id"`      // ID
	Code    string `json:"code"`    // 国家标识
	Country string `json:"country"` // 国家
	Domain  string `json:"domain"`  // 域名
}

// 查询负载列表
func GetDnsDomainList(cate int) (list []CmDnsDomainModel) {
	db.Table("cm_dns_domain").Where("cate = ?", cate).Where("status = ?", 1).Find(&list)
	return list
}

type UserLoadFlowModel struct {
	Uid      int    `json:"id"`
	Username string `json:"username"`
	Cate     int    `json:"cate"`
}

// 查询用户所属流量类型
func GetUserFlowCate(uid int) (cate int) {
	info := UserLoadFlowModel{}
	err := db.Table("cm_user_load_flow").Where("uid = ? ", uid).First(&info).Error
	if err == nil && info.Uid > 0 {
		cate = info.Cate
	} else {
		cate = 1
	}
	return cate
}

// 查询域名列表
type CmDnsFlowDomainModel struct {
	Id      int    `json:"id"`      // ID
	Code    string `json:"code"`    // 国家标识
	Country string `json:"country"` // 国家名
	Domain  string `json:"domain"`  // 域名
	Cate    int    `json:"cate"`
}

func GetDnsFlowDomainInfo(cate int) (list CmDnsFlowDomainModel) {
	db.Table("cm_dns_flow_domain").Where("cate =?", cate).Where("status = ?", 1).First(&list)
	return list
}

func GetDnsFlowDomainList(cate int) (list []CmDnsFlowDomainModel) {
	db.Table("cm_dns_flow_domain").Where("cate = ?", cate).Where("status =?", 1).Find(&list)
	return list
}

type UserLoadCateModel struct {
	Uid      int    `json:"uid"`
	Username string `json:"username"`
	Cate     int    `json:"cate"`
}

func GetUserIspCate(uid int) (cate int) {
	info := UserLoadCateModel{}
	err := db.Table("cm_user_load_type").Where("uid = ? ", uid).First(&info).Error
	if err == nil && info.Uid > 0 {
		cate = info.Cate
	} else {
		cate = 0
	}
	return cate
}
