package models

type UserMoneyLog struct {
	ID         int     `json:"id"`
	Uid        int     `json:"uid"`         //购买人的用户ID
	Ip         int64   `json:"ip"`          //IP数量
	Code       int     `json:"code"`        //标识 1自用购买  2兑换cdk 10邀请用户购买
	Money      float64 `json:"money"`       //
	Ratio      float64 `json:"ratio"`       //返佣比例 或者是单价
	Mark       int     `json:"mark"`        //符号标识 1增加 -1减少
	Cdkey      string  `json:"cdkey"`       // 生成cdk的时候用
	OrderId    string  `json:"order_id"`    // 邀请购买的关联订单
	InviterId  int     `json:"inviter_id"`  //上级邀请ID
	Status     int     `json:"status"`      //状态 1 正常   2冻结中
	CreateTime int     `json:"create_time"` //
	Today      int     `json:"today"`       //
}

type ResUserMoneyLog struct {
	Uid        int     `json:"uid"`         //购买人的用户ID
	Code       int     `json:"code"`        //标识 1自用转化  2     3、提现
	Flow       int     `json:"flow"`        // 兑换的流量
	Money      float64 `json:"money"`       //
	Cdkey      string  `json:"cdkey"`       // 生成cdk的时候用
	CreateTime string  `json:"create_time"` //
	Status     int     `json:"status"`      //状态 1 已经到账   2处理中
}

var userMoneyLogTable = "cm_user_money_log"

func CreateUserMoneyLog(data UserMoneyLog) (err error) {
	err = db.Table(userMoneyLogTable).Create(&data).Error
	return err
}

func GetMoneyListByMap(where map[string]interface{}) (err error, data []UserMoneyLog) {
	err = db.Table(userMoneyLogTable).Where(where).Find(&data).Error
	return
}

// 获取分页列表
func GetMoneyListPage(where map[string]interface{}, offset, limit int, sort string) (data []UserMoneyLog) {
	db.Table(userMoneyLogTable).Where(where).Offset(offset).Limit(limit).Order(sort, true).Find(&data)
	return
}

// 获取列表
func GetMoneyListBy(where map[string]interface{}, sort string) (data []UserMoneyLog) {
	db.Table(userMoneyLogTable).Where(where).Order(sort, true).Find(&data)
	return
}

func GetMoneyList(inviter_id int) (err error, data []UserMoneyLog) {
	err = db.Table(userMoneyLogTable).Where("inviter_id = ? and code = ? and mark = ?", inviter_id, 10, 1).Find(&data).Error
	return
}
func GetKeyUserMoneyListByUid(inviter_id int) (err error, data []UserMoneyLog) {
	err = db.Table(userMoneyLogTable).Where("uid = ? and inviter_id = ? and code = ? and mark = ? and status = ?", inviter_id, inviter_id, 14, 1, 1).Find(&data).Error
	return
}

func GetExchangeListByUid(uid int) (err error, data []UserMoneyLog) {
	err = db.Table(userMoneyLogTable).Where("uid = ? and mark = ? and code >= ?", uid, -1, 12).Find(&data).Error
	return
}

type UserOrderLog struct {
	ID         int     `json:"id"`
	Uid        int     `json:"uid"`         //购买人的用户ID
	Ip         int     `json:"ip"`          //IP数量
	Code       int     `json:"code"`        //标识 1自用购买  2兑换cdk 10邀请用户购买
	Money      float64 `json:"money"`       //
	Ratio      float64 `json:"ratio"`       //返佣比例 或者是单价
	Mark       int     `json:"mark"`        //符号标识 1增加 -1减少
	Cdkey      string  `json:"cdkey"`       // 生成cdk的时候用
	OrderId    string  `json:"order_id"`    // 邀请购买的关联订单
	InviterId  int     `json:"inviter_id"`  //上级邀请ID
	Status     int     `json:"status"`      //状态 1 正常   2冻结中
	CreateTime int     `json:"create_time"` //
	Username   string  `json:"username"`    //
	Email      string  `json:"email"`       //
}

type UserOrderLogV1 struct {
	ID         int     `json:"id"`          // ID
	Code       int     `json:"code"`        // 标识 1自用购买  2兑换cdk  3提现 10邀请用户购买 11: 用户首次购买 12：兑换ip  13：兑换流量
	Money      float64 `json:"money"`       // 用户支付金额   /    用户提现金额  / 用户兑换金额
	Ratio      float64 `json:"ratio"`       // 返佣比例 或者是单价
	CreateTime int     `json:"create_time"` // 创建时间
	RegTime    int     `json:"reg_time"`    // 被邀请者注册时间
	Username   string  `json:"username"`    // 被邀请者用户名
	Email      string  `json:"email"`       // 被邀请者邮箱
	Value      int64   `json:"value"`       // 购买的套餐值 / 兑换值
}

type ResUserOrderLog struct {
	Uid        int     `json:"uid"`         //购买人的用户ID
	Username   string  `json:"username"`    //购买人的用户名
	Email      string  `json:"email"`       //购买人的邮箱
	OrderId    string  `json:"order_id"`    //订单号
	PayMoney   float64 `json:"pay_money"`   //支付金额
	Money      float64 `json:"money"`       //
	Rate       string  `json:"rate"`        // 生成cdk的时候用
	Status     int     `json:"status"`      //状态 1 正常   2冻结中 3已退款
	StatusText string  `json:"status_text"` //状态 1 正常   2冻结中 3已退款
	CreateTime string  `json:"create_time"` //
}

type ResUserOrderLogV1 struct {
	Email      string  `json:"email"`       // 购买人的邮箱
	Money      float64 `json:"money"`       // 用户支付金额
	Ratio      float64 `json:"ratio"`       // 佣金比例
	Commission float64 `json:"commission"`  // 佣金
	CreateTime string  `json:"create_time"` // 下单时间
	RegTime    string  `json:"reg_time"`    // 注册时间
	Type       string  `json:"type"`        // 类型 isp flow
	Value      string  `json:"value"`       // 兑换值
	Status     int     `json:"status"`      // 状态 1待审核 2已打款 3已拒绝
}

// 获取分页列表
func GetMoneyListByInvitePage(invite_id, offset, limit int, sort string) (data []UserOrderLog) {
	db.Table(userMoneyLogTable+" as m").Select("m.*,i.username").Joins("left join "+userInviterTable+" as i on m.uid=i.uid").Where("m.inviter_id=?", invite_id).Where("m.code =?", "10").Where("m.status < ?", 3).Where("m.order_id <>?", "").Offset(offset).Limit(limit).Order(sort, true).Find(&data)
	return
}

// 获取分页列表
func GetMoneyListByInvitePageV1(invite_id, offset, limit int, sort string) (data []UserOrderLogV1) {
	db.Table(userMoneyLogTable+" as m").Select("m.*,i.email,i.reg_time").Joins("left join "+userInviterTable+" as i on m.uid=i.uid").Where("m.inviter_id=?", invite_id).Where("m.code =?", "10").Where("m.status = ?", 1).Where("m.order_id <>?", "").Offset(offset).Limit(limit).Order(sort, true).Find(&data)
	return
}

// 获取列表
func GetMoneyListByInvite(invite_id int, sort string) (data []UserOrderLog) {
	db.Table(userMoneyLogTable).Where("inviter_id=?", invite_id).Where("code =?", "10").Where("order_id <>?", "").Order(sort, true).Find(&data)
	return
}

// 查询用户佣金兑换记录
func GetUserExchangeList(uid, offset, limit int) (data []UserOrderLogV1) {
	db.Table(userMoneyLogTable).Where("uid = ?", uid).Where("mark = ?", -1).Where("code >= ?", 12).Order("id desc", true).Find(&data)
	return
}

// 查询用户佣金兑换记录
func GetUserExchangeCount(uid int) (total int) {
	db.Table(userMoneyLogTable).Where("uid = ?", uid).Where("mark = ?", -1).Where("code >= ?", 12).Where("status = 1").Count(&total)
	return
}

type UserMoneyLogV1 struct {
	ID         int     `json:"id"`
	Code       int     `json:"code"`        //标识 1自用购买  2兑换cdk 10邀请用户购买
	Money      float64 `json:"money"`       //
	Value      int64   `json:"value"`       //
	Ratio      float64 `json:"ratio"`       //返佣比例 或者是单价
	Mark       int     `json:"mark"`        //符号标识 1增加 -1减少
	Cdkey      string  `json:"cdkey"`       // 生成cdk的时候用
	OrderId    string  `json:"order_id"`    // 邀请购买的关联订单
	InviterId  int     `json:"inviter_id"`  //上级邀请ID
	Status     int     `json:"status"`      //状态 1 正常   2冻结中
	CreateTime int     `json:"create_time"` //
}

// 用户提现记录
type UserWithdrawalModel struct {
	Id         int     `json:"id"`
	Uid        int     `json:"uid"`
	Username   string  `json:"username"`
	Email      string  `json:"email"` //用户填写的邮箱
	Money      float64 `json:"money"`
	TrueMoney  float64 `json:"true_money"`
	Wallet     string  `json:"wallet"`
	OrderNo    string  `json:"order_no"`
	Ip         string  `json:"ip"`
	CreateTime int     `json:"create_time"`
	Status     int     `json:"status"`
	DealTime   int     `json:"deal_time"`
	Remark     string  `json:"remark"`
}

type ResWithdrawalLog struct {
	Id         int     `json:"id"`
	Uid        int     `json:"uid"`      // 用户ID
	Username   string  `json:"username"` // 用户名
	Money      float64 `json:"money"`    // 提现金额
	Status     int     `json:"status"`   // 状态  1待审核 2审核通过 3审核拒绝
	StatusText string  `json:"status_text"`
	Step       int     `json:"step"` // 步骤 1 信息提交中 2 信息核验中 3 有结果
	OrderNo    string  `json:"order_no"`
	CreateTime string  `json:"create_time"` // 申请时间
	DealTime   string  `json:"deal_time"`   // 审核时间
	Remark     string  `json:"remark"`      // 审核备注
	Wallet     string  `json:"wallet"`
}

// 添加提现记录
func AddWithdrawal(model UserWithdrawalModel) (err error) {
	err = db.Table("cm_user_withdrawal").Create(&model).Error
	return
}

// 获取列表
func GetWithdrawalListBy(uid int) (data []UserWithdrawalModel) {
	db.Table("cm_user_withdrawal").Where("uid=?", uid).Order("id desc", true).Find(&data)
	return
}

// 获取分页列表
func GetWithdrawalPageBy(uid int, offset, limit int) (data []UserWithdrawalModel) {
	dbs := db.Table("cm_user_withdrawal")
	if uid > 0 {
		dbs = dbs.Where("uid=?", uid)
	}
	dbs.Offset(offset).Limit(limit).Order("id desc", true).Find(&data)
	return
}

// 获取提现记录总数
func GetWithdrawalCount(uid int) (total int) {
	db.Table("cm_user_withdrawal").Where("uid=?", uid).Count(&total)
	return
}

// 获取分页列表
func GetWithdrawalById(id int) (data UserWithdrawalModel) {
	db.Table("cm_user_withdrawal").Where("id=?", id).Order("id desc", true).First(&data)
	return
}
