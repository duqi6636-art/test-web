package crons

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	emailSender "api-360proxy/web/service/email"
	"fmt"
	"strings"
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
		if userConfig.PauseIsp > 0 {
			return
		}
		executeIspRenew(uid, renewType, userConfig)
	case "flow":
		if userConfig.PauseFlow > 0 {
			return
		}
		executeFlowRenew(uid, renewType, userConfig)
	case "rotating":
		if userConfig.PauseRotating > 0 {
			return
		}
		executeRotatingRenew(uid, renewType, userConfig)
	case "static":
		if userConfig.PauseStatic > 0 {
			return
		}
		executeStaticRenew(uid, renewType, userConfig)
	case "unlimited":
		if userConfig.PauseUnlimited > 0 {
			return
		}
		executeUnlimitedRenew(uid, renewType, userConfig)
	}
}

func executeIspRenew(uid int, cate string, userConfig models.UserAutoRenewModel) {
	// 获取用户ISP续费配置
	detailList := models.GetUserAutoRenewDetailListJoin(uid, cate)
	confList := models.GetConfBalanceRenewList("", cate)

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
					sendAutoRenewEmail(detail.Cate, userConfig.Email, detail.RBalance) //发邮件
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
			opLogCode := cate + "_auto_renew"
			opLog := fmt.Sprintf("uid:%d;result:%s;kf:%s;value:%d;name:%s", uid, result, util.FtoS2(totalMoney, 3), value, confName)
			insertLogs(opLogCode, opLog)
		}
	}
}

func executeRotatingRenew(uid int, cate string, userConfig models.UserAutoRenewModel) {
	// 获取用户续费配置
	detailList := models.GetUserAutoRenewDetailListJoin(uid, cate)
	confList := models.GetConfBalanceRenewList("", cate)

	confMap := map[int]models.ConfBalanceRenewModel{}
	for _, conf := range confList {
		confMap[conf.Id] = conf
	}

	for _, detail := range detailList {
		nowTime := util.GetNowInt()
		dTime := detail.ExpireTime - nowTime
		value := detail.Value
		confInfo, ok := confMap[detail.ConfId]
		if !ok {
			continue
		}

		szBalance := detail.Balance * confInfo.UnitValue
		// 当用户剩余余额  <= 用户配置的自动续费余额时 或者是有效期小于等于配置的时间时，触发自动续费
		if detail.UserBalance <= szBalance || dTime <= (detail.SyDay*86400) {
			confName := confInfo.Name
			if confInfo.IsCustom == 1 {
				confName = fmt.Sprintf("%d %s", detail.Value/confInfo.UnitValue, confInfo.Unit)
			}

			totalMoney := confInfo.Price * float64(detail.Value/confInfo.UnitValue)

			expireDay := detail.ExpireDay
			if expireDay == 0 {
				expireDay = 30
			}
			expireTime := detail.ExpireTime + expireDay*86400
			if detail.ExpireTime < nowTime {
				expireTime = nowTime + expireDay*86400
			}
			result := ""
			// 判断扣费的余额用户是否余额足够
			if totalMoney > detail.RBalance {
				// 更新暂停状态
				upPause := map[string]interface{}{}
				pauseType := "pause_" + cate
				upPause[pauseType] = nowTime
				models.EditUserAutoRenew(uid, upPause)
				if userConfig.EmailSwitch == 1 { // 是有效邮箱才发送
					sendAutoRenewEmail(detail.Cate, userConfig.Email, detail.RBalance) //发邮件
				}
				result = "insufficient"
			} else {
				oldBalance := detail.UserBalance
				balanceNew := detail.RBalance - totalMoney
				res := models.DealUserBalance(uid, value, oldBalance, detail.Cate, totalMoney, expireTime)
				result = "fail"
				if res == nil {
					eee := models.AddUserBalanceLog(uid, 3, balanceNew, float64(oldBalance), confInfo.Cate, value, 1, -1, nowTime, "")
					fmt.Println("add user balance log", eee)
					result = "success"
				}

			}
			// 操作日志
			opLogCode := cate + "_auto_renew"
			opLog := fmt.Sprintf("uid:%d;result:%s;kf:%s;value:%d;name:%s", uid, result, util.FtoS2(totalMoney, 3), value, confName)
			insertLogs(opLogCode, opLog)
		}
	}
}

func executeFlowRenew(uid int, cate string, userConfig models.UserAutoRenewModel) {
	// 获取用户续费配置
	detailList := models.GetUserAutoRenewDetailListJoin(uid, cate)
	confList := models.GetConfBalanceRenewList("", cate)

	confMap := map[int]models.ConfBalanceRenewModel{}
	for _, conf := range confList {
		confMap[conf.Id] = conf
	}

	for _, detail := range detailList {
		confInfo, ok := confMap[detail.ConfId]
		if !ok {
			continue
		}
		nowTime := util.GetNowInt()
		dTime := detail.ExpireTime - nowTime
		value := detail.Value
		unitValue := confInfo.UnitValue

		szBalance := detail.Balance * unitValue
		// 当用户剩余余额  <= 用户配置的自动续费余额时 或者是有效期小于等于配置的时间时，触发自动续费
		if detail.UserBalance <= szBalance || dTime <= (detail.SyDay*86400) {
			limit := value / unitValue
			confName := confInfo.Name
			totalMoney := confInfo.Price * float64(limit)
			if confInfo.IsCustom == 1 {
				confName = fmt.Sprintf("%d %s", detail.Value/confInfo.UnitValue, confInfo.Unit)
				if limit >= 1000 {
					totalMoney = 0.77 * float64(limit)
				} else {
					totalMoney = 1 * float64(limit)
				}
			}

			expireDay := detail.ExpireDay
			if expireDay == 0 {
				expireDay = 30
			}
			expireTime := detail.ExpireTime + expireDay*86400
			if detail.ExpireTime < nowTime {
				expireTime = nowTime + expireDay*86400
			}
			result := ""
			// 判断扣费的余额用户是否余额足够
			if totalMoney > detail.RBalance {
				// 更新暂停状态
				upPause := map[string]interface{}{}
				pauseType := "pause_" + cate
				upPause[pauseType] = nowTime
				models.EditUserAutoRenew(uid, upPause)
				if userConfig.EmailSwitch == 1 { // 是有效邮箱才发送
					sendAutoRenewEmail(detail.Cate, userConfig.Email, detail.RBalance) //发邮件
				}
				result = "insufficient"
			} else {
				oldBalance := detail.UserBalance
				balanceNew := detail.RBalance - totalMoney
				res := models.DealUserBalance(uid, value, oldBalance, detail.Cate, totalMoney, expireTime)
				result = "fail"
				if res == nil {
					eee := models.AddUserBalanceLog(uid, 3, balanceNew, float64(oldBalance), confInfo.Cate, value, 1, -1, nowTime, "")
					fmt.Println("add user balance log", eee)
					result = "success"
				}

			}
			// 操作日志
			opLogCode := cate + "_auto_renew"
			opLog := fmt.Sprintf("uid:%d;result:%s;kf:%s;value:%d;name:%s", uid, result, util.FtoS2(totalMoney, 3), value, confName)
			insertLogs(opLogCode, opLog)
		}
	}
}

func executeStaticRenew(uid int, cate string, userConfig models.UserAutoRenewModel) {
	detailList := models.GetUserAutoRenewDetailListJoinStatic(uid, cate)
	confList := models.GetConfBalanceRenewList("", cate)
	confMap := map[int]models.ConfBalanceRenewModel{}
	for _, conf := range confList {
		confMap[conf.Id] = conf
	}

	nowTime := util.GetNowInt()
	// 获取用户详细配置
	userStaticList := models.GetUserStaticIpByBalance(uid)
	// 用户静态余额
	userStaticBalance := map[string]int64{}
	for _, val := range userStaticList {
		//uid-region-ispType-expireDay
		region := strings.ToUpper(val.PakRegion)
		str := fmt.Sprintf("%d-%s-%d", val.Uid, region, val.PakId)
		userStaticBalance[str] = int64(val.Balance)
	}

	// 获取用户余额充值
	userB := models.GetUserBalanceByUid(uid)
	userBalanceMoney := userB.Balance

	for _, detail := range detailList {
		value := detail.Value //待续费的值
		dTime := detail.ExpireTime - nowTime
		method := detail.Method

		if dTime <= (detail.SyDay * 86400) { // 当用户剩余余额  <= 用户配置有效期小于配置的时间时，触发自动续费
			confInfo, okc := confMap[detail.ConfId] //获取配置信息
			if !okc {
				continue
			}
			confName := confInfo.Name
			// 剩余静态余额
			strHas := fmt.Sprintf("%d-%s-%d", detail.Uid, detail.Country, confInfo.PakId)
			userBalance := userStaticBalance[strHas]

			// 扣除余额价价格
			kfMoney := confInfo.Price
			kfType := 0 //扣费类型 0余额不足不扣费 1扣除静态余额 2扣余额
			if method == "static" {
				if userBalance < 1 { //静态余额不足的时候扣除 余额充值的金额
					if userBalanceMoney < kfMoney { //用户余额小于扣费余额
						kfType = 0
					} else {
						kfType = 2
					}
				} else {
					kfType = 1
				}
			}

			if method == "balance" {
				if userBalanceMoney < kfMoney { //用户余额小于扣费余额
					if userBalance < 1 { //余额充值的金额不足的时候扣除 静态余额
						kfType = 0
					} else {
						kfType = 1
					}
				} else {
					kfType = 2
				}
			}
			result := ""
			if kfType == 0 { //余额不足
				upPause := map[string]interface{}{}
				pauseType := "pause_" + cate
				upPause[pauseType] = nowTime
				models.EditUserAutoRenew(uid, upPause)
				//余额不足，需要发邮件和站内信
				if userConfig.EmailSwitch == 1 { // 是有效邮箱才发送
					sendAutoRenewEmail(detail.Cate, userConfig.Email, userBalanceMoney) //发邮件
				}
				result = "insufficient"
			} else {
				if kfType == 1 { //静态余额扣费
					// 更新用户余额(余额充值的钱)
					// 更新余额（对应的IP余额）
					result = "static"
					res := models.DealUserStaticRecharge(uid, confInfo.PakId, nowTime, value, detail.Country, detail)
					if res == nil {
						newBalance := userBalance - 1
						userStaticBalance[strHas] = newBalance //更新map[]存储信息
						result = "success_static"
					}
				} else { // 余额扣费
					// 更新用户余额(余额充值的钱)
					// 更新IP有效期
					res := models.DealUserBalanceStatic(uid, value, kfMoney, detail.Country, nowTime, detail)
					result = "balance"
					if res == nil {
						balanceNew := userBalanceMoney - kfMoney
						eee := models.AddUserBalanceLog(uid, 3, balanceNew, userBalanceMoney, confInfo.Cate, value, 1, -1, nowTime, "")
						fmt.Println("add user balance log", eee)
						result = "success_balance"
					}
				}
			}
			// 操作日志
			opLogCode := cate + "_auto_renew"
			opLog := fmt.Sprintf("uid:%d;result:%s;kf:%s;value:%d;name:%s;staticIP:%s", uid, result, util.FtoS2(kfMoney, 3), value, confName, detail.StaticIp)
			insertLogs(opLogCode, opLog)
		}
	}
}

func executeUnlimitedRenew(uid int, cate string, userConfig models.UserAutoRenewModel) {
	// 获取用户续费配置
	detailList := models.GetUserAutoRenewDetailListJoinFlowDay(uid, cate)
	confList := models.GetConfBalanceRenewList("", cate)

	confMap := map[int]models.ConfBalanceRenewModel{}
	for _, conf := range confList {
		confMap[conf.Id] = conf
	}
	nowTime := util.GetNowInt()
	// 获取用户余额
	userB := models.GetUserBalanceByUid(uid)
	userBalanceMoney := userB.Balance
	packagePriceList := models.GetPackageUnlimitedList()
	bandwidthArr := map[string]float64{}
	for _, v := range packagePriceList {
		str := fmt.Sprintf("%d-%s-%d", v.PackageId, v.Cate, v.Config)
		bandwidthArr[str] = v.Money
	}

	for _, detail := range detailList {
		value := detail.Value //待续费的值
		dTime := detail.ExpireTime - nowTime

		if dTime <= (detail.SyDay * 86400) { // 当用户剩余余额  <= 用户配置有效期小于配置的时间时，触发自动续费
			confInfo, okc := confMap[detail.ConfId] //获取配置信息
			if !okc {                               //如果没有获取到配置信息，是否需要处理
				continue
			}
			confName := confInfo.Name
			result := ""
			// 剩余余额
			bandwidthKey := fmt.Sprintf("%d-%s-%d", confInfo.PakId, "bandwidth", detail.Bandwidth)
			bandwidthMoney, okb := bandwidthArr[bandwidthKey] //获取带宽配置价格
			if !okb {
				bandwidthMoney = 0
			}

			configKey := fmt.Sprintf("%d-%s-%d", confInfo.PakId, "config", detail.Config)
			configMoney, okb := bandwidthArr[configKey] //获取并发配置价格
			if !okb {
				configMoney = 0
			}
			// 扣除余额价价格
			kfMoney := confInfo.Price + bandwidthMoney + configMoney
			if detail.Money > 0 {
				kfMoney = detail.Money
			}

			if userBalanceMoney < kfMoney { //余额不足
				// 更新暂停状态
				upPause := map[string]interface{}{}
				pauseType := "pause_" + cate
				upPause[pauseType] = nowTime
				models.EditUserAutoRenew(uid, upPause)
				//余额不足，需要发邮件
				if userConfig.EmailSwitch == 1 { // 是有效邮箱才发送
					sendAutoRenewEmail(detail.Cate, userConfig.Email, userBalanceMoney) //发邮件
				}
				result = "insufficient"
			} else {
				expireTime := detail.ExpireTime //过期时间
				if detail.ExpireTime < nowTime {
					expireTime = nowTime
				}
				expireTime = expireTime + int(value)
				// 更新用户余额(余额充值的钱)
				// 更新IP有效期
				res := models.DealUserBalanceUnlimited(uid, value, kfMoney, expireTime, detail)
				if res == nil {
					balanceNew := userBalanceMoney - kfMoney
					eee := models.AddUserBalanceLog(uid, 3, balanceNew, userBalanceMoney, confInfo.Cate, value, 1, -1, nowTime, "")
					fmt.Println("add user balance log", eee)
					result = "success"
				}
			}
			// 操作日志
			opLogCode := cate + "_auto_renew"
			opLog := fmt.Sprintf("uid:%d;result:%s;kf:%s;value:%d;name:%s;hostip:%s", uid, result, util.FtoS2(kfMoney, 3), value, confName, detail.HostIp)
			insertLogs(opLogCode, opLog)
		}
	}
}

func sendAutoRenewEmail(cate, email string, balance float64) {
	isSend := 0
	if email != "" {
		default_mail := models.GetConfigVal("default_email")
		result := false
		//发送邮件
		emailType := 13
		vars := map[string]string{}
		vars["cate"] = cate
		vars["balance"] = "$" + util.FtoS2(balance, 3)
		vars["topUpLink"] = "https://center.cherryproxy.com/Dashboard/ProxyRecharge"

		if default_mail == "aws_mail" {
			result = emailSender.AwsSendEmail(email, emailType, vars, "auto_renew")
			fmt.Println("send result aws:", result)
		}
		if default_mail == "tencent_mail" {
			result = emailSender.TencentSendEmail(email, emailType, vars, "auto_renew")
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
