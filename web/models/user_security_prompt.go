package models

import "github.com/jinzhu/gorm"

type UserSecurityPrompt struct {
	Id           int `json:"id"`
	Uid          int `json:"uid"`
	DismissCount int `json:"dismiss_count"`
	Suppressed   int `json:"suppressed"`
	UpdateTime   int `json:"update_time"`
}

var userSecurityPromptTable = "cm_user_security_prompt"

func GetUserSecurityPrompt(uid int) (info UserSecurityPrompt) {
	db.Table(userSecurityPromptTable).Where("uid = ?", uid).First(&info)
	return
}

func IncUserSecurityPromptDismiss(uid int, now int) (info UserSecurityPrompt) {
	tx := db.Table(userSecurityPromptTable).Where("uid = ?", uid).Updates(map[string]interface{}{
		"dismiss_count": gorm.Expr("dismiss_count + 1"),
		"update_time":   now,
	})
	if tx.Error != nil {
		return GetUserSecurityPrompt(uid)
	}
	if tx.RowsAffected == 0 {
		info = UserSecurityPrompt{Uid: uid, DismissCount: 1, Suppressed: 0, UpdateTime: now}
		db.Table(userSecurityPromptTable).Create(&info)
		return info
	}
	info = GetUserSecurityPrompt(uid)
	if info.DismissCount >= 3 && info.Suppressed == 0 {
		db.Table(userSecurityPromptTable).Where("uid = ?", uid).Update("suppressed", 1)
		info.Suppressed = 1
	}
	return
}

func ResetUserSecurityPrompt(uid int, now int) {
	db.Table(userSecurityPromptTable).
		Where("uid = ?", uid).
		Updates(map[string]interface{}{
			"dismiss_count": 0,
			"suppressed":    0,
			"update_time":   now,
		})
}

type UserNewPrompt struct {
	Id           int `json:"id"`
	Uid          int `json:"uid"`
	DismissCount int `json:"dismiss_count"`
	Suppressed   int `json:"suppressed"`
	UpdateTime   int `json:"update_time"`
}

var userNewPromptTable = "cm_user_prompt"

func GetUserNewPrompt(uid int) (info UserNewPrompt) {
	db.Table(userNewPromptTable).Where("uid = ?", uid).First(&info)
	return
}

func IncUserNewPromptDismiss(uid int, now int) (info UserNewPrompt) {
	tx := db.Table(userNewPromptTable).Where("uid = ?", uid).Updates(map[string]interface{}{
		"dismiss_count": gorm.Expr("dismiss_count + 1"),
		"update_time":   now,
	})
	if tx.Error != nil {
		return GetUserNewPrompt(uid)
	}
	if tx.RowsAffected == 0 {
		info = UserNewPrompt{Uid: uid, DismissCount: 1, Suppressed: 0, UpdateTime: now}
		db.Table(userNewPromptTable).Create(&info)
		// 初次关闭即抑制
		db.Table(userNewPromptTable).Where("uid = ?", uid).Update("suppressed", 1)
		info.Suppressed = 1
		return info
	}
	info = GetUserNewPrompt(uid)
	if info.DismissCount >= 1 && info.Suppressed == 0 {
		db.Table(userNewPromptTable).Where("uid = ?", uid).Update("suppressed", 1)
		info.Suppressed = 1
	}
	return
}
