package models

import (
	"api-360proxy/web/db/clickhousedb"
	"api-360proxy/web/pkg/util"
)

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
	Supplier   string `json:"supplier"` //供应商 tencent  zenlayer
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

// 不限量预警

type UnlimitedEarlyWarning struct {
	Id     int    `json:"id"`
	Uid    int    `json:"uid"`
	Email  string `json:"email"`
	Status int    `json:"status"`
}

type UnlimitedEarlyWarningDetail struct {
	Id          int    `json:"id"`
	Uid         int    `json:"uid"`
	Ip          string `json:"ip"`
	Status      int    `json:"status"`
	InstanceId  string `json:"instance_id"`
	Cpu         int    `json:"cpu"`
	Memory      int    `json:"memory"`
	Bandwidth   int    `json:"bandwidth"`
	Concurrency int    `json:"concurrency"`
	Duration    int    `json:"duration"` // 持续时间
	SendTime    int64  `json:"send_time"`
	UpdateTime  int64  `json:"update_time"`
	CreateTime  int64  `json:"create_time"`
}

const (
	unlimitedEarlyWarningTable       = "cm_user_unlimited_early_warning"
	unlimitedEarlyWarningDetailTable = "cm_unlimited_early_warning_details"
)

func (u *UnlimitedEarlyWarning) Insert() {
	db.Table(unlimitedEarlyWarningTable).Create(u)
}

func (u *UnlimitedEarlyWarning) GetByUid() {
	db.Table(unlimitedEarlyWarningTable).Where("uid = ?", u.Uid).First(u)
}

func (u *UnlimitedEarlyWarning) Update() {
	db.Table(unlimitedEarlyWarningTable).Where("id = ?", u.Id).Save(u)
}

func GetUnlimitedEarlyWarningList(query string, args []interface{}) []UnlimitedEarlyWarning {
	var list = make([]UnlimitedEarlyWarning, 0)
	db.Table(unlimitedEarlyWarningTable).Where(query, args...).Find(&list)
	return list
}

func (u *UnlimitedEarlyWarningDetail) Insert() {
	db.Table(unlimitedEarlyWarningDetailTable).Create(u)
}

func (u *UnlimitedEarlyWarningDetail) GetByUidAndInstanceId() {
	db.Table(unlimitedEarlyWarningDetailTable).Where("uid = ? AND instance_id = ?", u.Uid, u.InstanceId).First(u)
}

func (u *UnlimitedEarlyWarningDetail) GetByIdAndUId() {
	db.Table(unlimitedEarlyWarningDetailTable).Where("id = ? AND uid = ?", u.Id, u.Uid).First(u)
}

func (u *UnlimitedEarlyWarningDetail) Update() {
	db.Table(unlimitedEarlyWarningDetailTable).Where("id = ? AND uid = ?", u.Id, u.Uid).Save(u)
}

func (u *UnlimitedEarlyWarningDetail) Delete() {
	db.Table(unlimitedEarlyWarningDetailTable).Where("id = ? AND uid = ?", u.Id, u.Uid).Delete(u)
}

func (u *UnlimitedEarlyWarningDetail) GetAll() []UnlimitedEarlyWarningDetail {
	var list = make([]UnlimitedEarlyWarningDetail, 0)
	db.Table(unlimitedEarlyWarningDetailTable).Where("uid = ? ", u.Uid).Find(&list)
	return list
}

// 不限量服务器数据 st_user_unlimited_cvm

type UserUnlimitedCvm struct {
	UID          uint32  `json:"uid"`           // 用户id
	Username     string  `json:"username"`      // 用户名
	Host         string  `json:"host"`          // host
	InsID        string  `json:"ins_id"`        // 实例ID
	Config       uint32  `json:"config"`        // 机器配置
	Bandwidth    uint32  `json:"bandwidth"`     // 机器带宽
	CpuAvg       float64 `json:"cpu_avg"`       // CPU平均值
	CpuMin       float64 `json:"cpu_min"`       // CPU最小值
	CpuMax       float64 `json:"cpu_max"`       // CPU最大值
	MemAvg       float64 `json:"mem_avg"`       // 内存平均值
	MemMin       float64 `json:"mem_min"`       // 内存最小值
	MemMax       float64 `json:"mem_max"`       // 内存最大值
	BandwidthAvg float64 `json:"bandwidth_avg"` // 带宽使用平均值
	BandwidthMin float64 `json:"bandwidth_min"` // 带宽使用最小值
	BandwidthMax float64 `json:"bandwidth_max"` // 带宽使用最大值
	OutAvg       float64 `json:"out_avg"`       // 出带宽利用率平均值
	OutMin       float64 `json:"out_min"`       // 出带宽利用率最小值
	OutMax       float64 `json:"out_max"`       // 出带宽利用率最大值
	TcpAvg       float64 `json:"tcp_avg"`       // TCP连接数平均值
	TcpMin       float64 `json:"tcp_min"`       // TCP连接数最小值
	TcpMax       float64 `json:"tcp_max"`       // TCP连接数最大值
	Today        uint32  `json:"today"`         // 当天时间
	Period       uint32  `json:"period"`        // 当前时间
	CreateTime   uint32  `json:"create_time"`   // 写入时间
}

const stUserUnlimitedCvm = "st_user_unlimited_cvm"

func GetUserUnlimitedCvmList(query string, args []interface{}) []UserUnlimitedCvm {
	var list = make([]UserUnlimitedCvm, 0)
	clickhousedb.ClickhouseDb.Table(stUserUnlimitedCvm).Where(query, args...).Find(&list)
	return list
}
