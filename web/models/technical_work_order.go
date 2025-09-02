package models

// ListTechnicalWorkOrderRsp 工单list
type ListTechnicalWorkOrderRsp struct {
	Id           int           `json:"id"`
	Number       string        `json:"number"`
	Uid          int           `json:"uid"`
	Username     string        `json:"username"`
	Type         int           `json:"type"`
	Status       int           `json:"status"`
	DealStatus   int           `json:"deal_status"`
	Files        string        `json:"files"`
	Email        string        `json:"email"`
	Title        string        `json:"title"`
	Content      string        `json:"content"`
	CreateTime   string        `json:"create_time"`
	FileInfo     []interface{} `json:"file_info"`
	FinishTime   string        `json:"finish_time"`
	ReadStatus   int           `json:"read_status"`
	ContactType  string        `json:"contact_type"`
	ContactValue string        `json:"contact_value"`
}

type TechnicalWorkOrder struct {
	Id           int    `json:"id"`
	Number       string `json:"number"`
	Uid          int    `json:"uid"`
	Username     string `json:"username"`
	Type         int    `json:"type"`
	Status       int    `json:"status"`
	DealStatus   int    `json:"deal_status"`
	Files        string `json:"files"`
	Email        string `json:"email"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	SendStatus   int    `json:"send_status"`
	SendTime     int    `json:"send_time"`
	LastTime     int64  `json:"last_time"`
	CreateTime   int64  `json:"create_time"`
	UserIp       string `json:"user_ip"`
	UserRegion   string `json:"user_region"`
	FileInfo     string `json:"file_info"`
	FinishTime   int64  `json:"finish_time"`
	ReadStatus   int    `json:"read_status"`
	ContactType  string `json:"contact_type"`  // whatsapp/discord/others
	ContactValue string `json:"contact_value"` // 联系方式值
}

const technicalWorkOrderTableName = "technical_work_order"

func GetTechnicalWorkOrderList(query string, args ...interface{}) []TechnicalWorkOrder {
	var data = make([]TechnicalWorkOrder, 0)
	db.Table(technicalWorkOrderTableName).Where(query, args...).Order("id desc").Find(&data)
	return data
}

func CreateTechnicalWorkOrder(data *TechnicalWorkOrder) error {
	return db.Table(technicalWorkOrderTableName).Create(data).Error
}

func UpdateTechnicalWorkOrder(query string, values map[string]interface{}, args ...interface{}) error {
	err := db.Table(technicalWorkOrderTableName).Where(query, args...).Updates(values).Error
	return err
}

func GetTechnicalWorkOrder(query string, args ...interface{}) TechnicalWorkOrder {
	var data = TechnicalWorkOrder{}
	db.Table(technicalWorkOrderTableName).Where(query, args...).First(&data)
	return data
}

type TechnicalWorkOrderDetailsListRsp struct {
	Id         int    `json:"id"`
	TwoId      int    `json:"twd_id"`
	Uid        int    `json:"uid"`
	Type       int    `json:"type"`
	Files      string `json:"files"`
	Content    string `json:"content"`
	CreateTime string `json:"create_time"`
}

type TechnicalWorkOrderDetails struct {
	Id         int    `json:"id"`
	TwoId      int    `json:"twd_id"`
	Uid        int    `json:"uid"`
	Type       int    `json:"type"`
	Files      string `json:"files"`
	Content    string `json:"content"`
	CreateTime int64  `json:"create_time"`
}

const technicalWorkOrderDetailsTableName = "md_technical_work_order_details"

func CreateTechnicalWorkOrderDetails(data *TechnicalWorkOrderDetails) error {
	return db.Table(technicalWorkOrderDetailsTableName).Create(data).Error
}

func GetTechnicalWorkOrderDetailsList(query string, args ...interface{}) []TechnicalWorkOrderDetails {
	var list = make([]TechnicalWorkOrderDetails, 0)
	db.Table(technicalWorkOrderDetailsTableName).Where(query, args...).Order("id desc").Find(&list)
	return list
}
