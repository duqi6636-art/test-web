package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"math"
	"strconv"
	"strings"
)

func GetActivityUserInfo(c *gin.Context) {

	if util.GetNowInt() > util.StoI(models.GetConfigVal("activity_invite_jump_time")) {
		JsonReturn(c, 40001, "", nil)
		return
	}

	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	//查询用户邀请记录
	err, inviteList := models.GetActivityInviterListByMap(map[string]interface{}{"inviter_id": user.Id})
	inviteNum := 0
	if err == nil && len(inviteList) > 0 {
		inviteNum = len(inviteList)
	}

	//查询用户佣金记录
	mWhere := map[string]interface{}{
		"inviter_id": user.Id,
		"status":     1,
	}
	err, moneyList := models.GetActivityMoneyListByMap(mWhere)
	totalMoney := 0.00 //累计总奖励
	todayMoney := 0.00 //今日佣金
	selfMoney := 0.00  //自用流量使用佣金
	today := util.GetTodayTime()
	if err == nil && len(inviteList) > 0 {
		for _, v := range moneyList {
			if v.Code == 10 {
				totalMoney = totalMoney + v.Money //
				if v.CreateTime >= today {
					todayMoney = todayMoney + v.Money
				}
			}
		}
	}
	mWhere2 := map[string]interface{}{
		"uid":    user.Id,
		"status": 1,
	}
	err, moneyUseList := models.GetActivityMoneyListByMap(mWhere2)
	if err == nil && len(moneyUseList) > 0 {
		for _, v := range moneyUseList {
			if v.Code == 12 || v.Code == 13 {
				selfMoney = selfMoney + v.Money
			}
		}
	}

	err, pUser := models.GetUserActivityInviterByMap(map[string]interface{}{"uid": user.Id})
	inviterCode := ""
	if err != nil || pUser.Id == 0 {
		uuid := GetUuid()
		invArr := strings.Split(uuid, "-")
		inviterCode = "act_" + invArr[0]
		// 添加邀请记录
		inviteInfo := models.UserInviter{}
		inviteInfo.Uid = user.Id
		inviteInfo.Username = user.Username
		inviteInfo.Email = user.Email
		inviteInfo.Ip = user.RegIp
		inviteInfo.RegCountry = user.RegCountry
		inviteInfo.InviterCode = strings.ToLower(inviterCode)
		inviteInfo.RegTime = user.CreateTime
		inviteInfo.InviterId = 0
		inviteInfo.InviterUsername = ""
		inviteInfo.Origin = 0
		inviteInfo.PayNum = 0
		inviteInfo.PayMoney = 0.0
		inviteInfo.Level = 1
		inviteInfo.Ratio = 0
		inviteInfo.CreateTime = util.GetNowInt()
		res := models.CreateUserActivityInviter(inviteInfo)
		fmt.Println("res", res)
	} else {
		inviterCode = pUser.InviterCode
	}
	links := strings.TrimSpace(models.GetConfigV("inviter_share_url")) + "?invite=" + inviterCode

	withList := models.GetActivityWithdrawalListBy(user.Id)
	withdrawal := 0.0
	for _, v := range withList {
		if v.Status == 2 {
			withdrawal = withdrawal + v.TrueMoney
		}
	}

	resData := map[string]interface{}{}
	resData["total"] = fmt.Sprintf("%.2f", totalMoney)           // 累计总奖励
	resData["today"] = fmt.Sprintf("%.2f", todayMoney)           // 今日佣金
	resData["money"] = fmt.Sprintf("%.2f", pUser.UsedMoney)      // 可用佣金
	resData["withdrawal"] = fmt.Sprintf("%.2f", withdrawal)      // 累计提现金额
	resData["amount_converted"] = fmt.Sprintf("%.2f", selfMoney) // 自用金额
	resData["invite_num"] = inviteNum                            // 累计邀请人数
	resData["invite"] = inviterCode                              //
	resData["invite_url"] = links                                //

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resData)
	return
}

func ActivityInviterList(c *gin.Context) {
	orderField := "id desc"
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	// 邀请记录
	where := map[string]interface{}{
		"inviter_id": uid,
	}
	inviteLists := models.GetActivityInviterListBy(where, orderField)

	//佣金记录
	mWhere := map[string]interface{}{
		"inviter_id": uid,
		"code":       10,
		"mark":       1,
		"status":     1,
	}
	_, moneyList := models.GetActivityMoneyListByMap(mWhere)
	userMoneyMap := map[int]float64{}
	for _, vm := range moneyList {
		money, ok := userMoneyMap[vm.Uid]
		if !ok {
			money = 0
		}
		userMoneyMap[vm.Uid] = money + vm.Money
	}

	lists := []models.ResUserInviter{}
	for _, v := range inviteLists {
		money, ok := userMoneyMap[v.Uid]
		if !ok {
			money = 0
		}
		info := models.ResUserInviter{}
		info.Uid = v.Uid
		info.Username = UsernameReplaceRep(v.Username)
		info.Number = v.PayNum
		info.RegTime = util.GetTimeStr(v.RegTime, "Y/m/d H:i:s")
		info.OrderMoney = v.PayMoney
		info.Money = util.StoF(fmt.Sprintf("%.2f", money))
		lists = append(lists, info)
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", lists)
	return
}

func ActivityMoneyList(c *gin.Context) {
	orderField := "id desc"
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	// 佣金记录
	where := map[string]interface{}{
		"uid":    uid,
		"code":   13,
		"mark":   -1,
		"status": 1,
	}
	moneyLists := models.GetActivityMoneyListBy(where, orderField)

	flowChar := int64(1024 * 1024 * 1024)
	lists := []models.ResUserMoneyLog{}
	for _, v := range moneyLists {
		info := models.ResUserMoneyLog{}

		flow := int(v.Ip / flowChar) //设置信息
		info.Uid = v.Uid
		info.Code = v.Code
		info.Money = v.Money
		info.Flow = flow
		info.CreateTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i:s")
		info.Cdkey = v.Cdkey
		info.Status = v.Status
		lists = append(lists, info)
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", lists)
	return
}

func ActivityInviterOrderList(c *gin.Context) {
	limitStr := c.DefaultPostForm("limit", "10")
	pageStr := c.DefaultPostForm("page", "1")

	if limitStr == "" {
		limitStr = "10"
	}
	if pageStr == "" {
		pageStr = "1"
	}
	orderField := "m.id desc"
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	limit := util.StoI(limitStr)
	page := util.StoI(pageStr)
	offset := (page - 1) * limit

	// 邀请支付记录
	moneyLists := models.GetActivityMoneyListByInvitePage(uid, offset, limit, orderField)

	lists := []models.ResUserOrderLog{}
	for _, v := range moneyLists {
		info := models.ResUserOrderLog{}
		info.Uid = v.Uid
		info.Username = v.Username
		info.Email = v.Email
		info.OrderId = v.OrderId
		info.PayMoney = float64(int(math.Round(v.Money / v.Ratio)))
		info.Money = v.Money
		info.Rate = strconv.FormatFloat(v.Ratio, 'f', 2, 64)
		info.Status = v.Status
		info.StatusText = TextReturn(c, "__T_ORDER_MONEY_LOG"+util.ItoS(v.Status))
		info.CreateTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i:s")
		lists = append(lists, info)
	}
	totalList := models.GetActivityMoneyListByInvite(uid, "id")
	totalPage := int(math.Ceil(float64(len(totalList)) / float64(limit)))
	result := map[string]interface{}{
		"total":      len(totalList),
		"total_page": totalPage,
		"lists":      lists,
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
	return
}

func ActivityWithdrawal(c *gin.Context) {
	if util.GetNowInt() > util.StoI(models.GetConfigVal("activity_inviter_withdrawal_end_time")) {
		JsonReturn(c, e.ERROR, "__T_WITHDRAWAL_TIME_END", nil)
		return
	}

	moneyStr := strings.TrimSpace(c.DefaultPostForm("money", "")) //提现金额
	wallet := strings.TrimSpace(c.DefaultPostForm("wallet", ""))  //钱包地址
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))    //联系邮箱
	resCode, msg, user := DealUser(c)                             //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	money := util.StoF(moneyStr)
	if money <= 0 {
		JsonReturn(c, e.ERROR, "__T_MONEY_ERROR", nil)
		return
	}
	if wallet == "" {
		JsonReturn(c, e.ERROR, "__T_WALLET_ERROR", nil)
		return
	}
	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", nil)
		return
	}
	if !util.CheckEmail(email) {
		JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", nil)
		return
	}
	surNum := int(money) % 10
	money = money - float64(surNum)
	//最低金额限制
	minMoneyStr := models.GetConfigVal("inviter_tx_money_min") //
	minMoney := 0.00
	if minMoneyStr == "" {
		minMoney = 50
		minMoneyStr = "50"
	} else {
		minMoney = util.StoF(minMoneyStr)
	}
	if money < minMoney {
		JsonReturn(c, e.ERROR, "__T_MONEY_NUMBER_MIN-- $"+minMoneyStr, nil)
		return
	}
	_, activityInfo := models.GetActivityInviterInfoByUid(user.Id)
	if activityInfo.UsedMoney <= 0 || activityInfo.UsedMoney < money {
		JsonReturn(c, e.ERROR, "__T_COMMISSION_LOW", nil)
		return
	}

	//写提现申请记录
	txLog := models.UserActivityWithdrawal{}
	txLog.Uid = user.Id
	txLog.Username = user.Username
	txLog.Email = email
	txLog.Money = money
	txLog.TrueMoney = money
	txLog.Wallet = wallet
	txLog.Ip = c.ClientIP()
	txLog.CreateTime = util.GetNowInt()
	txLog.Status = 1
	models.AddActivityWithdrawal(txLog)

	_, pUser := models.GetUserActivityInviterByMap(map[string]interface{}{"uid": user.Id})
	//写佣金记录
	moneyLog := models.UserActivityMoneyLog{}
	moneyLog.Uid = user.Id
	moneyLog.Code = 3 // 类型1 提现
	moneyLog.Money = money
	moneyLog.Ratio = 0
	moneyLog.Mark = -1
	moneyLog.Cdkey = ""
	moneyLog.OrderId = ""
	moneyLog.InviterId = pUser.InviterId
	moneyLog.Status = 2 //2处理中   1已打款  3已拒绝
	moneyLog.CreateTime = util.GetNowInt()
	res := models.CreateUserActivityMoneyLog(moneyLog)
	if res == nil {
		// 更新用户可用余额
		updateParams := make(map[string]interface{})
		updateParams["used_money"] = activityInfo.UsedMoney - money
		editError := models.UpdateActivityUserById(user.Id, updateParams)
		fmt.Println("update_user_used_money", editError)
		JsonReturn(c, e.SUCCESS, "__T_APPLY_SUCCESS", nil)
		return
	}
	JsonReturn(c, e.ERROR, "__T_APPLY_FAIL", nil)
	return
}

func ActivityWithdrawalLog(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	moneyLists := models.GetActivityWithdrawalListBy(uid)

	lists := []models.ResWithdrawalLog{}
	for _, v := range moneyLists {
		var dealTime = ""
		var remark = ""
		var text = ""
		var step = 0
		if v.Status != 1 {
			dealTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i:s")
			remark = v.Remark
			if v.Status == 2 {
				text = "Withdraw successfully"
			} else {
				text = "Withdrawal Failed"
			}
			step = 3
		} else {
			text = "Processing"
			step = 2
		}
		walletStr := v.Wallet
		num := len(walletStr)
		if num > 4 {
			num = 4
		}
		// 获取前四位
		first4 := walletStr[:num]
		// 获取后四位
		last4 := walletStr[len(walletStr)-num:]
		wallet := first4 + "****" + last4
		info := models.ResWithdrawalLog{}
		info.Id = v.Id
		info.Uid = v.Uid
		info.Username = v.Username
		info.Money = v.Money
		info.CreateTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i:s")
		info.Status = v.Status
		info.StatusText = text
		info.Wallet = wallet
		info.Step = step
		info.OrderNo = v.OrderNo
		info.DealTime = dealTime
		info.Remark = remark
		lists = append(lists, info)
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", lists)
	return
}

func ActivityWithdrawalDetail(c *gin.Context) {
	idStr := c.DefaultPostForm("id", "")
	id := util.StoI(idStr)
	moneyInfo := models.GetActivityWithdrawalById(id)
	var dealTime = ""
	var status = 0
	var statusText = ""
	var remark = ""
	nowTime := util.GetNowInt()
	if moneyInfo.Status == 1 {
		day := math.Ceil(float64(nowTime-moneyInfo.CreateTime) / (86400 * 7))
		status = int(day)
		statusText = TextReturn(c, "__T_TX_STEP"+util.ItoS(status))
	} else {
		dealTime = util.GetTimeStr(moneyInfo.CreateTime, "Y/m/d H:i:s")
		remark = moneyInfo.Remark
	}
	info := models.ResWithdrawalLog{}
	info.Id = moneyInfo.Id
	info.Uid = moneyInfo.Uid
	info.Username = moneyInfo.Username
	info.Money = moneyInfo.Money
	info.CreateTime = util.GetTimeStr(moneyInfo.CreateTime, "Y/m/d H:i:s")
	info.Status = status
	info.StatusText = statusText
	info.OrderNo = moneyInfo.OrderNo
	info.DealTime = dealTime
	info.Remark = remark

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", info)
	return
}

func ActivityExFlow(c *gin.Context) {
	numberStr := c.DefaultPostForm("number", "") // 兑换的数量
	number := util.StoI(numberStr)
	if numberStr == "" || number <= 0 {
		JsonReturn(c, e.ERROR, "__T_FAIL", nil)
		return
	}
	if util.GetNowInt() > util.StoI(models.GetConfigVal("activity_inviter_withdrawal_end_time")) {
		JsonReturn(c, e.ERROR, "__T_FAIL", nil)
		return
	}
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	_, activityInfo := models.GetActivityInviterInfoByUid(user.Id)
	money := 0.00
	code := 0
	flows := int64(number)
	// 计算需扣除的金额
	money = float64(number) * 2
	code = 13
	flows = int64(number) * 1024 * 1024 * 1024
	if money == 0.00 || activityInfo.UsedMoney <= 0 || activityInfo.UsedMoney < money {
		JsonReturn(c, e.ERROR, "__T_COMMISSION_LOW", nil)
		return
	}

	nowTime := util.GetNowInt()
	today := util.GetTodayTime()

	//写佣金记录
	moneyLog := models.UserActivityMoneyLog{}
	moneyLog.Uid = user.Id
	moneyLog.Code = code
	moneyLog.Money = money
	moneyLog.Ip = flows
	moneyLog.Ratio = 0
	moneyLog.Mark = -1
	moneyLog.Status = 1
	moneyLog.CreateTime = nowTime
	moneyLog.Today = today
	res := models.CreateUserActivityMoneyLog(moneyLog)
	if res == nil {
		// 获取用户信息
		userFlowInfo := models.GetUserFlowInfo(user.Id)

		//流量余额变动日志
		go models.AddUserFlowsChangeLog(models.UserFlowsChangeLog{
			Uid:         user.Id,
			OldFlows:    userFlowInfo.Flows,
			NewFlows:    userFlowInfo.Flows + flows,
			ChangeFlows: flows,
			ChangeType:  "activity_ex_flow",
			ChangeTime:  nowTime,
			Mark:        1,
		})

		if userFlowInfo.ID == 0 {
			//创建用户余额IP
			socksIp := models.UserFlow{}
			socksIp.Uid = user.Id
			socksIp.Email = user.Email
			socksIp.Username = user.Username
			socksIp.Flows = flows
			socksIp.AllFlow = flows
			socksIp.Status = 1
			socksIp.ExpireTime = 30*86400 + nowTime
			socksIp.CreateTime = nowTime
			models.CreateUserFlow(socksIp)
		} else {
			upParam := make(map[string]interface{})
			upParam["all_flow"] = flows + userFlowInfo.AllFlow
			upParam["flows"] = flows + userFlowInfo.Flows
			if userFlowInfo.ExpireTime < nowTime {
				upParam["expire_time"] = nowTime + 30*86400
			}
			models.EditUserFlow(userFlowInfo.ID, upParam)
		}

		// 更新用户可用余额
		updateParams := make(map[string]interface{})
		updateParams["used_money"] = activityInfo.UsedMoney - money
		editError := models.UpdateActivityUserById(user.Id, updateParams)
		fmt.Println("update_user_used_money", editError)
		JsonReturn(c, e.SUCCESS, "__T_SELF_USE_OK", nil)
		return
	}
	JsonReturn(c, e.ERROR, "__T_FAIL", nil)
	return
}
