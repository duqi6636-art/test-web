package models

// 不限量异步队列记录
type ListUserUnlimitedModel struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`         // 用户id
	Config     int    `json:"config"`      // 标识 1购买  2生成cdk
	Bandwidth  int    `json:"bandwidth"`   // 类型 isp flow 等 生成cdk类型
	StartTime  int    `json:"start_time"`  // 开始生效时间
	ExpireTime int    `json:"expire_time"` // 过期时间
	Day        int    `json:"day"`         // 购买天数
	InstanceId string `json:"instance_id"` //实例ID
	Ip         string `json:"ip"`          // 不限量IP
	OrderId    string `json:"order_id"`    // 购买的关联订单
	Status     int    `json:"status"`      // 状态 1 正常 2冻结
	CreateTime int    `json:"create_time"` // 操作时间(购买)
	ExecData   string `json:"exec_data"`   // 订单配置
	PakType    string `json:"pak_type"`    // 套餐类型
}

// 用户不限量配置修改记录
func AddLogUserUnlimited(uid int, config, bandwidth, expire_time int, day, startTime int, orderId string, createTime int, data, pak_type string) (err error) {
	info := ListUserUnlimitedModel{
		Uid:        uid,
		Config:     config,
		Bandwidth:  bandwidth,
		StartTime:  startTime,
		ExpireTime: expire_time,
		Day:        day,
		Ip:         "",
		InstanceId: "",
		OrderId:    orderId,
		Status:     0, //0待处理 1已开实例 2已经分配IP  3需要手动操作
		CreateTime: createTime,
		ExecData:   data,
		PakType:    pak_type,
	}
	var tableName = "list_user_unlimited"
	err = db.Table(tableName).Create(&info).Error
	return
}
