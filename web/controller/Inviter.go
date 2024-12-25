package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// 获取邀请信息
// @Summary 获取邀请信息
// @Description 获取邀请信息
// @Tags 邀请返佣
// @Accept  x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce  json
// @Success 0 {object} map[string]interface{} "total：累计总奖励，today：今日佣金，money：可用佣金，withdrawal：累计提现金额，amount_converted：自用金额，invite_num：累计邀请人数，invite：邀请码，invite_url：邀请链接，unit_price：佣金兑换比例，invite_ratio：佣金兑换比例"
// @Router /web/invite/info [post]
func GetUserInviteInfo(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	//查询用户邀请记录
	err, inviteList := models.GetInviterListByMap(map[string]interface{}{"inviter_id": user.Id})
	inviteNum := 0
	if err == nil && len(inviteList) > 0 {
		inviteNum = len(inviteList)
	}

	//查询用户佣金记录
	mWhere := map[string]interface{}{
		"inviter_id": user.Id,
		"status":     1,
	}
	err, moneyList := models.GetMoneyListByMap(mWhere)
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
	err, moneyUseList := models.GetMoneyListByMap(mWhere2)
	if err == nil && len(moneyUseList) > 0 {
		for _, v := range moneyList {
			if v.Code == 12 || v.Code == 13 {
				selfMoney = selfMoney + v.Money
			}
		}
	}

	_, pUser := models.GetUserInviterByMap(map[string]interface{}{"uid": user.Id})
	inviterCode := pUser.InviterCode
	links := strings.TrimSpace(models.GetConfigV("inviter_share_url")) + "?invite=" + inviterCode

	withList := models.GetWithdrawalListBy(user.Id)
	withdrawal := 0.0
	for _, v := range withList {
		if v.Status == 2 {
			withdrawal = withdrawal + v.TrueMoney
		}
	}
	// 获取比例
	ratio := models.GetExchangeRatio("flow")
	if ratio == 0 {
		ratio = 2
	}

	//查询用户邀请记录
	_, inviteInfo := models.GetUserInviterByUid(user.Id)
	where := map[string]interface{}{
		"cate":  "level",
		"level": inviteInfo.Level,
	}
	ratioInfo := models.GetConfLevelBy(where, "level asc")
	percent := ratioInfo.Ratio * 100

	resData := map[string]interface{}{}
	resData["total"] = fmt.Sprintf("%.2f", totalMoney)            // 累计总奖励
	resData["today"] = fmt.Sprintf("%.2f", todayMoney)            // 今日佣金
	resData["money"] = fmt.Sprintf("%.2f", user.UsedMoney)        // 可用佣金
	resData["withdrawal"] = fmt.Sprintf("%.2f", withdrawal)       // 累计提现金额
	resData["amount_converted"] = fmt.Sprintf("%.2f", selfMoney)  // 自用金额
	resData["invite_num"] = inviteNum                             // 累计邀请人数
	resData["invite"] = inviterCode                               //
	resData["invite_url"] = links                                 //
	resData["unit_price"] = ratio                                 // 佣金兑换比例
	resData["invite_ratio"] = fmt.Sprintf("%.0f", percent) + " %" // 佣金兑换比例

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resData)
	return
}

// 邀请 列表
// @BasePath /api/v1
// @Summary 邀请 列表
// @Description 邀请 列表
// @Tags 邀请返佣
// @Accept  x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce  json
// @Success 0 {object} []models.ResUserInviter{} "邀请列表"
// @Router /web/invite/record [post]
func InviterList(c *gin.Context) {
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
	inviteLists := models.GetInviterListBy(where, orderField)

	//佣金记录
	mWhere := map[string]interface{}{
		"inviter_id": uid,
		"code":       10,
		"mark":       1,
	}
	_, moneyList := models.GetMoneyListByMap(mWhere)
	userMoneyMap := map[int]float64{}
	for _, vm := range moneyList {
		money, ok := userMoneyMap[vm.Uid]
		if !ok {
			money = 0
		}
		userMoneyMap[vm.Uid] = money + vm.Money
	}

	//大客户佣金记录
	lists := []models.ResUserInviter{}
	keyUserWhite := models.GetKeyUserWhiteBy(uid) //获取主要用户名单配置
	if keyUserWhite.Id > 0 {
		_, keyUserInfo := models.GetKeyUserMoneyListByUid(uid)
		if len(keyUserInfo) > 0 {

			keyUserMoney := 0.0
			orderIdArr := []string{}
			for _, v := range keyUserInfo {
				keyUserMoney = keyUserMoney + v.Money
				orderIdArr = append(orderIdArr, v.OrderId)
			}
			keyUserOrderMoney := models.GetOrderTotalMoney(orderIdArr)
			KeyInfo := models.ResUserInviter{}
			KeyInfo.Uid = uid
			KeyInfo.Username = UsernameReplaceRep(user.Username)
			KeyInfo.Number = len(keyUserInfo)
			KeyInfo.RegTime = util.GetTimeStr(user.CreateTime, "Y/m/d H:i:s")
			KeyInfo.OrderMoney = util.StoF(fmt.Sprintf("%.2f", keyUserOrderMoney))
			KeyInfo.Money = util.StoF(fmt.Sprintf("%.2f", keyUserMoney))
			lists = append(lists, KeyInfo)
		}
	}

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

// 用户名加*
func UsernameReplaceRep(str string) string {
	re, _ := regexp.Compile("(\\d{2})(\\d{4})(\\d{2})")
	return re.ReplaceAllString(str, "$1****$3")
}

// 佣金操作 列表
// @BasePath /api/v1
// @Summary 佣金记录
// @Description 佣金记录
// @Tags 邀请返佣
// @Accept  x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce  json
// @Success 0 {object} []models.ResUserMoneyLog{} "佣金记录"
// @Router /web/invite/money_record [post]
func MoneyList(c *gin.Context) {
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
	moneyLists := models.GetMoneyListBy(where, orderField)

	flowChar := int64(1024 * 1024 * 1024)
	lists := []models.ResUserMoneyLog{}
	for _, v := range moneyLists {
		info := models.ResUserMoneyLog{}

		flow := int(v.Ip / flowChar) //设置信息
		info.Uid = v.Uid
		info.Code = v.Code
		info.Money = util.StoF(fmt.Sprintf("%.2f", v.Money))
		info.Flow = flow
		info.CreateTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i:s")
		info.Cdkey = v.Cdkey
		info.Status = v.Status
		lists = append(lists, info)
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", lists)
	return
}

// 邀请用户购买记录列表
// @BasePath /api/v1
// @Summary 邀请用户购买记录
// @Description 邀请用户购买记录
// @Tags 邀请返佣
// @Accept  x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param limit formData string true "每页显示数量"
// @Param page formData string true "页数"
// @Produce  json
// @Success 0 {object} map[string]interface{} "total：总数，lists []models.ResUserOrderLog{} 列表数据 total_page：总页数"
// @Router /web/invite/pay_record [post]
func InviterOrderList(c *gin.Context) {
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
	moneyLists := models.GetMoneyListByInvitePage(uid, offset, limit, orderField)

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
	totalList := models.GetMoneyListByInvite(uid, "id")
	totalPage := int(math.Ceil(float64(len(totalList)) / float64(limit)))
	result := map[string]interface{}{
		"total":      len(totalList),
		"total_page": totalPage,
		"lists":      lists,
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
	return
}

// 点击统计
//func StatsClick(c *gin.Context) {
//	code := strings.TrimSpace(c.DefaultPostForm("code", ""))
//	url := strings.TrimSpace(c.DefaultPostForm("url", ""))
//	session := strings.TrimSpace(c.DefaultPostForm("session", ""))
//
//	uid := 0
//	if session != "" {
//		var res = false
//		res, uid = GetUIDbySession(session)
//		fmt.Println(res)
//	}
//	ip := c.ClientIP()
//	nowTime := util.GetNowInt()
//
//	var res error
//	info := models.StatsClickModel{}
//	info.CreateTime = nowTime
//	info.Ip = ip
//	info.Uid = uid
//	info.Code = code
//	info.Url = url
//	res = models.AddStatsClick(info)
//
//	JsonReturn(c, 0, "__T_SUCCESS", res)
//	return
//}

// 用户佣金余额转换成提取余额
// @BasePath /api/v1
// @Summary 佣金余额兑换余额
// @Description 佣金余额兑换余额
// @Tags 邀请返佣
// @Accept  x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce  json
// @Success 0 {object} interface{} "成功"
// @Router /web/invite/ex_balance [post]
func ExBalance(c *gin.Context) {
	//moneyStr := strings.TrimSpace(c.DefaultPostForm("money",""))	//提现金额
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	//money := util.StoF(moneyStr)
	//if money == 0{
	//	JsonReturn(c, e.ERROR, "__T_MONEY_ERROR", nil)
	//	return
	//}

	money := user.UsedMoney

	//最低金额限制
	//minMoneyStr := models.GetConfigVal("inviter_ex_balance_min")	//
	//minMoney := 0.00
	//if minMoneyStr == "" {
	//	minMoney = 100
	//	minMoneyStr = "100"
	//}else{
	//	minMoney = util.StoF(minMoneyStr)
	//}
	//if money < minMoney {
	//	JsonReturn(c, e.ERROR, "__T_MONEY_NUMBER_MIN-- $"+ minMoneyStr, nil)
	//	return
	//}
	// 提现手续费
	//withRatioStr := models.GetConfigVal("inviter_ex_balance_ratio")
	//withRatio := util.StoF(withRatioStr)
	//if withRatio == 0.0 {
	//	withRatio = 0.1
	//}
	//fei := money * withRatio
	//total := money + fei

	total := money
	if user.UsedMoney <= 0 || user.UsedMoney < total {
		JsonReturn(c, e.ERROR, "__T_COMMISSION_LOW", nil)
		return
	}

	_, pUser := models.GetUserInviterByMap(map[string]interface{}{"uid": user.Id})
	//写佣金记录
	moneyLog := models.UserMoneyLog{}
	moneyLog.Uid = user.Id
	moneyLog.Code = 1 // 类型1 自用
	moneyLog.Money = money
	moneyLog.Ratio = 0
	moneyLog.Mark = -1
	moneyLog.Cdkey = ""
	moneyLog.OrderId = ""
	moneyLog.InviterId = pUser.InviterId
	moneyLog.Status = 1
	moneyLog.CreateTime = util.GetNowInt()
	res := models.CreateUserMoneyLog(moneyLog)
	if res == nil {
		// 更新用户可用余额
		updateParams := make(map[string]interface{})
		updateParams["used_money"] = user.UsedMoney - money
		updateParams["balance"] = user.Money + money
		editError := models.UpdateUserById(user.Id, updateParams)
		fmt.Println("update_user_used_money", editError)
		// 更新用户IP余额
		JsonReturn(c, e.SUCCESS, "__T_APPLE_SUCCESS", nil)
		return
	}
	JsonReturn(c, e.ERROR, "__T_APPLE_FAIL", nil)
	return
}

// 用户佣金余额转换成用户流量
// @BasePath /api/v1
// @Summary 佣金余额兑换余额
// @Description 佣金余额兑换余额
// @Tags 邀请返佣
// @Accept  x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param number formData string true "兑换数量"
// @Param cate formData string true "兑换类型 isp  /  flow"
// @Produce  json
// @Success 0 {object} interface{} "成功"
// @Router /web/invite/ex_flow [post]
func ExFlow(c *gin.Context) {
	cate := c.DefaultPostForm("cate", "flow")    // 兑换的类型   isp  /  flow
	numberStr := c.DefaultPostForm("number", "") // 兑换的数量
	number := util.StoI(numberStr)
	if cate == "" {
		cate = "flow"
	}
	if cate == "" || numberStr == "" || number <= 0 {
		JsonReturn(c, e.ERROR, "__T_FAIL", nil)
		return
	}

	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	// 获取比例
	ratio := models.GetExchangeRatio(cate)
	if ratio == 0.00 {
		JsonReturn(c, e.ERROR, "__T_FAIL", nil)
		return
	}
	money := 0.00
	code := 0
	flows := int64(number)
	// 计算需扣除的金额
	if cate == "isp" {
		money = float64(number) * ratio
		code = 12
	} else if cate == "flow" {
		money = float64(number) * ratio
		code = 13
		flows = int64(number) * 1024 * 1024 * 1024
	}

	if ratio == 0.00 || money == 0.00 || user.UsedMoney <= 0 || user.UsedMoney < money {
		JsonReturn(c, e.ERROR, "__T_COMMISSION_LOW", nil)
		return
	}

	nowTime := util.GetNowInt()
	today := util.GetTodayTime()
	////写佣金记录
	//moneyLog := models.UserMoneyLog{}
	//moneyLog.Uid = user.Id
	//moneyLog.Code = code
	//moneyLog.Money = money
	//moneyLog.Ip = flows
	//moneyLog.Ratio = ratio
	//moneyLog.Mark = -1
	//moneyLog.Status = 1
	//moneyLog.CreateTime = nowTime
	//moneyLog.Today = today
	//res := models.CreateUserMoneyLog(moneyLog)
	//if res == nil {
	//updateParams := make(map[string]interface{})
	//updateParams["used_money"] = user.UsedMoney - money
	//if cate == "isp" {
	//	// 更新用户isp余额
	//	updateParams["balance"] = user.Balance + number
	//} else if cate == "flow" {
	//	/// 获取用户信息
	//	//userFlowInfo := models.GetUserFlowInfo(user.Id)
	//	//if userFlowInfo.ID == 0 {
	//	//	//创建用户余额IP
	//	//	socksIp := models.UserFlow{}
	//	//	socksIp.Uid = user.Id
	//	//	socksIp.Email = user.Email
	//	//	socksIp.Username = user.Username
	//	//	socksIp.Flows = flows
	//	//	socksIp.AllFlow = flows
	//	//	socksIp.ExpireTime = 30*86400 + nowTime
	//	//	socksIp.CreateTime = nowTime
	//	//	models.CreateUserFlow(socksIp)
	//	//} else {
	//		//upParam := make(map[string]interface{})
	//		//upParam["all_flow"] = flows + userFlowInfo.AllFlow
	//		//upParam["flows"] = flows + userFlowInfo.Flows
	//		//if userFlowInfo.ExpireTime < nowTime {
	//		//	upParam["expire_time"] = nowTime + 30*86400
	//		//}
	//		//models.EditUserFlow(userFlowInfo.ID, upParam)
	//	}
	//	// 更新用户流量余额
	//	editError := models.UpdateUserById(user.Id, updateParams)
	//	fmt.Println("update_user_used_money", editError)
	//}
	// 更新用户信息
	//editError := models.UpdateUserById(user.Id, updateParams)
	//fmt.Println("update_user_used_money", editError)
	// 更新变成 异步

	//}

	dealInfo := models.PushInvite{}
	// 异步处理流量
	dealInfo.Cate = cate
	dealInfo.Code = code
	dealInfo.Uid = user.Id
	dealInfo.Value = flows
	dealInfo.Ratio = ratio
	dealInfo.Money = money
	dealInfo.CreateTime = nowTime
	dealInfo.Today = today
	listStr, _ := json.Marshal(dealInfo)
	resP := models.RedisLPUSH("list_invite_balance", string(listStr))
	fmt.Println(resP)

	JsonReturn(c, e.SUCCESS, "__T_EX_RECHARGE_OK", nil)
	return
	//JsonReturn(c, e.ERROR, "__T_APPLE_FAIL", nil)
	//return
}

// 用户提现
// @BasePath /api/v1
// @Summary 申请提现
// @Description 申请提现
// @Tags 邀请返佣
// @Accept  x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param money formData string true "提现金额"
// @Param wallet formData string true "钱包地址"
// @Param email formData string true "联系邮箱"
// @Produce  json
// @Success 0 {object} interface{}
// @Router /web/invite/withdrawal [post]
func Withdrawal(c *gin.Context) {
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
	// 提现手续费
	//withRatioStr := models.GetConfigVal("inviter_tx_money_ratio")
	//withRatio := util.StoF(withRatioStr)
	//if withRatio == 0.0 {
	//	withRatio = 0.1
	//}
	//fei := money * withRatio
	//total := money + fei
	//if user.UsedMoney <= 0 || user.UsedMoney < total {
	//	JsonReturn(c, e.ERROR, "__T_COMMISSION_LOW", nil)
	//	return
	//}
	if user.UsedMoney <= 0 || user.UsedMoney < money {
		JsonReturn(c, e.ERROR, "__T_COMMISSION_LOW", nil)
		return
	}

	//写提现申请记录
	txLog := models.UserWithdrawalModel{}
	txLog.Uid = user.Id
	txLog.Username = user.Username
	txLog.Email = email
	txLog.Money = money
	txLog.TrueMoney = money
	txLog.Wallet = wallet
	txLog.Ip = c.ClientIP()
	txLog.CreateTime = util.GetNowInt()
	txLog.Status = 1
	res := models.AddWithdrawal(txLog)

	_, pUser := models.GetUserInviterByMap(map[string]interface{}{"uid": user.Id})
	//写佣金记录
	moneyLog := models.UserMoneyLog{}
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
	res = models.CreateUserMoneyLog(moneyLog)
	if res == nil {
		// 更新用户可用余额
		updateParams := make(map[string]interface{})
		updateParams["used_money"] = user.UsedMoney - money
		editError := models.UpdateUserById(user.Id, updateParams)
		fmt.Println("update_user_used_money", editError)
		// 更新用户IP余额
		JsonReturn(c, e.SUCCESS, "__T_APPLY_SUCCESS", nil)
		return
	}
	JsonReturn(c, e.ERROR, "__T_APPLY_FAIL", nil)
	return
}

// 提现记录
// @BasePath /api/v1
// @Summary 提现记录
// @Description 提现记录
// @Tags 邀请返佣
// @Accept  x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce  json
// @Success 0 {object} []models.ResWithdrawalLog{}
// @Router /web/invite/withdrawal_record [post]
func WithdrawalLog(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	moneyLists := models.GetWithdrawalListBy(uid)

	nowTime := util.GetNowInt()
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
			day := math.Ceil(float64(nowTime-v.CreateTime) / (86400 * 15))
			step = int(day)
			if step > 2 {
				step = 2
			}
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

// 提现详情
// @BasePath /api/v1
// @Summary 提现记录详细信息
// @Description 提现记录详细信息
// @Tags 邀请返佣
// @Accept  x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param id formData string true "提现记录id"
// @Produce  json
// @Success 0 {object} models.ResWithdrawalLog{}
// @Router /web/invite/withdrawal_info [post]
func WithdrawalDetail(c *gin.Context) {
	idStr := c.DefaultPostForm("id", "")
	id := util.StoI(idStr)
	moneyInfo := models.GetWithdrawalById(id)
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

// 设置邀请码信息
// @BasePath /api/v1
// @Summary 设置邀请码
// @Description 设置邀请码
// @Tags 邀请返佣
// @Accept  x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param invite formData string true "邀请码"
// @Produce  json
// @Success 0 {object} map[string]interface{} "invite：邀请码；invite_url：邀请链接"
// @Router /web/invite/set_code [post]
func SetInviteCode(c *gin.Context) {
	code := strings.TrimSpace(c.DefaultPostForm("invite", ""))
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	if !util.CheckCode(code) {
		JsonReturn(c, -1, "__T_INVITE_CODE_FORMAT", nil)
		return
	}
	//查询用户邀请信息
	err, inviteInfo := models.GetUserInviterByCode(code)

	if err == nil && inviteInfo.ID > 0 {
		JsonReturn(c, e.ERROR, "__T_INVITE_CODE_HAS", nil)
		return
	}
	upInvite := models.UserInviter{}
	upInvite.InviterCode = code
	eRes := models.EditUserInviter(user.Id, upInvite)

	links := strings.TrimSpace(models.GetConfigV("inviter_share_url")) + "?invite=" + code
	resData := map[string]interface{}{}
	resData["invite"] = code      //
	resData["invite_url"] = links //
	if eRes == nil {
		JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resData)
		return
	} else {
		JsonReturn(c, e.SUCCESS, "__T_FAIL", nil)
		return
	}
}
