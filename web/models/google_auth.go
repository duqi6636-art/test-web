package models

type UserGoogleAuth struct {
	ID              int		`json:"id"`
	Uid          	int		`json:"uid"`
	Username        string	`json:"username"`
	GoogleKey       string	`json:"google_key"`
	Create_time     int		`json:"create_time"`
}

var googleAuthTable = "cm_user_google_auth"

func GetUserGoogleAuthBy(username string)(err error,data UserGoogleAuth)  {
	err = db.Table(googleAuthTable).Where("username = ?",username).Find(&data).Error
	return
}

func CreateUserGoogleAuth(data UserGoogleAuth) (err error)  {
	err = db.Table(googleAuthTable).Create(&data).Error
	return err
}

func DeleteUserGoogleAuth(uid int) (err error)  {
	err = db.Table(googleAuthTable).Where("uid=?",uid).Delete(&UserGoogleAuth{}).Error
	return err
}


