package models

type ProxyConfig struct {
	Id          int    `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL" json:"id,omitempty"`
	Uid         int    `gorm:"column:uid;default:NULL;comment:'uid'" json:"uid,omitempty"`
	Default     string `gorm:"column:default;comment:'基础信息'" json:"default,omitempty"`
	Columns     string `gorm:"column:columns;comment:'表头配置'" json:"columns,omitempty"`
	ProxyLists  string `gorm:"column:proxy_lists;comment:'端口配置'" json:"proxy_lists,omitempty"`
	BrowserType string `gorm:"column:browser_type;comment:'浏览器配置'" json:"browser_type,omitempty"`
	CreatedAt   int    `gorm:"column:created_at;default:NULL" json:"created_at,omitempty"`
}

func (p *ProxyConfig) TableName() string {
	return "proxy_config"
}

func AddProxyConfig(proxyConfig ProxyConfig) {
	db.Create(&proxyConfig)
}

func GetProxyConfig(uid int) (ProxyConfig, error) {
	var proxyConfig ProxyConfig
	err := db.Where("uid =?", uid).First(&proxyConfig).Error
	return proxyConfig, err
}

func UpdateProxyConfig(uid int, proxyConfig ProxyConfig) {
	db.Model(&ProxyConfig{}).Where("uid =?", uid).Updates(proxyConfig)
}
