package crons

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// 异步处理账户流量Cdk信息
func HandleCdkInfo() {
	time.Sleep(1)
	redisResult := models.RedisRPop("list_cdk_info")
	if redisResult != "" {
		pushInfo := models.PushCdkey{}
		err := json.Unmarshal([]byte(redisResult), &pushInfo)
		fmt.Println(err)
		if err == nil {
			value := pushInfo.Number //数值
			oldValue := int64(0)     //数值
			uid := pushInfo.Uid
			bindUid := pushInfo.BindUid
			cate := pushInfo.Cate
			cdkType := pushInfo.CdkType
			isOk := 1
			res1 := 0
			res2 := 0
			result := ""
			createTime := pushInfo.CreateTime
			dealCate := 0
			if cate == "flow" {
				username := pushInfo.BindUsername
				email := pushInfo.BindEmail
				if cdkType == "cdk" {
					dealCate = 1
					flowInfo := models.GetUserFlowInfo(uid)
					if flowInfo.ID == 0 {
						isOk = 0
					}
					redeemFlow := flowInfo.BuyFlow
					if redeemFlow > 0 {
						redeemFlow = flowInfo.BuyFlow - flowInfo.CdkFlow
						if redeemFlow > flowInfo.Flows { // 如果买的剩余的 大于 总剩余的就展示 总剩余的
							redeemFlow = flowInfo.Flows
						}
					}

					if (flowInfo.Flows - value) < 0 { //余额充足的情况才继续
						isOk = 0
					}
					if redeemFlow <= 0 || redeemFlow < value { ////余额充足的情况才继续
						isOk = 0
					}
					oldValue = flowInfo.Flows
					if isOk == 1 {
						res1 = AddExchangeRecord(pushInfo, 3)
						if res1 != 1 {
							isOk = 0
						}
					}
				}
				if cdkType == "to_user" {
					dealCate = 2
					flowInfo := models.GetUserFlowInfo(uid)
					if flowInfo.ID == 0 {
						isOk = 0
					}
					redeemFlow := flowInfo.BuyFlow
					if redeemFlow > 0 {
						redeemFlow = flowInfo.BuyFlow - flowInfo.CdkFlow
						if redeemFlow > flowInfo.Flows { // 如果买的剩余的 大于 总剩余的就展示 总剩余的
							redeemFlow = flowInfo.Flows
						}
					}
					if (flowInfo.Flows - value) < 0 { //余额充足的情况才继续
						isOk = 0
					}
					if redeemFlow <= 0 || redeemFlow < value { ////余额充足的情况才继续
						isOk = 0
					}

					oldValue = flowInfo.Flows
					if isOk == 1 {
						res1 = AddExchangeRecord(pushInfo, 3)
						if res1 != 1 {
							isOk = 0
						}
					}
				}
				if cdkType == "forbid" {
					dealCate = 3
					flowInfo := models.GetUserFlowInfo(uid)
					if flowInfo.ID == 0 {
						isOk = 0
					}

					if (flowInfo.CdkFlow + value) > flowInfo.BuyFlow { //剩余 + 恢复的 不能大于总购买的
						isOk = 0
					}
					oldValue = flowInfo.Flows
					if isOk == 1 {
						res1 = EditExchangeRecord(pushInfo)
						if res1 != 1 {
							isOk = 0
						}
					}
				}
				if isOk == 1 {
					res2 = DealUserFlow(value, uid, createTime, dealCate, bindUid, username, email)
				}
			}
			if cate == "isp" {
				agentInfo := models.GetAgentBalanceByUid(uid)
				if cdkType == "cdk" {
					dealCate = 1
					if agentInfo.Id == 0 {
						isOk = 0
					}
					if (agentInfo.Balance - int(value)) < 0 { //余额充足的情况才继续
						isOk = 0
					}
					oldValue = int64(agentInfo.Balance)
					if isOk == 1 {
						res1 = AddExchangeRecord(pushInfo, 2)
						if res1 != 1 {
							isOk = 0
						}
					}
				}
				if cdkType == "self" {
					dealCate = 2
					if agentInfo.Id == 0 {
						isOk = 0
					}
					if (agentInfo.Balance - int(value)) < 0 { //余额充足的情况才继续
						isOk = 0
					}
					oldValue = int64(agentInfo.Balance)
				}
				if cdkType == "forbid" {
					dealCate = 3
					if agentInfo.Id == 0 {
						isOk = 0
					}
					total := agentInfo.Total
					agentBalance := agentInfo.Balance
					if agentBalance < 0 {
						agentBalance = 0
					}
					newBalance := int(value) + agentBalance
					if newBalance > total {
						isOk = 0
					}

					oldValue = int64(agentInfo.Balance)
					if isOk == 1 {
						res1 = EditExchangeRecord(pushInfo)
					}
				}

				if isOk == 1 {
					res2 = DealUserIsp(value, uid, createTime, dealCate)
				}
			}
			result = util.ItoS(res1) + "_" + util.ItoS(res2)

			// 存日志
			remark := "auto" + "__" + util.ItoS(isOk)
			logInfo := models.PushCdkRecord{}
			logInfo.Cate = pushInfo.Cate
			logInfo.Cdkey = pushInfo.Cdkey
			logInfo.Value = oldValue //操作前余额
			logInfo.Number = pushInfo.Number
			logInfo.CdkType = pushInfo.CdkType
			logInfo.Uid = pushInfo.Uid
			logInfo.BindUsername = pushInfo.BindUsername
			logInfo.BindUid = pushInfo.BindUid
			logInfo.BindEmail = pushInfo.BindEmail
			logInfo.BindTime = pushInfo.BindTime
			logInfo.Ip = pushInfo.Ip
			logInfo.CreateTime = pushInfo.CreateTime
			logInfo.Result = result
			logInfo.Remark = remark
			errLog := models.CreateCdkLog(logInfo) // 加个 兑换变动日志
			fmt.Println(errLog)
		}
	} else {
		//fmt.Println("no data")
	}
}

// 兑换Cdk
func ExchangeCdk() {
	time.Sleep(1)
	redisResult := models.RedisRPop("list_cdk_exchange")
	if redisResult != "" {
		pushInfo := models.PushCdkey{}
		err := json.Unmarshal([]byte(redisResult), &pushInfo)
		fmt.Println(err)
		if err == nil {
			exInfo := pushInfo.ExInfo
			isOk := 1
			res1 := 0
			res2 := 0
			oldValue := int64(0)
			result := ""
			createTime := pushInfo.CreateTime
			value := pushInfo.Number
			uid := pushInfo.BindUid
			if exInfo.Id == 0 {
				isOk = 0
			}
			if exInfo.Status != 1 || exInfo.BindUid > 0 {
				isOk = 0
			}

			if exInfo.Expire > 0 && exInfo.Expire < createTime {
				isOk = 0
			}
			if exInfo.UseCycle >= 1 {
				_, couponUse := models.GetExchangeByUsePlatform(uid, exInfo.UseCycle, exInfo.Platform)
				if len(couponUse) >= exInfo.UseNumber {
					isOk = 0
				}
			}
			// 同一分组下只可兑换一次
			if exInfo.GroupId > 0 {
				cdkInfo, _ := models.GetExchangeByGroupId(uid, exInfo.GroupId)
				if cdkInfo.Id != 0 && exInfo.Id != cdkInfo.Id {
					isOk = 0
				}
			}
			if isOk == 1 {
				if exInfo.UseType == 2 {
					info := exInfo
					info.Id = 0
					info.Expire = exInfo.Expire
					info.BindUid = pushInfo.BindUid
					info.BindUsername = pushInfo.BindUsername
					info.CreateTime = createTime
					err1 := models.AddExchange(info)
					if err1 == nil {
						res1 = 1
					}
				} else {
					exParam := map[string]interface{}{}
					exParam["status"] = 2
					exParam["use_time"] = createTime
					exParam["bind_uid"] = pushInfo.BindUid
					exParam["bind_username"] = pushInfo.BindUsername
					err1 := models.EditExchangeByCode(pushInfo.Cdkey, exParam)
					if err1 == true {
						res1 = 1
					}
				}
				if res1 == 1 {
					// ISP
					if pushInfo.Cate == "isp" {
						_, userInfo := models.GetUserById(uid)
						if userInfo.Id > 0 {
							oldValue = int64(userInfo.Balance)
							// 处理用户余额问题
							go models.AddUserIpChangeLog(models.UserIpChangeLog{
								Uid:        uid,
								OldIp:      userInfo.Balance,
								NewIp:      userInfo.Balance + int(value),
								ChangeIp:   int(value),
								ChangeType: "cdk_exchange",
								ChangeTime: util.GetNowInt(),
								Mark:       1,
							})
							updateParams := make(map[string]interface{})
							updateParams["pay_ip"] = userInfo.PayIp + int(value)
							updateParams["balance"] = userInfo.Balance + int(value)
							err = models.EditUserById(uid, updateParams)
							if err == nil {
								res2 = 1
							}
						}
					}

					// 流量
					if pushInfo.Cate == "flow" {
						expireDay := 30
						if exInfo.Uid > 0 {
							cdkUserInfo := models.GetUserFlowInfo(exInfo.Uid) //查询生成cdk用户的流量信息
							expireDay = cdkUserInfo.DayMax
						}
						if expireDay == 0 {
							expireDay = 30
						}
						userInfo := models.GetUserFlowInfo(uid)
						//流量余额变动日志
						go models.AddUserFlowsChangeLog(models.UserFlowsChangeLog{
							Uid:         uid,
							OldFlows:    userInfo.Flows,
							NewFlows:    userInfo.Flows + value,
							ChangeFlows: value,
							ChangeType:  "cdk_exchange",
							ChangeTime:  util.GetNowInt(),
							Mark:        1,
						})
						if userInfo.ID == 0 {
							//创建用户余额
							socksIp := models.UserFlow{}
							socksIp.Uid = uid
							socksIp.Email = pushInfo.BindEmail
							socksIp.Username = pushInfo.BindUsername
							socksIp.Flows = value
							socksIp.AllFlow = value
							socksIp.ExFlow = value
							socksIp.Status = 1
							socksIp.ExpireTime = createTime + expireDay*86400
							socksIp.CreateTime = createTime
							err, _ = models.CreateUserFlow(socksIp)
						} else {
							oldValue = userInfo.Flows
							upParam := make(map[string]interface{})
							upParam["all_flow"] = userInfo.AllFlow + value
							upParam["ex_flow"] = userInfo.ExFlow + value
							upParam["flows"] = userInfo.Flows + value
							upParam["expire_time"] = createTime + expireDay*86400
							err = models.EditUserFlow(userInfo.ID, upParam)
						}
						if err == nil {
							res2 = 1
						}
					}

					// 动态ISP
					if pushInfo.Cate == "dynamic_isp" {
						expireDay := 90

						userInfo := models.GetUserDynamicIspInfo(uid)
						if userInfo.ID == 0 {
							//创建用户余额
							socksIp := models.UserDynamicIsp{}
							socksIp.Uid = uid
							socksIp.Email = pushInfo.BindEmail
							socksIp.Username = pushInfo.BindUsername
							socksIp.Flows = value
							socksIp.AllFlow = value
							socksIp.Status = 1
							socksIp.ExpireTime = createTime + expireDay*86400
							socksIp.CreateTime = createTime
							err, _ = models.CreateUserDynamicIsp(socksIp)
						} else {
							oldValue = userInfo.Flows
							upParam := make(map[string]interface{})
							upParam["all_flow"] = userInfo.AllFlow + value
							upParam["flows"] = userInfo.Flows + value
							upParam["expire_time"] = createTime + expireDay*86400
							err = models.EditUserDynamicIspByUid(uid, upParam)
						}
						//流量余额变动日志
						models.AddDynamicIspLog(uid, 0, pushInfo.BindUsername, value, userInfo.Flows, userInfo.Flows+value, pushInfo.Ip, exInfo.Mode, 1)
						if err == nil {
							res2 = 1
						}
					}

					// 不限量
					if pushInfo.Cate == "unlimited" {
						// 获取用户信息
						//flowDayInfo := models.GetUserFlowDayByUid(uid)
						expireTime := int(value) + createTime
						//dayNum := int(value)
						//if flowDayInfo.Id == 0 {
						//	//创建用户余额IP
						//	addInfo := models.UserFlowDay{}
						//	addInfo.Uid = uid
						//	addInfo.Username = pushInfo.BindUsername
						//	addInfo.Email = pushInfo.BindEmail
						//	addInfo.AllDay = dayNum
						//	addInfo.ExpireTime = expireTime
						//	addInfo.CreateTime = createTime
						//	addInfo.Status = 1
						//	err, _ = models.CreateUserFlowDay(addInfo)
						//} else {
						//	upParam := make(map[string]interface{})
						//	upParam["all_day"] = dayNum + flowDayInfo.AllDay //累计总购买时间
						//	upParam["pre_day"] = flowDayInfo.ExpireTime      //购买前时间
						//	if flowDayInfo.ExpireTime < createTime {
						//		upParam["expire_time"] = expireTime
						//	} else {
						//		expireTime = int(value) + flowDayInfo.ExpireTime
						//		upParam["expire_time"] = expireTime
						//	}
						//	err = models.EditUserFlowDay(flowDayInfo.Id, upParam)
						//}
						//// IP池信息
						configNum := models.GetConfigVal("unlimited_base_config")
						bandwidthNum := models.GetConfigVal("unlimited_base_bandwidth")
						config := util.StoI(configNum)
						if config == 0 {
							config = 200
						}
						bandwidth := util.StoI(bandwidthNum)
						if bandwidth == 0 {
							bandwidth = 200
						}

						// 创建IP池队列 异步处理
						models.AddLogUserUnlimited(uid, config, bandwidth, expireTime, int(value/86400), createTime, pushInfo.Cdkey, createTime, "", "flow_day")

						//poolInfo := models.ScoreGetPoolFlowDayByUid(uid)
						//if poolInfo.Id == 0 {
						//	poolInfo = models.ScoreGetPoolFlowDayByUid(0)
						//	if poolInfo.Id > 0 {
						//		poolParam := make(map[string]interface{})
						//		poolParam["uid"] = uid //用户信息
						//		poolParam["expire_time"] = expireTime
						//		err = models.EditPoolFlowDay(poolInfo.Id, poolParam)
						//	} else {
						//		dingMsg("预警提示 ，360不限量流量IP配置不足: 用户ID" + util.ItoS(uid) + "  CDK兑换")
						//	}
						//
						//} else {
						//	poolParam := make(map[string]interface{})
						//	poolParam["expire_time"] = expireTime
						//	err = models.EditPoolFlowDay(poolInfo.Id, poolParam)
						//}

						if err == nil {
							res2 = 1
						}
					}

					// 静态IP
					if strings.Contains(pushInfo.Cate, "static") {
						packageList := models.GetStaticPackageList()
						packArr := map[int]int{}
						for _, v := range packageList {
							packArr[v.Value] = v.Id
						}

						exDay := exInfo.Day
						if exDay == 0 {
							exDay = 7
						}
						pakId, ok := packArr[exDay]
						if !ok {
							pakId = packageList[0].Id
						}
						country := strings.ToLower(exInfo.Region)

						_, userStaticPak := models.GetUserStaticByPakRegion(uid, pakId, country) //查询用户静态套餐余额

						if userStaticPak.Id == 0 {
							//创建用户余额IP
							staticIp := models.UserStaticIp{}
							staticIp.Uid = uid
							staticIp.Username = pushInfo.BindUsername
							staticIp.Email = pushInfo.BindEmail
							staticIp.PakId = pakId
							staticIp.PakRegion = country
							staticIp.Balance = int(value)
							staticIp.AllBuy = int(value)
							staticIp.AllNum = 1
							staticIp.ExpireDay = exDay
							staticIp.CreateTime = pushInfo.CreateTime
							staticIp.Status = 1
							models.AddUserStatic(staticIp)
						} else {
							upParam := make(map[string]interface{})
							upParam["all_buy"] = int(value) + userStaticPak.AllBuy
							upParam["all_num"] = userStaticPak.AllNum + 1
							upParam["balance"] = int(value) + userStaticPak.Balance
							models.EditUserStatic(userStaticPak.Id, upParam)
						}
					}
				}

				result = util.ItoS(res1) + "_" + util.ItoS(res2)

				// 存日志
				remark := "auto" + "__" + util.ItoS(isOk)
				logInfo := models.PushCdkRecord{}
				logInfo.Cate = pushInfo.Cate
				logInfo.Cdkey = pushInfo.Cdkey
				logInfo.Value = oldValue //操作前余额
				logInfo.Number = pushInfo.Number
				logInfo.Country = pushInfo.Country
				logInfo.CdkType = "exchange"
				logInfo.Uid = pushInfo.Uid
				logInfo.BindUsername = pushInfo.BindUsername
				logInfo.BindUid = pushInfo.BindUid
				logInfo.BindEmail = pushInfo.BindEmail
				logInfo.BindTime = pushInfo.BindTime
				logInfo.Ip = pushInfo.Ip
				logInfo.CreateTime = pushInfo.CreateTime
				logInfo.Result = result
				logInfo.Remark = remark
				errLog := models.CreateCdkLog(logInfo) // 加个 兑换变动日志
				fmt.Println(errLog)

			}
		}
	}
}

// 处理用户流量余额 cate 1 减余额  2 加余额   3 恢复余额（生成的cdk 没用，禁用的时候）
func DealUserFlow(value int64, uid, createTime int, cate, bindUid int, username, email string) (res int) {
	/// 获取用户信息
	userInfo := models.GetUserFlowInfo(uid)
	if cate == 1 {
		//流量余额变动日志
		go models.AddUserFlowsChangeLog(models.UserFlowsChangeLog{
			Uid:         uid,
			OldFlows:    userInfo.Flows,
			NewFlows:    userInfo.Flows - value,
			ChangeFlows: value,
			ChangeType:  "cdk_1",
			ChangeTime:  util.GetNowInt(),
			Mark:        -1,
		})
		upParam := make(map[string]interface{})
		upParam["flows"] = userInfo.Flows - value
		upParam["cdk_flow"] = userInfo.CdkFlow + value
		err := models.EditUserFlow(userInfo.ID, upParam)
		if err == nil {
			return 1
		}
		return 2
	}
	if cate == 2 {
		//流量余额变动日志
		go models.AddUserFlowsChangeLog(models.UserFlowsChangeLog{
			Uid:         uid,
			OldFlows:    userInfo.Flows,
			NewFlows:    userInfo.Flows - value,
			ChangeFlows: value,
			ChangeType:  "cdk_2",
			ChangeTime:  util.GetNowInt(),
			Mark:        -1,
		})
		upParam := make(map[string]interface{})
		upParam["flows"] = userInfo.Flows - value
		upParam["cdk_flow"] = userInfo.CdkFlow + value
		err := models.EditUserFlow(userInfo.ID, upParam)

		var err2 error
		userInfo2 := models.GetUserFlowInfo(bindUid)
		//流量余额变动日志
		go models.AddUserFlowsChangeLog(models.UserFlowsChangeLog{
			Uid:         uid,
			OldFlows:    userInfo2.Flows,
			NewFlows:    userInfo2.Flows + value,
			ChangeFlows: value,
			ChangeType:  "cdk_2",
			ChangeTime:  util.GetNowInt(),
			Mark:        1,
		})
		if userInfo2.ID == 0 {
			//创建用户余额
			socksIp := models.UserFlow{}
			socksIp.Uid = bindUid
			socksIp.Email = email
			socksIp.Username = username
			socksIp.Flows = value
			socksIp.AllFlow = value
			socksIp.ExFlow = value
			socksIp.Status = 1
			socksIp.ExpireTime = createTime + userInfo.DayMax*86400
			//socksIp.ExpireTime = createTime + 30*86400
			socksIp.CreateTime = createTime
			err2, _ = models.CreateUserFlow(socksIp)
		} else {
			upParam2 := make(map[string]interface{})
			upParam2["all_flow"] = userInfo2.AllFlow + value
			upParam2["ex_flow"] = userInfo2.ExFlow + value
			upParam2["flows"] = userInfo2.Flows + value
			//upParam2["expire_time"] = createTime + 30*86400
			upParam2["expire_time"] = createTime + userInfo.DayMax*86400
			err2 = models.EditUserFlow(userInfo2.ID, upParam2)
		}
		if err == nil {
			if err2 == nil {
				return 1
			} else {
				return 2
			}
		} else {
			return 3
		}
		return 0
	}

	if cate == 3 {
		//流量余额变动日志
		go models.AddUserFlowsChangeLog(models.UserFlowsChangeLog{
			Uid:         uid,
			OldFlows:    userInfo.Flows,
			NewFlows:    userInfo.Flows + value,
			ChangeFlows: value,
			ChangeType:  "cdk_3",
			ChangeTime:  util.GetNowInt(),
			Mark:        1,
		})
		//if (userInfo.Flows - userInfo.ExFlow + value) <= userInfo.BuyFlow {	//剩余 + 恢复的
		cdkFlows := userInfo.CdkFlow - value
		if cdkFlows < 0 {
			cdkFlows = 0
		}
		upParam := make(map[string]interface{})
		upParam["flows"] = userInfo.Flows + value
		upParam["cdk_flow"] = cdkFlows
		err := models.EditUserFlow(userInfo.ID, upParam)
		if err == nil {
			return 1
		}
		//}
		return 0
	}
	return 0

}

// 处理用户ISP余额 cate 1 减余额  2 加余额   3 恢复余额（生成的cdk 没用，禁用的时候）
func DealUserIsp(value64 int64, uid, createTime int, cate int) (res int) {
	/// 获取用户信息
	value := int(value64)
	agentInfo := models.GetAgentBalanceByUid(uid)
	if cate == 1 {
		if (agentInfo.Balance - value) >= 0 { //余额充足的情况才继续
			upParam := make(map[string]interface{})
			upParam["balance"] = agentInfo.Balance - value
			err := models.EditAgentBalanceByUid(uid, upParam)
			if err == nil {
				return 1
			}
		}
		return 0
	}
	if cate == 2 {
		var err error
		_, userInfo := models.GetUserById(uid)
		if userInfo.Id > 0 {
			err1 := 0
			if (agentInfo.Balance - value) >= 0 { //余额充足的情况才继续
				upParam := make(map[string]interface{})
				upParam["balance"] = agentInfo.Balance - value
				err = models.EditAgentBalanceByUid(uid, upParam)
				if err == nil {
					err1 = 1
				}
			}
			if err1 == 1 {
				// 处理用户余额问题
				go models.AddUserIpChangeLog(models.UserIpChangeLog{
					Uid:        uid,
					OldIp:      userInfo.Balance,
					NewIp:      userInfo.Balance + value,
					ChangeIp:   value,
					ChangeType: "cdk_2",
					ChangeTime: util.GetNowInt(),
					Mark:       1,
				})
				updateParams := make(map[string]interface{})
				updateParams["pay_ip"] = userInfo.PayIp + value
				updateParams["balance"] = userInfo.Balance + value
				err = models.EditUserById(uid, updateParams)
				if err == nil {
					return 1
				} else {
					return 2
				}
			} else {
				return 3
			}
		}
		return 0
	}

	if cate == 3 {
		total := agentInfo.Total
		agentBalance := agentInfo.Balance
		if agentBalance < 0 {
			agentBalance = 0
		}
		newBalance := value + agentBalance
		if newBalance <= total {
			upParam := make(map[string]interface{})
			upParam["balance"] = agentInfo.Balance + value
			err := models.EditAgentBalanceByUid(uid, upParam)
			if err == nil {
				return 1
			}
		}
		return 2
	}
	return 0
}

// 添加兑换记录
func AddExchangeRecord(pushInfo models.PushCdkey, cate int) (res int) {
	//写入兑换券记录
	code := pushInfo.Cdkey
	info := models.ExchangeList{}
	info.Cid = 0
	info.Mode = pushInfo.Mode
	info.Cate = cate // 2 ISP  3 流量
	info.Uid = pushInfo.Uid
	info.Code = code
	info.Name = pushInfo.CdkType
	info.Title = "Exchange IPs"
	info.Status = 1
	if cate >= 2 {
		if pushInfo.CdkType == "self" {
			info.BindUid = pushInfo.BindUid
			info.BindUsername = pushInfo.BindUsername
			info.Status = 2
			info.UseTime = pushInfo.CreateTime
		}
	}

	if cate == 3 {
		info.Title = "Exchange Residential"
		if pushInfo.CdkType == "to_user" {
			info.BindUid = pushInfo.BindUid
			info.BindUsername = pushInfo.BindUsername
			info.Status = 2
			info.UseTime = pushInfo.CreateTime
		}
	}
	if cate == 4 {
		info.Title = "Exchange Rotating"
	}
	if cate == 5 {
		info.Title = "Exchange Unlimited"
	}
	if cate == 6 {
		info.Title = "Exchange Static"
		info.Region = strings.ToLower(pushInfo.Country)
		exArr := strings.Split(pushInfo.Cate, "-")
		exDay := util.StoI(exArr[1])
		if exDay == 0 {
			exDay = 7
		}
		info.Day = exDay
	}

	info.Value = pushInfo.Number
	info.UserType = "all"
	info.UseType = 1
	info.Expire = 0
	info.ExpiryDay = 0
	info.UseCycle = 0
	info.UseNumber = 0
	info.Platform = 0
	info.GroupId = 0
	info.CreateTime = pushInfo.CreateTime
	err := models.AddExchange(info)
	if err == nil {
		return 1
	}
	return 0
}

// 更新兑换记录
func EditExchangeRecord(pushInfo models.PushCdkey) (res int) {
	//写入兑换券记录
	code := pushInfo.Cdkey
	has, exInfo := models.GetExchangeInfo(code)
	if has != nil || exInfo.Id == 0 || exInfo.Status != 1 {
		return 2
	}
	data := map[string]interface{}{}
	data["status"] = 3
	data["use_time"] = pushInfo.CreateTime
	err := models.EditExchangeByCode(code, data)
	if err == true {
		return 1
	}
	return 0
}
