package models

import (
	"fmt"
	"time"
)

// LoginFailure 登录失败记录模型
type LoginFailure struct {
	ID         int    `json:"id" gorm:"primary_key;auto_increment"`
	Email      string `json:"email" gorm:"type:varchar(255);not null;index"`
	IP         string `json:"ip" gorm:"type:varchar(45);not null;index"`
	UserAgent  string `json:"user_agent" gorm:"type:text"`
	Platform   string `json:"platform" gorm:"type:varchar(50);default:'web'"`
	FailReason string `json:"fail_reason" gorm:"type:varchar(255)"` // 失败原因：password_error, account_not_exist, etc.
	CreateTime int64  `json:"create_time" gorm:"not null;index"`
	Today      int    `json:"today" gorm:"not null;index"` // YYYYMMDD格式的日期
	Status     int    `json:"status" gorm:"not null;index"`
}

// AddLoginFailure 添加登录失败记录
func AddLoginFailure(failure LoginFailure) error {
	// 按月分表
	date := time.Now().Format("200601")
	tableName := "login_failure_" + date

	// 检查表是否存在，不存在则创建
	if !db.HasTable(tableName) {
		createLoginFailureTable(tableName)
	}

	// 确保status默认为1（活跃失败）
	if failure.Status == 0 {
		failure.Status = 1
	}

	return db.Table(tableName).Create(&failure).Error
}

// GetConsecutiveFailuresByIP 获取IP连续失败次数（基于status）
// status: 1=活跃失败（计入连续失败），0=已解除（不计入连续失败）
func GetConsecutiveFailuresByIP(ip string) (int, error) {
	date := time.Now().Format("200601")
	tableName := "login_failure_" + date

	if !db.HasTable(tableName) {
		return 0, nil
	}

	var count int
	err := db.Table(tableName).Where("ip = ? AND status = 1", ip).Count(&count).Error
	return count, err
}

// GetConsecutiveFailuresByEmail 获取用户连续失败次数（基于status）
// status: 1=活跃失败（计入连续失败），0=已解除（不计入连续失败）
func GetConsecutiveFailuresByEmail(email string) (int, error) {
	date := time.Now().Format("200601")
	tableName := "login_failure_" + date

	if !db.HasTable(tableName) {
		return 0, nil
	}

	var count int
	err := db.Table(tableName).Where("email = ? AND status = 1", email).Count(&count).Error
	return count, err
}

// ResetConsecutiveFailuresByEmail 重置用户连续失败状态（将status设为0）
func ResetConsecutiveFailuresByEmail(email string) error {
	date := time.Now().Format("200601")
	tableName := "login_failure_" + date

	if !db.HasTable(tableName) {
		return nil
	}

	return db.Table(tableName).Where("email = ? AND status = 1", email).Update("status", 0).Error
}

// ResetConsecutiveFailuresByIP 重置IP连续失败状态（将status设为0）
func ResetConsecutiveFailuresByIP(ip string) error {
	date := time.Now().Format("200601")
	tableName := "login_failure_" + date

	if !db.HasTable(tableName) {
		return nil
	}

	return db.Table(tableName).Where("ip = ? AND status = 1", ip).Update("status", 0).Error
}

// createLoginFailureTable 创建登录失败记录表
func createLoginFailureTable(tableName string) {
	createSQL := fmt.Sprintf(`CREATE TABLE %s (
		id int NOT NULL AUTO_INCREMENT COMMENT 'ID',
		email varchar(255) NOT NULL DEFAULT '' COMMENT '邮箱',
		ip varchar(45) NOT NULL DEFAULT '' COMMENT 'IP地址',
		user_agent text COMMENT '用户代理',
		platform varchar(50) DEFAULT 'web' COMMENT '平台',
		fail_reason varchar(255) DEFAULT '' COMMENT '失败原因',
		create_time bigint NOT NULL COMMENT '创建时间戳',
		today int NOT NULL COMMENT '日期YYYYMMDD',
		status int NOT NULL DEFAULT 1 COMMENT '状态：1=活跃失败，0=已解除',
		PRIMARY KEY (id),
		KEY idx_email (email),
		KEY idx_ip (ip),
		KEY idx_create_time (create_time),
		KEY idx_today (today),
		KEY idx_status (status),
		KEY idx_email_status (email, status),
		KEY idx_ip_status (ip, status),
		KEY idx_email_time (email, create_time),
		KEY idx_ip_time (ip, create_time)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='登录失败记录'`, tableName)

	db.Exec(createSQL)
}
