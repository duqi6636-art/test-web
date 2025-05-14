package util

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// 获取当前时间戳
func GetTime(t string) string {
	now_time := time.Now().UnixNano()/1e6 + 120*1000
	if t == "13" {
		return strconv.FormatInt(now_time, 10)
	}
	now_time = time.Now().Unix()
	return strconv.FormatInt(now_time, 10)
}

// 获取当前时间 int
func GetNowInt() int {
	return int(time.Now().Unix())
}

func GetTodayHour() int {
	now := time.Now()
	// 将分钟、秒和纳秒置零，获取当前小时的起始时间
	currentHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
	// 转换为 Unix 时间戳
	timestamp := currentHour.Unix()
	return int(timestamp)
}

// 时间戳转 英文日期
func Time2DateEn(t int) string {
	now_time := time.Now()
	now_time = time.Unix(int64(t), 0)
	m := now_time.Month()
	month := m.String()
	year := now_time.Year()
	day := now_time.Day()
	date := month + " " + ItoS(day) + "," + ItoS(year)
	return date
}

// 获取当前时间年月日时分秒
func NowTimeYmd() string {
	formatLayout := "20060102030405"
	orderNo := time.Now().Format(formatLayout)
	return orderNo
}

// 获取当前时间
func GetNowTimeStr() string {
	return strconv.Itoa(int(time.Now().Unix()))
}

// 获取当前时间
func GetTimeStr(t int, formate string) string {
	now_time := time.Now()
	formateStr := "20060102150405"
	if formate == "Y" {
		formateStr = "2006"
	}
	if formate == "Ym" {
		formateStr = "200601"
	}
	if formate == "Ymd" {
		formateStr = "20060102"
	}
	if formate == "Y-m-d" {
		formateStr = "2006-01-02"
	}
	if formate == "Y.m.d" {
		formateStr = "2006.01.02"
	}
	if formate == "Y-m-d H:i" {
		formateStr = "2006-01-02 15:04"
	}
	if formate == "d/m/Y" {
		formateStr = "02/01/2006"
	}
	if formate == "d/m/Y H:i" {
		formateStr = "02/01/2006 15:04"
	}
	if formate == "d/m/Y H:i:s" {
		formateStr = "02/01/2006 15:04:05"
	}
	if formate == "Y/m/d H:i" {
		formateStr = "2006/01/02 15:04"
	}
	if formate == "Y/m/d H:i:s" {
		formateStr = "2006/01/02 15:04:05"
	}
	if formate == "Y-m-d H:i:s" {
		formateStr = "2006-01-02 15:04:05"
	}
	if formate == "d-m-Y" {
		formateStr = "02-01-2006"
	}
	if formate == "d-m-Y H:i:s" {
		formateStr = "02-01-2006 15:04:05"
	}
	if formate == "d-m-Y H:i" {
		formateStr = "02-01-2006 15:04"
	}
	if t > 0 {
		now_time = time.Unix(int64(t), 0)
	}
	return now_time.Format(formateStr)
}

func GetTime64Str(t int64, formate string) string {
	now_time := time.Now()
	formateStr := "20060102150405"
	if formate == "Ym" {
		formateStr = "200601"
	}
	if formate == "Y-m-d" {
		formateStr = "2006-01-02"
	}
	if formate == "Ymd" {
		formateStr = "20060102"
	}
	if formate == "Y-m-d H" {
		formateStr = "2006-01-02 15"
	}
	if formate == "Y-m-d H:i" {
		formateStr = "2006-01-02 15:04"
	}
	if formate == "Y/m/d H:i" {
		formateStr = "2006/01/02 15:04"
	}
	if formate == "Y/m/d" {
		formateStr = "2006/01/02"
	}
	if formate == "Y/m/d H:i:s" {
		formateStr = "2006/01/02 15:04:05"
	}
	if formate == "Y-m-d H:i:s" {
		formateStr = "2006-01-02 15:04:05"
	}
	if formate == "d-m-Y" {
		formateStr = "02-01-2006"
	}
	if formate == "d/m/Y" {
		formateStr = "2/01/2006"
	}
	if formate == "d/m/Y H:i" {
		formateStr = "2/01/2006 15:04"
	}
	if formate == "d/m/Y H:i:s" {
		formateStr = "2/01/2006 15:04:05"
	}
	if t > 0 {
		now_time = time.Unix(t, 0)
	}
	return now_time.Format(formateStr)
}

func GetTodayTime() int {
	the_time, _ := time.ParseInLocation("2006-01-02", GetTimeStr(0, "Y-m-d"), time.Local)
	return int(the_time.Unix())
}

// 获取ISO8601 时间格式
func GetIso8601Time(timestamp int64) string {
	tNow := time.Now().Unix()
	if timestamp == 0 {
		timestamp = tNow
	}

	// 将时间戳转换为time.Time类型
	t := time.Unix(timestamp, 0)

	// 使用time.Time的Format方法转换为ISO 8601格式
	iso8601 := t.Format("2006-01-02T15:04:05")

	return iso8601
}

// 获取今天0点0时0分的时间戳
func TodayStart() int {
	currentTime := time.Now()
	startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	return int(startTime.Unix())
}

// 获取今天23:59:59秒的时间戳
func TodayEnd() int {
	currentTime := time.Now()
	endTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 23, 59, 59, 0, currentTime.Location())
	return int(endTime.Unix())
}

// 根据时区 ，把A 时区的 时间日期换成B时区对映的时间戳
// timezone A时区
// YmdHis A时区时间
// timeZone2 B时区时间
func GetTimeByTimezone(timeZone, YmdHis, timeZone2 string) int64 {
	locA, err_a := time.LoadLocation(timeZone) // 加载用户时区 - 开始时间
	if err_a != nil {
		return time.Now().Unix()
	}
	timeInA, _ := time.ParseInLocation("2006-01-02 15:04:05", YmdHis, locA) //获取时区下的 时间

	locB, _ := time.LoadLocation(timeZone2) // 加载服务时区
	stamp := timeInA.In(locB).Unix()        //获取服务时区对映的时间戳
	return stamp
}

// int 秒 转时间量
func GetValTime(seconds int) map[string]interface{} {
	dv := 86400
	dh := 3600
	dm := 60
	day := seconds / dv
	hour := (seconds % dv) / dh
	min := (seconds % dh) / dm
	sec := seconds % dm
	return map[string]interface{}{
		"day":    day,
		"hour":   hour,
		"minute": min,
		"second": sec,
	}
}

// 时间日期转时间戳
func GetTimeStamp(date, formate string) (timeStamp string) {
	if formate == "" {
		formate = "Y-m-d H:i:s"
	}
	formateStr := "20060102150405"
	if formate == "Y-m-d" {
		formateStr = "2006-01-02"
	}
	if formate == "Ymd" {
		formateStr = "20060102"
	}
	if formate == "Y-m" {
		formateStr = "2006-01"
	}
	if formate == "Y-m-d H:i" {
		formateStr = "2006-01-02 15:04"
	}
	if formate == "Y-m-d H:i:s" {
		formateStr = "2006-01-02 15:04:05"
	}
	the_time, _ := time.ParseInLocation(formateStr, date, time.Local)
	times := the_time.Unix()

	if times > 0 {
		return strconv.FormatInt(times, 10)
	}
	return "0"
}

// 根据秒,自动返回计量单位
func ResAutoSec(sec int) string {
	switch {
	case sec > 0 && sec <= 60:
		return RtTimeFormat("sec", sec) + " seconds ago"
	case sec >= 60 && sec < 3600:
		return RtTimeFormat("min", sec) + " minutes ago"
	case sec >= 3600 && sec < 86400:
		return RtTimeFormat("hour", sec) + " hours ago"
	case sec >= 86400 && sec < (86400*7):
		return RtTimeFormat("day", sec) + " days ago"
	case sec >= (86400*7) && sec < (86400*30):
		return RtTimeFormat("week", sec) + " weeks ago"
	case sec >= (86400*30) && sec < (86400*365):
		return RtTimeFormat("month", sec) + " months ago"
	case sec >= (86400 * 365):
		return RtTimeFormat("year", sec) + " years ago"
	default:
		return ""
	}
}

// 根据秒,返回指定类型的计量单位
func RtTimeFormat(type_of string, sec int) string {
	var i float64
	switch type_of {
	case "year":
		i = math.Round(float64(sec / (86400 * 365)))
	case "month":
		i = math.Round(float64(sec / (86400 * 30)))
	case "week":
		i = math.Round(float64(sec / (86400 * 7)))
	case "day":
		i = math.Round(float64(sec / 86400))
	case "hour":
		i = math.Round(float64(sec / 3600))
	case "min":
		i = math.Round(float64(sec / 60))
	case "sec":
		i = math.Round(float64(sec))
	default:
		i = 0.00
	}
	return fmt.Sprintf("%0.f", i)
}

// 根据语言获取时间格式
func GetTimeByLang(t int, lang string) string {
	if lang == "zh-tw" || lang == "de" {
		return GetTimeStr(t, "Y/m/d")
	} else {
		return GetTimeStr(t, "d/m/Y")
	}
}

// 根据语言获取时间格式
func GetTimeHIByLang(t int, lang string) string {
	if lang == "zh-tw" || lang == "de" {
		return GetTimeStr(t, "Y/m/d H:i")
	} else {
		return GetTimeStr(t, "d/m/Y H:i")
	}
}

// 根据语言获取时间格式
func GetTimeHISByLang(t int, lang string) string {
	if lang == "zh-tw" || lang == "de" {
		return GetTimeStr(t, "Y/m/d H:i:s")
	} else {
		return GetTimeStr(t, "d/m/Y H:i:s")
	}
}
