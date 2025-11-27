package models

import "api-360proxy/web/pkg/util"

type Session struct {
	ID          int    `json:"id"`
	SessionId   string `json:"session_id"`
	Enable      string `json:"enable"`
	Uid         int    `json:"uid"`
	Username    string `json:"username"`
	LoginIp     string `json:"login_ip"`
	LoginTime   int    `json:"login_time"`
	Platform    string `json:"platform"`
	Remark      string `json:"remark"`
	ExpireTime  int    `json:"expire_time"`
	Version     string `json:"version"`
	DeviceToken string `json:"device_token"`
	UpdateTime  int    `json:"update_time"`
}

func AddLoginSession(ls Session) bool {

	err := db.Table("cm_session").Create(&ls).Error
	if err != nil {
		return false
	}
	return true
}

func DeleteSession(where interface{}) error {
	return db.Table("cm_session").Where(where).Delete(&Session{}).Error
}

func GetSessionBySn(sn string) (log Session, err error) {
	err = db.Table("cm_session").Where("session_id = ?", sn).Where("expire_time >= ?", util.GetNowInt()).First(&log).Error
	return
}

func GetSessionInfo(where map[string]interface{}) (err error, ses Session) {
	err = db.Table("cm_session").Where(where).First(&ses).Error
	return
}

func GetSessionByUsername(username, ip, platform string) (log Session, err error) {
	db1 := db.Table("cm_session").
		Where("username =?", username).
		Where("login_ip =?", ip)
	if platform != "" {
		db1 = db1.Where("platform =?", platform)
	}
	db1 = db1.Where("expire_time >= ?", util.GetNowInt()).
		First(&log)
	return
}
