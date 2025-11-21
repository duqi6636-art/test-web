package crons

import (
	"api-360proxy/web/pkg/setting"
	"github.com/robfig/cron"
)

// 执行计划任务
func GoCron() {
	// 创建一个 Cron 对象
	c := cron.New()

	if setting.AppConfig.SendEmailSwitch == 1 {
		_ = c.AddFunc("5 */10 * * * *", func() { //每10分钟的第5秒执行
			CheckDoSending()        // 设置 用户 发送邮件
			CheckLongIspDoSending() // 长效Isp设置 用户 发送邮件
		})
		_ = c.AddFunc("6 */5 * * * *", func() { //每5分钟的第6秒执行
			MarketDoSending() // 邮件营销
		})
		_ = c.AddFunc("* */2 * * * *", func() { // 每2分钟执行一次
			UnlimitedEarlyWarning() // 不限量邮件预警
		})

		_ = c.AddFunc("0 */5 * * * *", func() { // 每1分钟执行一次
			StaticRegionStatusWarning()
		})
	}
	c.Start()
	// 阻塞主线程，以等待定时任务的执行
	select {}
}
