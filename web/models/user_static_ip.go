package models

import "api-360proxy/web/pkg/util"

type ResIpStaticLogModel struct {
	Id         int    `json:"id"`
	Ip         string `json:"ip"`
	Port       int    `json:"port"`
	Country    string `json:"country"`
	State      string `json:"state"`
	City       string `json:"city"`
	ExpireTime string `json:"expire_time"`
	IsExpire   int    `json:"is_expire"`
	IsOffline  string `json:"is_offline"`
	IsReplace  int    `json:"is_replace"` //1:可替换 0:不可替换
	CreateTime string `json:"create_time"`
	Forward    string `json:"forward"`
	Account    string `json:"account"`
	Password   string `json:"password"`
	Remark     string `json:"remark"`
}

type RepayLogModel struct {
	Id         int    `json:"id"`
	Ip         string `json:"ip"`
	Country    string `json:"country"`
	ExpireTime string `json:"expire_time"`
	IsExpire   int    `json:"is_expire"`
	CreateTime string `json:"create_time"`
}

type IpStaticLogModel struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	Username   string `json:"username"`
	Code       string `json:"code"`
	Ip         string `json:"ip"`
	Port       int    `json:"port"`
	Country    string `json:"country"`
	State      string `json:"state"`
	City       string `json:"city"`
	ExpireDay  int    `json:"expire_day"`
	ExpireTime int    `json:"expire_time"`
	UpdateTime int    `json:"update_time"`
	CreateTime int    `json:"create_time"`
	UserIp     string `json:"user_ip"`
	Account    string `json:"account"`
	Password   string `json:"password"`
	Remark     string `json:"remark"`
	Replaced   int    `json:"replaced"` //1:已替换 0:未替换
	Status     int    `json:"status"`   //IP状态 1正常
	OrderId    string `json:"order_id"` //开通时传值给资源中台的信息
	IsNew      int    `json:"is_new"`   //是否新资源中台 1是
}

type ResStaticRecordModel struct {
	Id         int    `json:"id"`
	Ip         string `json:"ip"`
	Port       int    `json:"port"`
	Country    string `json:"country"`
	State      string `json:"state"`
	City       string `json:"city"`
	ExpireTime string `json:"expire_time"`
	IsExpire   int    `json:"is_expire"`
	CreateTime string `json:"create_time"`
	Remark     string `json:"remark"`
}

type ResUserStaticArea struct {
	Country string            `json:"country"`  // 地区
	Balance int               `json:"balance"`  // 剩余IP
	PakList []ResUserStaticIp `json:"pak_list"` // 套餐列表
}

// 查询静态IP信息
func GetIpStaticIpByUid(uid int) (err error, user []IpStaticLogModel) {
	err = db.Table("cm_log_static").Where("uid=?", uid).Find(&user).Error
	return
}

// 查询静态IP信息
func GetIpStaticIpBy(uid int, ip, status, orderBy, country string) (err error, user []IpStaticLogModel) {
	if orderBy == "" {
		orderBy = "id desc"
	}
	dbs := db.Table("cm_log_static").Where("uid=?", uid)
	if status != "" {
		now := util.GetNowInt()
		if status == "1" {
			dbs = dbs.Where("expire_time >= ?", now)
		}
		if status == "2" {
			dbs = dbs.Where("expire_time < ?", now)
		}
	}
	if country != "" {
		dbs = dbs.Where("country = ?", country)
	}
	if ip != "" {
		dbs = dbs.Where("ip like ? or remark like ?", "%"+ip+"%", "%"+ip+"%")
	}
	err = dbs.Order(orderBy).Find(&user).Error
	return
}

// 查询静态IP信息
func GetUsedByLog(uid, start, end int, isDel int) (err error, user []IpStaticLogModel) {
	orderBy := "id desc"
	table := "cm_log_static"
	if isDel == 1 {
		table = "cm_log_static_del"
	}
	dbs := db.Table(table).Where("uid=?", uid)
	if start > 0 {
		dbs = dbs.Where("create_time >= ?", start)
	}
	if end > 0 {
		dbs = dbs.Where("create_time <= ?", end)
	}
	err = dbs.Order(orderBy).Find(&user).Error
	return
}

// 查询静态IP信息
func GetIpStaticIpById(id int) (err error, user IpStaticLogModel) {
	err = db.Table("cm_log_static").Where("id=?", id).First(&user).Error
	return
}

// 查询静态IP信息
func GetIpStaticIp(uid int, ip string) (err error, user IpStaticLogModel) {
	err = db.Table("cm_log_static").Where("uid=?", uid).Where("ip=?", ip).First(&user).Error
	return
}

// 修改用户IP账号密码
func SetIpStaticIp(id int, upInfo map[string]interface{}) (err error) {
	err = db.Table("cm_log_static").Where("id = ?", id).Update(upInfo).Error
	return
}

type ResUserStaticInfo struct {
	Country string `json:"country"`  // 国家地区
	PakName string `json:"pak_name"` // 套餐类型
	Balance int    `json:"balance"`  // 剩余IP
}

type UserStaticIpModel struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	PakId      int    `json:"pak_id"`     // 套餐ID
	PakRegion  string `json:"pak_region"` // 套餐类型
	AllBuy     int    `json:"all_buy"`    // 总充值
	AllNum     int    `json:"all_num"`    //总充值次数
	Balance    int    `json:"balance"`    // 剩余IP
	ExpireDay  int    `json:"expire_day"` // 过期时间
	Sort       int    `json:"sort"`
	CreateTime int    `json:"create_time"`
}
