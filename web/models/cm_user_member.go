package models

import (
	"fmt"
	"strings"
)

/// 用户会员表
type CmUserMember struct {
	Id         int     `gorm:"column:id;primary_key;" json:"id,omitempty"`
	LevelName  string  `gorm:"column:level_name;type:varchar(60);comment:'会员等级名称'" json:"level_name,omitempty"`
	LevelID    int     `gorm:"column:level_id;comment:'会员等级ID'" json:"level_id,omitempty"`
	Uid        int     `gorm:"column:uid;comment:'用户ID'" json:"uid,omitempty"`
	Email      string  `gorm:"column:email;type:varchar(100);comment:'邮箱'" json:"email,omitempty"`
	Username   string  `gorm:"column:username;type:varchar(60);comment:'用户名'" json:"username,omitempty"`
	TotalMoney float64 `gorm:"column:total_money;type:float(20,2);comment:'累计购买金额'" json:"total_money,omitempty"`
	TotalTime  int     `gorm:"column:total_time;comment:'累计购买次数'" json:"total_time,omitempty"`
	Enable     int     `gorm:"column:enable;comment:'是否可用 1-可用'" json:"enable,omitempty"`
	CreateTime int     `gorm:"column:create_time;comment:'创建时间'" json:"create_time,omitempty"`
	UpdateTime int     `gorm:"column:update_time;comment:'最后更新时间'" json:"update_time,omitempty"`
	Remark     string  `gorm:"column:remark;type:varchar(255);comment:'备注'" json:"remark,omitempty"`
	Admin      string  `gorm:"column:admin;type:varchar(60);comment:'操作人'" json:"admin,omitempty"`
}

/// 根据用户ID查询用户会员
func GetUserMemberByUid(uid int) (data CmUserMember) {

	db.Model(&CmUserMember{}).Where("uid = ?", uid).Where("enable = ?", 1).First(&data)
	return
}

/// 添加会员
func AddUserMember(data CmUserMember) error {

	return db.Model(&CmUserMember{}).Create(&data).Error
}

/// 更新会员信息
func UpdateUserMember(uid int, info CmUserMember) error {

	return db.Model(&CmUserMember{}).Where("uid = ?", uid).Update(info).Error
}

// 批量添加会员
func AddUserMembers(valueArgs []string) (err error) {

	tableName := "cm_user_member"
	sql := "insert into " + tableName + " (`level_name`,`level_id`,`uid`,`email`,`username`,`total_money`,`total_time`,`enable`,`create_time`,`update_time`) values "

	if len(valueArgs) > 0 {
		smt := fmt.Sprintf("%s %s", sql, strings.Join(valueArgs, ","))
		err = db.Exec(smt).Error
	}
	return
}
