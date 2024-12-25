package models

type UserActivityMoneyLog struct {
	Id         int     `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL" json:"id"`
	Uid        int     `gorm:"column:uid;default:0;comment:'用户id'" json:"uid"`
	Ip         int64   `gorm:"column:ip;default:0;comment:'当前套餐值'" json:"ip"`
	Money      float64 `gorm:"column:money;default:0.00;comment:'套餐付款金额'" json:"money"`
	Ratio      float64 `gorm:"column:ratio;default:0.00;comment:'当前返佣比例'" json:"ratio"`
	Code       int     `gorm:"column:code;default:0;NOT NULL;comment:'标识 1自用购买  2兑换cdk  3提现 10邀请用户购买 11: 用户首次购买 12：兑换ip  13：兑换流量'" json:"code"`
	Mark       int     `gorm:"column:mark;default:0;NOT NULL;comment:'符号标识 1增加 -1减少'" json:"mark"`
	Cdkey      string  `gorm:"column:cdkey;default:;NOT NULL;comment:'cdkey'" json:"cdkey"`
	OrderId    string  `gorm:"column:order_id;default:;comment:'邀请购买的关联订单'" json:"order_id"`
	PayPlat    int     `gorm:"column:pay_plat;default:0;comment:'支付方式'" json:"pay_plat"`
	InviterId  int     `gorm:"column:inviter_id;default:0;comment:'上级邀请人ID'" json:"inviter_id"`
	Status     int     `gorm:"column:status;default:1;comment:'状态 1 正常   2冻结中'" json:"status"`
	DealTime   int     `gorm:"column:deal_time;default:0;comment:'处理时间'" json:"deal_time"`
	Remark     string  `gorm:"column:remark;default:;comment:'处理信息'" json:"remark"`
	CreateTime int     `gorm:"column:create_time;default:0;comment:'操作时间'" json:"create_time"`
	Today      int     `gorm:"column:today;default:0;comment:'当日凌晨时间戳'" json:"today"`
	IsFirst    int     `gorm:"column:is_first;default:0;comment:'是否 是首单 1 是'" json:"is_first"`
}

func (c *UserActivityMoneyLog) TableName() string {
	return "cm_user_activity_money_log"
}

func CreateUserActivityMoneyLog(data UserActivityMoneyLog) (err error) {
	err = db.Model(&UserActivityMoneyLog{}).Create(&data).Error
	return err
}

func GetActivityMoneyListByMap(where map[string]interface{}) (err error, data []UserActivityMoneyLog) {
	err = db.Model(&UserActivityMoneyLog{}).Where(where).Find(&data).Error
	return
}

// 获取分页列表
func GetActivityMoneyListPage(where map[string]interface{}, offset, limit int, sort string) (data []UserActivityMoneyLog) {
	db.Model(&UserActivityMoneyLog{}).Where(where).Offset(offset).Limit(limit).Order(sort, true).Find(&data)
	return
}

// 获取列表
func GetActivityMoneyListBy(where map[string]interface{}, sort string) (data []UserActivityMoneyLog) {
	db.Model(&UserActivityMoneyLog{}).Where(where).Order(sort, true).Find(&data)
	return
}

func GetActivityMoneyList(inviter_id int) (err error, data []UserActivityMoneyLog) {
	err = db.Model(&UserActivityMoneyLog{}).Where("inviter_id = ? and code = ? and mark = ?", inviter_id, 10, 1).Find(&data).Error
	return
}

func GetActivityExchangeListByUid(uid int) (err error, data []UserActivityMoneyLog) {
	err = db.Model(&UserActivityMoneyLog{}).Where("uid = ? and mark = ? and code >= ?", uid, -1, 12).Find(&data).Error
	return
}

var userActivityMoneyLogTable = "cm_user_activity_money_log"

// 获取分页列表
func GetActivityMoneyListByInvitePage(invite_id, offset, limit int, sort string) (data []UserOrderLog) {
	db.Table(userActivityMoneyLogTable+" as m").Select("m.*,i.username").Joins("left join "+userActivityInviterTable+" as i on m.uid=i.uid").Where("m.inviter_id=?", invite_id).Where("m.code =?", "10").Where("m.status < ?", 3).Where("m.order_id <>?", "").Offset(offset).Limit(limit).Order(sort, true).Find(&data)
	return
}

// 获取分页列表
func GetActivityMoneyListByInvitePageV1(invite_id, offset, limit int, sort string) (data []UserOrderLogV1) {
	db.Table(userActivityMoneyLogTable+" as m").Select("m.*,i.email,i.reg_time").Joins("left join "+userActivityInviterTable+" as i on m.uid=i.uid").Where("m.inviter_id=?", invite_id).Where("m.code =?", "10").Where("m.status = ?", 1).Where("m.order_id <>?", "").Offset(offset).Limit(limit).Order(sort, true).Find(&data)
	return
}

// 获取列表
func GetActivityMoneyListByInvite(invite_id int, sort string) (data []UserOrderLog) {
	db.Table(userActivityMoneyLogTable).Where("inviter_id=?", invite_id).Where("code =?", "10").Where("order_id <>?", "").Order(sort, true).Find(&data)
	return
}

// 查询用户佣金兑换记录
func GetActivityUserExchangeList(uid, offset, limit int) (data []UserOrderLogV1) {
	db.Table(userActivityMoneyLogTable).Where("uid = ?", uid).Where("mark = ?", -1).Where("code >= ?", 12).Order("id desc", true).Find(&data)
	return
}

// 查询用户佣金兑换记录
func GetActivityUserExchangeCount(uid int) (total int) {
	db.Table(userActivityMoneyLogTable).Where("uid = ?", uid).Where("mark = ?", -1).Where("code >= ?", 12).Where("status = 1").Count(&total)
	return
}
