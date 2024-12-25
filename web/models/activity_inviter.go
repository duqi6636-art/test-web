package models

type UserActivityInviter struct {
	Id              int     `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL" json:"id"`
	Uid             int     `gorm:"column:uid;default:NULL;comment:'用户ID'" json:"uid"`
	Username        string  `gorm:"column:username;default:;comment:'用户昵称'" json:"username"`
	Email           string  `gorm:"column:email;default:NULL;comment:'邮箱'" json:"email"`
	Ip              string  `gorm:"column:ip;default:NULL;comment:'ip'" json:"ip"`
	RegCountry      string  `gorm:"column:reg_country;default:NULL;comment:'注册时的国家'" json:"reg_country"`
	InviterCode     string  `gorm:"column:inviter_code;default:;comment:'邀请码'" json:"inviter_code"`
	RegTime         int     `gorm:"column:reg_time;default:NULL;comment:'用户注册时间'" json:"reg_time"`
	InviterId       int     `gorm:"column:inviter_id;default:NULL;comment:'用户填写的邀请码对应uid(上级)'" json:"inviter_id"`
	InviterUsername string  `gorm:"column:inviter_username;default:;comment:'用户填写的邀请码(上级)'" json:"inviter_username"`
	Origin          int     `gorm:"column:origin;default:0;comment:'推广类型 1链接  2填写推广码'" json:"origin"`
	PayNum          int     `gorm:"column:pay_num;default:0;comment:'交易单数'" json:"pay_num"`
	PayMoney        float64 `gorm:"column:pay_money;default:0.00;comment:'邀请金额'" json:"pay_money"`
	Level           int     `gorm:"column:level;default:1;comment:'级别'" json:"level"`
	Ratio           string  `gorm:"column:ratio;default:0.02;comment:'佣金比例'" json:"ratio"`
	IsInner         int     `gorm:"column:is_inner;default:0;comment:'内部设置 1内部特殊设置'" json:"is_inner"`
	CreateTime      int     `gorm:"column:create_time;default:0;comment:'创建时间'" json:"create_time"`
	Remarks         string  `gorm:"column:remarks;default:NULL;comment:'备注'" json:"remarks"`
	TotalMoney      float64 `gorm:"column:total_money;default:0.00;comment:'总佣金'" json:"total_money"`
	UsedMoney       float64 `gorm:"column:used_money;default:0.00;comment:'已使用佣金'" json:"used_money"`
}

var userActivityInviterTable = "cm_user_activity_inviter"

func GetUserActivityInviterByMap(where map[string]interface{}) (err error, data UserActivityInviter) {
	err = db.Table(userActivityInviterTable).Where(where).Order("id desc", true).First(&data).Error
	return
}

func GetActivityInviterListByMap(where map[string]interface{}) (err error, data []UserActivityInviter) {
	err = db.Table(userActivityInviterTable).Where(where).Find(&data).Error
	return
}

func GetActivityInviterInfoByUid(uid int) (err error, data UserActivityInviter) {
	err = db.Table(userActivityInviterTable).Where("uid=?", uid).Order("id desc", true).First(&data).Error
	return
}

func CreateUserActivityInviter(data UserInviter) (err error) {
	err = db.Table(userActivityInviterTable).Create(&data).Error
	return err
}

// 获取列表
func GetActivityInviterListBy(where map[string]interface{}, sort string) (data []UserInviter) {
	db.Table(userActivityInviterTable).Where(where).Order(sort, true).Find(&data)
	return
}

// 更新信息
func UpdateActivityUserById(uid int, user interface{}) bool {
	db.Table(userActivityInviterTable).Where("uid = ?", uid).Updates(user)
	return true
}
