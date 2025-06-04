package crons

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// 异步处理余额充值生成Cdk信息
func HandleCdkBalance() {
	time.Sleep(1)
	redisResult := models.RedisRPop("list_balance_cdk")
	if redisResult != "" {
		pushInfo := models.PushCdkey{}
		err := json.Unmarshal([]byte(redisResult), &pushInfo)
		fmt.Println(err)
		if err == nil {
			//value := pushInfo.Number //数值
			oldValue := 0.0 //原来余额值
			uid := pushInfo.Uid
			//bindUid := pushInfo.BindUid
			cateStr := pushInfo.Cate
			cdkType := pushInfo.CdkType
			isOk := 1
			res1 := 0
			res2 := 0
			result := ""
			createTime := pushInfo.CreateTime
			dealCate := 0
			value := pushInfo.Number
			cate := 2

			if cateStr == "isp" {
				cate = 2
			} else if cateStr == "flow" {
				cate = 3
			} else if cateStr == "dynamic_isp" {
				cate = 4
			} else if cateStr == "unlimited" {
				cate = 5
			} else if strings.Contains(cateStr, "static") {
				cate = 6
			}
			balanceInfo := models.GetUserBalanceByUid(uid)

			if balanceInfo.Id == 0 {
				isOk = 0
			}
			if (balanceInfo.Balance - pushInfo.Need) < 0 { //余额充足的情况才继续
				isOk = 0
			}
			oldValue = balanceInfo.Balance

			if cdkType == "cdk" {
				dealCate = 1
				if isOk == 1 {
					res1 = AddExchangeRecord(pushInfo, cate)
					if res1 != 1 {
						isOk = 0
					}
				}
			}
			if cdkType == "self" {
				dealCate = 2
				if isOk == 1 {
					res1 = AddExchangeRecord(pushInfo, cate)
					if res1 != 1 {
						isOk = 0
					}
				}
			}
			if cdkType == "forbid" {
				dealCate = 3
				if balanceInfo.Id == 0 {
					isOk = 0
				}
				total := balanceInfo.AllBuy
				agentBalance := balanceInfo.Balance
				if agentBalance < 0 {
					agentBalance = 0
				}
				newBalance := pushInfo.Need + agentBalance
				if newBalance > total {
					isOk = 0
				}

				if isOk == 1 {
					res1 = EditExchangeRecord(pushInfo)
				}
			}

			if isOk == 1 {
				res2 = DealUserBalance(pushInfo.Need, uid, createTime, dealCate)

				//if res2 == 1 {		//写入日志
				//	eee := models.AddUserBalanceLog(uid, 2, pushInfo.Need, cateStr, value, 1, -1, createTime)
				//	fmt.Println(eee)
				//}
				if res2 == 1 && cdkType == "self" { //写入日志
					// ISP
					if cateStr == "isp" {
						_, userInfo := models.GetUserById(uid)
						if userInfo.Id > 0 {
							// 处理用户余额问题
							go models.AddUserIpChangeLog(models.UserIpChangeLog{
								Uid:        uid,
								OldIp:      userInfo.Balance,
								NewIp:      userInfo.Balance + int(value),
								ChangeIp:   int(value),
								ChangeType: "balance_self",
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

					//流量
					if cateStr == "flow" {
						expireDay := 30

						userInfo := models.GetUserFlowInfo(uid)
						//流量余额变动日志
						go models.AddUserFlowsChangeLog(models.UserFlowsChangeLog{
							Uid:         uid,
							OldFlows:    userInfo.Flows,
							NewFlows:    userInfo.Flows + value,
							ChangeFlows: value,
							ChangeType:  "balance_self",
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

					//动态ISP
					if cateStr == "dynamic_isp" {
						expireDay := 90

						userInfo := models.GetUserDynamicIspInfo(uid)

						//流量余额变动日志

						models.AddDynamicIspLog(uid, 0, pushInfo.BindUsername, value, userInfo.Flows, userInfo.Flows+value, pushInfo.Ip, pushInfo.Mode, 1)
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
							upParam := make(map[string]interface{})
							upParam["all_flow"] = userInfo.AllFlow + value
							upParam["flows"] = userInfo.Flows + value
							upParam["expire_time"] = createTime + expireDay*86400
							err = models.EditUserDynamicIspByUid(uid, upParam)
						}
						if err == nil {
							res2 = 1
						}
					}

					// 不限量
					if cateStr == "unlimited" {
						//flowDayInfo := models.GetUserFlowDayByUid(uid)
						//preValue := flowDayInfo.ExpireTime
						expireTime := int(value) + createTime
						//dayNum := int(value)
						startTime := createTime
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
						//	err,_ = models.CreateUserFlowDay(addInfo)
						//} else {
						//
						//	if flowDayInfo.ExpireTime > createTime {
						//		startTime = flowDayInfo.ExpireTime
						//	}
						//	upParam := make(map[string]interface{})
						//	upParam["all_day"] = dayNum + flowDayInfo.AllDay //累计总购买时间
						//	upParam["pre_day"] = flowDayInfo.ExpireTime          //购买前时间
						//	if flowDayInfo.ExpireTime < createTime {
						//		upParam["expire_time"] = expireTime
						//	} else {
						//		expireTime = int(value) + flowDayInfo.ExpireTime
						//		upParam["expire_time"] = expireTime
						//	}
						//	err = models.EditUserFlowDay(flowDayInfo.Id, upParam)
						//}
						// IP池信息
						//poolInfo := models.ScoreGetPoolFlowDayByUid(uid)
						//if poolInfo.Id == 0 {
						//	poolInfo = models.ScoreGetPoolFlowDayByUid(0)
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
						models.AddLogUserUnlimited(uid, config, bandwidth, expireTime, int(value/86400), startTime, pushInfo.Cdkey, createTime, "", "flow_day")

						//} else {
						//	poolParam := make(map[string]interface{})
						//	poolParam["expire_time"] = expireTime
						//	err = models.EditPoolFlowDay(poolInfo.Id, poolParam)
						//}
						models.AddUnlimitedModel(uid, 0, pushInfo.BindUsername, value, 0, int64(expireTime), pushInfo.Ip, pushInfo.Mode, 1)

						if err == nil {
							res2 = 1
						}
					}

					// 静态IP
					if strings.Contains(cateStr, "static") {
						packageList := models.GetStaticPackageList()
						packArr := map[int]int{}
						for _, v := range packageList {
							packArr[v.Value] = v.Id
						}
						exArr := strings.Split(cateStr, "-")
						exDay := util.StoI(exArr[1])
						if exDay == 0 {
							exDay = 7
						}
						pakId, ok := packArr[exDay]
						if !ok {
							pakId = packageList[0].Id
						}
						country := strings.ToLower(pushInfo.Country)

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
			}

			result = util.ItoS(res1) + "_" + util.ItoS(res2)
			// 存日志
			remark := "auto" + "__" + util.ItoS(isOk)
			logInfo := models.PushCdkRecord{}
			logInfo.Cate = pushInfo.Cate
			logInfo.Cdkey = pushInfo.Cdkey
			logInfo.Balance = oldValue //操作前余额
			logInfo.Number = pushInfo.Number
			logInfo.Country = pushInfo.Country
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

// 处理用户余额 cate 1 减余额  2 加余额   3 恢复余额（生成的cdk 没用，禁用的时候）
func DealUserBalance(value float64, uid, createTime int, cate int) (res int) {
	/// 获取用户信息

	balanceInfo := models.GetUserBalanceByUid(uid)
	if cate == 1 || cate == 2 {
		if (balanceInfo.Balance - value) >= 0 { //余额充足的情况才继续
			upParam := make(map[string]interface{})
			upParam["balance"] = balanceInfo.Balance - value
			err := models.EditUserBalanceByUid(uid, upParam)
			if err == nil {
				return 1
			}
		}
		return 0
	}

	if cate == 3 {
		total := balanceInfo.AllBuy
		agentBalance := balanceInfo.Balance
		if agentBalance < 0 {
			agentBalance = 0
		}
		newBalance := value + agentBalance
		if newBalance <= total {
			upParam := make(map[string]interface{})
			upParam["balance"] = balanceInfo.Balance + value
			err := models.EditAgentBalanceByUid(uid, upParam)
			if err == nil {
				return 1
			}
		}
		return 2
	}
	return 0
}
