package models

// 用户不限量配置记录
type UserUnlimitedModel struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`         // 用户id
	Config     int    `json:"config"`      // 标识 1购买  2生成cdk
	Bandwidth  int    `json:"bandwidth"`   // 类型 isp flow 等 生成cdk类型
	StartTime  int    `json:"start_time"`  // 开始时间
	ExpireTime int    `json:"expire_time"` // 过期时间
	Ip         string `json:"ip"`          // 不限量IP
	OrderId    string `json:"order_id"`    // 购买的关联订单
	CreateTime int    `json:"create_time"` // 操作时间(购买)
}

// 用户不限量配置记录
type ResUserUnlimitedModel struct {
	//Id         int    `json:"id"`
	Config     string `json:"config"`      // 标识 1购买  2生成cdk
	Bandwidth  string `json:"bandwidth"`   // 类型 isp flow 等 生成cdk类型
	ExpireTime string `json:"expire_time"` // 过期时间
	Ip         string `json:"ip"`          // 不限量IP
	Status     int    `json:"status"`      // 状态 1待使用 2已使用 3已过期
}

//查询记录
func GetUserUnlimitedRecord(uid int) (data []UserUnlimitedModel) {
	db.Table("log_user_unlimited").Where("uid = ?", uid).Order("id DESC").Find(&data)
	return
}

