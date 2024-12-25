package models

type CmOrder struct {
	PakId      int     `json:"pak_id"`
	PakType    string  `json:"pak_type"`
	OrderId    string  `json:"order_id"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Username   string  `json:"username"`
	Uid        int     `json:"uid"`
	Value      int     `json:"value"`
	Give       int     `json:"give"`
	Gift       int     `json:"gift"`
	Discount   float64 `json:"discount"`
	TrueMoney  float64 `json:"true_money"`
	PayStatus  int     `json:"pay_status"`
	PayTime    int     `json:"pay_time"`
	PayPlat    int     `json:"pay_plat"`
	PaySn      string  `json:"pay_sn"`
	CreateTime int     `json:"create_time"`
	Plat       string  `json:"plat"`
	PakName    string  `json:"pak_name"`
	Invoice    string  `json:"invoice"`
	PakRegion  string  `json:"pak_region"`
	Reward     int     `json:"reward"`
}

type ResSocksOrder struct {
	PakId      int     `json:"pak_id"`
	OrderId    string  `json:"order_id"`
	Name       string  `json:"name"`
	Uid        int     `json:"uid"`
	Total      int     `json:"total"`
	Value      int     `json:"value"`
	Give       int     `json:"give"`
	Char       string  `json:"char"`
	TrueMoney  float64 `json:"true_money"`
	PayTime    string  `json:"pay_time"`
	PayPlat    string  `json:"pay_plat"`
	PaySn      string  `json:"pay_sn"`
	CreateTime string  `json:"create_time"`
	Plat       string  `json:"plat"`
	PakName    string  `json:"pak_name"`
	PakType    string  `json:"pak_type"`
	Invoice    string  `json:"invoice"`
	PayStatus  string  `json:"pay_status"`
}

var socksOrderTable = "cm_order"

// 分页获取的订单信息
func GetSocksOrderPage(uid, offset, limit int) (err error, data []CmOrder) {
	err = db.Table(socksOrderTable).Where("uid=?", uid).Where("pay_status=?", 3).Where("status=?", 1).Offset(offset).Limit(limit).Order(" id desc").Find(&data).Error
	return
}

func GetOrderListCount(uid int) (num int) {
	db.Table(socksOrderTable).Where("uid=?", uid).Where("pay_status=?", 3).Where("status=?", 1).Order("id desc", true).Count(&num)
	return
}

func GetOrderListBy(uid, start_time, end_time, pay_status int, pak_type string) (data []CmOrder) {
	bds := db.Table(socksOrderTable)
	if uid > 0 {
		bds = bds.Where("uid = ?", uid)
	}
	if pay_status > 0 {
		bds = bds.Where("pay_status = ?", pay_status)
	}
	if pak_type != "" {
		bds = bds.Where("pak_type = ?", pak_type)
	}
	if start_time > 0 && end_time > 0 {
		bds = bds.Where("create_time >= ? and create_time <= ?", start_time, end_time)
	}

	bds.Where("status=?", 1).Order(" id desc").Find(&data)
	return

}

func GetOrderCountByPakId(uid, pak_id, create_time int) (num int) {
	db.Table(socksOrderTable).Where("uid=?", uid).Where("pak_id=?", pak_id).Where("pay_status=?", 3).Where("status=?", 1).Where("create_time >=?", create_time).Order("id desc", true).Count(&num)
	return
}

func GetOrderListByLimit(limit int) (err error, cmOrder []CmOrder) {
	err = db.Table(socksOrderTable).Where("status = ?", 1).Where("pay_status >= ?", 1).Order("id desc", true).Limit(limit).Find(&cmOrder).Error
	return
}

func GetOrderTotalMoney(orderIds []string) float64 {
	var list []CmOrder
	var total_money float64
	db.Table(socksOrderTable).Where("order_id in (?)", orderIds).Select("true_money").Find(&list)
	for _, v := range list {
		total_money += v.TrueMoney
	}
	return total_money
}

/// 获取用户购买这个套餐的次数
func GetOrderCountWith(uid, pakId int) (num int) {
	db.Table(socksOrderTable).Where("uid=?", uid).
		Where("pak_id=?", pakId).
		Where("pay_status=?", 3).
		Where("status=?", 1).
		Order("id desc", true).Count(&num)
	return
}

type ResOrderNumber struct {
	PakId int `json:"pak_id"`
	Count int `json:"count"`
}

// 获取用户购买套餐的次数
func GetOrderCountWithUid(uid int) (info []ResOrderNumber) {
	db.Table(socksOrderTable).Select("pak_id, count(*) as count").Where("uid=?", uid).
		Where("pay_status=?", 3).
		Where("status=?", 1).
		Group("pak_id").Find(&info)
	return
}
