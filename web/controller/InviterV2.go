package controller
//
//import (
//	"api-360proxy/web/e"
//	"api-360proxy/web/models"
//	"api-360proxy/web/pkg/util"
//	"fmt"
//	"github.com/gin-gonic/gin"
//	"math"
//	"regexp"
//	"strconv"
//	"strings"
//)
//
//// 获取邀请信息
//func GetUserInviteInfo(c *gin.Context) {
//	resCode, msg, user := DealUser(c) //处理用户信息
//	if resCode != e.SUCCESS {
//		JsonReturn(c, resCode, msg, nil)
//		return
//	}
//	//查询用户邀请记录
//	err, inviteList := models.GetInviterListByMap(map[string]interface{}{"inviter_id": user.Id})
//	inviteNum := 0
//	if err == nil && len(inviteList) > 0 {
//		inviteNum = len(inviteList)
//	}
//
//	//查询用户佣金记录
//	mWhere := map[string]interface{}{
//		"inviter_id": user.Id,
//		"code":       10,
//		"mark":       1,
//	}
//	err, moneyList := models.GetMoneyListByMap(mWhere)
//	totalMoney := 0.00  //累计总奖励
//	freezeMoney := 0.00 //冻结余额
//	todayMoney := 0.00  //今日佣金
//	today := util.GetTodayTime()
//	if err == nil && len(inviteList) > 0 {
//		for _, v := range moneyList {
//			totalMoney = totalMoney + v.Money //
//			if v.Status == 2 {                //2冻结中的金额
//				freezeMoney = freezeMoney + v.Money
//			}
//			if v.CreateTime >= today {
//				todayMoney = todayMoney + v.Money
//			}
//		}
//	}
//
//	_, pUser := models.GetUserInviterByMap(map[string]interface{}{"uid": user.Id})
//	inviterCode := pUser.InviterCode
//	links := strings.TrimSpace(models.GetConfigV("inviter_share_url")) + "?invite=" + inviterCode
//
//	totalInfoList := models.GetInviterListBy(map[string]interface{}{"inviter_id": user.Id}, "id desc")
//	tMoney := 0.00 //累计总邀请金额
//	for _, v := range totalInfoList {
//		tMoney = tMoney + v.PayMoney
//	}
//	// 累计总佣金金额
//	totalMoney1 := tMoney * 0.05
//	if totalMoney != totalMoney1 {
//		totalMoney = totalMoney1
//	}
//	withList := models.GetWithdrawalListBy(user.Id)
//	withdrawal := 0.0
//	for _, v := range withList {
//		if v.Status == 2 {
//			withdrawal = withdrawal + v.TrueMoney
//		}
//	}
//	money_y := 0.0
//	if user.UsedMoney > 0 {
//		money_y = user.UsedMoney / 0.05
//	}
//
//	resData := map[string]interface{}{}
//	resData["invite_num"] = inviteNum  // 累计邀请人数
//	resData["freeze"] = fmt.Sprintf("%.2f",freezeMoney)    // 冻结余额
//	resData["today"] = fmt.Sprintf("%.2f",todayMoney)      // 今日佣金
//	resData["money"] = fmt.Sprintf("%.2f",user.UsedMoney)  // 可用佣金
//	resData["money_y"] = fmt.Sprintf("%.2f",money_y)       // 可用佣金 源
//	resData["withdrawal"] = fmt.Sprintf("%.2f",withdrawal) // 累计提现金额
//	resData["invite"] = inviterCode    //
//	resData["invite_url"] = links      //
//	resData["total_money"] = fmt.Sprintf("%.2f",tMoney)    // 累计总邀请金额
//	resData["total"] = fmt.Sprintf("%.2f",totalMoney)      // 累计总奖励
//
//	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resData)
//	return
//}
//
//// 邀请 列表
//func InviterList(c *gin.Context) {
//	limitStr := c.DefaultPostForm("limit", "10")
//	pageStr := c.DefaultPostForm("page", "1")
//
//	if limitStr == "" {
//		limitStr = "10"
//	}
//	if pageStr == "" {
//		pageStr = "1"
//	}
//	orderField := "id desc"
//	resCode, msg, user := DealUser(c) //处理用户信息
//	if resCode != e.SUCCESS {
//		JsonReturn(c, resCode, msg, nil)
//		return
//	}
//	uid := user.Id
//
//	limit := util.StoI(limitStr)
//	page := util.StoI(pageStr)
//	offset := (page - 1) * limit
//
//	// 邀请记录
//	where := map[string]interface{}{
//		"inviter_id": uid,
//	}
//	inviteLists := models.GetInviterListPage(where, offset, limit, orderField)
//
//	//佣金记录
//	//mWhere := map[string]interface{}{
//	//	"inviter_id"	:uid,
//	//	"code"			:10,
//	//	"mark"			:1,
//	//}
//	//err ,moneyList := models.GetMoneyListByMap(mWhere)
//
//	lists := []models.ResUserInviter{}
//	for _, v := range inviteLists {
//		//money := 0.00
//		//if err == nil && len(moneyList) > 0 {
//		//	for _,vm := range moneyList {
//		//		if vm.Uid == v.Uid {
//		//			money = money + vm.Money		//
//		//		}
//		//	}
//		//}
//		info := models.ResUserInviter{}
//		info.Uid = v.Uid
//		info.Username = UsernameReplaceRep(v.Username)
//		info.Number = v.PayNum
//		info.RegTime = util.GetTimeStr(v.RegTime, "Y/m/d H:i:s")
//		info.Money = v.PayMoney
//		lists = append(lists, info)
//	}
//	totalList := models.GetInviterListBy(where, orderField)
//	totalPage := int(math.Ceil(float64(len(totalList)) / float64(limit)))
//	result := map[string]interface{}{
//		"total":      len(totalList),
//		"total_page": totalPage,
//		"lists":      lists,
//	}
//	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
//	return
//}
//
//// 用户名加*
//func UsernameReplaceRep(str string) string {
//	re, _ := regexp.Compile("(\\d{2})(\\d{4})(\\d{2})")
//	return re.ReplaceAllString(str, "$1****$3")
//}
//
//// 佣金操作 列表
//func MoneyList(c *gin.Context) {
//	limitStr := c.DefaultPostForm("limit", "10")
//	pageStr := c.DefaultPostForm("page", "1")
//
//	if limitStr == "" {
//		limitStr = "10"
//	}
//	if pageStr == "" {
//		pageStr = "1"
//	}
//	orderField := "id desc"
//	resCode, msg, user := DealUser(c) //处理用户信息
//	if resCode != e.SUCCESS {
//		JsonReturn(c, resCode, msg, nil)
//		return
//	}
//	uid := user.Id
//
//	limit := util.StoI(limitStr)
//	page := util.StoI(pageStr)
//	offset := (page - 1) * limit
//
//	// 佣金记录
//	where := map[string]interface{}{
//		"uid":  uid,
//		"mark": -1,
//	}
//	moneyLists := models.GetMoneyListPage(where, offset, limit, orderField)
//
//	lists := []models.ResUserMoneyLog{}
//	for _, v := range moneyLists {
//		info := models.ResUserMoneyLog{}
//		statusText := ""
//		if v.Code == 3 { //2处理中   1已打款  3已拒绝
//			statusText = TextReturn(c, "__T_TX_MONEY_STATUS"+util.ItoS(v.Status))
//		} else {
//			//1 成功
//			statusText = TextReturn(c, "__T_ORDER_MONEY_LOG"+util.ItoS(v.Status))
//		}
//		info.Uid = v.Uid
//		info.Code = v.Code
//		info.CodeText = TextReturn(c, "__T_MONEY_LOG_CODE"+util.ItoS(v.Code))
//		info.Money = v.Money
//		info.CreateTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i:s")
//		info.Cdkey = v.Cdkey
//		info.Status = v.Status
//		info.StatusText = statusText
//		lists = append(lists, info)
//	}
//	totalList := models.GetMoneyListBy(where, orderField)
//	totalPage := int(math.Ceil(float64(len(totalList)) / float64(limit)))
//	result := map[string]interface{}{
//		"total":      len(totalList),
//		"total_page": totalPage,
//		"lists":      lists,
//	}
//	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
//	return
//}
//
//// 邀请用户购买记录列表
//func InviterOrderList(c *gin.Context) {
//	limitStr := c.DefaultPostForm("limit", "10")
//	pageStr := c.DefaultPostForm("page", "1")
//
//	if limitStr == "" {
//		limitStr = "10"
//	}
//	if pageStr == "" {
//		pageStr = "1"
//	}
//	orderField := "m.id desc"
//	resCode, msg, user := DealUser(c) //处理用户信息
//	if resCode != e.SUCCESS {
//		JsonReturn(c, resCode, msg, nil)
//		return
//	}
//	uid := user.Id
//
//	limit := util.StoI(limitStr)
//	page := util.StoI(pageStr)
//	offset := (page - 1) * limit
//
//	// 邀请支付记录
//	moneyLists := models.GetMoneyListByInvitePage(uid, offset, limit, orderField)
//
//	lists := []models.ResUserOrderLog{}
//	for _, v := range moneyLists {
//		info := models.ResUserOrderLog{}
//		info.Uid = v.Uid
//		info.Username = v.Username
//		info.Email = v.Email
//		info.OrderId = v.OrderId
//		info.PayMoney = float64(int(math.Round(v.Money / v.Ratio)))
//		info.Money = v.Money
//		info.Rate = strconv.FormatFloat(v.Ratio, 'f', 2, 64)
//		info.Status = v.Status
//		info.StatusText = TextReturn(c, "__T_ORDER_MONEY_LOG"+util.ItoS(v.Status))
//		info.CreateTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i:s")
//		lists = append(lists, info)
//	}
//	totalList := models.GetMoneyListByInvite(uid, "id")
//	totalPage := int(math.Ceil(float64(len(totalList)) / float64(limit)))
//	result := map[string]interface{}{
//		"total":      len(totalList),
//		"total_page": totalPage,
//		"lists":      lists,
//	}
//	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
//	return
//}
//
//// 点击统计
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
//
//// 用户佣金余额转换成提取余额
//func ExBalance(c *gin.Context) {
//	//moneyStr := strings.TrimSpace(c.DefaultPostForm("money",""))	//提现金额
//	resCode, msg, user := DealUser(c) //处理用户信息
//	if resCode != e.SUCCESS {
//		JsonReturn(c, resCode, msg, nil)
//		return
//	}
//	//money := util.StoF(moneyStr)
//	//if money == 0{
//	//	JsonReturn(c, e.ERROR, "__T_MONEY_ERROR", nil)
//	//	return
//	//}
//
//	money := user.UsedMoney
//
//	//最低金额限制
//	//minMoneyStr := models.GetConfigVal("inviter_ex_balance_min")	//
//	//minMoney := 0.00
//	//if minMoneyStr == "" {
//	//	minMoney = 100
//	//	minMoneyStr = "100"
//	//}else{
//	//	minMoney = util.StoF(minMoneyStr)
//	//}
//	//if money < minMoney {
//	//	JsonReturn(c, e.ERROR, "__T_MONEY_NUMBER_MIN-- $"+ minMoneyStr, nil)
//	//	return
//	//}
//	// 提现手续费
//	//withRatioStr := models.GetConfigVal("inviter_ex_balance_ratio")
//	//withRatio := util.StoF(withRatioStr)
//	//if withRatio == 0.0 {
//	//	withRatio = 0.1
//	//}
//	//fei := money * withRatio
//	//total := money + fei
//
//	total := money
//	if user.UsedMoney <= 0 || user.UsedMoney < total {
//		JsonReturn(c, e.ERROR, "__T_COMMISSION_LOW", nil)
//		return
//	}
//
//	_, pUser := models.GetUserInviterByMap(map[string]interface{}{"uid": user.Id})
//	//写佣金记录
//	moneyLog := models.UserMoneyLog{}
//	moneyLog.Uid = user.Id
//	moneyLog.Code = 1 // 类型1 自用
//	moneyLog.Money = money
//	moneyLog.Ratio = 0
//	moneyLog.Mark = -1
//	moneyLog.Cdkey = ""
//	moneyLog.OrderId = ""
//	moneyLog.InviterId = pUser.InviterId
//	moneyLog.Status = 1
//	moneyLog.CreateTime = util.GetNowInt()
//	res := models.CreateUserMoneyLog(moneyLog)
//	if res == nil {
//		// 更新用户可用余额
//		updateParams := make(map[string]interface{})
//		updateParams["used_money"] = user.UsedMoney - money
//		updateParams["balance"] = user.Money + money
//		editError := models.UpdateUserById(user.Id, updateParams)
//		fmt.Println("update_user_used_money", editError)
//		// 更新用户IP余额
//		JsonReturn(c, e.SUCCESS, "__T_APPLE_SUCCESS", nil)
//		return
//	}
//	JsonReturn(c, e.ERROR, "__T_APPLE_FAIL", nil)
//	return
//}
//
//// 用户提现
//func Withdrawal(c *gin.Context) {
//	moneyStr := strings.TrimSpace(c.DefaultPostForm("money", "")) //提现金额
//	wallet := strings.TrimSpace(c.DefaultPostForm("wallet", ""))  //钱包地址
//	email := strings.TrimSpace(c.DefaultPostForm("email", ""))    //联系邮箱
//	resCode, msg, user := DealUser(c)                             //处理用户信息
//	if resCode != e.SUCCESS {
//		JsonReturn(c, resCode, msg, nil)
//		return
//	}
//	money := util.StoF(moneyStr)
//	if money == 0 {
//		JsonReturn(c, e.ERROR, "__T_MONEY_ERROR", nil)
//		return
//	}
//	if wallet == "" {
//		JsonReturn(c, e.ERROR, "__T_WALLET_ERROR", nil)
//		return
//	}
//	if email == "" {
//		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", nil)
//		return
//	}
//	if !util.CheckEmail(email) {
//		JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", nil)
//		return
//	}
//	//最低金额限制
//	minMoneyStr := models.GetConfigVal("inviter_tx_money_min") //
//	minMoney := 0.00
//	if minMoneyStr == "" {
//		minMoney = 100
//		minMoneyStr = "100"
//	} else {
//		minMoney = util.StoF(minMoneyStr)
//	}
//	if money < minMoney {
//		JsonReturn(c, e.ERROR, "__T_MONEY_NUMBER_MIN-- $"+minMoneyStr, nil)
//		return
//	}
//	// 提现手续费
//	withRatioStr := models.GetConfigVal("inviter_tx_money_ratio")
//	withRatio := util.StoF(withRatioStr)
//	if withRatio == 0.0 {
//		withRatio = 0.1
//	}
//	fei := money * withRatio
//	total := money + fei
//	if user.UsedMoney <= 0 || user.UsedMoney < total {
//		JsonReturn(c, e.ERROR, "__T_COMMISSION_LOW", nil)
//		return
//	}
//
//	//写提现申请记录
//	txLog := models.UserWithdrawalModel{}
//	txLog.Uid = user.Id
//	txLog.Username = user.Username
//	txLog.Email = email
//	txLog.Money = money
//	txLog.TrueMoney = money
//	txLog.Wallet = wallet
//	txLog.Ip = c.ClientIP()
//	txLog.CreateTime = util.GetNowInt()
//	txLog.Status = 1
//	res := models.AddWithdrawal(txLog)
//
//	_, pUser := models.GetUserInviterByMap(map[string]interface{}{"uid": user.Id})
//	//写佣金记录
//	moneyLog := models.UserMoneyLog{}
//	moneyLog.Uid = user.Id
//	moneyLog.Code = 3 // 类型1 提现
//	moneyLog.Money = money
//	moneyLog.Ratio = 0
//	moneyLog.Mark = -1
//	moneyLog.Cdkey = ""
//	moneyLog.OrderId = ""
//	moneyLog.InviterId = pUser.InviterId
//	moneyLog.Status = 2 //2处理中   1已打款  3已拒绝
//	moneyLog.CreateTime = util.GetNowInt()
//	res = models.CreateUserMoneyLog(moneyLog)
//	if res == nil {
//		// 更新用户可用余额
//		updateParams := make(map[string]interface{})
//		updateParams["used_money"] = user.UsedMoney - money
//		editError := models.UpdateUserById(user.Id, updateParams)
//		fmt.Println("update_user_used_money", editError)
//		// 更新用户IP余额
//		JsonReturn(c, e.SUCCESS, "__T_APPLY_SUCCESS", nil)
//		return
//	}
//	JsonReturn(c, e.ERROR, "__T_APPLY_FAIL", nil)
//	return
//}
//
//// 提现记录
//func WithdrawalLog(c *gin.Context) {
//	limitStr := c.DefaultPostForm("limit", "10")
//	pageStr := c.DefaultPostForm("page", "1")
//
//	if limitStr == "" {
//		limitStr = "10"
//	}
//	if pageStr == "" {
//		pageStr = "1"
//	}
//
//	resCode, msg, user := DealUser(c) //处理用户信息
//	if resCode != e.SUCCESS {
//		JsonReturn(c, resCode, msg, nil)
//		return
//	}
//	uid := user.Id
//
//	limit := util.StoI(limitStr)
//	page := util.StoI(pageStr)
//	offset := (page - 1) * limit
//
//	moneyLists := models.GetWithdrawalPageBy(uid, offset, limit)
//
//	lists := []models.ResWithdrawalLog{}
//	for _, v := range moneyLists {
//		var dealTime = ""
//		var remark = ""
//		var text = ""
//		if v.Status != 1 {
//			dealTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i:s")
//			remark = v.Remark
//			if v.Status == 2 {
//				text = "Withdraw successfully"
//			} else {
//				text = "Withdrawal Failed"
//			}
//		} else {
//			text = "Processing"
//		}
//		info := models.ResWithdrawalLog{}
//		info.Id = v.Id
//		info.Uid = v.Uid
//		info.Username = v.Username
//		info.Money = v.Money
//		info.CreateTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i:s")
//		info.Status = v.Status
//		info.StatusText = text
//		info.OrderNo = v.OrderNo
//		info.DealTime = dealTime
//		info.Remark = remark
//		lists = append(lists, info)
//	}
//	totalList := models.GetWithdrawalListBy(uid)
//	totalPage := int(math.Ceil(float64(len(totalList)) / float64(limit)))
//	result := map[string]interface{}{
//		"total":      len(totalList),
//		"total_page": totalPage,
//		"lists":      lists,
//	}
//	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
//	return
//}
//
//// 提现详情
//func WithdrawalDetail(c *gin.Context) {
//	idStr := c.DefaultPostForm("id", "")
//	id := util.StoI(idStr)
//	moneyInfo := models.GetWithdrawalById(id)
//	var dealTime = ""
//	var status = 0
//	var statusText = ""
//	var remark = ""
//	nowTime := util.GetNowInt()
//	if moneyInfo.Status == 1 {
//		day := math.Ceil(float64(nowTime-moneyInfo.CreateTime) / (86400 * 7))
//		status = int(day)
//		statusText = TextReturn(c, "__T_TX_STEP"+util.ItoS(status))
//	} else {
//		dealTime = util.GetTimeStr(moneyInfo.CreateTime, "Y/m/d H:i:s")
//		remark = moneyInfo.Remark
//	}
//	info := models.ResWithdrawalLog{}
//	info.Id = moneyInfo.Id
//	info.Uid = moneyInfo.Uid
//	info.Username = moneyInfo.Username
//	info.Money = moneyInfo.Money
//	info.CreateTime = util.GetTimeStr(moneyInfo.CreateTime, "Y/m/d H:i:s")
//	info.Status = status
//	info.StatusText = statusText
//	info.OrderNo = moneyInfo.OrderNo
//	info.DealTime = dealTime
//	info.Remark = remark
//
//	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", info)
//	return
//}
//
//// 设置邀请码信息
//func SetInviteCode(c *gin.Context) {
//	code := strings.TrimSpace(c.DefaultPostForm("invite", ""))
//	resCode, msg, user := DealUser(c) //处理用户信息
//	if resCode != e.SUCCESS {
//		JsonReturn(c, resCode, msg, nil)
//		return
//	}
//	if !util.CheckCode(code) {
//		JsonReturn(c, -1, "__T_INVITE_CODE_FORMAT", nil)
//		return
//	}
//	//查询用户邀请信息
//	err, inviteInfo := models.GetUserInviterByCode(code)
//
//	if err == nil && inviteInfo.ID > 0 {
//		JsonReturn(c, e.ERROR, "__T_INVITE_CODE_HAS", nil)
//		return
//	}
//	upInvite := models.UserInviter{}
//	upInvite.InviterCode = code
//	eRes := models.EditUserInviter(user.Id, upInvite)
//
//	links := strings.TrimSpace(models.GetConfigV("inviter_share_url")) + "?invite=" + code
//	resData := map[string]interface{}{}
//	resData["invite"] = code      //
//	resData["invite_url"] = links //
//	if eRes == nil {
//		JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resData)
//		return
//	} else {
//		JsonReturn(c, e.SUCCESS, "__T_FAIL", nil)
//		return
//	}
//}
