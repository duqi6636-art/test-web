package controller

import (
	"api-360proxy/web/e"
	"github.com/gin-gonic/gin"
	"time"
)

func GetChatTime(c *gin.Context) {
	now := time.Now()
	var result = map[string]interface{}{}

	// 返回完整的时间格式
	result["serverTime"] = now.Format("2006-01-02 15:04:05")

	// 分别返回小时、分钟、秒
	result["hour"] = now.Hour()     // 小时 (0-23)
	result["minute"] = now.Minute() // 分钟 (0-59)
	result["second"] = now.Second() // 秒 (0-59)

	// 额外返回一些可能有用的时间信息
	result["year"] = now.Year()                // 年
	result["month"] = int(now.Month())         // 月 (1-12)
	result["day"] = now.Day()                  // 日 (1-31)
	result["weekday"] = now.Weekday().String() // 星期几

	// 返回时间戳（毫秒）
	result["timestamp"] = now.UnixMilli()

	JsonReturn(c, e.SUCCESS, "", result)
	return
}
