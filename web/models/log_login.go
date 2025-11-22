package models

import (
	"api-360proxy/web/constants"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type LogLogin struct {
	ID           int32  `json:"id"`
	UID          int    `json:"uid"`
	UserLogin    string `json:"user_login"`
	Ip           string `json:"ip"`
	LoginTime    int64  `json:"login_time"`
	Platform     string `json:"platform"`
	Browser      string `json:"browser"`
	OsInfo       string `json:"os_info"`
	Version      string `json:"version"`
	RegTime      int    `json:"reg_time"`
	Country      string `json:"country"`
	Language     string `json:"language"`
	Lang         string `json:"lang"`
	Today        int    `json:"today"`
	Cate         string `json:"cate"`
	TimeZone     string `json:"time_zone"`
	DeviceNumber string `json:"device_number"`
	UserAgent    string `json:"user_agent"`
	IsPay        int    `json:"is_pay"` //是否是已购买
}

func AddLoginLog(info LogLogin) bool {
	var tableName = "log_login" + time.Now().Format("200601")
	if !db.HasTable(tableName) {
		createLoginLogTable(tableName)
	}
	return db.Table(tableName).Create(&info).Error == nil
}

// 创建表
func createLoginLogTable(tableName string) {
	createTable := `CREATE TABLE ` + tableName + `(
		id int NOT NULL AUTO_INCREMENT COMMENT 'ID',
  		uid int DEFAULT '0' COMMENT '用户id',
  		user_login varchar(50) DEFAULT '' COMMENT '登录用户名',
  		ip varchar(60) DEFAULT '' COMMENT '登录ip',
  		login_time int DEFAULT '0' COMMENT '登录时间',
  		platform varchar(30) DEFAULT '' COMMENT '登录终端',
  		cate varchar(50) DEFAULT '' COMMENT '日志类型',
  		browser varchar(255) DEFAULT '' COMMENT '浏览器信息',
		user_agent varchar(255) DEFAULT '' COMMENT '浏览器UA',
  		os_info varchar(255) DEFAULT '' COMMENT '系统信息',
  		version varchar(50) DEFAULT '' COMMENT '版本',
  		channel varchar(50) DEFAULT '' COMMENT '渠道',
  		country varchar(255) DEFAULT '' COMMENT 'IP所在国家',
  		language varchar(255) DEFAULT '' COMMENT '电脑语言',
  		lang varchar(50) DEFAULT '' COMMENT '客户端语言',
  		time_zone varchar(255) DEFAULT '' COMMENT '时区',
  		today int(11) NOT NULL DEFAULT '0' COMMENT '零时时间戳',
  		is_pay int(11) NOT NULL DEFAULT '0' COMMENT '是否付费',
  		reg_time int DEFAULT '0' COMMENT '注册时间',
		device_number varchar(100) DEFAULT NULL COMMENT '设备号',
  		PRIMARY KEY (id) USING BTREE,
  		KEY uid (uid) USING BTREE,
  		KEY country (country) USING BTREE,
		KEY user_login (user_login) USING BTREE,
		KEY login_time (login_time) USING BTREE,
  		KEY today (today)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='登录日志'`
	db.Exec(createTable)
}

// 查询当日登陆数据
//func GettodayLogin(uid int) int {
//	today := util.GetTodayTime()
//	num := 0
//	db.Table(tableName).Where("today = ? and uid = ?", today, uid).Count(&num)
//	return num
//}

// GetLoginCountInTimeWindow 获取指定时间窗口内的登录次数
// timeWindow: 时间窗口（秒），例如 600 表示10分钟
func GetLoginCountInTimeWindow(timeWindow int64) (int64, error) {
	now := time.Now()
	startTime := now.Unix() - timeWindow
	endTime := now.Unix()

	// 获取当前月份的表名
	tableName := "log_login" + now.Format("200601")

	if !db.HasTable(tableName) {
		return 0, nil
	}

	log.Println("startTime", startTime, "endTime", endTime)

	var count int64
	err := db.Table(tableName).
		Where("login_time >= ? AND login_time <= ?", startTime, endTime).
		Where("cate = ?", "login"). // 只统计登录，不包括自动登录和退出
		Count(&count).Error
	return count, err
}

// GetAverage20MinLoginCountForLast30Days 获取近30天同一20分钟时间段的平均登录次数
func GetAverage20MinLoginCountForLast30Days() (float64, error) {
	now := time.Now()

	// 计算当前时间在一天中的20分钟段位置
	// 一天有 24*60/20 = 72 个20分钟段
	currentMinute := now.Hour()*60 + now.Minute()
	current20MinSlot := currentMinute / 20

	var totalCount int64

	// 遍历过去30天
	for i := 1; i < 30; i++ {
		targetDate := now.AddDate(0, 0, -i)
		tableName := "log_login" + targetDate.Format("200601")

		if !db.HasTable(tableName) {
			continue
		}

		// 计算目标日期的20分钟段开始和结束时间
		startMinute := current20MinSlot * 20
		endMinute := startMinute + 20

		dayStart := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
			startMinute/60, startMinute%60, 0, 0, targetDate.Location())
		dayEnd := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
			endMinute/60, endMinute%60, 0, 0, targetDate.Location())

		var dayCount int64
		err := db.Table(tableName).
			Where("login_time >= ? AND login_time < ?", dayStart.Unix(), dayEnd.Unix()).
			Where("cate = ?", "login").
			Count(&dayCount).Error

		if err != nil {
			continue
		}

		totalCount += dayCount
	}
	return float64(totalCount), nil
}

// CheckGlobalLoginCaptchaTrigger 检查全局登录人机验证触发条件
// 返回值：needCaptcha bool, reason string, error
func CheckGlobalLoginCaptchaTrigger() (bool, string, error) {
	// 获取20分钟内的登录次数
	timeWindowStr := GetConfigV(constants.ConfigKeyGlobalTimeWindow)
	timeWindow := constants.DefaultGlobalTimeWindow // 使用常量默认值
	if timeWindowStr != "" {
		if val, err := strconv.Atoi(timeWindowStr); err == nil && val > 0 {
			timeWindow = val
		}
	}
	currentCount, err := GetLoginCountInTimeWindow(int64(timeWindow))
	if err != nil {
		return false, "", fmt.Errorf("获取20分钟内登录次数失败: %v", err)
	}

	// 获取近30天同一20分钟时间段的总登录次数
	avgCount, err := GetAverage20MinLoginCountForLast30Days()
	if err != nil {
		return false, "", fmt.Errorf("获取近30天平均登录次数失败: %v", err)
	}

	// 如果近30天没有数据，使用默认阈值
	if avgCount == 0 {
		avgCount = 1
		//return false, "", nil
	}

	// 触发条件：20分钟内全局登录次数大于近30天同一20分钟段的三倍
	threshold := avgCount * 3
	needCaptcha := float64(currentCount) > threshold

	// 如果需要触发人机验证，记录触发状态
	if needCaptcha {
		err = SetGlobalCaptchaTrigger(currentCount, avgCount, threshold)
		if err != nil {
			// 记录错误但不影响主要逻辑
			fmt.Printf("记录全局人机验证触发状态失败: %v\n", err)
		}
		minutes := timeWindow / 60
		runtime := map[string]any{
			"minutes": minutes,
		}
		fallbackTpl := fmt.Sprintf("紧急：【cherry】当前【%d】分钟内登录接口次数异常，请立即查看！", minutes)
		SendProductAlertWithRule("global_login_trigger", runtime, fallbackTpl)
		log.Println("needCaptcha 为 ture")
	}

	reason := fmt.Sprintf("当前20分钟登录次数: %d, 近30天平均: %.2f, 触发阈值: %.2f",
		currentCount, avgCount, threshold)

	return needCaptcha, reason, nil
}

// GetLoginCountFromTime 获取从指定时间开始的时间窗口内的登录次数
func GetLoginCountFromTime(startTime int64, timeWindow int64) (int64, error) {
	endTime := startTime + timeWindow
	now := time.Now().Unix()

	// 如果结束时间超过当前时间，则以当前时间为准
	if endTime > now {
		endTime = now
	}

	// 获取当前月份的表名
	tableName := "log_login" + time.Unix(startTime, 0).Format("200601")

	if !db.HasTable(tableName) {
		return 0, nil
	}

	log.Println("GetLoginCountFromTime - startTime", startTime, "endTime", endTime)

	var count int64
	err := db.Table(tableName).
		Where("login_time >= ? AND login_time <= ?", startTime, endTime).
		Where("cate = ?", "login"). // 只统计登录，不包括自动登录和退出
		Count(&count).Error
	return count, err
}

// GetAvgLoginCountForSame20MinSlot 获取过去 N 天同一 20 分钟时间段的平均登录次数
func GetAvgLoginCountForSame20MinSlot(lastDays int) (float64, error) {
	now := time.Now()

	// 当前时间在一天中的第几个 20 分钟段（共 72 段）
	currentMinute := now.Hour()*60 + now.Minute()
	currentSlot := currentMinute / 20 // 第几个 20 分钟段

	var totalCount int64
	var validDays int // 实际成功查询的天数

	for i := 1; i <= lastDays; i++ {
		targetDate := now.AddDate(0, 0, -i)
		tableName := "log_login" + targetDate.Format("200601")

		// 判断表是否存在
		if !db.HasTable(tableName) {
			continue
		}

		// 计算当天该时间段的开始和结束时间
		startMinute := currentSlot * 20
		endMinute := startMinute + 20

		startTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
			startMinute/60, startMinute%60, 0, 0, targetDate.Location())
		endTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
			endMinute/60, endMinute%60, 0, 0, targetDate.Location())

		var count int64
		err := db.Table(tableName).
			Where("login_time >= ? AND login_time < ?", startTime.Unix(), endTime.Unix()).
			Where("cate = ?", "login").
			Count(&count).Error

		if err != nil {
			continue
		}

		totalCount += count
		validDays++
	}

	// 避免除零错误
	if validDays == 0 {
		return 0, fmt.Errorf("no valid data found in last %d days", lastDays)
	}

	// 平均值
	avg := float64(totalCount) / float64(validDays)
	return avg, nil
}

// CheckGlobalLoginCaptchaRelease 检查全局登录人机验证解除条件
// 返回值：shouldRelease bool, reason string, error
func CheckGlobalLoginCaptchaRelease() (bool, string, error) {
	// 首先检查是否有活跃的触发状态
	activeState, err := GetActiveCaptchaTrigger()
	if err != nil || activeState == nil {
		// 没有活跃的触发状态，无需解除
		return false, "没有活跃的全局人机验证状态", nil
	}

	// 检查触发时间后的20分钟内的登录次数
	now := time.Now().Unix()
	triggerTime := activeState.TriggerTime

	// 计算从触发时间开始的20分钟时间窗口
	timeWindowStr := GetConfigV(constants.ConfigKeyGlobalTimeWindow)
	timeWindow := int64(constants.DefaultResetTimeWindow) // 20分钟 = 1200秒
	if timeWindowStr != "" {
		if val, err := strconv.Atoi(timeWindowStr); err == nil && val > 0 {
			timeWindow = int64(val)
		}
	}

	// 获取从触发时间开始的20分钟内的登录次数
	currentCount, err := GetLoginCountFromTime(triggerTime, timeWindow)
	if err != nil {
		return false, "", fmt.Errorf("获取触发时间后20分钟内登录次数失败: %v", err)
	}

	// 获取近30天同一20分钟时间段的平均登录次数
	avgCount, err := GetAvgLoginCountForSame20MinSlot(30)
	if err != nil {
		return false, "", fmt.Errorf("获取近30天平均登录次数失败: %v", err)
	}

	// 如果近30天没有数据，使用默认阈值
	if avgCount == 0 {
		avgCount = 1
	}

	// 解除条件：触发时间后20分钟内全局登录次数小于近30天同时段平均登录次数
	shouldRelease := float64(currentCount) < avgCount

	// 如果满足解除条件，更新状态
	if shouldRelease {
		err = ReleaseCaptchaTrigger()
		if err != nil {
			fmt.Printf("解除全局人机验证状态失败: %v\n", err)
		}
	}

	reason := fmt.Sprintf("触发时间: %d, 当前时间: %d, 触发后20分钟内登录次数: %d, 近30天平均: %.2f",
		triggerTime, now, currentCount, avgCount)

	return shouldRelease, reason, nil
}

type dingMsgV struct {
	MsgType string                 `json:"msgtype"`
	Text    map[string]string      `json:"text"`
	At      map[string]interface{} `json:"at"`
}

// 统一的产品侧规则驱动预警，支持模板与回退

func SendProductAlertWithRule(ruleKey string, runtime map[string]any, fallbackTpl string) {
	rule, ruleErr := GetAlertRule(ruleKey)
	if ruleErr == nil && rule.ID > 0 && strings.TrimSpace(rule.WebhookURL) != "" {
		msg, renderErr := RenderMessage(strings.TrimSpace(rule.Context), runtime)
		if renderErr != nil || strings.TrimSpace(msg) == "" {
			msg = fallbackTpl
		}
		if sendErr := SendDingTalkURL(strings.TrimSpace(rule.WebhookURL), msg); sendErr != nil {
			fmt.Println(sendErr)
		}
	}
}

// 直接使用完整 webhook URL 发送；若提供 secret 则自动加签

func SendDingTalkURL(webhookURL, content string) error {
	url := webhookURL
	isAtAll := false
	phoneArr := []string{}

	atArr := map[string]interface{}{
		"atMobiles": phoneArr,
		"isAtAll":   isAtAll,
	}
	body := dingMsgV{
		MsgType: "text",
		Text:    map[string]string{"content": content},
		At:      atArr,
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// 将 rule.Context 作为模板

func RenderMessage(contextTpl string, runtimeVars map[string]interface{}) (string, error) {
	data := map[string]interface{}{}
	if runtimeVars != nil {
		for k, v := range runtimeVars {
			data[k] = v
		}
	}
	tplText := strings.TrimSpace(contextTpl)
	tpl, err := template.New("alert").Option("missingkey=zero").Parse(tplText)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
