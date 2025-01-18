package models

// 支付反馈
type FeedbackWeb struct {
	ID         int    `json:"id,omitempty"`
	Type       int    `json:"type,omitempty"` // 反馈类型 0普通反馈  1申请协助
	Cate       string `json:"cate,omitempty"`
	Email      string `json:"email,omitempty"`
	Img        string `json:"img,omitempty"`
	Content    string `json:"content,omitempty"`
	Uid        int    `json:"uid,omitempty"`
	Lang       string `json:"lang,omitempty"`
	Ip         string `json:"ip,omitempty"`
	Country    string `json:"country,omitempty"`
	Config     string `json:"config,omitempty"`
	Bandwidth  string `json:"bandwidth,omitempty"`
	Platform   string `json:"platform,omitempty"`
	CreateTime int    `json:"create_time,omitempty"`
	Occupation string `json:"occupation,omitempty"` // 职业
	Nickname   string `json:"nickname,omitempty"`   // 用户昵称
	UserEmail  string `json:"user_email,omitempty"` // 用户邮箱
}

func AddFeedBackWeb(fb FeedbackWeb) error {
	return db.Table("cm_feedback").Create(&fb).Error
}
