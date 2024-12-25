package models

import "api-360proxy/web/pkg/util"

type EmailCode struct {
	ID            int    `json:"id"`
	EmailCode     string `json:"email_code"`
	Email         string `json:"email"`
	Type          string `json:"type"`
	Enable        string `json:"enable"`
	ExpireTime    int    `json:"expire_time"`
	UpdateTime    int    `json:"update_time"`
	CreateTime    int    `json:"create_time"`
	Ip			  string `json:"ip"`
}

var emailCodeTable = "cm_email_code"
func AddVerifyCode(vCode EmailCode) bool {
	err := db.Table(emailCodeTable).Create(&vCode).Error
	return err == nil
}

func GetAvailableVerifyCode(account, types string) (int, string) {
	var vc EmailCode
	err := db.Table(emailCodeTable).Select("id,email_code").Where("email = ? and type = ? and expire_time >= ? and enable = true",  account, types, util.GetNowInt()).First(&vc).Error
	if err != nil {
		return 0, ""
	}
	return vc.ID, vc.EmailCode
}

func GetVerifyCountByEmail(account string,start int) (num int) {
	dbRead.Table(emailCodeTable).Where("email = ?",  account).Where("create_time >= ?",start).Count(&num)
	return
}
func GetVerifyCountByIp(ip string,start int) (num int) {
	dbRead.Table(emailCodeTable).Where("ip = ?",  ip).Where("create_time >= ?",start).Count(&num)
	return
}

func CheckVerifyCode(code,  account, types string)  (int, string)  {
	start := util.GetNowInt()
	vc := EmailCode{}
	err := db.Table(emailCodeTable).Where("email_code = ? and email = ? and type = ? ", code,  account, types).Where("expire_time >= ?",start).Where("enable=?","true").First(&vc).Error
	if err != nil {
		return 0, ""
	}
	return vc.ID, vc.EmailCode
}

func UpdateCodeStatus(where interface{}) bool {
	vc := EmailCode{}
	now := util.GetNowInt()
	vc.UpdateTime = now
	vc.Enable = "false"
	err := db.Table(emailCodeTable).Where(where).Update(&vc).Error
	return err == nil
}
func UpdateVerifyCodeById(id int, maps map[string]interface{}) bool {
	return db.Model(&EmailCode{}).Where("id = ? ", id).Update(maps).Error == nil
}


type LogEmail struct {
	Id         int    `json:"id"`
	Email      string `json:"email"`       // 用户邮箱
	Param      string `json:"param"`       // 参数
	TplSn      string `json:"tpl_sn"`      // 模板标识
	Result     string `json:"result"`      // 结果
	Ip     	   string `json:"ip"`      // 结果
	CreateTime int    `json:"create_time"` // 操作时间
}

func AddLogEmail(email,param,tplSn,result,ip string) {
	nowTime := util.GetNowInt()
	verMap := LogEmail{
		Email			:email,
		Param			:param,
		TplSn			:tplSn,
		Result			:result,
		Ip				:ip,
		CreateTime		:nowTime,
	}
	db.Table("cm_log_email").Create(&verMap)
	return
}

