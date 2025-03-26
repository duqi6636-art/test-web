package crons

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"time"
)

// 积分兑换流量
func HandleScore() {
	time.Sleep(1)
	redisResult := models.RedisRPop("list_score")
	if redisResult != "" {
		pushInfo := models.PushScoreFlow{}
		err := json.Unmarshal([]byte(redisResult), &pushInfo)
		fmt.Println(err)
		if err == nil {
			isOk := 1
			res1 := 0
			res2 := 0
			status := 1
			createTime := pushInfo.CreateTime
			value := pushInfo.Flow
			score := pushInfo.Score
			uid := pushInfo.Uid
			userScore := models.GetUserScoreInfo(uid)
			confScore := models.GetConfScoreByFlow(value)

			if userScore.Id == 0 || confScore.Id == 0 {
				isOk = 0
			}
			if userScore.Score < score {
				isOk = 0
			}
			if userScore.Score < confScore.Score {
				isOk = 0
			}

			flows := int64(confScore.Flow * 1024 * 1024 * 1024)
			if isOk == 1 {
				// 更新用户积分
				uScore := userScore.Score - confScore.Score
				upScore := map[string]interface{}{}
				upScore["score"] = uScore
				resS := models.EditUserScore(uid, upScore)
				if resS == nil {
					res1 = 1
				}

				// 更新用户流量
				userInfo := models.GetUserFlowInfo(uid)
				//流量余额变动日志
				go models.AddUserFlowsChangeLog(models.UserFlowsChangeLog{
					Uid:         uid,
					OldFlows:    userInfo.Flows,
					NewFlows:    userInfo.Flows + flows,
					ChangeFlows: flows,
					ChangeType:  "score_exchange",
					ChangeTime:  util.GetNowInt(),
					Mark:        1,
				})
				newExpire := createTime + 30*86400
				if userInfo.ID == 0 {
					//创建用户余额
					socksIp := models.UserFlow{}
					socksIp.Uid = uid
					socksIp.Email = userScore.Email
					socksIp.Username = userScore.Username
					socksIp.Flows = flows
					socksIp.AllFlow = flows
					socksIp.ExFlow = flows
					socksIp.Status = 1
					socksIp.ExpireTime = newExpire
					socksIp.CreateTime = createTime
					err, _ = models.CreateUserFlow(socksIp)
				} else {
					upParam := make(map[string]interface{})
					upParam["all_flow"] = userInfo.AllFlow + flows
					upParam["ex_flow"] = userInfo.ExFlow + flows
					upParam["flows"] = userInfo.Flows + flows

					if userInfo.ExpireTime < newExpire { //有效期按照 长的来
						upParam["expire_time"] = newExpire
					}
					err = models.EditUserFlow(userInfo.ID, upParam)
				}
				if err == nil {
					res2 = 1
				}
			} else {
				status = 2
			}

			result := util.ItoS(res1) + "_" + util.ItoS(res2)
			logInfo := models.LogUserScore{}
			logInfo.Uid = pushInfo.Uid
			logInfo.Name = confScore.Name
			logInfo.Score = score
			logInfo.Money = 0
			logInfo.Code = 3
			logInfo.Mark = -1
			logInfo.Value = flows
			logInfo.Status = status
			logInfo.Ip = pushInfo.Ip
			logInfo.Remark = result
			logInfo.CreateTime = pushInfo.CreateTime
			errLog := models.CreateLogUserScore(logInfo) // 写日志
			fmt.Println(errLog)
		}
	}
}

// 积分兑换不限量流量
func HandleScoreFlowDay() {
	time.Sleep(1)
	redisResult := models.RedisRPop("list_score_flow_day")
	if redisResult != "" {
		pushInfo := models.PushScoreFlowDay{}
		err := json.Unmarshal([]byte(redisResult), &pushInfo)
		fmt.Println(err)
		if err == nil {
			isOk := 1
			res1 := 0
			res2 := 0
			status := 1
			createTime := pushInfo.CreateTime
			value := pushInfo.Day
			score := pushInfo.Score
			uid := pushInfo.Uid
			userScore := models.GetUserScoreInfo(uid)
			confScore := models.GetConfScoreByDay(value)

			if userScore.Id == 0 || confScore.Id == 0 {
				isOk = 0
			}
			if userScore.Score < score {
				isOk = 0
			}
			if userScore.Score < confScore.Score {
				isOk = 0
			}

			seconds := confScore.Day * 24 * 60 * 60
			if isOk == 1 {
				// 更新用户积分
				uScore := userScore.Score - confScore.Score
				upScore := map[string]interface{}{}
				upScore["score"] = uScore
				resS := models.EditUserScore(uid, upScore)
				if resS == nil {
					res1 = 1
				}

				// 更新用户流量信息
				//flowDayInfo := models.GetUserFlowDayByUid(uid)
				nowTime := util.GetNowInt()
				expireTime := seconds + nowTime
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
				models.AddLogUserUnlimited(uid, config, bandwidth, expireTime, int(value/86400), nowTime, "score_ex_"+util.ItoS(nowTime), createTime, "")

				//if flowDayInfo.Id == 0 {
				//	//创建用户余额IP
				//	addInfo := models.UserFlowDay{}
				//	addInfo.Uid = userScore.Uid
				//	addInfo.Username = userScore.Username
				//	addInfo.Email = userScore.Email
				//	addInfo.AllDay = seconds
				//	addInfo.ExpireTime = expireTime
				//	addInfo.CreateTime = nowTime
				//	addInfo.Status = 1
				//	err, _ = models.CreateUserFlowDay(addInfo)
				//
				//} else {
				//	upParam := make(map[string]interface{})
				//	upParam["all_day"] = seconds + flowDayInfo.AllDay //累计总购买时间
				//	upParam["pre_day"] = flowDayInfo.ExpireTime          //购买前时间
				//	if flowDayInfo.ExpireTime < nowTime {
				//		upParam["expire_time"] = expireTime
				//	} else {
				//		expireTime = seconds + flowDayInfo.ExpireTime
				//		upParam["expire_time"] = expireTime
				//	}
				//	err = models.EditUserFlowDay(flowDayInfo.Id, upParam)
				//}
				//// IP池信息
				//poolInfo := models.ScoreGetPoolFlowDayByUid(uid)
				//if poolInfo.Id == 0 {
				//	poolInfo = models.ScoreGetPoolFlowDayByUid(0)
				//	if poolInfo.Id > 0 {
				//		poolParam := make(map[string]interface{})
				//		poolParam["uid"] = uid //用户信息
				//		poolParam["expire_time"] = expireTime
				//		err = models.EditPoolFlowDay(poolInfo.Id, poolParam)
				//	} else {
				//		dingMsg("预警提示 ，360不限量流量IP配置不足: 用户ID" + util.ItoS(uid) + "    积分兑换")
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
			} else {
				status = 2
			}

			result := util.ItoS(res1) + "_" + util.ItoS(res2)
			logInfo := models.LogUserScore{}
			logInfo.Uid = pushInfo.Uid
			logInfo.Name = confScore.Name
			logInfo.Score = score
			logInfo.Money = 0
			logInfo.Code = 3
			logInfo.Mark = -1
			logInfo.Value = int64(seconds)
			logInfo.Status = status
			logInfo.Ip = pushInfo.Ip
			logInfo.Remark = result
			logInfo.CreateTime = createTime
			errLog := models.CreateLogUserScore(logInfo) // 写日志
			fmt.Println(errLog)
		}
	}
}

// 发送钉钉通知
func dingMsg(msg string) {
	phoneArr := []string{}
	url := models.GetConfigVal("dd_push_pay_refund")
	if url == "" {
		url = "https://oapi.dingtalk.com/robot/send?access_token=33712688eb42deb9b0fdf1882c277777c74ec91e642873cbcf4aaf658851528e"
	}
	isAtAll := false
	textArr := map[string]interface{}{
		"content": msg,
	}
	atArr := map[string]interface{}{
		"atMobiles": phoneArr, //被@人的手机号（在content里添加@人的手机号）
		"isAtAll":   isAtAll,  //是否@所有人
	}
	data_info := map[string]interface{}{
		"msgtype": "text",
		"text":    textArr,
		"at":      atArr,
	}
	//data_string, _ := json.Marshal(data_info)
	info := util.HttpPost(url, data_info, "application/json;charset=utf-8")
	fmt.Println(info)
}
