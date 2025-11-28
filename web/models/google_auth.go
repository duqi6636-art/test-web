package models

type UserGoogleAuth struct {
	ID          int    `json:"id"`
	Uid         int    `json:"uid"`
	GoogleKey   string `json:"google_key"`
	Username    string `json:"username"`
	Cate        string `json:"cate"`
	IsOpen      int    `json:"is_open"`
	Create_time int    `json:"create_time"`
}

var googleAuthTable = "cm_user_google_auth"

func GetUserAuthByUsername(username, cate string) (err error, data UserGoogleAuth) {
	err = db.Table(googleAuthTable).Where("username = ?", username).Where("cate = ?", cate).Find(&data).Error
	return
}

func GetUserGoogleAuthBy(username string) (err error, data UserGoogleAuth) {
	err = db.Table(googleAuthTable).Where("username = ?", username).Find(&data).Error
	return
}

func CreateUserGoogleAuth(data UserGoogleAuth) (err error) {
	err = db.Table(googleAuthTable).Create(&data).Error
	return err
}

func DeleteUserGoogleAuth(uid int) (err error) {
	err = db.Table(googleAuthTable).Where("uid=?", uid).Delete(&UserGoogleAuth{}).Error
	return err
}

func GetAuthByUid(uid int) (data []UserGoogleAuth, err error) {
	err = db.Table(googleAuthTable).Where("uid = ?", uid).Find(&data).Error
	return
}

func EditAuthById(id int, info interface{}) bool {
	err := db.Table(googleAuthTable).Where("id = ?", id).Updates(info).Error
	if err == nil {
		return true
	}
	return false
}

func GetUserAuthByUid(uid int, cate string) (err error, data UserGoogleAuth) {
	err = db.Table(googleAuthTable).Where("uid = ?", uid).Where("cate = ?", cate).Find(&data).Error
	return
}

func DeleteUserGoogleAuthByCate(uid int, cate string) (err error) {
	err = db.Table(googleAuthTable).Where("uid=?", uid).Where("cate=?", cate).Delete(&UserGoogleAuth{}).Error
	return err
}
