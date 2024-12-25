package models

import "api-360proxy/web/pkg/util"

// 流量
type EmailMarketModel struct {
	Id       int    `json:"id"`
	Uid      int    `json:"uid"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Cate     string `json:"cate"`
	Code     string `json:"code"`
	IsCoupon int    `json:"is_coupon"`
	IsEmail  int    `json:"is_email"`
}

// 发送队列
func GetEmailMarketListBy() (data []EmailMarketModel) {
	var tableName = "log_email_market"

	db.Table(tableName).Where("is_send =? and is_email =?", 0, 1).Limit(50).Find(&data)

	return
}

// 更新发送结果
func EditEmailMarket(id int, isSend int) {
	var tableName = "log_email_market"
	//info := []EmailMarketModel{}
	//err := dbLog.Table(tableName).Where("id =?").Find(&info).Error

	upParams := map[string]interface{}{
		"is_send":     isSend,
		"update_time": util.GetNowInt(),
	}
	db.Table(tableName).Where("id =? ", id).Update(upParams)
	return
}
