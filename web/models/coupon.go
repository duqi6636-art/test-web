package models

import (
	"api-360proxy/web/pkg/util"
	"time"
)

type Coupon struct {
	Id         int     `json:"id"`
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Title      string  `json:"title"`
	Type       string  `json:"type"`                           // "1" => "减价格", "2" => "加（IP/流量）", "3" => "折扣-减价格", "4" => "折扣-加（IP/流量）"
	Value      float64 `json:"value" gorm:"value"`             // 优惠金额/折扣比例/增加余额 折扣或折现值
	Number     int     `json:"number" gorm:"number"`           // 发放数
	Cate       string  `json:"cate" gorm:"cate"`               // 类型
	UserType   string  `json:"user_type" gorm:"user_type"`     // 用户类型 payed 已付费 no_pay 未付费 agent 代理商
	Status     int     `json:"status" gorm:"status"`           // 状态:1生效 2过期
	UseType    int     `json:"use_type" gorm:"use_type"`       // 使用类型 1单次使用(单个券应单个用户使用)  2重复使用(单个券对应多个用户)
	Meals      string  `json:"meals" gorm:"meals"`             // 绑定套餐id
	ExpiryDay  int     `json:"expiry_day" gorm:"expiry_day"`   // 过期时间(领取的时间戳+ 这个天数)
	UseCycle   int     `json:"use_cycle" gorm:"use_cycle"`     // 可用周期 1终身  30一月 90一季 180 半年 360一年
	UseNumber  int     `json:"use_number" gorm:"use_number"`   // 可用次数
	Platform   int     `json:"platform" gorm:"platform"`       // 平台ID
	GroupId    int     `json:"group_id" gorm:"group_id"`       // 分组ID
	Remark     string  `json:"remark" gorm:"remark"`           // 备注
	Admin      string  `json:"admin" gorm:"admin"`             // 创建人
	CreateTime int     `json:"create_time" gorm:"create_time"` // 创建时间
	UsedNum    int     `json:"used_num" gorm:"used_num"`       // 使用数量
	LastTime   int     `json:"last_time" gorm:"last_time"`     // 上次统计时间
	Cron       string  `json:"cron" gorm:"cron"`               // 角标
}
type CouponList struct {
	Id           int     `json:"id"`
	Cid          int     `json:"cid"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	BindUid      int     `json:"bind_uid"`
	BindUsername string  `json:"bind_username"`
	Status       int     `json:"status"`   // 状态:1生效 2已使用
	UseTime      int     `json:"use_time"` // 使用时间
	Title        string  `json:"title"`
	Type         string  `json:"type"`       // "1" => "减价格", "2" => "加（IP/流量）", "3" => "折扣-减价格", "4" => "折扣-加（IP/流量）"
	Value        float64 `json:"value"`      // type!=money时 使用
	UserType     string  `json:"user_type"`  // "all" => "所有", "payed" => "已付费", "no_pay" => "未付费", "agent" => "代理商"
	UseType      int     `json:"use_type"`   // 使用类型 1单次使用(单个券应单个用户使用)  2重复使用(单个券对应多个用户)
	Meals        string  `json:"meals"`      //
	Expire       int     `json:"expire"`     // 过期时间
	ExpiryDay    int     `json:"expiry_day"` // 过期天数  0永不过期
	UseCycle     int     `json:"use_cycle"`
	UseNumber    int     `json:"use_number"`
	Platform     int     `json:"platform"` //平台ID
	GroupId      int     `json:"group_id"` //分组ID
	CreateTime   int     `json:"create_time"`
	Cron         string  `json:"cron"`      // 角标
	Cate         string  `json:"cate"`      // 类型
	Condition    string  `json:"condition"` // 标题文案描述
	PayType      string  `json:"pay_type"`  // 支付类型
}

type ResCouponList struct {
	Id         int    `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Status     int    `json:"status"`   // 状态:1生效 2已使用
	UseTime    string `json:"use_time"` // 使用时间
	Title      string `json:"title"`
	Type       string `json:"type"`        // money优惠券 ip赠送IP
	Value      string `json:"value"`       // 返回值信息 根据不同的类型判断返回
	Meals      string `json:"meals"`       //
	PayType    string `json:"pay_type"`    // 支付类型
	ExpireTime int    `json:"expire_time"` // 过期时间
	Expire     string `json:"expire"`      // 过期时间
	CreateTime string `json:"create_time"` // 获得时间
	PakId      string `json:"pak_id"`      // 加个最大的套餐id
	PakType    string `json:"pak_type"`    // 套餐类型
	Cron       string `json:"cron"`        // 角标
	Unit       string `json:"unit"`        // 单位
}

type ResCouponInfo struct {
	Id         int      `json:"id"`
	Code       string   `json:"code"`
	Name       string   `json:"name"`
	Status     int      `json:"status"`   // 状态:1生效 2已使用
	UseTime    string   `json:"use_time"` // 使用时间
	Title      string   `json:"title"`
	Type       string   `json:"type"`        // money优惠券 ip赠送IP
	Value      string   `json:"value"`       // 返回值信息 根据不同的类型判断返回
	MealArr    []string `json:"meal_arr"`    //套餐列表
	Expire     string   `json:"expire"`      // 过期时间
	CreateTime string   `json:"create_time"` // 获得时间
	CouponType string   `json:"coupon_type"` // 优惠券类型
}

var couponListTable = "cm_coupon_list"

// 发送券
func AddCoupon(info CouponList) (err error) {
	err = db.Table(couponListTable).Create(&info).Error
	return
}

// 更新信息
func EditCouponBind(code string, uid int, data map[string]interface{}) bool {
	err := db.Table(couponListTable).Where("code = ?", code).Where("status=?", 1).Limit(1).Updates(data).Error
	if err != nil {
		return false
	}
	return true
}

// 获取 券
func GetCoupon(code string) (err error, info CouponList) {
	err = dbRead.Table(couponListTable).Where("code = ? and status = ?", code, 1).Where("bind_uid=?", 0).First(&info).Error
	return
}

// 获取 券
func GetCouponList(uid int, types string) (err error, info []CouponList) {
	dbt := db.Table(couponListTable).Where("bind_uid = ?", uid)
	if types != "" {
		dbt = dbt.Where("`type` = ?", types)
	}
	err = dbt.Find(&info).Error
	return
}

// 获取优惠券
func GetCouponById(id int) (err error, coupon Coupon) {
	dbt := db.Table("cm_coupon").Where("id = ?", id).Where("status = ?", 1)
	err = dbt.First(&coupon).Error
	return
}

// 获取信息
func GetCouponByCode(code string) (err error, coupon Coupon) {
	dbt := db.Table("cm_coupon").Where("code = ?", code).Where("status = ?", 1)
	err = dbt.First(&coupon).Error
	return
}

// 获取根据用户类型获取可用优惠券
func GetCouponListByUserType(uid int, userType string) (err error, num int) {

	nowTime := util.GetNowInt()
	db := db.Table(couponListTable).Where("bind_uid = ?", uid).Where("status=?", 1).Where("expire > ?", nowTime)
	if userType != "" {
		db = db.Where("`user_type` = ?", userType)
	}
	err = db.Count(&num).Error
	return
}

// 获取 自动发放的券
func GetCouponListByCate(uid int, cate string) (info []CouponList) {
	dbs := db.Table(couponListTable).Where("cate=?", cate).Where("bind_uid = ?", uid)
	if uid == 0 {
		dbs = dbs.Where("status=?", 1)
	}
	dbs.Find(&info)
	return
}

// 获取 券
func GetCardList(uid int, types string) (err error, info []CouponList) {
	nowTime := int(time.Now().Unix())
	dbt := dbRead.Table(couponListTable).Where("bind_uid = ?", uid)
	if types != "" {
		dbt = dbt.Where("`type` = ?", types)
	}
	err = dbt.Where("create_time >= ? and create_time <= ?", nowTime-365*86400, nowTime).Order("create_time desc", true).Find(&info).Error
	return
}

// 获取 根据cid获取 cdk
func GetCouponByUsePlatform(uid, useCycle int, platform int) (err error, info []CouponList) {
	nowTime := int(time.Now().Unix())
	dbt := dbRead.Table(couponListTable).Where("bind_uid = ?", uid)

	if platform > 0 {
		dbt = dbt.Where("platform = ?", platform)
	}
	if useCycle >= 1 {
		start := nowTime - (useCycle * 86400)
		dbt = dbt.Where("create_time >= ?", start)
	}
	err = dbt.Where("status >= ?", 1).Find(&info).Error
	return
}

// 获取 当前用户时间段已使用的cdk类型数据
func GetCdkByGroupId(uid, group_id int) (info CouponList, err error) {
	err = dbRead.Table(couponListTable).Where("bind_uid = ?", uid).Where("group_id = ?", group_id).Where("status >= ?", 1).First(&info).Error
	return
}

// 获取 券
func GetCouponListByPakId(uid, pak_id int, types string) (err error, info []CouponList) {
	dbt := dbRead.Table(couponListTable).Where("bind_uid = ? and status = ? and expire > ?", uid, 1, int(time.Now().Unix()))
	if types != "" {
		dbt = dbt.Where("`type` = ?", types)
	}
	if pak_id > 0 {
		dbt = dbt.Where("FIND_IN_SET(? ,meals)", pak_id)
	}
	err = dbt.Find(&info).Error
	return
}

// 获取 这个用户未使用的券
func GetAvailableCouponListByUid(uid int) (err error, info []CouponList) {

	dbt := dbRead.Table(couponListTable).Where("bind_uid = ? and status = ? and expire > ?", uid, 1, int(time.Now().Unix()))
	err = dbt.Find(&info).Error

	return
}

// 更新信息
func EditCouponByCode(code string, data map[string]interface{}) bool {
	err := db.Table(couponListTable).Where("code = ?", code).Updates(data).Error
	if err != nil {
		return false
	}
	return true
}

// 获取 根据cid获取 cdk
func GetCouponByUse(cid, uid, useCycle int) (err error, info []CouponList) {
	nowTime := int(time.Now().Unix())
	dbt := db.Table(couponListTable).Where("cid = ?", cid).Where("bind_uid = ?", uid)
	if useCycle > 1 {
		start := nowTime - (useCycle * 86400)
		dbt = dbt.Where("create_time >= ?", start)
	}
	err = dbt.Where("status >= ?", 1).Find(&info).Error
	return
}

// 获取 根据cid获取 cdk
func GetCouponByCid(uid, cid int) (info CouponList) {
	dbt := db.Table(couponListTable).Where("bind_uid = ?", uid).Where("cid = ?", cid)
	dbt.Where("status = ?", 1).First(&info)
	return
}

// 获取 根据cid获取 cdk
func GetCouponByCidCount(uid, cid int) (info []CouponList) {
	dbt := db.Table(couponListTable).Where("bind_uid = ?", uid).Where("cid = ?", cid)
	dbt.Where("status = ?", 1).Find(&info)
	return
}

// 优惠券点击上报
type CouponPopupClick struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	Code       string `json:"code"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	CreateTime int    `json:"create_time"`
	Ip         string `json:"ip"`
}

func GetClickLog(uid int) (info CouponPopupClick) {
	db.Table("coupon_click_log").Where("uid = ?", uid).Where("create_time = ?", util.GetTodayTime()).First(&info)
	return
}

func AddPopupClickLog(uid int, username, email, code, userIp string, nowTime int) (err error) {
	log := CouponPopupClick{
		Uid:        uid,
		Username:   username,
		Email:      email,
		Code:       code,
		Ip:         userIp,
		CreateTime: nowTime,
	}

	var tableName = "coupon_click_log"
	if !db.HasTable(tableName) {
		createClickLogLogTable(tableName)
	}
	err = db.Table(tableName).Create(&log).Error
	return
}

// 创建表
func createClickLogLogTable(tableName string) {
	createTables := `CREATE TABLE ` + tableName + `(
		id int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
		uid int(11) NOT NULL COMMENT '用户ID',
		username varchar(30) DEFAULT '' COMMENT '用户名',
		email varchar(60) DEFAULT '' COMMENT '用户名',
		code varchar(50) DEFAULT '' COMMENT '标识码',
		ip varchar(50) DEFAULT '' COMMENT '用户IP',
		create_time int(11) DEFAULT '0' COMMENT '操作时间',
		PRIMARY KEY (id),
		KEY uid (uid) USING BTREE,
		KEY code (code) USING BTREE,
		KEY create_time (create_time) USING BTREE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='优惠券弹窗点击记录';`
	db.Exec(createTables)
}
