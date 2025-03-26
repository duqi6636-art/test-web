package models

import (
	"api-360proxy/web/pkg/util"
)

type UserFlow struct {
	ID         int    `json:"id"`
	Uid        int    `json:"uid"`
	Email      string `json:"email"`       // 用户邮箱
	Username   string `json:"username"`    // 用户邮箱
	AllFlow    int64  `json:"all_flow"`    // 总流量数据
	BuyFlow    int64  `json:"buy_flow"`    // 总购买 流量
	ExFlow     int64  `json:"ex_flow"`     // 兑换得到 流量
	CdkFlow    int64  `json:"cdk_flow"`    // 已经生成cdk的 流量
	Flows      int64  `json:"flows"`       // 剩余流量
	PreFlow    int64  `json:"pre_flow"`    // 上次购买流量
	SendOpen   int    `json:"send_open"`   // 流量阀值 开关
	SendFlows  int64  `json:"send_flows"`  // 流量阀值
	SendUnit   string `json:"send_unit"`   // 流量阀值 单位
	Day        int    `json:"day"`         // 过期天数
	DayMax     int    `json:"day_max"`     // 过期天数-最大
	ExpireTime int    `json:"expire_time"` // 过期时间
	CreateTime int    `json:"create_time"`
	Status     int    `json:"status"` //状态 1正常2禁用
}

var userFlowTable = "cm_user_flow"

// 查询用户流量表
func GetUserFlowInfo(uid int) (data UserFlow) {
	db.Table(userFlowTable).Where("uid = ?", uid).Find(&data)
	return
}

// 创建用户信息
func CreateUserFlow(data UserFlow) (err error, uid int) {
	err = db.Table(userFlowTable).Create(&data).Error
	return err, data.Uid
}

// 修改用户余额
func EditUserFlow(id int, params interface{}) (err error) {
	err = db.Table(userFlowTable).Where("id =?", id).Update(params).Error
	return err
}

// 修改用户余额
func EditUserFlowByUid(uid int, params interface{}) (err error) {
	err = db.Table(userFlowTable).Where("uid =?", uid).Where("status =?", 1).Update(params).Error
	return err
}

// 获取 阀值列表信息
func UserFlowLists() (data []UserFlow, err error) {
	err = db.Table(userFlowTable).Where("flows < send_flows").Where("send_flows >?", 0).Where("send_has =?", 0).Find(&data).Error
	return
}

type LogFlowsAccount struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`         // 用户id
	AccountId  int    `json:"account_id"`  // 子账户ID
	Username   string `json:"username"`    // 子账户名
	PreFlows   int64  `json:"pre_flows"`   // 变动前的 余额
	PreLimit   int64  `json:"pre_limit"`   // 变动前的 配置额度
	Flows      int64  `json:"flows"`       // 变动后的 余额
	LimitFlow  int64  `json:"limit_flow"`  // 变动后的 配置额度
	Cate       string `json:"cate"`        // 处理IP
	Ip         string `json:"ip"`          // 处理IP
	Remark     string `json:"remark"`      // 处理信息
	CreateTime int    `json:"create_time"` // 操作时间
}

// 创建用户信息
func CreateLogFlowsAccount(uid, accountId int, preFlow, preLimit, flows, limitFlow int64, username, ip, cate, remark string) (err error) {
	data := LogFlowsAccount{}
	data.Uid = uid
	data.AccountId = accountId
	data.Username = username
	data.PreFlows = preFlow
	data.PreLimit = preLimit
	data.Flows = flows
	data.LimitFlow = limitFlow
	data.Cate = cate
	data.Ip = ip
	data.Remark = remark
	data.CreateTime = util.GetNowInt()
	err = db.Table("log_flows_account").Create(&data).Error
	return err
}

// 长效ISP流量操作记录添加
func CreateLogLongIspFlowsAccount(uid, accountId int, preFlow, preLimit, flows, limitFlow int64, username, ip, cate, remark string) (err error) {
	data := LogFlowsAccount{}
	data.Uid = uid
	data.AccountId = accountId
	data.Username = username
	data.PreFlows = preFlow
	data.PreLimit = preLimit
	data.Flows = flows
	data.LimitFlow = limitFlow
	data.Cate = cate
	data.Ip = ip
	data.Remark = remark
	data.CreateTime = util.GetNowInt()
	err = db.Table("log_long_isp_flows_account").Create(&data).Error
	return err
}

type UserFlowDayModel struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	AllDay     int    `json:"all_day"` // 总购买	时间（s）
	ExpireTime int    `json:"expire_time"`
	PreDay     int    `json:"pre_day"` // 上次购买的过期时间
	CreateTime int    `json:"create_time"`
	Remark     string `json:"remark"` //
	Hostname   string `json:"hostname"`
	Status     int    `json:"status"` // 状态 1正常2禁用
}

var userFlowDayTable = "cm_user_flow_day"

// 查询用户 不限时套餐信息
func GetUserFlowDayByUid_copy(uid int) (user UserFlowDayModel) {
	db.Table(userFlowDayTable).Where("uid =?", uid).First(&user)
	return
}

// 查询用户 不限时套餐信息
func GetUserFlowDayByUid(uid int) (user []UserFlowDayModel) {
	db.Table(userFlowDayTable).Where("uid =?", uid).Order("expire_time desc").Find(&user)
	return
}

// 不限量流量套餐
type UserFlowDay struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	AllDay     int    `json:"all_day"`     // 总购买	时间（s）
	ExpireTime int    `json:"expire_time"` // 过期时间
	PreDay     int    `json:"pre_day"`     // 上次购买的过期时间
	CreateTime int    `json:"create_time"`
	Remark     string `json:"remark"` //
	Status     int    `json:"status"` // 状态
}

// 查询用户流量过期时间表
func GetUserFlowDay(uid int) (data UserFlowDay) {
	db.Table(userFlowDayTable).Where("uid = ?", uid).Find(&data)
	return
}

// 创建用户信息
func CreateUserFlowDay(data UserFlowDay) (err error, uid int) {
	err = db.Table(userFlowDayTable).Create(&data).Error
	return err, data.Uid
}

// 修改用户余额
func EditUserFlowDay(id int, params interface{}) (err error) {
	err = db.Table(userFlowDayTable).Where("id =?", id).Update(params).Error
	return err
}

// 修改
func EditUserFlowDayBy(where interface{}, params interface{}) (err error) {
	err = db.Table(userFlowDayTable).Where(where).Update(params).Error
	return err
}
