package models

type LongIspUserAccount struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	Account    string `json:"account"`
	Password   string `json:"password"`
	Status     int    `json:"status"` // 账户状态: 1=正常,0= 禁用，-1：删除
	Remark     string `json:"remark"`
	CreateTime int    `json:"create_time"`
	UpdateTime int    `json:"update_time"`
	LimitFlow  int64  `json:"limit_flow"`  // 流量限制
	Master     int    `json:"master"`      // 是否主账号
	Flows      int64  `json:"flows"`       // 剩余流量
	FlowUnit   string `json:"flow_unit"`   // 流量单位
	ExpireTime int    `json:"expire_time"` // 过期时间
	UseFlow    int64  `json:"use_flow"`    // 已使用流量

}

var user_account_long_isp_table = "cm_user_account_dynamic_isp"

// 添加代理账户
func AddLongIspProxyAccount(data LongIspUserAccount) (err error, id int) {
	err = db.Table(user_account_long_isp_table).Create(&data).Error
	id = data.Id
	return
}

// 查询用户关联的代理账户信息
func GetLongIspUserAccount(uid int, account string) (err error, userAccount LongIspUserAccount) {
	dbs := db.Table(user_account_long_isp_table)
	if uid > 0 {
		dbs = dbs.Where("uid = ?", uid)
	}
	if account != "" {
		dbs = dbs.Where("account = ?", account)
	}
	err = dbs.First(&userAccount).Error
	return
}

// 查询用户关联的代理账户信息
func GetLongIspUserAccountMaster(uid int) (err error, userAccount LongIspUserAccount) {
	dbs := db.Table(user_account_long_isp_table)
	if uid > 0 {
		dbs = dbs.Where("uid = ?", uid)
	}
	dbs = dbs.Where("master = ?", 1)
	err = dbs.First(&userAccount).Error
	return
}

// 更新用户信息
func UpdateLongIspUserAccountById(id int, user interface{}) bool {
	db.Table(user_account_long_isp_table).Where("id = ?", id).Updates(user)
	return true
}

// 查询用户关联的代理账户信息列表
func GetLongIspUserAccountAllList(uid int, account string) (err error, lists []LongIspUserAccount) {
	dbs := db.Table(user_account_long_isp_table)
	if uid > 0 {
		dbs = dbs.Where("uid = ?", uid)
	}
	if account != "" {
		dbs = dbs.Where("account like ?", "%"+account+"%")
	}

	dbs = dbs.Where("status >= ?", 0)
	err = dbs.Find(&lists).Error
	return
}

// 查询用户关联的代理账户信息列表
func GetLongIspUserAccountList(uid int, account string) (err error, lists []LongIspUserAccount) {
	dbs := db.Table(user_account_long_isp_table)
	if uid > 0 {
		dbs = dbs.Where("uid = ?", uid)
	}
	if account != "" {
		dbs = dbs.Where("account like ?", "%"+account+"%")
	}

	dbs = dbs.Where("status >= ?", 0)
	dbs = dbs.Where("master = ?", 0)
	err = dbs.Find(&lists).Error
	return
}

// 查询用户关联的代理账户信息
func GetLongIspUserAccountById(id int) (userAccount LongIspUserAccount, err error) {
	dbs := db.Table(user_account_long_isp_table)
	if id > 0 {
		dbs = dbs.Where("id =?", id)
	}
	err = dbs.First(&userAccount).Error
	return
}

// 查询用户关联的代理账户信息
func GetLongIspUserAccountNeqId(id int, account string) (err error, userAccount LongIspUserAccount) {
	dbs := db.Table(user_account_long_isp_table)
	if id > 0 {
		dbs = dbs.Where("id <> ?", id)
	}
	if account != "" {
		dbs = dbs.Where("account = ?", account)
	}
	err = dbs.First(&userAccount).Error
	return
}
