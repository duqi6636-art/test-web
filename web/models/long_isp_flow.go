package models

type UserDynamicIsp struct {
	ID         int    `json:"id"`
	Uid        int    `json:"uid"`
	Email      string `json:"email"`       // 用户邮箱
	Username   string `json:"username"`    // 用户名
	AllFlow    int64  `json:"all_flow"`    // 总充值
	Flows      int64  `json:"flows"`       // 剩余流量
	ExpireTime int    `json:"expire_time"` // 过期时间
	CreateTime int    `json:"create_time"`
	SendOpen   int    `json:"send_open"`  // 流量阀值 开关
	SendFlows  int64  `json:"send_flows"` // 流量阀值
	SendUnit   string `json:"send_unit"`  // 流量阀值 单位
	//ExFlow     int64  `json:"ex_flow"`    // 兑换得到 流量
	Status     int    `json:"status"` 	  // 状态 1正常2禁用
}

// GetFlowsStruct 获取账户 流量信息返回
type GetFlowsStruct struct {
	Flow       int64  `json:"flow"`        // 账户流量
	ExpireDate string `json:"expire_date"` // 账户流量到期日
	Expired    int    `json:"expired"`     // 流量到期时间戳
	FlowGb     string `json:"flow_gb"`     // 流量单位为GB
	FlowMb     string `json:"flow_mb"`     // 流量单位为MB
	SendFlow   string `json:"send_flow"`   // 已发送邮件
	Status     int    `json:"status"`      // 是否冻结流量
	SendOpen   int    `json:"send_open"`   // 是否开通邮件提醒
	SendUnit   string `json:"send_unit"`   // 邮件提醒单位
	UnitGb     string `json:"unit_gb"`     // 单位为GB
	UnitMb     string `json:"unit_mb"`     // 单位为MB
}

var userDynamicIspTable = "cm_user_dynamic_isp"

// 查询用户流量表
func GetUserDynamicIspInfo(uid int) (data UserDynamicIsp) {
	db.Table(userDynamicIspTable).Where("uid = ?", uid).Find(&data)
	return
}
func EditUserDynamicIspByUid(uid int, params interface{}) (err error) {
	err = db.Table(userDynamicIspTable).Where("uid =?", uid).Update(params).Error
	return err
}
// 创建用户信息
func CreateUserDynamicIsp(data UserDynamicIsp) (err error, uid int) {
	err = db.Table(userDynamicIspTable).Create(&data).Error
	return err, data.Uid
}
// 获取长效Isp流量阀值列表信息
func UserLongIspSendFlowLists() (data []UserFlow, err error) {
	err = db.Table(userDynamicIspTable).Where("flows < send_flows").Where("send_flows >?", 0).Where("send_has =?", 0).Find(&data).Error
	return
}