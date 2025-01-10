package models

import (
	"api-360proxy/web/pkg/util"
	"fmt"
	"time"
)

type ResUserStaticIpRegion struct {
	Id          int               `json:"id"`
	PakName     string            `json:"pak_name"`   // 套餐类型
	Balance     int               `json:"balance"`    // 剩余IP
	ExpireDay   int               `json:"expire_day"` // 过期时间
	CountryList []ResStaticRegion `json:"country_list"`
}

type ResStaticRegion struct {
	Country  string             `json:"country"`
	Code     string             `json:"code"`
	Name     string             `json:"name"`
	IpNumber int                `json:"ip_number"`
	Balance  int                `json:"balance"`
	StatList []StaticStateModel `json:"state_list"`
}

type StaticStateModel struct {
	Code     string            `json:"code"`
	Name     string            `json:"name"`
	IpNumber int               `json:"ip_number"`
	Sort     int               `json:"sort"`
	CityList []StaticCityModel `json:"city_list"`
}

type StaticCityModel struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	IpNumber int    `json:"ip_number"`
	Sort     int    `json:"sort"`
}

type StaticIpCountryModel struct {
	Country  string `json:"country"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	IpNumber int    `json:"ip_number"`
}

// 获取静态国家
func GetStaticIpCountry() (area []StaticIpCountryModel) {
	dbs := db.Table("cm_static_ip_country").Where("status =?", 1)
	dbs.Find(&area)
	return
}

type StaticRegionModel struct {
	Country  string `json:"country"`
	State    string `json:"state"`
	City     string `json:"city"`
	IpNumber int    `json:"ip_number"`
	Sort     int    `json:"sort"`
}

// 获取静态国家
func GetStaticRegion() (area []StaticRegionModel) {
	db.Table("cm_static_region").Where("status =?", 1).Order("sort desc").Find(&area)
	return
}

// 获取列表
func GetStaticIpPool() (area []StaticIpPoolModel) {
	dbs := db.Table("cm_static_ip_pool").Where("uid =?", 0)

	dbs.Where("status =?", 1).Find(&area)
	return
}

type StaticIpInfo struct {
	Ip      string `json:"ip"`
	Sn      string `json:"sn"`
	Ping    int    `json:"ping"`
	Code    string `json:"code"`
	Country string `json:"country"`
	State   string `json:"state"`
	City    string `json:"city"`
	Port    string `json:"port"`
}

type StaticIpPoolModel struct {
	Id      int    `json:"id"`
	Ip      string `json:"ip"`
	Port    int    `json:"port"`
	Uid     int    `json:"uid"`
	Country string `json:"country"`
	Code    string `json:"code"`
	State   string `json:"state"`
	City    string `json:"city"`
	Status  int    `json:"status"`
}

// 随机排序
func GetStaticIpPoolRand(country, state, city string) (area []StaticIpPoolModel) {
	dbs := db.Table("cm_static_ip_pool").Where("uid =?", 0)

	if country != "" && country != "all" {
		dbs = dbs.Where("country=?", country)
	}
	if state != "" && state != "all" {
		dbs = dbs.Where("state=?", state)
	}
	if city != "" && city != "all" {
		dbs = dbs.Where("city=?", city)
	}
	dbs.Where("status =?", 1).Order("rand()").Limit(150).Find(&area)
	return
}

// 查询静态IP信息
func GetStaticIpById(id int) (err error, info StaticIpPoolModel) {
	err = db.Table("cm_static_ip_pool").Where("id=?", id).First(&info).Error
	return
}
// 获取下线IP列表
func GetStaticOfflineIps() (area []StaticIpPoolModel) {
	dbs := db.Table("cm_static_ip_pool").Where("status != ?", 1)
	dbs.Find(&area)
	return
}
// 查询静态IP信息
func GetStaticIpByIp(ip string) (err error, info StaticIpPoolModel) {
	err = db.Table("cm_static_ip_pool").Where("ip=?", ip).First(&info).Error
	return
}

// 删除提取记录
type IpStaticLogDelModel struct {
	IpStaticLogModel
	DelTime int    `json:"del_time"`
	DelIp   string `json:"del_ip"`
}

// 静态长效续费记录记录
type IpStaticRepayModel struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	StaticId   int    `json:"static_id"`
	Username   string `json:"username"`
	Code       string `json:"code"`
	Ip         string `json:"ip"`
	Port       int    `json:"port"`
	Country    string `json:"country"`
	State      string `json:"state"`
	City       string `json:"city"`
	ExpireDay  int    `json:"expire_day"`  //续费天数
	ExpireTime int    `json:"expire_time"` //续费后的过期时间
	CreateTime int    `json:"create_time"` // 续费时间
	UserIp     string `json:"user_ip"`
}

type IpExtractModel struct {
	Id          int     `json:"id"`
	Uid         int     `json:"uid"`
	Username    string  `json:"username"`
	UserBalance float64 `json:"user_balance"`
	Cate        string  `json:"cate"`
	Ip          string  `json:"ip"`
	UserIp      string  `json:"user_ip"`
	Unit        float64 `json:"unit"`
	ExtractFrom string  `json:"extract_from"`
	CreateTime  int64   `json:"create_time"`
}

// 添加用户使用记录
func AddIpExtract(log IpExtractModel) (err error) {
	date := time.Now().Format("200601")
	var logTableName = "cm_log_extract" + date
	if !db.HasTable(logTableName) {
		createExtractLogTable(logTableName)
	}
	err = db.Table(logTableName).Create(&log).Error
	return
}

func createExtractLogTable(logTableName string) {
	createTable := `CREATE TABLE ` + logTableName + `(
		id int unsigned NOT NULL AUTO_INCREMENT,
  		uid int DEFAULT NULL,
  		username varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT NULL,
  		user_balance int DEFAULT NULL,
  		ip varchar(60) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT '',
  		user_ip varchar(45) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT '',
  		unit float(11,2) DEFAULT '0.00' COMMENT '单价',
  		cate varchar(30) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT 'ip' COMMENT '类型',
		extract_from varchar(30) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT '' COMMENT '提取端：client，web',
  		create_time int DEFAULT NULL,
		PRIMARY KEY (id),
		KEY uid (uid) USING BTREE,
		KEY username (username) USING BTREE,
		KEY cate (cate) USING BTREE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='ip提取记录'`
	db.Exec(createTable)
}

// 扣费
func StaticKf(code, user_ip string, ipInfo StaticIpPoolModel, userInfo Users, balanceInfo UserStaticIpModel) error {
	nowTime := util.GetNowInt()
	// 开始扣费
	tx := db.Begin()
	uid := userInfo.Id
	balance := balanceInfo.Balance - 1
	err1 := tx.Table(UserStaticIpTable).Where("id = ?", balanceInfo.Id).Updates(map[string]interface{}{"balance": balance, "last_use_time": nowTime}).Error
	if err1 != nil {
		tx.Rollback()
	}
	expireDay := balanceInfo.ExpireDay

	ipLogInfo := IpStaticLogModel{}
	ipLogInfo.Uid = uid
	ipLogInfo.Username = userInfo.Username
	ipLogInfo.Ip = ipInfo.Ip
	ipLogInfo.Code = code
	ipLogInfo.Port = ipInfo.Port
	ipLogInfo.Country = ipInfo.Country
	ipLogInfo.State = ipInfo.State
	ipLogInfo.City = ipInfo.City
	ipLogInfo.ExpireDay = expireDay
	ipLogInfo.ExpireTime = expireDay*86400 + nowTime
	ipLogInfo.CreateTime = nowTime
	ipLogInfo.UserIp = user_ip
	ipLogInfo.Account = userInfo.Username
	ipLogInfo.Password = util.RandStr("r", 8)
	err1 = tx.Table("cm_log_static").Create(&ipLogInfo).Error

	expire_rime := expireDay*86400 + nowTime
	upPool := map[string]interface{}{"uid": uid, "expired": expire_rime}
	err1 = tx.Table("cm_static_ip_pool").Where("id = ?", ipInfo.Id).Updates(upPool).Error
	//添加扣费日志
	ip_log := IpExtractModel{
		Uid:         uid,
		CreateTime:  time.Now().Unix(),
		Username:    userInfo.Username,
		UserBalance: 0,
		Ip:          ipInfo.Ip,
		UserIp:      user_ip,
		ExtractFrom: "web",
		Cate:        balanceInfo.PakRegion,
		Unit:        0,
	}
	err1 = AddIpExtract(ip_log)
	tx.Commit()
	return err1
}

// 续费扣费
func Recharge(user_ip string, ipLog IpStaticLogModel, balanceInfo UserStaticIpModel) error {
	nowTime := util.GetNowInt()
	// 开始扣费
	tx := db.Begin()

	balance := balanceInfo.Balance - 1
	err1 := tx.Table(UserStaticIpTable).Where("id = ?", balanceInfo.Id).Updates(map[string]interface{}{"balance": balance, "last_use_time": nowTime}).Error
	if err1 != nil {
		tx.Rollback()
	}
	expireTime := ipLog.ExpireTime
	if ipLog.ExpireTime < nowTime {
		expireTime = nowTime
	}
	expire_day := ipLog.ExpireDay + balanceInfo.ExpireDay
	expire_rime := expireTime + balanceInfo.ExpireDay*86400
	up := map[string]interface{}{"expire_day": expire_day, "expire_time": expire_rime, "update_time": nowTime}
	err1 = tx.Table("cm_log_static").Where("id = ?", ipLog.Id).Updates(up).Error
	//添加续费日志
	ipRepayModel := IpStaticRepayModel{}
	ipRepayModel.StaticId = ipLog.Id
	ipRepayModel.Uid = ipLog.Uid
	ipRepayModel.Username = ipLog.Username
	ipRepayModel.Code = ipLog.Code
	ipRepayModel.Ip = ipLog.Ip
	ipRepayModel.Port = ipLog.Port
	ipRepayModel.Country = ipLog.Country
	ipRepayModel.State = ipLog.State
	ipRepayModel.City = ipLog.City
	ipRepayModel.ExpireDay = expire_day
	ipRepayModel.ExpireTime = expire_rime
	ipRepayModel.CreateTime = nowTime
	ipRepayModel.UserIp = user_ip
	err1 = tx.Table("cm_ip_static_repay").Create(&ipRepayModel).Error
	tx.Commit()
	return err1
}

// 过期删除记录
func DelStaticLog(ip string, ipLog IpStaticLogModel) error {
	nowTime := util.GetNowInt()
	// 开始扣费
	tx := db.Begin()

	err1 := tx.Table("cm_log_static").Where("id = ?", ipLog.Id).Delete(&IpStaticLogModel{}).Error
	if err1 != nil {
		tx.Rollback()
	}
	ipDelModel := IpStaticLogDelModel{}
	ipDelModel.Id = ipLog.Id
	ipDelModel.Uid = ipLog.Uid
	ipDelModel.Username = ipLog.Username
	ipDelModel.Code = ipLog.Code
	ipDelModel.Ip = ipLog.Ip
	ipDelModel.Port = ipLog.Port
	ipDelModel.Country = ipLog.Country
	ipDelModel.State = ipLog.State
	ipDelModel.City = ipLog.City
	ipDelModel.ExpireDay = ipLog.ExpireDay
	ipDelModel.ExpireTime = ipLog.ExpireTime
	ipDelModel.UpdateTime = ipLog.UpdateTime
	ipDelModel.CreateTime = ipLog.CreateTime
	ipDelModel.UserIp = ipLog.UserIp
	ipDelModel.DelTime = nowTime
	ipDelModel.DelIp = ip
	err1 = tx.Table("cm_log_static_del").Create(&ipDelModel).Error
	tx.Commit()
	return err1
}
// 添加记录
func AddDelStaticLog(ip string, ipLog IpStaticLogModel) error {
	nowTime := util.GetNowInt()
	ipDelModel := IpStaticLogDelModel{}
	ipDelModel.Id = ipLog.Id
	ipDelModel.Uid = ipLog.Uid
	ipDelModel.Username = ipLog.Username
	ipDelModel.Code = ipLog.Code
	ipDelModel.Ip = ipLog.Ip
	ipDelModel.Port = ipLog.Port
	ipDelModel.Country = ipLog.Country
	ipDelModel.State = ipLog.State
	ipDelModel.City = ipLog.City
	ipDelModel.ExpireDay = ipLog.ExpireDay
	ipDelModel.ExpireTime = ipLog.ExpireTime
	ipDelModel.UpdateTime = ipLog.UpdateTime
	ipDelModel.CreateTime = ipLog.CreateTime
	ipDelModel.UserIp = ipLog.UserIp
	ipDelModel.DelTime = nowTime
	ipDelModel.DelIp = ip
	err1 := db.Table("cm_log_static_del").Create(&ipDelModel).Error
	fmt.Println(err1)
	return err1
}