package models

// 余额兑换配置
type ConfBalanceConfigModel struct {
	Id          int     `json:"id"`
	Name        string  `json:"name"`         //等级名称
	Cate        string  `json:"cate"`         //类型
	Unit        string  `json:"unit"`         //兑换单价
	Price       float64 `json:"price"`        //兑换单价
	PriceOrigin float64 `json:"price_origin"` //原兑换单价
	Value       int64   `json:"value"`        //兑换值
	Min         int     `json:"min"`          //单次最小兑换值
	Max         int     `json:"max"`          //单次最大兑换值
	Num         int     `json:"num"`          //兑换数量最大值
}

// 余额兑换配置
type ResConfBalanceConfigModel struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`         //等级名称
	Cate        string  `json:"cate"`         //类型
	Unit        string  `json:"unit"`         //兑换单价
	Price       float64 `json:"price"`        //兑换单价
	PriceOrigin float64 `json:"price_origin"` //原兑换单价
	Min         int     `json:"min"`          //单次最小兑换值
	Max         int     `json:"max"`          //单次最大兑换值
	Num         int     `json:"num"`          //兑换数量最大值
}

// 获取信息
func GetBalanceConfigById(id int) (data ConfBalanceConfigModel) {
	db1 := db.Table("conf_balance_pay").Where("id = ?", id)
	db1.Where("status = ?", 1).First(&data)
	return
}

// 获取信息
func GetBalanceConfigList(code string, uid, level int) (data []ConfBalanceConfigModel) {
	if code == "" {
		code = "all"
	}
	db1 := db.Table("conf_balance_pay")
	if code != "" {
		db1 = db1.Where("code = ?", code)
	}
	if uid > 0 {
		db1 = db1.Where("uid = ?", uid)
	}
	if level > 1 {
		db1 = db1.Where("level = ?", level)
	}
	db1.Where("status = ?", 1).Find(&data)
	return
}

// 用户IP 余额
type UserBalanceModel struct {
	Id       int     `json:"id"`
	Uid      int     `json:"uid"`
	Email    string  `json:"email"`
	Username string  `json:"username"`
	Balance  float64 `json:"balance"`
	AllBuy   float64 `json:"all_buy"`
	Status   int     `json:"status"`
}

var userBalanceTable = "cm_user_balance"

func GetUserBalanceByUid(uid int) (data UserBalanceModel) {
	db.Table(userBalanceTable).Where("uid=?", uid).Find(&data)
	return
}

func EditUserBalanceByUid(uid int, params interface{}) (err error) {
	err = db.Table(userBalanceTable).Where("uid=?", uid).Update(params).Error
	return err
}

// 用户余额日志记录
type UserBalanceLogModel struct {
	Id         int     `json:"id"`
	Uid        int     `json:"uid"`         // 用户id
	Money      float64 `json:"money"`       // 金额数
	PreValue   float64 `json:"pre_value"`   // 操作前余额数
	Code       int     `json:"code"`        // 标识 1购买  2生成cdk
	Cate       string  `json:"cate"`        // 类型 isp flow 等 生成cdk类型
	Value      int64   `json:"value"`       // IP数量 或者是流量数
	Number     int     `json:"number"`      // 一次生成cdk数量
	Mark       int     `json:"mark"`        // 符号标识 1增加 -1减少
	OrderId    string  `json:"order_id"`    // 购买的关联订单
	PayPlat    int     `json:"pay_plat"`    // 支付方式
	Status     int     `json:"status"`      // 状态 1 正常 2冻结
	CreateTime int     `json:"create_time"` // 操作时间
}

// 返回用户余额日志记录
type ResUserBalanceLogModel struct {
	Id         int    `json:"id"`
	Money      string `json:"money"`       // 金额数
	Name       string `json:"name"`        // 名称
	Cate       string `json:"cate"`        // 类型
	Value      int64  `json:"value"`       // IP数量 或者是流量数
	Unit       string `json:"unit"`        // 单位
	Number     int    `json:"number"`      // 一次生成cdk数量
	CreateTime string `json:"create_time"` // 操作时间
}

var tableUserBalanceLogName = "cm_user_balance_log"

// 用户余额日志记录
// code 标识 1购买  2生成cdk 3自用充值
// balance 余额数
// preMoney 操作前余额数
func AddUserBalanceLog(uid, code int, balance, preMoney float64, cate string, value int64, number, mark, createTime int,country string) (err error) {
	info := UserBalanceLogModel{
		Uid:        uid,
		Money:      balance,
		PreValue:   preMoney,
		Code:       code,
		Cate:       cate,
		Value:      value,
		Number:     number,
		Mark:       mark,
		OrderId:    country,	// 符号标识 购买是订单号 ，使用时是其他参数
		PayPlat:    0,
		Status:     1,
		CreateTime: createTime,
	}
	err = db.Table(tableUserBalanceLogName).Create(&info).Error
	return
}

// 获取用户余额日志记录
func GetUserBalanceLogBy(uid, code int, cate string, mark, startTime, endTime int) (data []UserBalanceLogModel) {
	dbBl := db.Table(tableUserBalanceLogName).Where("uid=?", uid)
	if code > 0 {
		dbBl = dbBl.Where("code >=?", code)
	}
	if cate != "" {
		dbBl = dbBl.Where("cate=?", cate)
	}
	if mark != 0 {
		dbBl = dbBl.Where("mark=?", mark)
	}
	if startTime > 0 {
		dbBl = dbBl.Where("create_time>?", startTime)
	}
	if endTime > 0 {
		dbBl = dbBl.Where("create_time<?", endTime)
	}
	dbBl.Order("id desc").Find(&data)
	return
}
