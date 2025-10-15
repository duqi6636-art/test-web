package models

import "github.com/jinzhu/gorm"

// ConfBalanceRenewModel 余额自动续费配置
type ConfBalanceRenewModel struct {
	Id             int     `json:"id"`
	Name           string  `json:"name"`             // 名称
	Cate           string  `json:"cate"`             // 类型
	CateName       string  `json:"cate_name"`        // 类型
	Value          int     `json:"value"`            // 兑换套餐值
	Price          float64 `json:"price"`            // 兑换单价
	Unit           string  `json:"unit"`             // 兑换单位
	UnitValue      int64   `json:"unit_value"`       // 兑换值
	Min            int     `json:"min"`              // 最小续费值
	Max            int     `json:"max"`              // 最大续费值
	IsCustom       int     `json:"is_custom"`        // 是否自定义
	SingleIspPrice float64 `json:"single_isp_price"` // 单IP价格
	DoubleIspPrice float64 `json:"double_isp_price"` // 双IP价格
	LocalIspPrice  float64 `json:"local_isp_price"`  // 本地IP价格
	PakId          int     `json:"pak_id"`           // 关联的套餐ID
}

// ResConfBalanceRenewModel 余额兑换配置
type ResConfBalanceRenewModel struct {
	Id         string  `json:"id"`
	Name       string  `json:"name"`        // 名称
	NameExtra  string  `json:"name_extra"`  // 名称中加号后面的部分
	Value      int     `json:"value"`       // 兑换套餐值
	Price      float64 `json:"price"`       // 单价
	TotalPrice float64 `json:"total_price"` // 展示总价
	Unit       string  `json:"unit"`        // 单位
	IsCustom   int     `json:"is_custom"`   // 是否自定义
	IsRenew    int     `json:"is_renew"`    // 是否已设置
	Balance    int     `json:"balance"`     // 余额剩余
	Day        int     `json:"day"`         // 剩余几天有效期自动续费
}

// ResConfStaticBalanceRenewModel 余额兑换配置
type ResConfStaticBalanceRenewModel struct {
	Id             string  `json:"id"`
	Name           string  `json:"name"`             // 名称
	Value          int     `json:"value"`            // 兑换套餐值
	TotalPrice     float64 `json:"total_price"`      // 展示总价
	Unit           string  `json:"unit"`             // 单位
	SingleIspPrice float64 `json:"single_isp_price"` // 单IP价格
	DoubleIspPrice float64 `json:"double_isp_price"` // 双IP价格
	LocalIspPrice  float64 `json:"local_isp_price"`  // 本地IP价格
}

var confBalanceAutoRenewTable = "conf_balance_renew"

// GetConfBalanceRenewById 获取信息
func GetConfBalanceRenewById(id int) (data ConfBalanceRenewModel) {
	db.Table(confBalanceAutoRenewTable).Where("id = ?", id).First(&data)
	return
}

// GetConfBalanceRenewList 获取信息
func GetConfBalanceRenewList(code, cate string) (data []ConfBalanceRenewModel) {
	if code == "" {
		code = "renew"
	}
	db1 := db.Table(confBalanceAutoRenewTable).Where("code = ?", code)
	if cate != "" {
		db1 = db1.Where("cate = ?", cate)
	}
	db1.Order("sort desc,id asc").Where("status = ?", 1).Find(&data)
	return
}

// UserAutoRenewDetailModel 详细配置
type UserAutoRenewDetailModel struct {
	Id         int     `json:"id"`
	Uid        int     `json:"uid"`
	Username   string  `json:"username"`
	Email      string  `json:"email"`
	ConfId     int     `json:"conf_id"`    // 配置ID
	Cate       string  `json:"cate"`       // 类型
	Value      int64   `json:"value"`      // 续费的数量
	Balance    int64   `json:"balance"`    // 余额剩余
	ExpireDay  int     `json:"expire_day"` // 续费的有效期 比如流量 30天  静态有 7天 10天
	SyDay      int     `json:"sy_day"`     // 剩余几天有效期自动续费
	Money      float64 `json:"money"`      // 展示价格 ，如果配置了这个值就展示这个，没有配置，则展示基础配置的价格
	Method     string  `json:"method"`     // 续费方式 static  静态余额 balance 余额充值
	ExId       int     `json:"ex_id"`      // 提取IP信息ID
	ExIp       string  `json:"ex_ip"`      // 提取IP信息
	Status     int     `json:"status"`     // 状态 1正常 2禁用
	Ip         string  `json:"ip"`         // 最后操作IP
	UpdateTime int     `json:"update_time"`
	CreateTime int     `json:"create_time"`
}

type UserAutoRenewDetailJoinModel struct {
	Id          int     `json:"id"`
	Uid         int     `json:"uid"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	ConfId      int     `json:"conf_id"`      // 配置ID
	Cate        string  `json:"cate"`         // 类型
	Value       int64   `json:"value"`        // 续费的数量
	Balance     int64   `json:"balance"`      // 余额剩余
	ExpireDay   int     `json:"expire_day"`   // 续费的有效期 比如流量 30天  静态有 7天 10天
	SyDay       int     `json:"sy_day"`       // 剩余几天有效期自动续费
	Method      string  `json:"method"`       // 续费方式 static  静态余额 balance 余额充值
	ExId        int     `json:"ex_id"`        // 提取IP信息ID
	UserBalance int64   `json:"user_balance"` // 用户余额
	ExpireTime  int     `json:"expire_time"`  // 用户流量余额有效期
	RBalance    float64 `json:"r_balance"`    // 用户余额充值的余额
}

var userAutoRenewDetailTable = "cm_user_auto_renew_detail"

// GetUserAutoRenewDetailListJoin 获取详细列表信息
func GetUserAutoRenewDetailListJoin(uid int, cate string) (data []UserAutoRenewDetailJoinModel) {
	fields := "auto.*,us.balance as user_balance,b.balance as r_balance"
	joinTable := CmUserTable
	if cate == "flow" {
		fields = "auto.*,us.flows as user_balance,us.expire_time,b.balance as r_balance"
		joinTable = userFlowTable
	}
	if cate == "rotating" {
		fields = "auto.*,us.flows as user_balance,us.expire_time,b.balance as r_balance"
		joinTable = userDynamicIspTable
	}

	// cm_users 使用 id 字段关联，其他用户相关表使用 uid 字段关联
	joinKey := "id"
	if joinTable != CmUserTable {
		joinKey = "uid"
	}

	dbs := db.Table(userAutoRenewDetailTable + " as auto").
		Select(fields).
		Joins("left join cm_user_balance as b on auto.uid=b.uid").
		Joins("left join " + joinTable + " as us on auto.uid=us." + joinKey)

	if uid > 0 {
		dbs = dbs.Where("auto.uid =?", uid)
	}
	if cate != "" {
		dbs = dbs.Where("auto.cate =?", cate)
	}
	dbs = dbs.Where("auto.status=?", 1).
		Where("us.status=?", 1).
		Where("b.status=?", 1).
		Find(&data)
	return
}

// GetUserAutoRenewDetail 获取详细信息
func GetUserAutoRenewDetail(uid int, cate string, exId int) (data UserAutoRenewDetailModel) {
	dbs := db.Table(userAutoRenewDetailTable).Where("uid =?", uid).Where("cate =?", cate)
	if exId > 0 {
		dbs = dbs.Where("ex_id =?", exId)
	}
	dbs.First(&data)
	return
}

// GetUserAutoRenewDetailList 获取详细列表信息
func GetUserAutoRenewDetailList(uid int, cate string) (data []UserAutoRenewDetailModel) {
	dbs := db.Table(userAutoRenewDetailTable)
	if uid > 0 {
		dbs = dbs.Where("uid =?", uid)
	}
	if cate != "" {
		dbs = dbs.Where("cate =?", cate)
	}
	dbs.Find(&data)
	return
}

// AddUserAutoRenewDetail 添加用户设置信息
func AddUserAutoRenewDetail(info UserAutoRenewDetailModel) (err error) {
	err = db.Table(userAutoRenewDetailTable).Create(&info).Error
	return
}

// EditUserAutoRenewDetail 更新详细信息
func EditUserAutoRenewDetail(id int, param interface{}) (err error) {
	err = db.Table(userAutoRenewDetailTable).Where("id = ?", id).Update(param).Error
	return
}

type UserAutoRenewAlreadyModel struct {
	Cate string `json:"cate"`    // 类型
	Name string `json:"balance"` // 名称
}

type UserAutoRenewModel struct {
	Id             int    `json:"id"`
	Uid            int    `json:"uid"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Open           int    `json:"open"`            // 开关
	EmailSwitch    int    `json:"email_switch"`    // 邮件预警开关 0关闭 1开启
	IsEmail        int    `json:"is_email"`        // 是否是有效邮箱
	Method         string `json:"method"`          // 续费方式 static  静态余额 balance 余额充值
	Ip             string `json:"ip"`              // 操作IP
	UpdateTime     int    `json:"update_time"`     // 最后操作时间
	CreateTime     int    `json:"create_time"`     // 设置时间
	PauseIsp       int    `json:"pause_isp"`       //ISP代理暂停时间戳
	PauseStatic    int    `json:"pause_static"`    // 静态IP暂停时间戳
	PauseFlow      int    `json:"pause_flow"`      // 住宅代理暂停时间戳
	PauseRotating  int    `json:"pause_rotating"`  // 轮转ISP暂停时间戳
	PauseUnlimited int    `json:"pause_unlimited"` // 不限量代理暂停时间戳
}

var userAutoRenewTable = "cm_user_auto_renew"

// GetAutoRenewRecordByOpen 获取开启 未暂停的列表信息
func GetAutoRenewRecordByOpen() (data []UserAutoRenewModel) {
	db.Table(userAutoRenewTable).Where("open =? ", 1).Find(&data)
	return
}

// GetUserAutoRenewInfo 查询用户设置信息
func GetUserAutoRenewInfo(uid int) (data UserAutoRenewModel) {
	db.Table(userAutoRenewTable).Where("uid = ?", uid).First(&data)
	return
}

// EditUserAutoRenew 更新用户设置信息
func EditUserAutoRenew(uid int, param interface{}) (err error) {
	err = db.Table(userAutoRenewTable).Where("uid = ?", uid).Update(param).Error
	return
}

// AddUserAutoRenew 添加用户设置信息
func AddUserAutoRenew(info UserAutoRenewModel) (err error) {
	err = db.Table(userAutoRenewTable).Create(&info).Error
	return
}

// DealUserBalance 处理用户余额
// uid 用户Id
// value 增加的用户余额
// oldBalance 原来用户余额数
// tMoney 扣除的余额金额
// expireTime 设置用户过期时间
func DealUserBalance(uid int, value, oldBalance int64, cate string, tMoney float64, expireTime int) (err error) {
	// 开始扣费
	tx := db.Begin()

	//扣除余额
	editMap := map[string]interface{}{}
	editMap["balance"] = gorm.Expr("balance - ?", tMoney)
	editMap["all_buy"] = gorm.Expr("all_buy + ?", tMoney)
	err1 := tx.Table(userBalanceTable).Where("uid = ?", uid).Updates(editMap).Error
	if err1 != nil {
		tx.Rollback()
		return err1
	}
	if cate == "isp" {
		userIpMap := map[string]interface{}{}
		userIpMap["balance"] = gorm.Expr("balance + ?", value)
		userIpMap["pay_ip"] = gorm.Expr("pay_ip + ?", value)
		err2 := tx.Table(CmUserTable).Where("id = ?", uid).Updates(userIpMap).Error
		if err2 != nil {
			tx.Rollback()
			return err2
		}
	}
	if cate == "flow" {
		userIpMap := map[string]interface{}{}
		userIpMap["pre_flow"] = oldBalance
		userIpMap["expire_time"] = expireTime
		userIpMap["flows"] = gorm.Expr("flows + ?", value)
		userIpMap["all_flow"] = gorm.Expr("all_flow + ?", value)
		err2 := tx.Table(userFlowTable).Where("uid = ?", uid).Updates(userIpMap).Error
		if err2 != nil {
			tx.Rollback()
			return err2
		}
	}
	if cate == "rotating" {
		userIpMap := map[string]interface{}{}
		userIpMap["pre_flow"] = oldBalance
		userIpMap["expire_time"] = expireTime
		userIpMap["flows"] = gorm.Expr("flows + ?", value)
		userIpMap["all_flow"] = gorm.Expr("all_flow + ?", value)
		err2 := tx.Table(userDynamicIspTable).Where("uid = ?", uid).Updates(userIpMap).Error
		if err2 != nil {
			tx.Rollback()
			return err2
		}
	}
	tx.Commit()
	return err1
}
