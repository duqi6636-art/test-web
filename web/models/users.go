package models

type Users struct {
	Id                int     `json:"id"`
	Username          string  `json:"username"`           // 用户名
	Email             string  `json:"email"`              // 登录邮箱
	Password          string  `json:"password"`           // 登录密码；sp_password加密
	PlaintextPassword string  `json:"plaintext_password"` // 明文密码
	Nickname          string  `json:"nickname"`           // 用户昵称
	GoogleUid         string  `json:"google_uid"`         // 谷歌登录唯一标识
	Platform          string  `json:"platform"`           // pc wap android ios web
	RegRegion         string  `json:"reg_region"`         // 注册省份
	RegCountry        string  `json:"reg_country"`        // 注册时的国家
	RegIp             string  `json:"reg_ip"`             // 注册ip
	Status            int     `json:"status"`             // 1正常 -1删除 -2禁用
	Balance           int     `json:"balance"`            // 剩余IP数
	PayIp             int     `json:"pay_ip"`             // 累计购买IP数
	Money             float64 `json:"money"`              // 累计支付金额
	IsPay             string  `json:"is_pay"`             // 是否付费过
	PayNumber         int     `json:"pay_number"`         // 累计支付次数
	PayMoney          float64 `json:"pay_money"`          // 累计支付金额
	Origin            string  `json:"origin"`             // 来源
	LastLoginIp       string  `json:"last_login_ip"`      // 最后登录ip
	LastLoginTime     int     `json:"last_login_time"`    // 最后登录时间
	LastUseTime       int     `json:"last_use_time"`      // 最后使用时间
	LockTime          int     `json:"lock_time"`          // 账户锁定时间大于当前时间,不允许登录
	Remark            string  `json:"remark"`             // 备注
	Version           int     `json:"version"`            // APP包内部版本号
	UpdateTime        int     `json:"update_time"`        // 更新时间
	CreateTime        int     `json:"create_time"`        // 创建时间
	BiddingKeyword    string  `json:"bidding_keyword"`    // 来源关键词
	BiddingCode       string  `json:"bidding_code"`       // 来源唯一标识
	BiddingPlat       string  `json:"bidding_plat"`       // 竞价来源 平台 如 ggtg bingtg
	BiddingDomain     string  `json:"bidding_domain"`     // 来源竞价域名
	TotalMoney        float64 `json:"total_money"`        // 总佣金 金额
	UsedMoney         float64 `json:"used_money"`         // 可用的佣金金额
	Sn                string  `json:"sn"`                 // SN
	Port              int     `json:"port"`               // 端口
	Areas             string  `json:"areas"`              // 新增：用户设置的国家地区代码
	Wallet            string  `json:"wallet"`             // 钱包地址
	FrozenHour        int     `json:"frozen_hour"`        // IP冻结时间
	IpStatus          int     `json:"ip_status"`          // IP状态
	GithubId          int     `json:"github_id"`          // github唯一标识
	MpGithubId        int     `json:"mp_github_id"`       // 代理管理器github唯一标识
	DeviceToken       string  `json:"device_token"`
}

// 用户信息返回信息
type ResUser struct {
	Session string      `json:"session"`
	User    ResUserInfo `json:"user"`
}

// 用户信息返回信息
type ResUserInfo struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	Email        string `json:"email"`
	Balance      int64  `json:"balance"`
	IsPay        int    `json:"is_pay"`        //是否是已购买
	GoogleAuth   int    `json:"google_auth"`   //是否 已经绑定  1 已经绑定
	Invite       string `json:"invite"`        //邀请码
	InviteUrl    string `json:"invite_url"`    //邀请地址
	IpNum        int    `json:"ip_num"`        //剩余IP数
	AgentBalance int    `json:"agent_balance"` //代理余额
	Flow         string `json:"flow"`          //用户剩余流量 GB
	FlowRedeem   string `json:"flow_redeem"`   //可兑换 流量余额 GB
	FlowUnit     string `json:"flow_unit"`     //用户剩余流量单位 GB
	FlowMb       string `json:"flow_mb"`       //用户剩余流量 MB
	FlowMbUnit   string `json:"flow_mb_unit"`  //用户剩余流量单位 MB
	FlowDate     string `json:"flow_date"`     //用户剩余流量到期时间
	FlowExpire   int    `json:"flow_expire"`   //用户剩余流量是否到期
	IsAct        int    `json:"is_act"`        //用户剩余流量是否到期
}

var CmUserTable = "cm_users"

// 查询用户
func GetUserCountByIp(ip string, start int) (num int) {
	db.Table(CmUserTable).Where("reg_ip =?", ip).Where("create_time > ?", start).Count(&num)
	return
}

// 查询用户
func GetUserInfo(maps interface{}) (err error, user Users) {
	err = db.Table(CmUserTable).Where(maps).Where("status>=?", 1).First(&user).Error
	return
}

// 用户ID
func GetUserById(uid int) (err error, user Users) {
	err = db.Table(CmUserTable).Where("id = ? ", uid).Where("status >= ?", 1).First(&user).Error
	return
}

// 邮箱
func GetUserByEmail(name string) (err error, user Users) {
	err = db.Table(CmUserTable).Where("email = ? ", name).Where("status >= ?", 1).First(&user).Error
	return
}

// 根据谷歌ID查询用户信息
func GetUserByGuid(name string) (err error, user Users) {
	err = db.Table(CmUserTable).Where("google_uid = ? ", name).Where("status >= ?", 1).First(&user).Error
	return
}

func GetUserByGithubId(name int) (err error, user Users) {
	err = db.Table(CmUserTable).Where("github_id = ? ", name).Where("status >= ?", 1).First(&user).Error
	return
}

func GetUserByMpGithubId(name int) (err error, user Users) {
	err = db.Table(CmUserTable).Where("mp_github_id = ? ", name).Where("status >= ?", 1).First(&user).Error
	return
}

// 用户名
func GetUserByUsername(name string) (err error, user Users) {
	err = db.Table(CmUserTable).Where("username = ? ", name).Where("status >= ?", 1).First(&user).Error
	return
}

// 查询用户消费金额大于
func GetUserListByPayMoney(payMoney float64) (err error, users []Users) {
	err = db.Table(CmUserTable).Where("pay_money>=?", payMoney).Where("status>=?", 1).Find(&users).Error
	return
}

// 添加用户
func AddUser(user Users) (err error, uid int) {
	err = db.Table(CmUserTable).Create(&user).Error
	return err, user.Id
}

// 更新用户信息
func UpdateUserById(id int, user interface{}) bool {
	db.Table(CmUserTable).Model(&Users{}).Where("id = ?", id).Updates(user)
	return true
}

func ExistUserByUuid(uuid string) bool {
	var user Users
	return !db.Table(CmUserTable).Where("uuid = ?", uuid).First(&user).RecordNotFound()
}

func EditUserByMap(where interface{}, params interface{}) (err error) {
	err = db.Table(CmUserTable).Where(where).Update(params).Error
	return err
}

func EditUserById(id int, params interface{}) (err error) {
	err = db.Table(CmUserTable).Where("id =?", id).Update(params).Error
	return err
}

// 获取加严记录 根据设备号
func GetUsersBySaltCount(salt string, start int) (num int) {
	db.Table(CmUserTable).Where("device_token =?", salt).Where("create_time > ?", start).Count(&num)
	return
}
