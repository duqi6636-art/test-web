package models

/// 会员等级表
type CmMemberLevel struct {
	Id            int     `gorm:"column:id;primary_key;" json:"id,omitempty"`
	Name          string  `gorm:"column:name;type:varchar(60);comment:'会员名称'" json:"name,omitempty"`
	MinMoney         float64 `gorm:"column:min_money;comment:'最小金额'" json:"min_money,omitempty"`
	MaxMoney         float64 `gorm:"column:max_money;comment:'最大金额'" json:"max_money,omitempty"`
	Ratio         float64 `gorm:"column:ratio;comment:'积分比例'" json:"ratio,omitempty"`
	UpdateTime int `gorm:"column:update_time;comment:'创建时间'" json:"update_time,omitempty"`
	Remark       string     `gorm:"column:remark;type:varchar(255);comment:'备注'" json:"remark,omitempty"`
	Admin       string     `gorm:"column:admin;type:varchar(60);comment:'操作人'" json:"admin,omitempty"`
}

/// 根据用户已花费金额获取等级
func GetMemberLevelByMoney(money float64) (data CmMemberLevel) {

	db.Model(&CmMemberLevel{}).Where("min_money <=?", money).Where("max_money >?", money).First(&data)
	return data
}

/// 查询会员等级
func GetMemberLevelById(id int) (data CmMemberLevel) {

	db.Model(&CmMemberLevel{}).Where("id <=?", id).First(&data)
	return data
}