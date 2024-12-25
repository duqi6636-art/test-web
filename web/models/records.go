package models

type Records struct {
	Id        int    `json:"id"`
	DomainId  int    `json:"domain_id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	Ttl       int    `json:"ttl"`
	Prio      int    `json:"prio"`
	Disabled  int    `json:"disabled"`
	Ordername string `json:"ordername"`
}

// 查询DNS解析数据
func GetDnsInfo(name string) (info Records) {
	dnsDb.Table("records").Where("name = ?", name).First(&info)
	return
}

// 写入dns解析数据
func AddDnsInfo(info Records) error {
	return dnsDb.Table("records").Create(&info).Error
}
