package models

// Advert
type Advert struct {
	Img 			string 	`json:"img"`
	LeftImg			string 	`json:"left_img"`
	Name 			string 	`json:"name"`
	JumpUrl   		string 	`json:"jump_url"`
	JumpType  		int 	`json:"jump_type"`		//跳转类型: 1=外部页面, 2=内部页面, 3=客服
	ShowCountdown  	int 	`json:"show_countdown"`		//是否展示倒计时
	Start  			int 	`json:"start"`			//活动开始时间
	End  			int 	`json:"end"`			//活动结束时间
}

var advertTable = "cm_advert"

var field = "img,left_img,name,jump_url,jump_type"

// 获取Banner 单条
func GetAdvertInfo(lang,cate,user_type,device,place string) (ad Advert,err error) {
	dbs := db.Table(advertTable).Select(field).Where("lang =?",lang)
	if cate != "" {	//类型
		dbs = dbs.Where("cate = ?",cate)
	}
	if user_type != "" {	//用户类型
		dbs = dbs.Where("user_type =?",user_type)
	}
	if device != "" {	//设备
		dbs = dbs.Where("platform =?",device)
	}
	if place != "" {	//位置
		dbs = dbs.Where("place =?",place)
	}
	err = dbs.Where("status = ?",1).Order("sort desc,id desc").First(&ad).Error
	return
}

// 获取Banner 多条
func GetAdvertList(lang,cate ,user_type,device,place string) (ad []Advert,err error) {
	dbs := db.Table(advertTable).Select(field).Where("lang =?",lang)
	if cate != "" {	//类型
		dbs = dbs.Where("cate = ?",cate)
	}
	if user_type != "" {	//用户类型
		dbs = dbs.Where("user_type =?",user_type)
	}
	if device != "" {	//设备
		dbs = dbs.Where("platform =?",device)
	}
	if place != "" {	//位置
		dbs = dbs.Where("place =?",place)
	}
	err = dbs.Where("status = ?",1).Order("sort desc,id desc").Find(&ad).Error
	return
}

