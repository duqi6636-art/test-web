package crons

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"time"
)

// 异步处理子账户流量信息
func DealFlowInfo() {
	time.Sleep(1)
	redisResult := models.RedisRPop("list_account_flow")
	if redisResult != "" {
		pushInfo := models.PushAccount{}
		err := json.Unmarshal([]byte(redisResult), &pushInfo)
		fmt.Println(err)
		if err == nil {
			cate := pushInfo.Cate
			flows := int64(0)             //流量数
			limitFlows := int64(0)        //分配流量数
			flowUnit := pushInfo.FlowUnit //分配流量单位
			uid := pushInfo.Uid
			isOk := 1
			accountInfo, err_a := models.GetUserAccountById(pushInfo.AccountId)
			if err_a != nil || accountInfo.Id == 0 {
				isOk = 0
			}
			flowInfo := models.GetUserFlowInfo(uid)
			if flowInfo.ID == 0 {
				isOk = 0
			}
			if cate == "add" {
				if flowInfo.ExpireTime < pushInfo.CreateTime {
					isOk = 0
				}
				flows = pushInfo.Flows
				limitFlows = pushInfo.LimitFlow
				uid = pushInfo.Uid
				if isOk == 1 {
					userFlow := flowInfo.Flows - flows
					if userFlow >= 0 { //用户余额大于 0的时候才继续执行
						editFlow := map[string]interface{}{
							"flows": userFlow,
						}
						//流量余额变动日志
						go models.AddUserFlowsChangeLog(models.UserFlowsChangeLog{
							Uid:         uid,
							OldFlows:    flowInfo.Flows,
							NewFlows:    flowInfo.Flows - flows,
							ChangeFlows: flows,
							ChangeType:  "add_child_account",
							ChangeTime:  util.GetNowInt(),
							Mark:        -1,
						})
						err := models.EditUserFlow(flowInfo.ID, editFlow)
						fmt.Println("editFlow", flowInfo.ID, err)

						upMap := map[string]interface{}{}
						upMap["flows"] = flows //余额
						upMap["limit_flow"] = limitFlows
						upMap["flow_unit"] = flowUnit
						upMap["expire_time"] = pushInfo.ExpireTime
						models.UpdateUserAccountById(accountInfo.Id, upMap) //更新子账号信息
						fmt.Println("editAccount", accountInfo.Id, err)
					} else {
						isOk = 0
					}
				}
			} else if cate == "edit" {
				if flowInfo.ExpireTime < pushInfo.CreateTime {
					isOk = 0
				}
				flows = pushInfo.Flows
				limitFlows = pushInfo.LimitFlow
				uid = pushInfo.Uid
				if isOk == 1 {
					userFlow := flowInfo.Flows - flows
					if userFlow >= 0 && limitFlows != accountInfo.LimitFlow { //用户余额大于 0的时候才继续执行 且用户 分配余额有变动的时候
						editFlow := map[string]interface{}{
							"flows": userFlow,
						}
						//流量余额变动日志
						go models.AddUserFlowsChangeLog(models.UserFlowsChangeLog{
							Uid:         uid,
							OldFlows:    flowInfo.Flows,
							NewFlows:    userFlow,
							ChangeFlows: flows,
							ChangeType:  "edit_child_account",
							ChangeTime:  util.GetNowInt(),
							Mark:        -1,
						})

						err := models.EditUserFlow(flowInfo.ID, editFlow)
						fmt.Println("editFlow", flowInfo.ID, err)

						upMap := map[string]interface{}{}
						upMap["flows"] = accountInfo.Flows + flows //子账户流量余额
						upMap["limit_flow"] = limitFlows
						upMap["flow_unit"] = flowUnit
						upMap["expire_time"] = pushInfo.ExpireTime
						models.UpdateUserAccountById(accountInfo.Id, upMap) //更新子账号信息
						fmt.Println("editAccount", accountInfo.Id, err)
					} else {
						isOk = 0
					}
				}
			} else if cate == "del" {
				if accountInfo.Status < 0 {
					isOk = 0
				}

				flows = pushInfo.Flows
				limitFlows = pushInfo.LimitFlow
				uid = pushInfo.Uid
				if isOk == 1 {
					userFlow := flowInfo.Flows + flows
					editFlow := map[string]interface{}{
						"flows": userFlow,
					}
					//流量余额变动日志
					go models.AddUserFlowsChangeLog(models.UserFlowsChangeLog{
						Uid:         uid,
						OldFlows:    flowInfo.Flows,
						NewFlows:    userFlow,
						ChangeFlows: flows,
						ChangeType:  "del_child_account",
						ChangeTime:  util.GetNowInt(),
						Mark:        1,
					})
					err := models.EditUserFlow(flowInfo.ID, editFlow)
					fmt.Println("editFlow", flowInfo.ID, err)

					upMap := map[string]interface{}{}
					upMap["status"] = -1
					upMap["update_time"] = pushInfo.CreateTime
					models.UpdateUserAccountById(accountInfo.Id, upMap) //更新子账号信息 删除子账号
					fmt.Println("editAccount", accountInfo.Id, err)
				}
			}

			remark := "auto" + "__" + util.ItoS(isOk)
			errLog := models.CreateLogFlowsAccount(uid, accountInfo.Id, accountInfo.Flows, accountInfo.LimitFlow, flows, limitFlows, accountInfo.Account, pushInfo.Ip, cate, remark) // 加个 流量变动日志  存 变动前的数据 剩余 ，和 配置额度   和变动后的配置额度    时间，IP
			fmt.Println(errLog)
		}

	} else {
		//fmt.Println("no data")
	}

}

// 异步处理佣金自用流量
func DealInviteFlow() {
	time.Sleep(1)
	redisResult := models.RedisRPop("list_invite_balance")
	if redisResult != "" {
		pushInfo := models.PushInvite{}
		err := json.Unmarshal([]byte(redisResult), &pushInfo)
		fmt.Println(err)
		if err == nil {
			cate := pushInfo.Cate
			value := pushInfo.Value
			money := pushInfo.Money
			uid := pushInfo.Uid
			isOk := 1
			ratio := models.GetExchangeRatio(cate)
			err_a, user := models.GetUserById(uid)
			if err_a != nil || user.Id == 0 {
				isOk = 0
			}
			if ratio == 0.00 || money == 0.00 || user.UsedMoney <= 0 || user.UsedMoney < money {
				isOk = 0
			}
			if isOk == 1 {
				////写佣金记录
				moneyLog := models.UserMoneyLog{}
				moneyLog.Uid = user.Id
				moneyLog.Code = pushInfo.Code
				moneyLog.Money = money
				moneyLog.Ip = value
				moneyLog.Ratio = ratio
				moneyLog.Mark = -1
				moneyLog.Status = 1
				moneyLog.CreateTime = pushInfo.CreateTime
				moneyLog.Today = pushInfo.Today
				models.CreateUserMoneyLog(moneyLog)

				updateParams := make(map[string]interface{})
				updateParams["used_money"] = user.UsedMoney - money
				if cate == "isp" {
					// 更新用户isp余额
					updateParams["balance"] = user.Balance + int(value)
				} else if cate == "flow" {
					/// 获取用户信息
					userFlowInfo := models.GetUserFlowInfo(user.Id)
					//流量余额变动日志
					go models.AddUserFlowsChangeLog(models.UserFlowsChangeLog{
						Uid:         uid,
						OldFlows:    userFlowInfo.Flows,
						NewFlows:    userFlowInfo.Flows + value,
						ChangeFlows: value,
						ChangeType:  "invite_money",
						ChangeTime:  util.GetNowInt(),
						Mark:        1,
					})
					newExpire := pushInfo.CreateTime + 30*86400
					if userFlowInfo.ID == 0 {
						//创建用户余额IP
						socksIp := models.UserFlow{}
						socksIp.Uid = user.Id
						socksIp.Email = user.Email
						socksIp.Username = user.Username
						socksIp.Flows = value
						socksIp.AllFlow = value
						socksIp.ExpireTime = newExpire
						socksIp.Status = 1
						socksIp.CreateTime = pushInfo.CreateTime
						models.CreateUserFlow(socksIp)
					} else {
						upParam := make(map[string]interface{})
						upParam["all_flow"] = value + userFlowInfo.AllFlow
						upParam["flows"] = value + userFlowInfo.Flows

						if userFlowInfo.ExpireTime < newExpire { //有效期按照 长的来
							upParam["expire_time"] = newExpire
						}

						models.EditUserFlow(userFlowInfo.ID, upParam)
					}
				}
				//更新用户信息
				editError := models.UpdateUserById(user.Id, updateParams)
				fmt.Println("update_user_used_money", editError)
			}

			data := models.PushInviteLog{}
			data.Cate = cate
			data.Code = pushInfo.Code
			data.Uid = pushInfo.Uid
			data.Value = pushInfo.Value
			data.Ratio = pushInfo.Ratio
			data.Money = pushInfo.Money
			data.CreateTime = pushInfo.CreateTime
			data.Today = pushInfo.Today
			data.Result = util.ItoS(isOk)
			models.CreateInviteLog(data) //写个日志
		}

	} else {
		//fmt.Println("no data")
	}

}
