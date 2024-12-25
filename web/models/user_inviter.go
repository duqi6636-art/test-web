package models

type UserInviter struct {
	ID              int     `json:"id"`
	Uid             int     `json:"uid"`
	Username        string  `json:"username"`         //用户昵称
	Email           string  `json:"email"`            //用户邮箱
	Ip              string  `json:"ip"`               //注册IP
	RegCountry      string  `json:"reg_country"`      //注册国家
	InviterCode     string  `json:"inviter_code"`     //用户唯一邀请码
	RegTime         int     `json:"reg_time"`         //
	InviterId       int     `json:"inviter_id"`       //上级邀请ID
	InviterUsername string  `json:"inviter_username"` //上级邀请码
	Origin          int     `json:"origin"`           //推广类型 1链接  2填写推广码
	PayNum          int     `json:"pay_num"`          //交易单数
	PayMoney        float64 `json:"pay_money"`        //交易金额
	Level           int     `json:"level"`            //佣金级别
	Ratio           float64 `json:"ratio"`            //佣金比例
	IsInner         int     `json:"is_inner"`         //内部设置 1内部特殊设置
	CreateTime      int     `json:"create_time"`      //
}

type ResUserInviter struct {
	Uid        int     `json:"uid"`
	Username   string  `json:"username"` //用户昵称
	Number     int     `json:"number"`   //交易笔数
	RegTime    string  `json:"reg_time"` //
	OrderMoney float64 `json:"order_money"`
	Money      float64 `json:"money"`
}

var userInviterTable = "cm_user_inviter"

func GetUserInviterByMap(where map[string]interface{}) (err error, data UserInviter) {
	err = db.Table(userInviterTable).Where(where).Order("id desc", true).First(&data).Error
	return
}
func GetUserInviterByUid(uid int) (err error, data UserInviter) {
	err = db.Table(userInviterTable).Where("uid=?", uid).Order("id desc", true).First(&data).Error
	return
}

// 邀请信息
func GetUserInviterByCode(code string) (err error, data UserInviter) {
	err = db.Table(userInviterTable).Where("inviter_code=?", code).Order("id desc", true).First(&data).Error
	return
}
func GetInviterListByMap(where map[string]interface{}) (err error, data []UserInviter) {
	err = db.Table(userInviterTable).Where(where).Find(&data).Error
	return
}

func CreateUserInviter(data UserInviter) (err error) {
	err = db.Table(userInviterTable).Create(&data).Error
	return err
}

func EditUserInviter(uid int, data UserInviter) (err error) {
	err = db.Table(userInviterTable).Where("uid = ?", uid).Update(data).Error
	return err
}

// 获取分页列表
func GetInviterListPage(where map[string]interface{}, offset, limit int, sort string) (data []UserInviter) {
	db.Table(userInviterTable).Where(where).Offset(offset).Limit(limit).Order(sort).Find(&data)
	return
}

// 获取列表
func GetInviterListBy(where map[string]interface{}, sort string) (data []UserInviter) {
	db.Table(userInviterTable).Where(where).Order(sort, true).Find(&data)
	return
}

// 佣金等级
type ConfLevel struct {
	Id    int     `json:"id"`
	Name  string  `json:"name"`  //等级名称
	Ratio float64 `json:"ratio"` //佣金比例
	Max   float64 `json:"max"`   //最大金额
	Cate  string  `json:"cate"`  // 类型
}

// 获取信息
func GetConfLevelByMoney(money float64, sort string) (data ConfLevel) {
	dbs := db.Table("conf_level")
	if money == 0 {
		dbs = dbs.Where("id = ?", 1)
	} else {
		dbs = dbs.Where("max > ?", money)
	}
	dbs.Order(sort, true).First(&data)
	return
}

// 获取信息
func GetConfLevelBy(where interface{}, sort string) (data ConfLevel) {
	db.Table("conf_level").Where(where).Order(sort, true).First(&data)
	return
}

// 获取信息
func GetExchangeRatio(cate string) (ratio float64) {
	info := ConfLevel{}
	db.Table("conf_level").Where("cate = ?", cate).First(&info)
	ratio = info.Ratio
	return
}
