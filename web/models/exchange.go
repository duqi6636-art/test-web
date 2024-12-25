package models

import "time"

type ExchangeList struct {
	Id           int    `json:"id"`
	Cid          int    `json:"cid"`
	Cate         int    `json:"cate"`
	Uid          int    `json:"uid"`
	Mode         string `json:"mode"` //使用模式 agent 代理商  invite 邀请  score积分 balance 余额
	Code         string `json:"code"`
	Name         string `json:"name"`
	BindUid      int    `json:"bind_uid"`
	BindUsername string `json:"bind_username"`
	Status       int    `json:"status"`   // 状态:1生效 2已使用
	UseTime      int    `json:"use_time"` // 使用时间
	Title        string `json:"title"`
	Region       string `json:"region"`     // 静态兑换地区
	Day          int    `json:"day"`        // 静态兑换套餐天数
	Value        int64  `json:"value"`      //
	UserType     string `json:"user_type"`  // 用户类型 payed 已付费 no_pay 未付费
	UseType      int    `json:"use_type"`   // 使用类型 1单次使用(单个券应单个用户使用)  2重复使用(单个券对应多个用户)
	Expire       int    `json:"expire"`     // 过期时间
	ExpiryDay    int    `json:"expiry_day"` // 过期天数  0永不过期
	UseCycle     int    `json:"use_cycle"`
	UseNumber    int    `json:"use_number"`
	Platform     int    `json:"platform"` //平台ID
	GroupId      int    `json:"group_id"` //分组ID
	CreateTime   int    `json:"create_time"`
}

type ResExchangeList struct {
	Id         int    `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Status     int    `json:"status"`   // 状态:1生效 2已使用
	UseTime    string `json:"use_time"` // 使用时间
	Title      string `json:"title"`
	Type       string `json:"type"`        // money优惠券 ip赠送IP
	Condition  string `json:"condition"`   //
	Value      string `json:"value"`       // 返回值信息 根据不同的类型判断返回
	Meals      string `json:"meals"`       //
	Expire     string `json:"expire"`      // 过期时间
	CreateTime string `json:"create_time"` // 获得时间
}

var exchangeListTable = "cm_exchange_list"

// 发送券
func AddExchange(info ExchangeList) (err error) {
	err = db.Table(exchangeListTable).Create(&info).Error
	return
}

// 获取 券信息
func GetExchangeInfo(code string) (err error, info ExchangeList) {
	//err = dbRead.Table(exchangeListTable).Where("code = ?", code).Where("bind_uid=?", 0).First(&info).Error
	err = dbRead.Table(exchangeListTable).Where("code = ?", code).First(&info).Error
	return
}

// 获取 券信息
func GetExchangeInfoWithId(id string) (err error, info ExchangeList) {

	err = db.Table(exchangeListTable).Where("id = ?", id).First(&info).Error
	return
}

// 更新信息
func EditExchangeById(id int, data map[string]interface{}) bool {
	err := db.Table(exchangeListTable).Where("id = ?", id).Updates(data).Error
	if err != nil {
		return false
	}
	return true
}

// 更新信息
func EditExchangeByCode(code string, data map[string]interface{}) bool {
	err := db.Table(exchangeListTable).Where("code = ?", code).Updates(data).Error
	if err != nil {
		return false
	}
	return true
}

// 获取 根据cid获取 cdk
func GetExchangeByUsePlatform(uid, useCycle int, platform int) (err error, info []ExchangeList) {
	nowTime := int(time.Now().Unix())
	dbt := dbRead.Table(exchangeListTable).Where("bind_uid = ?", uid).Where("cate=?", 1)

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
func GetExchangeByGroupId(uid, group_id int) (info ExchangeList, err error) {
	err = dbRead.Table(exchangeListTable).Where("bind_uid = ?", uid).Where("cate=?", 1).Where("group_id = ?", group_id).Where("status >= ?", 1).First(&info).Error
	return
}

type ResExchange struct {
	Code         string `json:"code"`
	Value        int64  `json:"value"`
	BindUsername string `json:"bind_username"`
	BindUid      int    `json:"bind_uid"`
	Status       int    `json:"status"`
	CreateTime   string `json:"create_time"`
	BindTime     string `json:"bind_time"`
}

// 获取 用户生成的券
func GetExchangeByCate(uid, cate, offset, limit int) (info []ExchangeList, err error) {
	err = dbRead.Table(exchangeListTable).Where("uid = ?", uid).Where("cate=?", cate).Where("status >=?", 0).Offset(offset).Limit(limit).Order("id desc", true).Find(&info).Error
	return
}

func GetExchangeListByCate(uid, cate int) (info []ExchangeList, err error) {
	err = dbRead.Table(exchangeListTable).Where("uid = ?", uid).Where("cate=?", cate).Where("status >=?", 0).Order("id desc", true).Find(&info).Error
	return
}
