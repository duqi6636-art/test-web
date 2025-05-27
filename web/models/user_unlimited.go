package models

import "api-360proxy/web/pkg/util"

// 不限量流量IP
type PoolFlowDayModel struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	Ip         string `json:"ip"`
	Port       int    `json:"port"`
	Port2      int    `json:"port2"`
	Config     int    `json:"config"`    //配置
	Bandwidth  int    `json:"bandwidth"` //带宽
	Country    string `json:"country"`
	ExpireTime int    `json:"expire_time"`
}

// 不限量流量IP 不用接口对外展示
type PoolFlowDayDetailModel struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	Ip         string `json:"ip"`
	Port       int    `json:"port"`
	Port2      int    `json:"port2"`
	Country    string `json:"country"`
	InstanceId string `json:"instance_id"`
	Config     int    `json:"config"`    //配置
	Bandwidth  int    `json:"bandwidth"` //带宽
	Region     string `json:"region"`
	ExpireTime int    `json:"expire_time"`
}

// 20250308 逻辑 由原来的一个换成多个 获取有效使用的不限量列表
func ListPoolFlowDayByUid(uid int) (info []PoolFlowDayModel) {
	nowTime := util.GetNowInt()
	db.Table("cm_pool_flow_day").Where("uid = ? ", uid).Where("expire_time >=? ", nowTime).Where("status =? ", 1).Find(&info)
	return
}

// 获取 不限量列表-全部
func ListPoolFlowDayByUidAll(uid int) (info []PoolFlowDayModel) {
	db.Table("cm_pool_flow_day").Where("uid = ? ", uid).Where("status =? ", 1).Find(&info)
	return
}

func GetPoolFlowDayByIp(uid int, ip string) (info PoolFlowDayDetailModel) {
	db.Table("cm_pool_flow_day").Where("uid = ? ", uid).Where("ip = ? ", ip).Where("status =? ", 1).First(&info)
	return
}

func GetPoolFlowDayByUid_copy(uid int) (info PoolFlowDayModel) {
	nowTime := util.GetNowInt()
	db.Table("cm_pool_flow_day").Where("uid = ? ", uid).Where("expire_time >=? ", nowTime).Where("status =? ", 1).First(&info)
	return
}
func ScoreGetPoolFlowDayByUid(uid int) (info PoolFlowDayModel) {
	db.Table("cm_pool_flow_day").Where("uid = ? ", uid).Where("status =? ", 1).First(&info)
	return
}

// 修改IP池信息
func EditPoolFlowDay(id int, params interface{}) (err error) {
	err = db.Table("cm_pool_flow_day").Where("id =?", id).Update(params).Error
	return err
}

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
	Config       string `json:"config"`        // 标识 1购买  2生成cdk
	Bandwidth    string `json:"bandwidth"`     // 类型 isp flow 等 生成cdk类型
	ConfigNum    int    `json:"config_num"`    // 标识 并发配置
	BandwidthNum int    `json:"bandwidth_num"` // 类型 带宽
	ExpireTime   string `json:"expire_time"`   // 过期时间
	Ip           string `json:"ip"`            // 不限量IP
	Port         int    `json:"port"`          // 不限量IP
	Status       int    `json:"status"`        // 状态 1待使用 2已使用 3已过期
}

// 查询记录
func GetUserUnlimitedRecord(uid int) (data []UserUnlimitedModel) {
	db.Table("log_user_unlimited").Where("uid = ?", uid).Order("id DESC").Find(&data)
	return
}
