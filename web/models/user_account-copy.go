package models
/*
type UserAccount struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	Account    string `json:"account"`
	Password   string `json:"password"`
	Status     int    `json:"status"` // 账户状态: 1=正常,0= 禁用，-1：删除
	Remark     string `json:"remark"`
	CreateTime int    `json:"create_time"`
	UpdateTime int    `json:"update_time"`
	LimitFlow  int64  `json:"limit_flow"` // 流量限制
	UseFlow    int64  `json:"use_flow"`   // 流量限制
	Master     int    `json:"master"`     // 是否主账号
}

var user_account_table = "cm_user_account"

type ResUserAccount struct {
	Id         int    `json:"id"`
	Account    string `json:"account"`
	Password   string `json:"password"`
	LimitFlow  string `json:"limit_flow"` // 流量限制
	UseFlow    string `json:"use_flow"`   // 流量限制
	Master     int    `json:"master"`     // 是否主账号
	Status     int    `json:"status"`     // 账户状态: 1=正常,0= 禁用，-1：删除
	Exceed     int    `json:"exceed"`     // 是否用超  1超出
	Remark     string `json:"remark"`
	CreateTime string `json:"create_time"`
}

// 添加代理账户
func AddProxyAccount(data UserAccount) (err error) {
	err = db.Table(user_account_table).Create(&data).Error
	return
}

// 查询用户关联的代理账户信息
func GetUserAccount(uid int, account string) (err error, userAccount UserAccount) {
	dbs := db.Table(user_account_table)
	if uid > 0 {
		dbs = dbs.Where("uid = ?", uid)
	}
	if account != "" {
		dbs = dbs.Where("account = ?", account)
	}
	err = dbs.First(&userAccount).Error
	return
}

// 更新用户信息
func UpdateUserAccountById(id int, user interface{}) bool {
	db.Table(user_account_table).Where("id = ?", id).Updates(user)
	return true
}

// 查询用户关联的代理账户信息列表
func GetUserAccountList(uid int, account string, status int) (err error, lists []UserAccount) {
	dbs := db.Table(user_account_table)
	if uid > 0 {
		dbs = dbs.Where("uid = ?", uid)
	}
	if account != "" {
		dbs = dbs.Where("account like ?", "%"+account+"%")
	}
	if status == 10 {
		dbs = dbs.Where("status >= ?", 0)
	} else {
		dbs = dbs.Where("status = ?", status)
	}
	err = dbs.Where("master = ?", 0).Find(&lists).Error
	return
}

// 查询用户关联的代理账户信息
func GetUserAccountById(id int) (userAccount UserAccount, err error) {
	dbs := db.Table(user_account_table)
	if id > 0 {
		dbs = dbs.Where("id =?", id)
	}
	err = dbs.First(&userAccount).Error
	return
}

// 查询用户关联的代理账户信息
func GetUserAccountNeqId(id int, account string) (err error, userAccount UserAccount) {
	dbs := db.Table(user_account_table)
	if id > 0 {
		dbs = dbs.Where("id <> ?", id)
	}
	if account != "" {
		dbs = dbs.Where("account = ?", account)
	}
	err = dbs.First(&userAccount).Error
	return
}

type IpRecord struct {
	Uid int `json:"uid"`
	//Account    string `json:"account"`
	AccountId  int    `json:"account_id"`
	Port       string `json:"port"`
	Address    string `json:"address"`
	MarkUserIp string `json:"mark_user_ip"`
	MarkNeedIp string `json:"mark_need_ip"`
	IpType     string `json:"ip_type"`
	Region     string `json:"region"`
	RemoteAddr string `json:"remote_addr"` //下级代理地址
	UseTime    int64  `json:"use_time"`
	Useflow    int64  `json:"useflow"`
}

 */