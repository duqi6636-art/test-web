package models

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

const (
	userTable            = "cm_users"
	RegWindowSeconds     = int64(45 * 60) // 默认注册窗口：45分钟
	RegTriggerMultiplier = 2.0            // 默认触发倍数
	RegReleaseMultiplier = 1.5            // 默认解除倍数
	RegBaselineDays      = 30             // 默认近30天基准
)

// 从数据库读取配置（如果没有则使用默认值）
func getRegWindowSeconds() int64 {
	v := GetConfigV("reg_window_seconds")
	if v == "" {
		return RegWindowSeconds
	}
	if val, err := strconv.ParseInt(v, 10, 64); err == nil && val > 0 {
		return val
	}
	return RegWindowSeconds
}

func getRegTriggerMultiplier() float64 {
	v := GetConfigV("reg_trigger_multiplier")
	if v == "" {
		return RegTriggerMultiplier
	}
	if val, err := strconv.ParseFloat(v, 64); err == nil && val > 0 {
		return val
	}
	return RegTriggerMultiplier
}

func getRegReleaseMultiplier() float64 {
	v := GetConfigV("reg_release_multiplier")
	if v == "" {
		return RegReleaseMultiplier
	}
	if val, err := strconv.ParseFloat(v, 64); err == nil && val > 0 {
		return val
	}
	return RegReleaseMultiplier
}

func getRegBaselineDays() int {
	v := GetConfigV("reg_baseline_days")
	if v == "" {
		return RegBaselineDays
	}
	if val, err := strconv.Atoi(v); err == nil && val > 0 {
		return val
	}
	return RegBaselineDays
}

//  统计最近 windowMin 分钟内指定平台的注册数量

func CountWebRegister(platform string, windowMin int64) (int64, error) {
	if platform == "" {
		platform = "web" // 默认 web 平台
	}

	now := time.Now().Unix()
	threshold := now - windowMin

	var count int64
	err := db.Table(userTable).
		Where("platform = ?", platform).
		Where("create_time >= ?", threshold).
		Count(&count).Error

	return count, err
}

// 统计最近 windowMin 分钟内指定平台的注册数量

func CountWebRegisterFromTime(platform string, startTime, window int64) (int64, error) {
	if platform == "" {
		platform = "web" // 默认 web 平台
	}

	end := startTime + window

	var count int64
	err := db.Table(userTable).
		Where("platform = ?", platform).
		Where("create_time >= ?", startTime).
		Where("create_time < ?", end).
		Count(&count).Error

	return count, err
}

// AvgRegisterSamePeriod 过去 N 天的同一时段窗口平均注册数
func AvgRegisterSamePeriod(platform string, days int, windowMin int64) (float64, error) {
	if platform == "" {
		platform = "web"
	}
	if days <= 0 {
		days = 30
	}
	if windowMin <= 0 {
		windowMin = 45 * 60 // 默认45分钟（单位：秒）
	}

	now := time.Now()
	// 注意：windowMin 为秒，需要乘以 time.Second
	windowStartToday := now.Add(-time.Duration(windowMin) * time.Second)

	// 过去 N 天最早的对齐窗口开始时间
	earliestWindowStart := windowStartToday.AddDate(0, 0, -days).Unix()

	// 为一次性查询做准备（把过去 N 天所有可能数据拉回来）
	var rows []struct {
		CreateTime int64 `gorm:"column:create_time"`
	}

	err := db.Table(userTable).
		Select("create_time").
		Where("platform = ?", platform).
		Where("create_time >= ?", earliestWindowStart).
		Find(&rows).Error

	if err != nil {
		return 0, err
	}

	// Go 层聚合
	var total int64

	for i := 1; i <= days; i++ {
		// 每天的窗口
		start := windowStartToday.AddDate(0, 0, -i).Unix()
		end := now.AddDate(0, 0, -i).Unix()
		var dayCount int64

		// 遍历 rows（单次查询的全部数据）进行统计
		for _, r := range rows {
			if r.CreateTime >= start && r.CreateTime < end {
				dayCount++
			}
		}
		total += dayCount
	}
	log.Println("total", total)
	// 平均值
	avg := float64(total) / float64(days)
	return avg, nil
}

// 检查全局注册人机验证触发条件
// 触发条件：当前45分钟注册量 > 最近30天同时间窗口平均注册量的2倍

func CheckGlobalRegisterCaptchaTrigger(platform string) (bool, string, error) {
	windowSec := getRegWindowSeconds()
	triggerMul := getRegTriggerMultiplier()
	baselineDays := getRegBaselineDays()

	currentCount, err := CountWebRegister(platform, windowSec)
	if err != nil {
		return false, "", fmt.Errorf("获取45分钟注册量失败: %v", err)
	}

	avgCount, err := AvgRegisterSamePeriod(platform, baselineDays, windowSec)
	if err != nil {
		// 若近30天没有有效数据，降级为阈值=1，避免除零
		avgCount = 1
	}
	if avgCount == 0 {
		avgCount = 1
	}

	threshold := avgCount * triggerMul
	needCaptcha := float64(currentCount) > threshold

	if needCaptcha {
		if err := SetGlobalRegCaptchaTrigger(currentCount, avgCount, threshold); err != nil {
			// 记录错误但不影响主要逻辑
			fmt.Printf("记录注册人机触发状态失败: %v\n", err)
		}
	}

	reason := fmt.Sprintf("当前%dm注册数: %d, 近%d天平均: %.2f, 触发阈值: %.2f (倍数: %.1f)",
		int(windowSec/60), currentCount, baselineDays, avgCount, threshold, triggerMul)

	return needCaptcha, reason, nil
}

// 检查全局注册人机验证解除条件
// 解除条件：触发后45分钟窗口内注册量 < 近30天该时段注册量的1.5倍

func CheckGlobalRegisterCaptchaRelease(platform string) (bool, string, error) {
	activeState, err := GetActiveRegCaptchaTrigger()
	if err != nil || activeState == nil {
		return false, "没有活跃的注册人机验证状态", nil
	}

	windowSec := getRegWindowSeconds()
	releaseMul := getRegReleaseMultiplier()
	baselineDays := getRegBaselineDays()

	triggerTime := activeState.TriggerTime
	currentCount, err := CountWebRegisterFromTime(platform, triggerTime, windowSec)
	if err != nil {
		return false, "", fmt.Errorf("获取触发后45分钟注册数失败: %v", err)
	}

	avgCount, err := AvgRegisterSamePeriod(platform, baselineDays, windowSec)
	if err != nil {
		avgCount = 1
	}
	if avgCount == 0 {
		avgCount = 1
	}

	shouldRelease := float64(currentCount) < (avgCount * releaseMul)
	if shouldRelease {
		if err := ReleaseRegCaptchaTrigger(); err != nil {
			fmt.Printf("解除注册人机验证状态失败: %v\n", err)
		}
	}

	reason := fmt.Sprintf("触发时间: %d, 触发后%dm注册数: %d, 近%d天平均: %.2f, 解除阈值(%.1f倍): %.2f",
		triggerTime, int(windowSec/60), currentCount, baselineDays, avgCount, releaseMul, avgCount*releaseMul)

	return shouldRelease, reason, nil
}
