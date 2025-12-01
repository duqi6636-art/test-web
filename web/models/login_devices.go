package models

var loginDeviceTable = "login_devices"

type ResLoginDevices struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`
	Cate       string `json:"cate"`
	Device     string `json:"device"`
	Platform   string `json:"platform"`
	Ip         string `json:"ip"`
	Country    string `json:"country"`
	State      string `json:"state"`
	City       string `json:"city"`
	Online     int    `json:"online"`
	Trust      int    `json:"trust"` //是否信任
	CreateTime string `json:"create_time"`
}

type LoginDevices struct {
	ID         int    `json:"id"`
	Uid        int    `json:"uid"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Cate       string `json:"cate"`
	Device     string `json:"device"`
	DeviceNo   string `json:"device_no"`
	Platform   string `json:"platform"`
	Ip         string `json:"ip"`
	Country    string `json:"country"`
	State      string `json:"state"`
	City       string `json:"city"`
	Session    string `json:"session"` //当前登录session信息
	Trust      int    `json:"trust"`   //设备状态 是否授信任 0 不受信任 ，大于0存时间戳，为设置时间
	UpdateTime int    `json:"update_time"`
	CreateTime int    `json:"create_time"`
}

func AddLoginDevice(user LoginDevices) (err error) {
	err = db.Table(loginDeviceTable).Create(&user).Error
	return
}

func GetLoginDeviceBy(uid int, device_no string) (err error, data LoginDevices) {
	err = db.Table(loginDeviceTable).Where("uid = ?", uid).Where("device_no = ?", device_no).Where("status = ?", 1).First(&data).Error
	return
}

func EditLoginDeviceInfo(id int, data map[string]interface{}) (err error) {
	err = db.Table(loginDeviceTable).Where("id = ?", id).Update(data).Error
	return
}

func ListLoginDevices(uid int) (data []LoginDevices) {
	db.Table(loginDeviceTable).Where("uid = ?", uid).Where("status = ?", 1).Order("id desc").Find(&data)
	return
}

func GetLoginDeviceById(id int) (data LoginDevices) {
	db.Table(loginDeviceTable).Where("id = ?", id).Order("id desc").First(&data)
	return
}

func GetLoginDeviceByIp(uid int, ip string) (err error, data LoginDevices) {
	err = db.Table(loginDeviceTable).Where("uid = ?", uid).Where("ip = ?", ip).Where("status = ?", 1).First(&data).Error
	return
}
