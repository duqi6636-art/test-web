package models

// CmUnlimitedPortPool undefined
type CmUnlimitedPortPool struct {
	ID          int    `json:"id" gorm:"id"`
	Ip          string `json:"ip" gorm:"ip"`                     // 机器ip
	Port        int    `json:"port" gorm:"port"`                 // 端口
	CreatedTime int    `json:"created_time" gorm:"created_time"` // 创建时间
	Status      string `json:"status" gorm:"status"`             // 状态
	Admin       string `json:"admin" gorm:"admin"`               // 操作人
}

// TableName 表名称
func (*CmUnlimitedPortPool) TableName() string {
	return "cm_unlimited_port_pool"
}

func GetUnlimitedPortPool(num int64) (data []CmUnlimitedPortPool, err error) {
	err = db.Model(CmUnlimitedPortPool{}).Where("status =?", 1).Limit(num).Order("port").Find(&data).Error
	return
}
