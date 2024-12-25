package crons

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"time"
)

// 异步处理长效Isp子账户流量信息
func DealLongIspFlowInfo() {
	time.Sleep(1)
	redisResult := models.RedisRPop("list_account_long_isp_flow")
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
			accountInfo, errA := models.GetLongIspUserAccountById(pushInfo.AccountId)
			if errA != nil || accountInfo.Id == 0 {
				isOk = 0
			}
			flowInfo := models.GetUserDynamicIspInfo(uid)
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
						err := models.EditUserDynamicIspByUid(flowInfo.Uid, editFlow)
						fmt.Println("editFlow", flowInfo.ID, err)

						upMap := map[string]interface{}{}
						upMap["flows"] = flows //余额
						upMap["limit_flow"] = limitFlows
						upMap["flow_unit"] = flowUnit
						upMap["expire_time"] = pushInfo.ExpireTime

						models.UpdateLongIspUserAccountById(accountInfo.Id, upMap) //更新子账号信息

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
						err := models.EditUserDynamicIspByUid(flowInfo.Uid, editFlow)
						fmt.Println("editFlow", flowInfo.ID, err)

						upMap := map[string]interface{}{}
						upMap["flows"] = accountInfo.Flows + flows //子账户流量余额
						upMap["limit_flow"] = limitFlows
						upMap["flow_unit"] = flowUnit
						upMap["expire_time"] = pushInfo.ExpireTime

						models.UpdateLongIspUserAccountById(accountInfo.Id, upMap) //更新子账号信息

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
					err := models.EditUserDynamicIspByUid(flowInfo.Uid, editFlow)
					fmt.Println("editFlow", flowInfo.ID, err)

					upMap := map[string]interface{}{}
					upMap["status"] = -1
					upMap["update_time"] = pushInfo.CreateTime

					models.UpdateLongIspUserAccountById(accountInfo.Id, upMap)

					//更新子账号信息 删除子账号
					fmt.Println("editAccount", accountInfo.Id, err)
				}
			}

			remark := "auto" + "__" + util.ItoS(isOk)
			errLog := models.CreateLogLongIspFlowsAccount(uid, accountInfo.Id, accountInfo.Flows, accountInfo.LimitFlow, flows, limitFlows, accountInfo.Account, pushInfo.Ip, cate, remark) // 加个 流量变动日志  存 变动前的数据 剩余 ，和 配置额度   和变动后的配置额度    时间，IP
			fmt.Println(errLog)
		}

	} else {
		//fmt.Println("no data")
	}

}
