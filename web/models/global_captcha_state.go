package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"time"
)

// 表名常量
var globalCaptchaStateTable = "global_captcha_states"

// GlobalCaptchaState 全局验证码触发状态表
// 表中仅维护一条全局记录
type GlobalCaptchaState struct {
	ID           int64   `gorm:"primaryKey;autoIncrement"`
	TriggerTime  int64   `gorm:"not null;index"`           // 触发时间
	ReleaseTime  int64   `gorm:"index;default:0"`          // 解除时间
	Status       int     `gorm:"not null;index;default:0"` // 1=触发中, 0=已解除
	TriggerCount int64   `gorm:"not null;default:0"`       // 当前触发时的登录数
	AvgCount     float64 `gorm:"not null;default:0"`       // 平均值
	Threshold    float64 `gorm:"not null;default:0"`       // 阈值
	CreateTime   int64   `gorm:"autoCreateTime"`           // 创建时间
	UpdateTime   int64   `gorm:"autoUpdateTime"`           // 更新时间
}

// InitTables 初始化数据库表结构
func InitTables() {
	db.AutoMigrate(&GlobalCaptchaState{})
}

// SetGlobalCaptchaTrigger 设置（或更新）全局验证码触发状态
// 若不存在记录则创建，存在则更新为最新触发状态
func SetGlobalCaptchaTrigger(triggerCount int64, avgCount, threshold float64) error {
	now := time.Now().Unix()
	var state GlobalCaptchaState

	// 尝试获取唯一记录
	err := db.Table(globalCaptchaStateTable).First(&state).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 若不存在则新建一条唯一记录
			state = GlobalCaptchaState{
				ID:           1,
				TriggerTime:  now,
				Status:       1,
				TriggerCount: triggerCount,
				AvgCount:     avgCount,
				Threshold:    threshold,
			}
			return db.Table(globalCaptchaStateTable).Create(&state).Error
		}
		return err
	}

	// 若存在则直接更新该记录
	return db.Table(globalCaptchaStateTable).
		Where("id = ?", state.ID).
		Updates(map[string]interface{}{
			"trigger_time":  now,
			"status":        1,
			"release_time":  0,
			"trigger_count": triggerCount,
			"avg_count":     avgCount,
			"threshold":     threshold,
			"update_time":   now,
		}).Error
}

// GetActiveCaptchaTrigger 获取当前活跃的全局验证码状态
// 若未触发则返回 nil
func GetActiveCaptchaTrigger() (*GlobalCaptchaState, error) {
	var state GlobalCaptchaState
	err := db.Table(globalCaptchaStateTable).Where("id = 1").First(&state).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if state.Status == 1 && state.ReleaseTime == 0 {
		return &state, nil
	}
	return nil, nil
}

// ReleaseCaptchaTrigger 解除全局验证码触发状态
func ReleaseCaptchaTrigger() error {
	now := time.Now().Unix()
	return db.Table(globalCaptchaStateTable).
		Where("id = 1").
		Updates(map[string]interface{}{
			"status":       0,
			"release_time": now,
			"update_time":  now,
		}).Error
}

// IsGlobalCaptchaActive 判断是否存在活跃的全局验证码触发状态
func IsGlobalCaptchaActive() (bool, error) {
	var state GlobalCaptchaState
	err := db.Table(globalCaptchaStateTable).Where("id = 1").First(&state).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return state.Status == 1 && state.ReleaseTime == 0, nil
}
