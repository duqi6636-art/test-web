package models

type UserActivityWithdrawal struct {
	Id         int     `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL" json:"id"`
	Uid        int     `gorm:"column:uid;default:0;comment:'用户id'" json:"uid"`
	Username   string  `gorm:"column:username;default:;comment:'用户名'" json:"username"`
	Email      string  `gorm:"column:email;default:;comment:'联系邮箱-用户提现填写的'" json:"email"`
	Money      float64 `gorm:"column:money;default:0.00;comment:'提现金额'" json:"money"`
	TrueMoney  float64 `gorm:"column:true_money;default:0.00;comment:'真实到账金额，扣除手续费后的'" json:"true_money"`
	Wallet     string  `gorm:"column:wallet;default:;NOT NULL;comment:'提现钱包'" json:"wallet"`
	OrderNo    string  `gorm:"column:order_no;default:;comment:'交易编号'" json:"order_no"`
	Ip         string  `gorm:"column:ip;default:;NOT NULL;comment:'申请IP'" json:"ip"`
	CreateTime int     `gorm:"column:create_time;default:0;comment:'申请时间'" json:"create_time"`
	Status     int     `gorm:"column:status;default:1;comment:'状态 1待审核 2已打款 3已拒绝'" json:"status"`
	DealTime   int     `gorm:"column:deal_time;default:0;comment:'处理时间'" json:"deal_time"`
	Remark     string  `gorm:"column:remark;default:;comment:'处理信息'" json:"remark"`
	DealAdmin  string  `gorm:"column:deal_admin;default:;comment:'处理人'" json:"deal_admin"`
}

func (c *UserActivityWithdrawal) TableName() string {
	return "cm_user_activity_withdrawal"
}

// 添加提现记录
func AddActivityWithdrawal(model UserActivityWithdrawal) (err error) {
	err = db.Model(&UserActivityWithdrawal{}).Create(&model).Error
	return
}

// 获取列表
func GetActivityWithdrawalListBy(uid int) (data []UserActivityWithdrawal) {
	db.Model(&UserActivityWithdrawal{}).
		Where("uid=?", uid).
		Order("id desc", true).Find(&data)
	return
}

// 获取分页列表
func GetActivityWithdrawalPageBy(uid int, offset, limit int) (data []UserActivityWithdrawal) {
	dbs := db.Model(&UserActivityWithdrawal{})
	if uid > 0 {
		dbs = dbs.Where("uid=?", uid)
	}
	dbs.Offset(offset).Limit(limit).Order("id desc", true).Find(&data)
	return
}

// 获取提现记录总数
func GetActivityWithdrawalCount(uid int) (total int) {
	db.Model(&UserActivityWithdrawal{}).Where("uid=?", uid).Count(&total)
	return
}

// 获取分页列表
func GetActivityWithdrawalById(id int) (data UserWithdrawalModel) {
	db.Model(&UserActivityWithdrawal{}).
		Where("id=?", id).
		Order("id desc", true).
		First(&data)
	return
}
