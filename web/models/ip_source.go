package models

type IpSourceModel struct {
	Id				int 	`json:"id"`
	Ip        		string  `json:"ip"`
	Source   		string  `json:"source"`
	Code   			string  `json:"code"`
	Keyword   		string  `json:"keyword"`
	CreateTime	 	int     `json:"create_time"`
}

func AddIpSource(data IpSourceModel) (err error) {
	err = db.Table("cm_ip_source").Create(&data).Error
	return
}

func GetIpSourceBy(keyword string) (err error ,info IpSourceModel) {
	err = db.Table("cm_ip_source").Where("keyword =?",keyword).First(&info).Error
	return
}

func EditIpSource(id int,data interface{}) (err error){
	err = db.Table("cm_ip_source").Where("id =?",id).Update(data).Error
	return err
}
