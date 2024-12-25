package models

import "api-360proxy/web/pkg/util"

type CmUsersAccountHistoryPassword struct {
	Id        int    `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	Uid       int    `gorm:"column:uid;default:NULL;comment:'用户的id'"`
	Password  string `gorm:"column:password;default:NULL;comment:'历史密码'"`
	CreatedAt int    `gorm:"column:created_at;default:NULL"`
	AccountId int    `gorm:"column:account_id;default:NULL;comment:'子账号id'"`
}

func GetHistoryPasswordArr(uid int) []string {
	var historyPasswordArr []CmUsersAccountHistoryPassword
	db.Table("cm_users_account_history_password").Select("password").Where("uid =?", uid).Find(&historyPasswordArr)
	var historyPasswordArrStr []string
	for _, v := range historyPasswordArr {
		historyPasswordArrStr = append(historyPasswordArrStr, v.Password)
	}
	return historyPasswordArrStr
}

func AddHistoryPassword(uid int, password string, accountId int) error {
	historyPassword := CmUsersAccountHistoryPassword{
		Uid:       uid,
		Password:  password,
		AccountId: accountId,
		CreatedAt: util.GetNowInt(),
	}
	return db.Table("cm_users_account_history_password").Create(&historyPassword).Error
}
