package models

import "api-360proxy/web/pkg/util"

// UserAutoRenewOrderModel 用户自动续费扣款顺序配置
type UserAutoRenewOrderModel struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`         // 用户ID
	Username   string `json:"username"`    // 用户名
	RenewType  string `json:"renew_type"`  // 续费类型: isp, static, flow, rotating, unlimited
	Priority   int    `json:"priority"`    // 优先级，数字越小优先级越高
	Status     int    `json:"status"`      // 状态 1启用 0禁用
	Ip         string `json:"ip"`          // 操作IP
	UpdateTime int    `json:"update_time"` // 更新时间
	CreateTime int    `json:"create_time"` // 创建时间
}

// UserAutoRenewOrderResModel 用户自动续费扣款顺序配置
type UserAutoRenewOrderResModel struct {
	Id        int    `json:"id"`
	Uid       int    `json:"uid"`        // 用户ID
	Username  string `json:"username"`   // 用户名
	RenewType string `json:"renew_type"` // 续费类型: isp, static, flow, rotating, unlimited
}

var userAutoRenewOrderTable = "cm_user_auto_renew_order"

// GetUserAutoRenewOrderList 获取用户扣款顺序配置列表（按优先级排序）
func GetUserAutoRenewOrderList(uid int) (data []UserAutoRenewOrderResModel) {
	db.Table(userAutoRenewOrderTable).
		Where("uid = ? AND status = ?", uid, 1).
		Order("priority ASC, id ASC").
		Find(&data)
	return
}

// AddUserAutoRenewOrder 添加用户扣款顺序配置
func AddUserAutoRenewOrder(info UserAutoRenewOrderModel) (err error) {
	err = db.Table(userAutoRenewOrderTable).Create(&info).Error
	return
}

// BatchUpdateUserAutoRenewOrder 批量更新用户扣款顺序配置
func BatchUpdateUserAutoRenewOrder(uid int, orders []UserAutoRenewOrderModel) (err error) {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 先删除用户所有配置
	if err = tx.Table(userAutoRenewOrderTable).Where("uid = ?", uid).Delete(&UserAutoRenewOrderModel{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 批量插入新配置
	for _, order := range orders {
		if err = tx.Table(userAutoRenewOrderTable).Create(&order).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetDefaultAutoRenewOrder 获取默认扣款顺序配置
func GetDefaultAutoRenewOrder() []string {
	// 默认顺序
	return []string{"isp", "flow", "static", "rotating", "unlimited"}
}

// InitUserAutoRenewOrder 初始化用户默认扣款顺序
func InitUserAutoRenewOrder(uid int, username string, ip string) error {
	nowTime := util.GetNowInt()
	defaultOrder := GetDefaultAutoRenewOrder()

	for i, renewType := range defaultOrder {
		order := UserAutoRenewOrderModel{
			Uid:        uid,
			Username:   username,
			RenewType:  renewType,
			Priority:   i + 1,
			Status:     1,
			Ip:         ip,
			UpdateTime: nowTime,
			CreateTime: nowTime,
		}

		if err := AddUserAutoRenewOrder(order); err != nil {
			return err
		}
	}

	return nil
}
