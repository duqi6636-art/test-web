package models

type StatsClick struct {
	Code       string `json:"code"`     // 来源标识
	Param      string `json:"param"`    // 来源标识
	Language   string `json:"language"` // 语言
	Platform   string `json:"platform"` // 设备
	Uid        int    `json:"uid"`      // 用户ID
	CreateTime int    `json:"create_time"`
	Today      int    `json:"today"`
	Ip         string `json:"ip"` // 用户IP
}

// 创建用户信息
func CreateStatsClick(data StatsClick) (err error) {
	err = db.Table("cm_stats_click").Create(&data).Error
	return
}
