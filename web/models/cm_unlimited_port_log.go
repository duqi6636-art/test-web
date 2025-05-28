package models

import "api-360proxy/web/pkg/util"

// CmUnlimitedPortLog undefined
type CmUnlimitedPortLog struct {
	ID          int    `json:"id" gorm:"id"`
	Uid         int    `json:"uid" gorm:"uid"`                   // 用户id
	Username    string `json:"username" gorm:"username"`         // username
	Ip          string `json:"ip" gorm:"ip"`                     // 机器ip\r\n
	Port        int    `json:"port" gorm:"port"`                 // 端口
	Region      string `json:"region" gorm:"region"`             // 国家
	Minute      int    `json:"minute" gorm:"minute"`             // 时间
	ExpiredTime int    `json:"expired_time" gorm:"c"`            // 过期时间
	CreatedTime int    `json:"created_time" gorm:"created_time"` // 创建时间
}

// TableName 表名称
func (*CmUnlimitedPortLog) TableName() string {
	return "cm_unlimited_port_log"
}

var userFlowDayPortTable = "cm_unlimited_port_log"

func InsertUnlimitedPortLog(logs CmUnlimitedPortLog) error {
	err := db.Table("cm_unlimited_port_log").Create(&logs).Error
	return err
}

// 查询用户 不限时端口套餐信息
func GetUserFlowDayPortByUid(uid int) (user []CmUnlimitedPortLog) {
	db.Table(userFlowDayPortTable).
		Where("uid =?", uid).
		Where("expired_time >?", util.GetNowInt()).
		Order("created_time desc").Find(&user)
	return
}

// 查询用户 可用不限时端口套餐信息
func GetUserCanFlowDayPortByUid(uid int, num int, expiredTime int) (user []CmUnlimitedPortLog) {
	db.Table(userFlowDayPortTable).Where("uid =?", uid).
		Where("expired_time = ?", expiredTime).
		Where("status  = ?", 1).
		Order("expired_time desc").
		Limit(num).
		Find(&user)
	return
}

func GetUserUnlimitedPortByUid(uid int) (user []CmUnlimitedPortLog) {
	db.Table(userFlowDayPortTable).Where("uid =?", uid).
		Where("expired_time >?", util.GetNowInt()).
		Order("created_time desc").Find(&user)
	return
}
