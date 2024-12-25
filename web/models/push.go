package models

//推送数据
type PushAccount struct {
	Cate       string `json:"cate"`
	Uid        int    `json:"uid"`
	AccountId  int    `json:"account_id"`
	Flows      int64  `json:"flows"`
	LimitFlow  int64  `json:"limit_flow"`
	FlowUnit   string `json:"flow_unit"`
	Ip         string `json:"ip"`
	ExpireTime int    `json:"expire_time"`
	CreateTime int    `json:"create_time"`
}

//推送数据 自用流量
type PushInvite struct {
	Cate       string  `json:"cate"`
	Code       int     `json:"code"`
	Uid        int     `json:"uid"`
	Value      int64   `json:"value"`
	Ratio      float64 `json:"ratio"`
	Money      float64 `json:"money"`
	CreateTime int     `json:"create_time"`
	Today      int     `json:"today"` //
}

//推送数据 自用流量
type PushInviteLog struct {
	PushInvite
	Result string `json:"result"` //
}

var logInviteTable = "log_invite_money"

// 推送兑换记录信息
func CreateInviteLog(data PushInviteLog) (err error) {
	err = db.Table(logInviteTable).Create(&data).Error
	return err
}

//异步处理cdk数据
type PushCdkey struct {
	Mode         string       `json:"mode"` //使用模式 agent 代理商  invite 邀请  score积分 balance 余额
	Cate         string       `json:"cate"`
	Cdkey        string       `json:"cdkey"`
	Number       int64        `json:"number"`
	Need         float64      `json:"need"`     //所需余额
	Country      string       `json:"country"`  //静态兑换国家  //20241204 增加
	CdkType      string       `json:"cdk_type"` // 类型 add  生成  ex 兑换
	Uid          int          `json:"uid"`
	BindUsername string       `json:"bind_username"` // 接收人用户名
	BindEmail    string       `json:"bind_email"`    // 接收人邮箱
	BindUid      int          `json:"bind_uid"`      // 接收人用户ID
	BindTime     int          `json:"bind_time"`     // 充值时间
	Ip           string       `json:"ip"`            // 当前IP
	CreateTime   int          `json:"create_time"`
	ExInfo       ExchangeList `json:"ex_info"`
}

type PushCdkRecord struct {
	Cate         string  `json:"cate"`
	Cdkey        string  `json:"cdkey"`
	Value        int64   `json:"value"`   //操作前余额
	Balance      float64 `json:"balance"` //操作前余额  余额充值
	Number       int64   `json:"number"`
	CdkType      string  `json:"cdk_type"` // 类型 add  生成  ex 兑换
	Country      string  `json:"country"`  //静态兑换国家  //20241204 增加
	Uid          int     `json:"uid"`
	BindUsername string  `json:"bind_username"` // 接收人用户名
	BindUid      int     `json:"bind_uid"`      // 接收人用户ID
	BindEmail    string  `json:"bind_email"`    // 接收人邮箱
	BindTime     int     `json:"bind_time"`     // 充值时间
	Ip           string  `json:"ip"`            // 当前IP
	CreateTime   int     `json:"create_time"`
	Result       string  `json:"result"`
	Remark       string  `json:"remark"`
}

var logCdkTable = "log_cdk_record"

// 推送兑换记录信息
func CreateCdkLog(data PushCdkRecord) (err error) {
	err = db.Table(logCdkTable).Create(&data).Error
	return err
}

// 异步推送 积分兑换-流量
type PushScoreFlow struct {
	Uid        int    `json:"uid"`
	Name       string `json:"name"`
	Score      int    `json:"score"` //所需积分
	Flow       int    `json:"flow"`  //可得流量值（GB）
	Ip         string `json:"ip"`    // 当前IP
	CreateTime int    `json:"create_time"`
}

// 异步推送 积分兑换-不限量流量
type PushScoreFlowDay struct {
	Uid        int    `json:"uid"`
	Name       string `json:"name"`
	Score      int    `json:"score"` //所需积分
	Day        int    `json:"day"`   //可得天数
	Ip         string `json:"ip"`    // 当前IP
	CreateTime int    `json:"create_time"`
}
