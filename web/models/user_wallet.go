package models

type UserWalletModel struct {
	Id         		int    	`json:"id"`
	Wallet       	string 	`json:"wallet"`
	Uid             int     `json:"uid"`
	Email       	string 	`json:"email"`
	Status          int     `json:"status"`
	CreateTime 		int  	`json:"create_time"`
	Ip       		string 	`json:"ip"`
}

func AddUserWallet(model UserWalletModel) (err error) {
	err = db.Table("log_user_wallet").Create(&model).Error
	return
}
