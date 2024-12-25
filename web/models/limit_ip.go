package models

type LimitIp struct {
	Ip             string `json:"ip"`
	CreateTime     int    `json:"create_time"`
	CreateTimeShow string `json:"create_time_show"`
}

func CreateLimitIp(ip LimitIp) error {
	if err := db.Create(&ip).Error; err != nil {
		return err
	}
	return nil
}

