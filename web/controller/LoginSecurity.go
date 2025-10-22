package controller

import (
	"api-360proxy/web/constants"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"time"
)

// LoginSecurityConfig 登录安全配置结构
type LoginSecurityConfig struct {
	// 全局限制配置
	GlobalTimeWindow   int64 `json:"global_time_window"`   // 全局时间窗口（秒）
	GlobalFailureLimit int   `json:"global_failure_limit"` // 全局失败次数限制

	// 用户级限制配置
	UserTimeWindow   int64 `json:"user_time_window"`   // 用户时间窗口（秒）
	UserFailureLimit int   `json:"user_failure_limit"` // 用户失败次数限制

	// IP级限制配置
	IPTimeWindow   int64 `json:"ip_time_window"`   // IP时间窗口（秒）
	IPFailureLimit int   `json:"ip_failure_limit"` // IP失败次数限制
}

// GetLoginSecurityConfig 获取登录安全配置
func GetLoginSecurityConfig() LoginSecurityConfig {
	config := LoginSecurityConfig{
		// 使用常量定义的默认配置值
		GlobalTimeWindow:   constants.DefaultGlobalTimeWindow,
		GlobalFailureLimit: constants.DefaultGlobalFailureLimit,
		UserTimeWindow:     constants.DefaultUserTimeWindow,
		UserFailureLimit:   constants.DefaultUserFailureLimit,
		IPTimeWindow:       constants.DefaultIPTimeWindow,
		IPFailureLimit:     constants.DefaultIPFailureLimit,
	}

	// 从数据库配置中读取，如果存在的话
	if val := models.GetConfigVal(constants.ConfigKeyGlobalTimeWindow); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			config.GlobalTimeWindow = parsed
		}
	}

	if val := models.GetConfigVal(constants.ConfigKeyGlobalFailureLimit); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.GlobalFailureLimit = parsed
		}
	}

	if val := models.GetConfigVal(constants.ConfigKeyUserTimeWindow); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			config.UserTimeWindow = parsed
		}
	}

	if val := models.GetConfigVal(constants.ConfigKeyUserFailureLimit); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.UserFailureLimit = parsed
		}
	}

	if val := models.GetConfigVal(constants.ConfigKeyIPTimeWindow); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			config.IPTimeWindow = parsed
		}
	}

	if val := models.GetConfigVal(constants.ConfigKeyIPFailureLimit); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.IPFailureLimit = parsed
		}
	}

	return config
}

// RecordLoginFailure 记录登录失败
func RecordLoginFailure(c *gin.Context, email, failReason string) {
	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()
	platform := c.DefaultPostForm("platform", "web")

	failure := models.LoginFailure{
		Email:      email,
		IP:         ip,
		UserAgent:  userAgent,
		Platform:   platform,
		FailReason: failReason,
		CreateTime: time.Now().Unix(),
		Today:      util.GetTodayTime(),
	}

	// 记录失败日志
	models.AddLoginFailure(failure)
}

// CheckLoginSecurity 检查登录安全限制（基于连续失败次数）
// 返回值：needCaptcha bool, reasons []string, error
func CheckLoginSecurity(email, ip string) (bool, []string, error) {
	config := GetLoginSecurityConfig()
	var result = []string{}
	needCaptcha := false

	// 1. 检查全局登录人机验证触发条件（基于20分钟登录次数与近30天同时段平均值比较）
	globalNeedCaptcha, globalReason, err := models.CheckGlobalLoginCaptchaTrigger()
	if err != nil {
		return false, nil, fmt.Errorf("检查全局登录人机验证触发条件失败: %v", err)
	}

	if globalNeedCaptcha {
		needCaptcha = true
		result = append(result, "global")
		// 可选：记录触发原因到日志
		AddLogs("CheckLoginSecurity", globalReason)
	}

	// 2. 检查用户级连续失败次数（基于status字段）
	emailFailures, err := models.GetConsecutiveFailuresByEmail(email)
	if err != nil {
		return false, nil, fmt.Errorf("检查用户连续失败次数失败: %v", err)
	}

	if emailFailures >= config.UserFailureLimit {
		needCaptcha = true
		result = append(result, "user")
	}

	// 3. 检查IP级连续失败次数（基于status字段）
	ipFailures, err := models.GetConsecutiveFailuresByIP(ip)
	if err != nil {
		return false, nil, fmt.Errorf("检查IP连续失败次数失败: %v", err)
	}

	if ipFailures >= config.IPFailureLimit {
		needCaptcha = true
		result = append(result, "ip")
	}

	return needCaptcha, result, nil
}

// ResetLoginFailureStateByType 根据类型重置登录失败状态
// verifyTypes: 支持的类型 "user", "ip", "global"
// 可以传入多个类型，用逗号分隔，如 "user,ip"
func ResetLoginFailureStateByType(c *gin.Context, email string, verifyTypes string) {
	ip := c.ClientIP()

	// 解析重置类型
	types := strings.Split(strings.ToLower(verifyTypes), ",")

	for _, resetType := range types {
		resetType = strings.TrimSpace(resetType)

		switch resetType {
		case "user":
			// 只重置用户级失败状态
			err := models.ResetConsecutiveFailuresByEmail(email)
			if err != nil {
				AddLogs("ResetConsecutiveFailuresByEmail", fmt.Sprintf("重置用户级登录失败状态失败: %s", err.Error()))
			}

		case "ip":
			// 只重置IP级失败状态
			err := models.ResetConsecutiveFailuresByIP(ip)
			if err != nil {
				AddLogs("ResetConsecutiveFailuresByIP", fmt.Sprintf("重置IP级登录失败状态失败: %s", err.Error()))
			}
		case "check_release":
			// 检查全局人机验证解除条件
			shouldRelease, reason, err := models.CheckGlobalLoginCaptchaRelease()
			if err != nil {
				AddLogs("CheckGlobalLoginCaptchaRelease", fmt.Sprintf("检查全局人机验证解除条件失败: %s", err.Error()))
			} else if shouldRelease {
				AddLogs("GlobalCaptchaReleased", fmt.Sprintf("全局人机验证已解除: %s", reason))
			} else {
				AddLogs("GlobalCaptchaNotReleased", fmt.Sprintf("全局人机验证解除条件未满足: %s", reason))
			}
		case "debug":
			// 显示当前登录统计信息（用于调试）
			// 从配置中获取时间窗口（秒），默认1200秒（20分钟）
			timeWindowStr := models.GetConfigV(constants.ConfigKeyLoginResetTimeWindow)
			timeWindow := 1200 // 默认值：10分钟
			if timeWindowStr != "" {
				if val, err := strconv.Atoi(timeWindowStr); err == nil && val > 0 {
					timeWindow = val
				}
			}

			recentCount, err1 := models.GetLoginCountInTimeWindow(int64(timeWindow))
			avgCount, err2 := models.GetAverage20MinLoginCountForLast30Days()
			globalNeedCaptcha, _, err3 := models.CheckGlobalLoginCaptchaTrigger()

			fmt.Printf("=== 登录安全统计信息 ===\n")
			fmt.Printf("配置的时间窗口: %d秒 (%.1f分钟)\n", timeWindow, float64(timeWindow)/60)
			if err1 != nil {
				fmt.Printf("获取%d秒内登录次数失败: %v\n", timeWindow, err1)
			} else {
				fmt.Printf("%d秒内登录次数: %d\n", timeWindow, recentCount)
			}

			if err2 != nil {
				fmt.Printf("获取30天平均登录次数失败: %v\n", err2)
			} else {
				fmt.Printf("近30天同时段平均登录次数: %.2f\n", avgCount)
			}

			if err3 != nil {
				fmt.Printf("判断重置条件失败: %v\n", err3)
			} else {
				fmt.Printf("是否应该触发全局重置: %t\n", globalNeedCaptcha)
			}
			fmt.Printf("========================\n")
		default:
			fmt.Printf("不支持的重置类型: %s\n", resetType)
		}
	}
}
