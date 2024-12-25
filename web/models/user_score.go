package models

type UserScore struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	AllScore   int    `json:"all_score"`   // 累计总积分
	Score      int    `json:"score"`       // 剩余积分
	ExpireTime int    `json:"expire_time"` // 过期时间
	CreateTime int    `json:"create_time"`
}

var userScoreTable = "cm_user_score"

// 查询用户流量表
func GetUserScoreInfo(uid int) (data UserScore) {
	db.Table(userScoreTable).Where("uid = ?", uid).Find(&data)
	return
}

// 查询用户流量表
func EditUserScore(uid int, param interface{}) (err error) {
	err = db.Table(userScoreTable).Where("uid = ?", uid).Update(param).Error
	return
}

type LogUserScore struct {
	Id         int     `json:"id"`
	Uid        int     `json:"uid"`         // 用户id
	Name       string  `json:"name"`        // 名称
	Score      int     `json:"score"`       // 积分值
	Money      float64 `json:"money"`       // 支付金额
	Code       int     `json:"code"`        // 标识 1购买 2反馈获得  3兑换cdk
	Mark       int     `json:"mark"`        // 符号标识 1增加 -1减少
	Value      int64   `json:"value"`       // 兑换的流量值
	Status     int     `json:"status"`      // 状态 1 正常   2冻结中
	Ip         string  `json:"ip"`          // 操作IP
	Remark     string  `json:"remark"`      // 备注
	CreateTime int     `json:"create_time"` // 创建时间
}

var logUserScoreTable = "log_user_score"

func CreateLogUserScore(data LogUserScore) (err error) {
	err = db.Table(logUserScoreTable).Create(&data).Error
	return err
}

// 获取积分操作记录
func GetLogUserScore(uid int) (data []LogUserScore) {
	db.Table(logUserScoreTable).Where("uid =?", uid).Where("status =?", 1).Order("id desc").Find(&data)
	return
}

// 积分兑换流量配置
type ConfScoreFlow struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`  // 等级名称
	Flow  int    `json:"flow"`  // 获得流量
	Score int    `json:"score"` // 所需积分
}
// 积分兑换不限量流量配置
type ConfScoreFlowDay struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`  // 等级名称
	Day  int    `json:"day"`  // 获得天数
	Score int    `json:"score"` // 所需积分
}
// 获取信息
func GetConfScoreByFlow(flow int) (data ConfScoreFlow) {
	dbs := db.Table("conf_score_flow")
	dbs = dbs.Where("flow = ?", flow)
	dbs.First(&data)
	return
}
// 获取信息
func GetConfScoreByDay(day int) (data ConfScoreFlowDay) {
	dbs := db.Table("conf_score_flow_day")
	dbs = dbs.Where("day = ?", day)
	dbs.First(&data)
	return
}
// 免费获得积分反馈
type ScoreFeedback struct {
	ID         int    `json:"id"`
	Uid        int    `json:"uid"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Company    string `json:"company"`
	Address    string `json:"address"`
	Content    string `json:"content"`
	Lang       string `json:"lang"`
	Ip         string `json:"ip"`
	Status     int    `json:"status"`
	CreateTime int    `json:"create_time"`
}

var feedbackScoreTable = "cm_feedback_score"

// 获取记录
func GetScoreFeedback(uid int) (fb ScoreFeedback) {
	db.Table(feedbackScoreTable).Where("uid =?", uid).First(&fb)
	return
}

// 添加
func AddScoreFeedback(fb ScoreFeedback) error {
	return db.Table(feedbackScoreTable).Create(&fb).Error
}

// 更新
func EditScoreFeedback(id int, param interface{}) error {
	return db.Table(feedbackScoreTable).Where("id =?", id).Update(param).Error
}
