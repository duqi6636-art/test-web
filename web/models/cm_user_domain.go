package models

type CmUserDomain struct {
	Id     int    `json:"id"`     // ID
	Domain string `json:"domain"` // 域名
	Title  string `json:"title"`  // 提示
}

// 查询记录
func GetUserDomain(uid int) (info []CmUserDomain, err error) {
	err = db.Table("cm_user_domain").Where("uid = ? and status = ?", uid, 1).Find(&info).Error
	return
}

type AddCmUserDomain struct {
	Id         int    `json:"id"`          // ID
	Uid        int    `json:"uid"`         // 用户id
	Username   string `json:"username"`    // 用户名
	Domain     string `json:"domain"`      // 域名
	Title      string `json:"title"`       // 提示
	Status     int    `json:"status"`      // 状态：1：正常，0：禁用
	CreateTime int    `json:"create_time"` // 创建时间
}

// 创建域名
func CreateDomain(info AddCmUserDomain) error {
	return db.Table("cm_user_domain").Create(&info).Error
}