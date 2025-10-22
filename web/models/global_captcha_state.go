package models

import "time"

// GlobalCaptchaState 全局人机验证状态模型
type GlobalCaptchaState struct {
	ID           int     `json:"id" gorm:"primary_key;auto_increment"`
	TriggerTime  int64   `json:"trigger_time" gorm:"not null;index"` // 触发时间戳
	ReleaseTime  int64   `json:"release_time" gorm:"index"`          // 解除时间戳，0表示未解除
	Status       int     `json:"status" gorm:"not null;index"`       // 状态：1=已触发，0=已解除
	TriggerCount int64   `json:"trigger_count" gorm:"not null"`      // 触发时的登录次数
	AvgCount     float64 `json:"avg_count" gorm:"not null"`          // 触发时的30天平均次数
	Threshold    float64 `json:"threshold" gorm:"not null"`          // 触发阈值
	CreateTime   int64   `json:"create_time" gorm:"not null;index"`
	UpdateTime   int64   `json:"update_time" gorm:"not null"`
}

// createGlobalCaptchaStateTable 创建全局人机验证状态表
func createGlobalCaptchaStateTable() {
	if !db.HasTable("global_captcha_state") {
		db.CreateTable(&GlobalCaptchaState{})
	}
}

// SetGlobalCaptchaTrigger 设置全局人机验证触发状态
func SetGlobalCaptchaTrigger(triggerCount int64, avgCount, threshold float64) error {
	// 确保表存在
	createGlobalCaptchaStateTable()

	now := time.Now().Unix()

	// 检查是否已经有活跃的触发状态
	var existingState GlobalCaptchaState
	err := db.Where("status = ? AND release_time = 0", 1).First(&existingState).Error

	if err == nil {
		// 已存在活跃状态，更新触发时间
		return db.Model(&existingState).Updates(map[string]interface{}{
			"trigger_time":  now,
			"trigger_count": triggerCount,
			"avg_count":     avgCount,
			"threshold":     threshold,
			"update_time":   now,
		}).Error
	} else {
		// 创建新的触发状态
		state := GlobalCaptchaState{
			TriggerTime:  now,
			ReleaseTime:  0,
			Status:       1,
			TriggerCount: triggerCount,
			AvgCount:     avgCount,
			Threshold:    threshold,
			CreateTime:   now,
			UpdateTime:   now,
		}
		return db.Create(&state).Error
	}
}

// GetActiveCaptchaTrigger 获取当前活跃的人机验证触发状态
func GetActiveCaptchaTrigger() (*GlobalCaptchaState, error) {
	// 确保表存在
	createGlobalCaptchaStateTable()

	var state GlobalCaptchaState
	err := db.Where("status = ? AND release_time = 0", 1).First(&state).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// ReleaseCaptchaTrigger 解除全局人机验证状态
func ReleaseCaptchaTrigger() error {
	// 确保表存在
	createGlobalCaptchaStateTable()

	now := time.Now().Unix()

	// 更新所有活跃状态为已解除
	return db.Model(&GlobalCaptchaState{}).
		Where("status = ? AND release_time = 0", 1).
		Updates(map[string]interface{}{
			"status":       0,
			"release_time": now,
			"update_time":  now,
		}).Error
}

// IsGlobalCaptchaActive 检查全局人机验证是否处于活跃状态
func IsGlobalCaptchaActive() (bool, error) {
	state, err := GetActiveCaptchaTrigger()
	if err != nil {
		return false, nil // 没有活跃状态视为未触发
	}
	return state != nil, nil
}
