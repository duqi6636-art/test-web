package models

type UserStaticIp struct {
	ID         int    `json:"id"`
	Uid        int    `json:"uid"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	PakId      int    `json:"pak_id"`     // 套餐ID
	PakRegion  string `json:"pak_region"` // 套餐类型
	AllBuy     int    `json:"all_buy"`    // 总充值
	AllNum     int    `json:"all_num"`    //总充值次数
	Balance    int    `json:"balance"`    // 剩余IP
	ExpireDay  int    `json:"expire_day"` // 过期时间
	Status     int    `json:"status"`     // 状态 1正常 2停用
	CreateTime int    `json:"create_time"`
}

type ResUserStaticIp struct {
	Id        int    `json:"id"`
	PakName   string `json:"pak_name"`   // 套餐类型
	Balance   int    `json:"balance"`    // 剩余IP
	ExpireDay int    `json:"expire_day"` // 过期时间
	Status    int    `json:"status"`     // 过期时间
}

var UserStaticIpTable = "cm_user_static_ip"

// 创建用户静态余额
func AddUserStatic(data UserStaticIp) (err error) {
	err = db.Table(UserStaticIpTable).Create(&data).Error
	return
}

// 修改用户余额
func EditUserStatic(id int, params interface{}) (err error) {
	err = db.Table(UserStaticIpTable).Where("id =?", id).Update(params).Error
	return err
}

// 查询用户IP余额 ID
func GetUserStaticIpById(uid, id int) (err error, user UserStaticIpModel) {
	err = db.Table(UserStaticIpTable).Where("id = ?", id).Where("uid = ? AND status >= 1", uid).First(&user).Error
	return
}

// 查询用户IP余额 套餐 和 地区
func GetUserStaticByPakRegion(uid, pakId int, region string) (err error, user UserStaticIpModel) {
	err = db.Table(UserStaticIpTable).Where("uid = ?", uid).Where("pak_id = ? AND status >= 1", pakId).Where("pak_region = ?", region).First(&user).Error
	return
}

type UserStaticPakNum struct {
	Balance int `json:"balance"` // 剩余IP
	PakId   int `json:"pak_id"`  // 过期时间
}

// 查询用户IP余额 套餐
func GetUserStaticByPak(uid, pakId int) (info UserStaticPakNum, err error) {
	err = db.Table(UserStaticIpTable).
		Select("sum(balance) as balance, pak_id").
		Where("uid = ?", uid).
		Where("pak_id = ?", pakId).
		Where("status >= ?", 1).
		Group("pak_id").First(&info).Error
	return
}

// 查询用户IP余额
func GetUserStaticIp(uid int) (err error, user []UserStaticIp) {
	err = db.Table(UserStaticIpTable).Where("uid = ?", uid).Find(&user).Error
	return
}

// 查询用户IP余额
func GetUserStaticList(uid int) (err error, user []UserStaticIp) {
	err = db.Table(UserStaticIpTable).Where("uid = ?", uid).Where("status >= ?", 1).Find(&user).Error
	return
}

func GetUserStaticIpByRegion(uid int, region string) (err error, user []UserStaticIpModel) {
	err = db.Table(UserStaticIpTable).Where("uid = ?", uid).Where("pak_region = ?", region).Where("status =?", 1).Find(&user).Error
	return
}

// 查询用户IP余额 UID AREA
func GetUserStaticIpByArea(uid int, region string) (err error, user UserStaticIpModel) {
	err = db.Table(UserStaticIpTable).Where("uid = ?", uid).Where("pak_region = ?", region).Where("balance > ?", 0).Order("sort desc", true).First(&user).Error
	return
}
