package models

import (
	"log"
	"time"
)

var alertRulesTable = "cm_alert_rules"

type AlertRule struct {
	ID            int       `gorm:"primaryKey" json:"id"`
	RuleKey       string    `gorm:"column:rule_key;uniqueIndex" json:"rule_key"`
	Name          string    `json:"name"`
	CronSpec      string    `json:"cron_spec"`
	Severity      string    `json:"severity"`
	SyncChannel   string    `json:"sync_channel"`
	WebhookURL    string    `gorm:"column:webhook_url" json:"webhook_url"`
	DingTalkGroup string    `json:"dingtalk_group"`
	Params        string    `json:"params"`  // JSON string for rule specific params
	ApiUrl        string    `json:"api_url"` // 接口地址
	Context       string    `json:"context"` // 内容模板
	Enabled       bool      `json:"enabled"`
	ApiURL        string    `json:"api_url"`
	SendInterval  string    `json:"send_interval"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func GetAlertRule(key string) (rule AlertRule, err error) {
	dbs := db.Table(alertRulesTable).Where("enabled = ?", 1).Where("cate = ?", 2)
	if key != "" {
		dbs = dbs.Where("rule_key = ?", key)
	}
	err = dbs.First(&rule).Error
	if err != nil {
		log.Println("load rule fail:", err)
		return
	}
	return
}
