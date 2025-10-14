package crons

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	emailSender "api-360proxy/web/service/email"
	"fmt"
	"time"
)

// DoAutoRenew 自动续费 处理 ISP
func DoAutoRenew() {
	DoAutoRenewWithOrder()
}

// DoAutoRenewWithOrder 按照用户设置的扣款顺序进行自动续费
func DoAutoRenewWithOrder() {
	// 获取所有开启自动续费的用户
	openList := models.GetAutoRenewRecordByOpen()
	if len(openList) == 0 {
		fmt.Println("no auto renew users")
		return
	}

	// 按用户分组处理
	for _, userConfig := range openList {
		uid := userConfig.Uid
		if userConfig.PauseIsp > 0 {
			continue
		}

		// 获取用户的扣款顺序配置
		orderList := models.GetUserAutoRenewOrderList(uid)

		// 如果用户没有配置扣款顺序，使用默认顺序
		if len(orderList) == 0 {
			err := models.InitUserAutoRenewOrder(uid, userConfig.Username, "")
			if err != nil {
				fmt.Printf("初始化用户 %d 扣款顺序失败: %v\n", uid, err)
				continue
			}
			orderList = models.GetUserAutoRenewOrderList(uid)
		}

		// 按照优先级顺序处理每种续费类型
		for _, order := range orderList {
			if order.Status != 1 {
				continue // 跳过禁用的续费类型
			}
			executeRenewByType(uid, order.RenewType, userConfig)
			// 每种类型处理后稍作延迟
			time.Sleep(2 * time.Second)
		}

		// 每个用户处理完后稍作延迟
		time.Sleep(5 * time.Second)
	}
}

// 执行指定类型的续费
func executeRenewByType(uid int, renewType string, userConfig models.UserAutoRenewModel) {
	switch renewType {
	case "isp":
		executeIspRenew(uid, userConfig)
		//case "static":
		//	return executeStaticRenew(uid, userConfig)
		//case "flow":
		//	return executeFlowRenew(uid, userConfig)
		//case "rotating":
		//	return executeRotatingRenew(uid, userConfig)
		//case "unlimited":
		//	return executeUnlimitedRenew(uid, userConfig)
	}
}

// 检查ISP续费余额
func executeIspRenew(uid int, userConfig models.UserAutoRenewModel) {
	// 获取用户ISP续费配置
	detailList := models.GetUserAutoRenewDetailListJoin(uid, "isp")
	confList := models.GetConfBalanceRenewList("", "isp")

	confMap := map[int]models.ConfBalanceRenewModel{}
	for _, conf := range confList {
		confMap[conf.Id] = conf
	}

	for _, detail := range detailList {
		if detail.UserBalance <= detail.Balance { // 用户余额小于自动续费余额 需要续费
			confInfo, ok := confMap[detail.ConfId]
			if !ok {
				continue
			}
			confName := confInfo.Name
			if confInfo.IsCustom == 1 {
				confName = fmt.Sprintf("%d %s", detail.Value, confInfo.Unit)
			}

			value := detail.Value
			totalMoney := confInfo.Price * float64(value)
			nowTime := util.GetNowInt()
			// 余额不足
			result := ""
			if totalMoney > detail.RBalance {
				// 更新暂停状态
				upPause := map[string]interface{}{}
				upPause["pause_isp"] = nowTime
				models.EditUserAutoRenew(uid, upPause)

				if userConfig.EmailSwitch == 1 { // 是有效邮箱才发送
					SendAutoRenewEmail(detail.Cate, userConfig.Email, detail.RBalance) //发邮件
				}
				result = "insufficient"
			} else {
				oldBalance := detail.UserBalance
				// 更新用户余额(余额充值的钱)
				// 更新余额（对应的IP余额）
				balanceNew := detail.RBalance - totalMoney
				res := models.DealUserBalance(uid, value, oldBalance, detail.Cate, totalMoney, 0)
				result = "fail"
				if res == nil {
					// 扣费日志
					eee := models.AddUserBalanceLog(uid, 3, balanceNew, float64(oldBalance), confInfo.Cate, value, 1, -1, nowTime, "")
					fmt.Println("add user balance log", eee)
					result = "success"
				}
			}
			// 操作日志
			opLogCode := "isp_auto_renew"
			opLog := fmt.Sprintf("uid:%d;result:%s;kf:%s;value:%d;name:%s", uid, result, util.FtoS2(totalMoney, 3), value, confName)
			insertLogs(opLogCode, opLog)
		}
	}
}

// 余额不足发送邮件

func SendAutoRenewEmail(cate, email string, balance float64) {
	isSend := 0
	if email != "" {
		default_mail := models.GetConfigVal("default_email")
		result := false
		//发送邮件
		code := "11"
		vars := map[string]string{}
		vars["cate"] = cate
		vars["balance"] = "$" + util.FtoS2(balance, 3)
		vars["topUpLink"] = "https://center.cherryproxy.com/Dashboard/ProxyRecharge"

		if default_mail == "aws_mail" {
			result = emailSender.AwsSendEmailMarket(email, code, vars, "auto_renew")
			fmt.Println("send result aws:", result)
		}
		if default_mail == "tencent_mail" {
			result = emailSender.TencentSendEmailMarket(email, code, vars, "auto_renew")
			fmt.Println("send result tencent:", result)
		}
		if result == true {
			isSend = 1
		} else {
			isSend = 2
		}
	} else {
		isSend = 3
	}
	str := fmt.Sprintf("%s-%f-%d", email, balance, isSend)
	insertLogs("auto_renew_send_email", str) //写日志
}

// insertLogs 日志记录
func insertLogs(code, data string) {
	models.AddLog(models.LogModel{
		Code:       code,
		Text:       data,
		CreateTime: util.GetTimeStr(util.GetNowInt(), "Y-m-d H:i:s"),
	})

}
