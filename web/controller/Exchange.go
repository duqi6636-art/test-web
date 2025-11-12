package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
	"math"
	"strings"
	"time"
)

// 兑换券
// @BasePath /api/v1
// @Summary 兑换券
// @Description 兑换券
// @Tags 个人中心 - 企业套餐
// @Accept  x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param code query string true "兑换码"
// @Produce  json
// @Success 0 {object} map[string]interface{} "user_balance：用户兑换后ip余额；ip_num：用户兑换后ip余额；balance：用户兑换后流量余额；flow：兑换的流量"
// @Router /web/ex/get_exchange [post]
func ExchangeCdk(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	code := strings.TrimSpace(c.DefaultPostForm("code", ""))
	if code == "" {
		JsonReturn(c, -1, "__T_EX_EMPTY", nil)
		return
	}

	nowTime := util.GetNowInt()
	balance := int64(0) //余额
	//flows := int64(0) //余额

	var err error
	exInfo := models.ExchangeList{}
	err, exInfo = models.GetExchangeInfo(code)
	if err != nil && exInfo.Id == 0 {
		JsonReturn(c, -1, "__T_EX_CODE_ERROR", nil)
		return
	}
	if exInfo.Status != 1 || exInfo.BindUid > 0 {
		JsonReturn(c, -1, "__T_EX_USED", nil)
		return
	}

	if exInfo.Expire > 0 && exInfo.Expire < nowTime {
		JsonReturn(c, -1, "__T_EX_EXPIRED", nil)
		return
	}
	if exInfo.UseCycle >= 1 {
		_, couponUse := models.GetExchangeByUsePlatform(uid, exInfo.UseCycle, exInfo.Platform)
		if len(couponUse) >= exInfo.UseNumber {
			JsonReturn(c, -1, "__T_COUPON_USED_LIMIT", nil)
			return
		}
	}
	// 同一分组下只可兑换一次
	if exInfo.GroupId > 0 {
		cdkInfo, _ := models.GetExchangeByGroupId(uid, exInfo.GroupId)
		if cdkInfo.Id != 0 && exInfo.Id != cdkInfo.Id {
			JsonReturn(c, -1, "__T_COUPON_USED_LIMIT", nil)
			return
		}
	}

	// 用户类型和券类型不匹配
	is_pay := "no_pay"
	if user.IsPay == "true" {
		is_pay = "payed"
	}
	if exInfo.UserType != "all" && is_pay != exInfo.UserType {
		JsonReturn(c, -1, "__T_COUPON_COMPARE", nil)
		return
	}
	flows := "0"
	userFlows := int64(0)
	userFlowInfo := models.GetUserFlowInfo(user.Id) //用户流量信息

	dynamicIspFlows := "0"
	userDynamic := int64(0)
	userDynamicInfo := models.GetUserDynamicIspInfo(user.Id) //用户轮转流量信息

	day := 0

	if exInfo.Cate == 3 {
		userFlows = userFlowInfo.Flows + exInfo.Value
		flows, _ = DealFlowChar(userFlows, "GB")
	} else if exInfo.Cate == 4 {
		userDynamic = userDynamicInfo.Flows + exInfo.Value
		dynamicIspFlows, _ = DealFlowChar(userDynamic, "GB")
	} else if exInfo.Cate == 5 {
		//flowDay := models.GetUserFlowDayByUid(uid) //不限量流量信息
		//expTime := flowDay.ExpireTime + int(exInfo.Value)
		//dayInfo := 0.00
		//if flowDay.Id > 0 {
		//	if flowDay.ExpireTime > nowTime {
		//		expTime = flowDay.ExpireTime + int(exInfo.Value)
		//		duration := expTime - nowTime
		//		dayInfo = math.Ceil(float64(duration) / 86400)
		//	} else {
		//		dayInfo = 1
		//	}
		//} else {
		//	dayInfo = 1
		//}
		//day = int(dayInfo)
	} else {
		balance = exInfo.Value
	}

	cate := "isp"
	if exInfo.Cate == 3 {
		cate = "flow"
	} else if exInfo.Cate == 4 {
		cate = "dynamic_isp"
	} else if exInfo.Cate == 5 {
		cate = "unlimited"
	} else if exInfo.Cate == 6 {
		cate = "static"
	}

	// 异步处理生成 cdk   ----start
	info := models.PushCdkey{}
	info.Mode = exInfo.Mode
	info.Cate = cate
	info.Cdkey = code
	info.Number = exInfo.Value
	info.Country = exInfo.Region
	info.CdkType = exInfo.Name
	info.Uid = exInfo.Uid
	info.BindUsername = user.Username
	info.BindUid = user.Id
	info.BindEmail = user.Email
	info.BindTime = nowTime
	info.Ip = c.ClientIP()
	info.CreateTime = nowTime
	info.ExInfo = exInfo
	listStr, _ := json.Marshal(info)
	resP := models.RedisLPUSH("list_cdk_exchange", string(listStr))
	fmt.Println(resP)
	// 异步处理生成 cdk   ----end

	//if exInfo.UseType == 2 {
	//	info := exInfo
	//	info.Id = 0
	//	info.Expire = exInfo.Expire
	//	info.BindUid = userInfo.Id
	//	info.BindUsername = userInfo.Username
	//	info.CreateTime = nowTime
	//	err = models.AddExchange(info)
	//	if err == nil {
	//		res = true
	//	}
	//} else {
	//	exParam := map[string]interface{}{}
	//	exParam["status"] = 2
	//	exParam["use_time"] = nowTime
	//	exParam["bind_uid"] = uid
	//	exParam["bind_username"] = username
	//	res = models.EditExchangeByCode(code, exParam)
	//}
	data := map[string]interface{}{
		"user_balance": user.Balance + int(balance), //旧版使用的 ，后期 新版过渡完可以去掉
		"ip_num":       user.Balance + int(balance),
		"balance":      userFlows,
		"flow":         flows,
		"dynamic_isp":  dynamicIspFlows,
		"flow_day":     day,
	}
	if models.GetConfigVal("dns_domain") != "" {
		b := CreateUserDomain(user)
		if !b {
			fmt.Println("域名生成失败")
		}
	}
	JsonReturn(c, 0, "__T_EX_SUCCESS", data)
	return

}

// 生成兑换券
// @BasePath /api/v1
// @Summary 生成兑换券
// @Description 生成兑换券
// @Tags 个人中心 - 企业套餐
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param number formData string true "兑换数量"
// @Produce json
// @Success 0 {object} map[string]interface{} "cdk：兑换码；balance：用户余额"
// @Router /web/ex/generate [post]
func Generate(c *gin.Context) {
	number := strings.TrimSpace(c.DefaultPostForm("number", "")) // 兑换ip数量
	resCode, msg, user := DealUser(c)                            //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	numbers := util.StoI(number)
	if numbers <= 0 {
		JsonReturn(c, e.ERROR, "__T_AMOUNT_ERROR", nil)
		return
	}
	// 查询用户代理商余额记录
	agentInfo := models.GetAgentBalanceByUid(uid)
	if agentInfo.Id == 0 {
		JsonReturn(c, e.ERROR, "__PC_NO_IP_INFO", nil)
		return
	}

	if agentInfo.Status == 2 {
		JsonReturn(c, e.ERROR, "__T_PACKAGE_FORBIDDEN", nil)
		return
	}
	minNumberStr := models.GetConfigV("generate_min_money") //生成 最小数量
	minNumber := 0
	if minNumberStr == "" {
		minNumber = 50
		minNumberStr = "50"
	} else {
		minNumber = util.StoI(minNumberStr)
	}
	if numbers < minNumber {
		JsonReturn(c, e.ERROR, "__T_AMOUNT_MIN-- "+" "+minNumberStr, nil)
		return
	}
	balStr := models.GetConfigV("user_balance_min") //用户余额最少剩余多少才可以生成
	bal := 0
	if balStr == "" {
		bal = 200
	} else {
		bal = util.StoI(balStr)
	}

	if agentInfo.Balance < bal || agentInfo.Balance < numbers {
		JsonReturn(c, e.ERROR, "__PC_NO_IP_INFO", nil)
		return
	}

	str := util.ItoS(user.Id) + "_" + util.GetNowTimeStr() + "_" + util.RandStr("r", 8)
	cdkStr := util.Md5(str)
	code := strings.ToUpper(cdkStr[0:6] + cdkStr[20:])

	nowTime := util.GetNowInt()
	// 异步处理生成 cdk   ----start
	info := models.PushCdkey{}
	info.Mode = "agent"
	info.Cate = "isp"
	info.Cdkey = code
	info.Number = int64(numbers)
	info.CdkType = "cdk"
	info.Uid = user.Id
	info.BindUsername = user.Username
	info.BindEmail = user.Email
	info.Ip = c.ClientIP()
	info.CreateTime = nowTime
	listStr, _ := json.Marshal(info)
	resP := models.RedisLPUSH("list_cdk_info", string(listStr))
	fmt.Println(resP)
	// 异步处理生成 cdk   ----end

	balance := agentInfo.Balance - numbers
	resInfo := map[string]interface{}{}
	resInfo["cdk"] = code
	resInfo["balance"] = balance
	JsonReturn(c, e.SUCCESS, "__T_EX_GENERATE_OK", resInfo)
	return

	//nowTime := util.GetNowInt()
	////写入兑换券记录
	//cdkStr := GetUuid()
	//cdkArr := strings.Split(cdkStr, "-")
	//lens := len(cdkArr)
	//code := strings.ToUpper(GenValidateCode(6) + cdkArr[lens-1])
	//info := models.ExchangeList{}
	//info.Cid = 0
	//info.Cate = 2
	//info.Uid = uid
	//info.Code = code
	//info.Name = "ISP"
	//info.BindUid = 0
	//info.BindUsername = ""
	//info.Status = 1
	//info.UseTime = 0
	//info.Title = "Exchange" + util.ItoS(numbers) + " IPs"
	//info.Value = numbers
	//info.UserType = "all"
	//info.UseType = 1
	//info.Expire = 0
	//info.ExpiryDay = 0
	//info.UseCycle = 0
	//info.UseNumber = 0
	//info.Platform = 0
	//info.GroupId = 0
	//info.CreateTime = nowTime
	//err := models.AddExchange(info)

	//data := map[string]interface{}{}
	//data["code"] = code
	//if err == nil {
	//	// 写入日志
	//	models.AddAgentExchange(user, agentInfo, numbers)
	//	// 更新用户可用余额
	//	agentBalance := make(map[string]interface{})
	//	agentBalance["balance"] = agentInfo.Balance - numbers
	//	editError := models.EditAgentBalanceBy(map[string]interface{}{"id": agentInfo.Id}, agentBalance)
	//	fmt.Println("update_user_agent_balance", editError)
	//	JsonReturn(c, e.SUCCESS, "__T_EX_GENERATE_OK", data)
	//	return
	//}
	//JsonReturn(c, e.ERROR, "error", nil)
	//return
}

// 批量生成兑换券
// @BasePath /api/v1
// @Summary 批量生成兑换券
// @Description 批量生成兑换券
// @Tags 个人中心 - 企业套餐
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param number formData string true "兑换数量"
// @Produce json
// @Success 0 {object} map[string]interface{} "cdk：兑换码；balance：用户余额"
// @Router /web/ex/generate_batch [post]
func BatchGenerate(c *gin.Context) {
	number := strings.TrimSpace(c.DefaultPostForm("number", "")) // 兑换ip数量
	quantity_str := strings.TrimSpace(c.DefaultPostForm("quantity", ""))
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	numbers := util.StoI(number)
	quantity := util.StoI(quantity_str)
	if numbers <= 0 {
		JsonReturn(c, e.ERROR, "__T_AMOUNT_ERROR", nil)
		return
	}
	// 查询用户代理商余额记录
	agentInfo := models.GetAgentBalanceByUid(uid)
	if agentInfo.Id == 0 {
		JsonReturn(c, e.ERROR, "__PC_NO_IP_INFO", nil)
		return
	}

	if agentInfo.Status == 2 {
		JsonReturn(c, e.ERROR, "__T_PACKAGE_FORBIDDEN", nil)
		return
	}
	minNumberStr := models.GetConfigV("generate_min_money") //生成 最小数量
	minNumber := 0
	if minNumberStr == "" {
		minNumber = 50
		minNumberStr = "50"
	} else {
		minNumber = util.StoI(minNumberStr)
	}
	if numbers < minNumber {
		JsonReturn(c, e.ERROR, "__T_AMOUNT_MIN-- "+" "+minNumberStr, nil)
		return
	}
	balStr := models.GetConfigV("user_balance_min") //用户余额最少剩余多少才可以生成
	bal := 0
	if balStr == "" {
		bal = 200
	} else {
		bal = util.StoI(balStr)
	}

	if agentInfo.Balance < bal || agentInfo.Balance < (numbers*quantity) {
		JsonReturn(c, e.ERROR, "__PC_NO_IP_INFO", nil)
		return
	}

	var codes []string
	for i := 0; i < quantity; i++ {

		str := util.ItoS(user.Id) + "_" + util.GetNowTimeStr() + "_" + util.RandStr("r", 8)
		cdkStr := util.Md5(str)
		code := strings.ToUpper(cdkStr[0:6] + cdkStr[20:])

		nowTime := util.GetNowInt()
		// 异步处理生成 cdk   ----start
		info := models.PushCdkey{}
		info.Mode = "agent"
		info.Cate = "isp"
		info.Cdkey = code
		info.Number = int64(numbers)
		info.CdkType = "cdk"
		info.Uid = user.Id
		info.BindUsername = user.Username
		info.BindEmail = user.Email
		info.Ip = c.ClientIP()
		info.CreateTime = nowTime
		listStr, _ := json.Marshal(info)
		resP := models.RedisLPUSH("list_cdk_info", string(listStr))
		fmt.Println(resP)
		// 异步处理生成 cdk   ----end
		codes = append(codes, code)
	}

	balance := agentInfo.Balance - (numbers * quantity)
	resInfo := map[string]interface{}{}
	resInfo["cdk"] = codes
	resInfo["balance"] = balance
	JsonReturn(c, e.SUCCESS, "__T_EX_GENERATE_OK", resInfo)
	return

}

// 直接提现到本账户ip余额
// @BasePath /api/v1
// @Summary 直接提现到本账户ip余额
// @Description 直接提现到本账户ip余额
// @Tags 个人中心 - 企业套餐
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param number formData string true "兑换数量"
// @Produce json
// @Success 0 {object} map[string]interface{} "cdk：兑换码；balance：用户余额"
// @Router /web/ex/direct_conversion [post]
func DirectConversion(c *gin.Context) {
	number := strings.TrimSpace(c.DefaultPostForm("number", "")) // 兑换ip数量
	resCode, msg, user := DealUser(c)                            //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	numbers := util.StoI(number)
	if numbers <= 0 {
		JsonReturn(c, e.ERROR, "__T_AMOUNT_ERROR", nil)
		return
	}
	// 查询用户代理商余额记录
	agentInfo := models.GetAgentBalanceByUid(user.Id)
	if agentInfo.Id == 0 && agentInfo.Balance > 0 {
		JsonReturn(c, e.ERROR, "__PC_NO_IP_INFO", nil)
		return
	}

	if agentInfo.Balance < numbers {
		JsonReturn(c, e.ERROR, "__PC_NO_IP_INFO", nil)
		return
	}
	if agentInfo.Status == 2 {
		JsonReturn(c, e.ERROR, "__T_PACKAGE_FORBIDDEN", nil)
		return
	}
	nowTime := util.GetNowInt()
	// 异步处理生成 cdk   ----start
	info := models.PushCdkey{}
	info.Mode = "agent"
	info.Cate = "isp"
	info.Cdkey = ""
	info.Number = int64(numbers)
	info.CdkType = "self"
	info.Uid = user.Id
	info.BindUid = user.Id
	info.BindUsername = user.Username
	info.BindEmail = user.Email
	info.Ip = c.ClientIP()
	info.CreateTime = nowTime
	listStr, _ := json.Marshal(info)
	resP := models.RedisLPUSH("list_cdk_info", string(listStr))
	fmt.Println(resP)
	// 异步处理生成 cdk   ----end

	balance := agentInfo.Balance - numbers
	resInfo := map[string]interface{}{}
	resInfo["cdk"] = ""
	resInfo["balance"] = balance
	if models.GetConfigVal("dns_domain") != "" {
		b := CreateUserDomain(user)
		if !b {
			fmt.Println("域名生成失败")
		}
	}
	JsonReturn(c, e.SUCCESS, "__T_EX_SUCCESS", resInfo)
	return
}

// 兑换券 列表
// @BasePath /api/v1
// @Summary 兑换券列表
// @Description 兑换券列表
// @Tags 个人中心 - 企业套餐
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param limit formData string true "每页显示数量"
// @Param page formData string true "当前页码"
// @Produce json
// @Success 0 {object} map[string]interface{} "total：总数量；total_page：总页数；lists：兑换券列表（值为[]models.ResExchange{}对象）"
// @Router /web/ex/generate_list [post]
func ExchangeList(c *gin.Context) {
	limitStr := c.DefaultPostForm("limit", "10")
	pageStr := c.DefaultPostForm("page", "1")

	if limitStr == "" {
		limitStr = "10"
	}
	if pageStr == "" {
		pageStr = "1"
	}

	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	limit := util.StoI(limitStr)
	page := util.StoI(pageStr)
	offset := (page - 1) * limit

	exLists, _ := models.GetExchangeByCate(uid, 2, offset, limit)
	lists := []models.ResExchange{}
	for _, v := range exLists {
		info := models.ResExchange{}
		info.Code = v.Code
		info.Value = v.Value
		info.BindUsername = v.BindUsername
		info.BindUid = v.BindUid
		info.Status = v.Status
		info.CreateTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i:s")
		lists = append(lists, info)
	}
	totalList, _ := models.GetExchangeListByCate(uid, 2)
	totalPage := int(math.Ceil(float64(len(totalList)) / float64(limit)))
	result := map[string]interface{}{
		"total":      len(totalList),
		"total_page": totalPage,
		"lists":      lists,
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
	return
}

// 禁用 代理商cdk券
// @BasePath /api/v1
// @Summary 禁用cdk
// @Description 禁用cdk
// @Tags 个人中心 - 企业套餐
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param code formData string true "兑换码"
// @Produce json
// @Success 0 {object} map[string]interface{} "status：状态（3：禁用）,use_time：使用时间"
// @Router /web/ex/forbid [post]
func ForbidEx(c *gin.Context) {
	code := strings.TrimSpace(c.DefaultPostForm("code", ""))
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	if code == "" {
		JsonReturn(c, -1, "__T_EX_EMPTY", nil)
		return
	}
	uid := user.Id
	time.Sleep(1)

	err, exInfo := models.GetExchangeInfo(code)
	if err != nil && exInfo.Id == 0 {
		JsonReturn(c, -1, "__T_EX_CODE_ERROR", nil)
		return
	}
	if exInfo.Cate == 1 {
		JsonReturn(c, -1, "__T_EX_CODE_ERROR", nil)
		return
	}
	if exInfo.Status == 3 {
		JsonReturn(c, -1, "__T_EX_FORBID", nil)
		return
	}
	if exInfo.Status == 2 || exInfo.BindUid > 0 {
		JsonReturn(c, -1, "__T_EX_USED", nil)
		return
	}

	exValue := int(exInfo.Value)

	cate := ""
	if exInfo.Cate == 2 {
		// 查询用户代理商余额记录
		agentBalance := 0
		agentInfo := models.GetAgentBalanceByUid(uid)
		if agentInfo.Id == 0 {
			JsonReturn(c, e.ERROR, "__PC_NO_IP_INFO", nil)
			return
		}

		total := agentInfo.Total
		if agentInfo.Balance < 0 {
			agentInfo.Balance = 0
		}
		agentBalance = agentInfo.Balance
		newBalance := exValue + agentBalance
		if newBalance > total {
			JsonReturn(c, e.ERROR, "error", nil)
			return
		}
		cate = "isp"
	} else if exInfo.Cate == 3 {
		cate = "flow"
	} else {
		cate = "other"
	}
	nowTime := util.GetNowInt()
	// 异步处理生成 cdk   ----start
	info := models.PushCdkey{}
	info.Cate = cate
	info.Cdkey = code
	info.Number = int64(exValue)
	info.CdkType = "forbid"
	info.Uid = user.Id
	info.CreateTime = nowTime
	listStr, _ := json.Marshal(info)
	resP := models.RedisLPUSH("list_cdk_info", string(listStr))
	fmt.Println(resP)
	// 异步处理生成 cdk   ----end

	data := map[string]interface{}{}
	data["status"] = 3
	data["use_time"] = nowTime
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", data)
	return

	//nowTime := util.GetNowInt()
	//data := map[string]interface{}{}
	//data["status"] = 3
	//data["use_time"] = nowTime
	//res := models.EditExchangeByCode(code, data)
	//if res == true {
	//	where := map[string]interface{}{"id": agentInfo.Id}
	//	info := map[string]interface{}{"balance": exValue + agentBalance}
	//	err := models.EditAgentBalanceBy(where, info)
	//	fmt.Println("update_user_agent_balance", err)
	//	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", data)
	//	return
	//}
	//JsonReturn(c, e.ERROR, "error", nil)
	//return
}

// 流量生成 cdk
// @BasePath /api/v1
// @Summary 流量生成 cdk
// @Description 流量生成 cdk
// @Tags 个人中心 - 企业套餐
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param flow formData string true "流量数量"
// @Produce json
// @Success 0 {object} map[string]interface{} "flows：流量（单位：GB）；cdk：兑换码；balance：用户余额;flow：剩余流量（单位：GB）;flow_redeem：可兑换流量（单位：GB）"
// @Router /web/ex/flow_cdk [post]
func GenerateFlowCdk(c *gin.Context) {
	flow_str := strings.TrimSpace(c.DefaultPostForm("flow", ""))
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	flow := util.StoI(flow_str)
	if flow <= 0 {
		JsonReturn(c, e.ERROR, "__T_FLOW_NUMBER_ERROR", nil)
		return
	}
	nowTime := util.GetNowInt()

	minNumberStr := models.GetConfigV("cdk_flow_min") //
	minNumber := 0
	if minNumberStr == "" {
		minNumber = 1
		minNumberStr = "1"
	} else {
		minNumber = util.StoI(minNumberStr)
	}
	if flow < minNumber {
		JsonReturn(c, e.ERROR, "__T_INVITE_FLOW_NUMBER_MIN", nil)
		return
	}
	userFlowInfo := models.GetUserFlowInfo(user.Id)
	if userFlowInfo.ID == 0 {
		JsonReturn(c, -1, "__T_NO_FLOW_INFO", gin.H{})
		return
	}
	if userFlowInfo.ExpireTime < nowTime {
		JsonReturn(c, -1, "__T_FLOW_EXPIRED", gin.H{})
		return
	}
	if userFlowInfo.Status == 2 {
		JsonReturn(c, e.ERROR, "__T_PACKAGE_FORBIDDEN", nil)
		return
	}
	redeemFlow := userFlowInfo.BuyFlow
	userFlows := userFlowInfo.Flows
	if redeemFlow > 0 {
		redeemFlow = userFlowInfo.BuyFlow - userFlowInfo.CdkFlow
		if redeemFlow > userFlows { // 如果买的剩余的大于 总剩余的 就展示总剩余的
			redeemFlow = userFlows
		}
	}

	exFlow := int64(flow) * 1024 * 1024 * 1024
	if redeemFlow <= 0 || redeemFlow < exFlow {
		JsonReturn(c, e.ERROR, "__T_FLOW_NO_ENOUGH", nil)
		return
	}

	str := util.ItoS(user.Id) + "_" + util.GetNowTimeStr() + "_" + util.RandStr("r", 8)
	cdkStr := util.Md5(str)
	code := strings.ToUpper(cdkStr[0:6] + cdkStr[20:])

	// 异步处理生成 cdk   ----start
	info := models.PushCdkey{}
	info.Mode = "agent"
	info.Cate = "flow"
	info.Cdkey = code
	info.Number = exFlow
	info.CdkType = "cdk"
	info.Uid = user.Id
	info.Ip = c.ClientIP()
	info.CreateTime = nowTime
	listStr, _ := json.Marshal(info)
	resP := models.RedisLPUSH("list_cdk_info", string(listStr))
	fmt.Println(resP)
	// 异步处理生成 cdk   ----end

	flows := redeemFlow - exFlow
	flowStr, _ := DealFlowChar(flows, "GB")
	userFlowsNew := userFlows - exFlow
	userFlowStr, _ := DealFlowChar(userFlowsNew, "GB")

	resInfo := map[string]interface{}{}
	//resInfo["flows"] = flowStr
	resInfo["flows"] = flowStr //旧版 个人中心
	resInfo["cdk"] = code
	// 新版个人中心展示
	resInfo["balance"] = userFlowsNew
	resInfo["flow"] = userFlowStr
	resInfo["flow_redeem"] = flowStr
	if models.GetConfigVal("dns_domain") != "" {
		b := CreateUserDomain(user)
		if !b {
			fmt.Println("域名生成失败")
		}
	}
	JsonReturn(c, e.SUCCESS, "__T_EX_GENERATE_OK", resInfo)
	return
}

// 批量流量生成 cdk
// @BasePath /api/v1
// @Summary 批量流量生成 cdk
// @Description 批量流量生成 cdk
// @Tags 个人中心 - 企业套餐
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param flow formData string true "流量数量"
// @Param quantity formData string true "cdk数"
// @Produce json
// @Success 0 {object} map[string]interface{} "flows：流量（单位：GB）；cdk：兑换码；balance：用户余额;flow：剩余流量（单位：GB）;flow_redeem：可兑换流量（单位：GB）"
// @Router /web/ex/flow_cdk_batch [post]
func BatchGenerateFlowCdk(c *gin.Context) {
	flow_str := strings.TrimSpace(c.DefaultPostForm("flow", ""))
	quantity_str := strings.TrimSpace(c.DefaultPostForm("quantity", ""))
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	flow := util.StoI(flow_str)
	quantity := util.StoI(quantity_str)
	if flow <= 0 {
		JsonReturn(c, e.ERROR, "__T_FLOW_NUMBER_ERROR", nil)
		return
	}
	nowTime := util.GetNowInt()

	minNumberStr := models.GetConfigV("cdk_flow_min") //
	minNumber := 0
	if minNumberStr == "" {
		minNumber = 1
		minNumberStr = "1"
	} else {
		minNumber = util.StoI(minNumberStr)
	}
	if flow < minNumber {
		JsonReturn(c, e.ERROR, "__T_INVITE_FLOW_NUMBER_MIN", nil)
		return
	}
	userFlowInfo := models.GetUserFlowInfo(user.Id)
	if userFlowInfo.ID == 0 {
		JsonReturn(c, -1, "__T_NO_FLOW_INFO", gin.H{})
		return
	}
	if userFlowInfo.ExpireTime < nowTime {
		JsonReturn(c, -1, "__T_FLOW_EXPIRED", gin.H{})
		return
	}
	if userFlowInfo.Status == 2 {
		JsonReturn(c, e.ERROR, "__T_PACKAGE_FORBIDDEN", nil)
		return
	}
	redeemFlow := userFlowInfo.BuyFlow
	userFlows := userFlowInfo.Flows
	if redeemFlow > 0 {
		redeemFlow = userFlowInfo.BuyFlow - userFlowInfo.CdkFlow
		if redeemFlow > userFlows { // 如果买的剩余的大于 总剩余的 就展示总剩余的
			redeemFlow = userFlows
		}
	}

	exFlow := int64(flow) * 1024 * 1024 * 1024
	if redeemFlow <= 0 || redeemFlow < (exFlow*int64(quantity)) {
		JsonReturn(c, e.ERROR, "__T_FLOW_NO_ENOUGH", nil)
		return
	}

	var codes []string
	for i := 0; i < quantity; i++ {

		str := util.ItoS(user.Id) + "_" + util.GetNowTimeStr() + "_" + util.RandStr("r", 8)
		cdkStr := util.Md5(str)
		code := strings.ToUpper(cdkStr[0:6] + cdkStr[20:])

		// 异步处理生成 cdk   ----start
		info := models.PushCdkey{}
		info.Mode = "agent"
		info.Cate = "flow"
		info.Cdkey = code
		info.Number = exFlow
		info.CdkType = "cdk"
		info.Uid = user.Id
		info.Ip = c.ClientIP()
		info.CreateTime = nowTime
		listStr, _ := json.Marshal(info)
		resP := models.RedisLPUSH("list_cdk_info", string(listStr))
		fmt.Println(resP)
		// 异步处理生成 cdk   ----end
		codes = append(codes, code)
	}

	flows := redeemFlow - (exFlow * int64(quantity))
	flowStr, _ := DealFlowChar(flows, "GB")
	userFlowsNew := userFlows - (exFlow * int64(quantity))
	userFlowStr, _ := DealFlowChar(userFlowsNew, "GB")

	resInfo := map[string]interface{}{}
	//resInfo["flows"] = flowStr
	resInfo["flows"] = flowStr //旧版 个人中心
	resInfo["cdk"] = codes
	// 新版个人中心展示
	resInfo["balance"] = userFlowsNew
	resInfo["flow"] = userFlowStr
	resInfo["flow_redeem"] = flowStr
	if models.GetConfigVal("dns_domain") != "" {
		b := CreateUserDomain(user)
		if !b {
			fmt.Println("域名生成失败")
		}
	}
	JsonReturn(c, e.SUCCESS, "__T_EX_GENERATE_OK", resInfo)
	return
}

// 给他人充值
// @BasePath /api/v1
// @Summary 给他人充值
// @Description 给他人充值
// @Tags 个人中心 - 企业套餐
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param flow formData string true "流量数量"
// @Param username formData string true "用户名/邮箱"
// @Produce json
// @Success 0 {object} map[string]interface{} "flows：流量（单位：GB）；cdk：兑换码；balance：用户余额;flow：剩余流量（单位：GB）;flow_redeem：可兑换流量（单位：GB）"
// @Router /web/ex/flow_recharge [post]
func FlowRechargeUser(c *gin.Context) {
	flow_str := strings.TrimSpace(c.DefaultPostForm("flow", ""))
	username := strings.TrimSpace(c.DefaultPostForm("username", ""))
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	nowTime := util.GetNowInt()

	flow := util.StoI(flow_str)
	if flow <= 0 {
		JsonReturn(c, e.ERROR, "__T_FLOW_NUMBER_ERROR", nil)
		return
	}

	var toUser = models.Users{}
	var err error
	if find := strings.Contains(username, "@"); find {
		err, toUser = models.GetUserByEmail(username)
	} else {
		err, toUser = models.GetUserByUsername(username)
	}
	if err != nil || toUser.Id == 0 {
		JsonReturn(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
		return
	}

	if user.Id == toUser.Id { //不能给自己充值
		JsonReturn(c, e.ERROR, "__T_NO_SELF_RECHARGE", nil)
		return
	}

	minNumberStr := models.GetConfigV("cdk_flow_min") //
	minNumber := 0
	if minNumberStr == "" {
		minNumber = 1
		minNumberStr = "1"
	} else {
		minNumber = util.StoI(minNumberStr)
	}
	if flow < minNumber {
		JsonReturn(c, e.ERROR, "__T_INVITE_FLOW_NUMBER_MIN", nil)
		return
	}
	userFlowInfo := models.GetUserFlowInfo(user.Id)
	if userFlowInfo.ID == 0 {
		JsonReturn(c, -1, "__T_NO_FLOW_INFO", gin.H{})
		return
	}
	if userFlowInfo.ExpireTime < nowTime {
		JsonReturn(c, -1, "__T_FLOW_EXPIRED", gin.H{})
		return
	}
	if userFlowInfo.Status == 2 {
		JsonReturn(c, e.ERROR, "__T_PACKAGE_FORBIDDEN", nil)
		return
	}
	redeemFlow := userFlowInfo.BuyFlow
	userFlows := userFlowInfo.Flows
	if redeemFlow > 0 {
		redeemFlow = userFlowInfo.BuyFlow - userFlowInfo.CdkFlow
		if redeemFlow > userFlows { // 如果买的剩余的大于 总剩余的 就展示总剩余的
			redeemFlow = userFlows
		}
	}

	exFlow := int64(flow) * 1024 * 1024 * 1024
	if redeemFlow <= 0 || redeemFlow < exFlow {
		JsonReturn(c, e.ERROR, "__T_FLOW_NO_ENOUGH", nil)
		return
	}

	str := util.ItoS(user.Id) + "_" + util.GetNowTimeStr()
	cdkStr := util.Md5(str)
	code := strings.ToUpper(cdkStr[0:6] + cdkStr[20:])

	// 异步处理生成 cdk   ----start
	info := models.PushCdkey{}
	info.Mode = "agent"
	info.Cate = "flow"
	info.Cdkey = code
	info.Number = exFlow
	info.CdkType = "to_user"
	info.Uid = user.Id
	info.BindEmail = toUser.Email
	info.BindUsername = toUser.Username
	info.BindUid = toUser.Id
	info.BindTime = nowTime
	info.Ip = c.ClientIP()
	info.CreateTime = nowTime
	listStr, _ := json.Marshal(info)
	resP := models.RedisLPUSH("list_cdk_info", string(listStr))
	fmt.Println(resP)
	// 异步处理生成 cdk   ----end

	flows := redeemFlow - exFlow
	flowStr, _ := DealFlowChar(flows, "GB")
	userFlowsNew := userFlows - exFlow
	userFlowStr, _ := DealFlowChar(userFlowsNew, "GB")

	resInfo := map[string]interface{}{}
	//resInfo["flows"] = flowStr
	resInfo["flows"] = flowStr //旧版 个人中心
	resInfo["cdk"] = code
	// 新版个人中心展示
	resInfo["balance"] = userFlowsNew
	resInfo["flow"] = userFlowStr
	resInfo["flow_redeem"] = flowStr
	if models.GetConfigVal("dns_domain") != "" {
		b := CreateUserDomain(user)
		if !b {
			fmt.Println("域名生成失败")
		}
	}
	JsonReturn(c, e.SUCCESS, "__T_EX_RECHARGE_OK", resInfo)
	return
}

// 流量cdk/给他人账户充值 记录列表
// @BasePath /api/v1
// @Summary 流量cdk/给他人账户充值 记录列表
// @Description 流量cdk/给他人账户充值 记录列表
// @Tags 个人中心 - 企业套餐
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} map[string]interface{} "cdk_lists：流量cdk列表（值为[]models.ResExchange{}对象）；to_lists：给他人充值列表（值为[]models.ResExchange{}对象）"
// @Router /web/ex/flow_record [post]
func ExchangeFlowList(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	exLists, _ := models.GetExchangeListByCate(uid, 3)
	cdkLists := []models.ResExchange{}
	toLists := []models.ResExchange{}
	for _, v := range exLists {
		info := models.ResExchange{}
		info.Code = v.Code
		info.Value = v.Value / 1024 / 1024 / 1024
		info.BindUsername = v.BindUsername
		info.BindUid = v.BindUid
		info.Status = v.Status
		info.CreateTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i:s")

		if v.Name == "cdk" {
			cdkLists = append(cdkLists, info)
		}
		if v.Name == "to_user" {
			toLists = append(toLists, info)
		}
	}

	result := map[string]interface{}{
		"cdk_lists": cdkLists,
		"to_lists":  toLists,
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
	return
}

// cdk生成列表
// @BasePath /api/v1
// @Summary cdk生成列表
// @Description cdk生成列表
// @Tags 个人中心 - 企业套餐
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} map[string]interface{} "cdk_lists：流量cdk列表（值为[]models.ResExchange{}对象）；to_lists：给他人充值列表（值为[]models.ResExchange{}对象）"
// @Router /center/cdk/ex/generate_list [post]
func NewGenerateList(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	mode := strings.TrimSpace(c.DefaultPostForm("mode", "agent"))   // agent 代理商  balance-余额充值
	cateStr := strings.TrimSpace(c.DefaultPostForm("cate", "flow")) //isp  flow dynamic_isp unlimited
	timeType := c.DefaultPostForm("time_type", "0")                 // 0-生成时间 1-兑换时间
	startDate := c.DefaultPostForm("start_date", "")
	endDate := c.DefaultPostForm("end_date", "")
	cdkEmailStr := c.DefaultPostForm("cdk_email", "")
	statusStr := c.DefaultPostForm("status", "") //  0-all 1-未兑换 2-已兑换
	start := 0
	end := 0
	if startDate != "" {
		start = util.StoI(util.GetTimeStamp(startDate, "Y-m-d"))
	}
	if endDate != "" {
		end = util.StoI(util.GetTimeStamp(endDate, "Y-m-d"))
	}
	if mode == "" {
		mode = "agent"
	}
	cate := 2
	if cateStr == "flow" {
		cate = 3
	} else if cateStr == "isp" {
		cate = 2
	} else if cateStr == "dynamic_isp" { //动态流量套餐
		cate = 4
	} else if cateStr == "unlimited" {
		cate = 5
	} else if strings.Contains(cateStr, "static") {
		cate = 6
	}

	configList := models.GetBalanceConfigList("all", 0, 0)
	configMap := map[string]models.ConfBalanceConfigModel{}
	for _, v := range configList {
		configMap[v.Cate] = v
	}

	nowTime := util.GetNowInt()
	lists := []models.ResGenerateList{}
	cdkLists := models.GetGenerateListByCate(uid, start, end, cate, mode, timeType, cdkEmailStr, statusStr, cateStr)

	aes_key := util.Md5(AesKey)
	for _, v := range cdkLists {
		useTime := ""
		if v.UseTime > 0 {
			useTime = util.GetTimeStr(v.UseTime, "Y/m/d H:i:s")
		}

		balanceStr := ""
		if v.Cate == 2 {
			if v.Balance > 0 {
				balanceStr = fmt.Sprintf("%d %s", v.Balance, "IPs")
			} else {
				balanceStr = ""
			}
		} else if v.Cate == 3 || v.Cate == 4 {
			if v.Balance > 0 {
				balanceStr = fmt.Sprintf("%d %s", v.Balance/1024/1024/1024, "GB")
			} else {
				balanceStr = ""
			}
		} else if v.Cate == 5 {
			if v.Balance > int64(nowTime) {
				balance := v.Balance
				duration := balance - int64(nowTime)
				dayUnit := ""
				dayInfo := 0.00
				if duration > 86400 {
					dayInfo = math.Ceil(float64(duration) / 86400)
					dayUnit = "Days"
					if dayInfo == 1 {
						dayUnit = "Day"
					}
				} else {
					dayInfo = math.Ceil(float64(duration) / 3600)
					dayUnit = "Hours"
					if dayInfo == 1 {
						dayUnit = "Hour"
					}
				}
				balanceStr = fmt.Sprintf("%d %s", int(dayInfo), dayUnit)
			} else {
				balanceStr = ""
			}
		} else if v.Cate == 6 {
			if v.Balance > 0 {
				balanceStr = fmt.Sprintf("%d %s", v.Balance, "IP")
			} else {
				balanceStr = ""
			}
		}
		confInfo, ok := configMap[cateStr]
		valueUnit := int64(1)
		unit, name := "", ""
		if ok {
			unit = confInfo.Unit
			name = confInfo.Name
			valueUnit = confInfo.Value
		}

		value := v.Value
		if v.Value > 0 && valueUnit > 0 {
			value = v.Value / valueUnit
		}

		valueStr := fmt.Sprintf("%d %s", value, unit)
		usageMode := 0
		if v.Name == "to_user" {
			usageMode = 1
		}
		idStr, err := util.AesEnCode([]byte(util.ItoS(v.Id)), []byte(aes_key))
		if err != nil {
			idStr = util.ItoS(v.Id)
		}
		info := models.ResGenerateList{}
		info.Id = idStr
		info.CdkKey = v.Code
		info.ExchangeType = name
		info.Value = valueStr
		info.Status = v.Status
		info.UsageMode = usageMode
		info.GenerateTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i:s")
		info.RedemptionTime = useTime
		info.Balance = balanceStr
		info.Email = hideEmail(v.Email)
		info.Remark = v.GenerateRemark
		lists = append(lists, info)
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", lists)
	return
}

// / 获取兑换列表
func NewRedemptionList(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	mode := strings.TrimSpace(c.DefaultPostForm("mode", "agent"))   // agent 代理商  balance-余额充值
	cateStr := strings.TrimSpace(c.DefaultPostForm("cate", "flow")) //isp  flow dynamic_isp unlimited
	startDate := c.DefaultPostForm("start_date", "")
	endDate := c.DefaultPostForm("end_date", "")
	cdkEmailStr := c.DefaultPostForm("cdk", "")
	statusStr := c.DefaultPostForm("status", "") //  0-all 1-cdk兑换 2-储值
	start := 0
	end := 0
	if startDate != "" {
		start = util.StoI(util.GetTimeStamp(startDate, "Y-m-d"))
	}
	if endDate != "" {
		end = util.StoI(util.GetTimeStamp(endDate, "Y-m-d"))
	}
	if mode == "" {
		mode = "agent"
	}
	cate := 2
	if cateStr == "flow" {
		cate = 3
	} else if cateStr == "isp" {
		cate = 2
	} else if cateStr == "dynamic_isp" { //动态流量套餐
		cate = 4
	} else if cateStr == "unlimited" {
		cate = 5
	} else if strings.Contains(cateStr, "static") {
		cate = 6
	}
	configList := models.GetBalanceConfigList("all", 0, 0)
	configMap := map[string]models.ConfBalanceConfigModel{}
	for _, v := range configList {
		configMap[v.Cate] = v
	}

	cdkLists := models.GetRedemptionList(uid, start, end, cate, mode, cdkEmailStr, statusStr)
	lists := []models.ResRedemptionList{}
	for _, v := range cdkLists {
		useTime := ""
		if v.UseTime > 0 {
			useTime = util.GetTimeStr(v.UseTime, "Y/m/d H:i:s")
		}

		confInfo, ok := configMap[cateStr]
		valueUnit := int64(1)
		unit, name := "", ""
		if ok {
			unit = confInfo.Unit
			name = confInfo.Name
			valueUnit = confInfo.Value
		}

		value := v.Value
		if v.Value > 0 && valueUnit > 0 {
			value = v.Value / valueUnit
			//fmt.Println("--------------",value)
		}
		valueStr := fmt.Sprintf("%d %s", value, unit)

		usageMode := 0
		if v.Name == "to_user" {
			usageMode = 1
		}

		info := models.ResRedemptionList{}
		info.Id = v.Id
		info.ExchangeType = name
		info.RedemptionTime = useTime
		info.Value = valueStr
		info.CdkKey = v.Code
		info.UsageMode = usageMode
		info.Remark = v.RedemptionRemark
		lists = append(lists, info)
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", lists)

	return
}

// @BasePath /api/v1
// @Summary CDK	获取用户CDK统计使用列表
// @Schemes
// @Description CDK	获取用户CDK使用列表
// @Tags 个人中心-CDK
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param mode formData string false "模式  agent 代理商  balance 余额"
// @Param cate formData string false "类型  isp  flow"
// @Param username formData string false "用户名"
// @Param collect formData string false "是否收藏 1收藏  0全部"
// @Param start_date formData string false "开始日期"
// @Param end_date formData string false "结束日期"
// @Produce json
// @Success 0 {array} models.ResUserUsageList{}"
// @Router /web/cdk/stats_use [post]
func GetUserCdkStats(c *gin.Context) {
	mode := strings.TrimSpace(c.DefaultPostForm("mode", "agent")) // agent 代理商  balance-余额充值
	cate := strings.TrimSpace(c.DefaultPostForm("cate", "flow"))  //isp  flow dynamic_isp unlimited
	startDate := c.DefaultPostForm("start_date", "")
	endDate := c.DefaultPostForm("end_date", "")
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))
	collect := com.StrTo(c.DefaultPostForm("collect", "0")).MustInt()
	sortType := strings.TrimSpace(c.DefaultPostForm("sort_type", "")) // 0-余额 1-数量 2-次数 3-最后兑换时间
	sort := strings.TrimSpace(c.DefaultPostForm("sort", ""))          // 排序升降 0-升序 1-降序
	if mode == "" {
		mode = "agent"
	}
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	start := 0
	end := 0
	if startDate != "" {
		start = util.StoI(util.GetTimeStamp(startDate, "Y-m-d"))
	}
	if endDate != "" {
		end = util.StoI(util.GetTimeStamp(endDate, "Y-m-d"))
	}
	if cate == "" {
		cate = "flow"
	}

	nowTime := util.GetNowInt()
	cdkLists := models.GetCdkUseStatsJoin(uid, mode, cate, email, collect, start, end, sortType, sort)
	lists := []models.ResUserUsageList{}
	aes_key := util.Md5(AesKey)
	for _, v := range cdkLists {
		lastTimeStr := ""
		if v.LastTime > 0 {
			lastTimeStr = util.GetTimeStr(v.LastTime, "Y/m/d H:i:s")
		}
		balanceStr := ""
		valueStr := ""
		if cate == "flow" || cate == "dynamic_isp" {
			if v.Value > 0 {
				value := v.Value / 1024 / 1024 / 1024
				valueStr = fmt.Sprintf("%d %s", value, "GB")
			}
			if v.Flows > 0 {
				balanceStr = util.ItoS(v.Flows/1024/1024/1024) + " GB"
			}
		} else if cate == "unlimited" {
			if v.Value > 0 {
				days := v.Value / 86400
				valueStr = fmt.Sprintf("%d %s", days, "Day")
			}
			if v.Flows > nowTime {
				balance := v.Flows
				duration := balance - nowTime
				dayUnit := ""
				dayInfo := 0.00
				if duration > 86400 {
					dayInfo = math.Ceil(float64(duration) / 86400)
					dayUnit = "Days"
					if dayInfo == 1 {
						dayUnit = "Day"
					}
				} else {
					dayInfo = math.Ceil(float64(duration) / 3600)
					dayUnit = "Hours"
					if dayInfo == 1 {
						dayUnit = "Hour"
					}
				}
				balanceStr = fmt.Sprintf("%d %s", int(dayInfo), dayUnit)
			}
		} else {
			if v.Value > 0 {
				valueStr = fmt.Sprintf("%d %s", v.Value, "IPs")
			}
			if v.Balance > 0 {
				balanceStr = util.ItoS(v.Balance) + " IPs"
			}
		}
		idStr, err := util.AesEnCode([]byte(util.ItoS(v.Id)), []byte(aes_key))
		if err != nil {
			idStr = util.ItoS(v.Id)
		}
		info := models.ResUserUsageList{}
		info.Id = idStr
		info.Email = hideEmail(v.Email)
		info.Balance = balanceStr
		info.Value = valueStr
		info.Times = v.Number
		info.Collect = v.IsCollect
		info.Remark = v.Remark
		info.LastRedemptionTime = lastTimeStr

		lists = append(lists, info)
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", lists)
	return
}

// @BasePath /api/v1
// @Summary CDK	记录列表下载
// @Schemes
// @Description CDK	记录列表下载
// @Tags 个人中心-CDK
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param mode formData string false "模式  agent 代理商  balance 余额"
// @Param cate formData string false "类型  isp  flow"
// @Param username formData string false "用户名"
// @Param collect formData string false "是否收藏 1收藏  0全部"
// @Param start_date formData string false "开始日期"
// @Param end_date formData string false "结束日期"
// @Produce json
// @Success 0 {array} map[string]interface{} ""
// @Router /web/cdk/stats_download [post]
func GetUserCdkStatsDownload(c *gin.Context) {
	mode := strings.TrimSpace(c.DefaultPostForm("cate", "agent")) // agent 代理商  balance-余额充值
	cate := strings.TrimSpace(c.DefaultPostForm("cate", "flow"))  //isp  flow dynamic_isp unlimited
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))
	collect := com.StrTo(c.DefaultPostForm("collect", "0")).MustInt()
	start_date := c.DefaultPostForm("start_date", "")
	end_date := c.DefaultPostForm("end_date", "")
	sortType := strings.TrimSpace(c.DefaultPostForm("sort_type", "")) // 0-余额 1-数量 2-次数 3-最后兑换时间
	sort := strings.TrimSpace(c.DefaultPostForm("sort", ""))          // 排序升降 0-升序 1-降序
	if mode == "" {
		mode = "agent"
	}
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	start := 0
	end := 0
	if start_date != "" {
		start = util.StoI(util.GetTimeStamp(start_date, "Y-m-d"))
	}
	if end_date != "" {
		end = util.StoI(util.GetTimeStamp(end_date, "Y-m-d"))
	}
	if cate == "" {
		cate = "isp"
	}
	nowTime := util.GetNowInt()
	cdkLists := models.GetCdkUseStatsJoin(uid, mode, cate, email, collect, start, end, sortType, sort)
	lists := []models.ResUserUsageList{}
	for _, v := range cdkLists {
		lastTimeStr := ""
		if v.LastTime > 0 {
			lastTimeStr = util.GetTimeStr(v.LastTime, "Y/m/d H:i:s")
		}
		balanceStr := ""
		valueStr := ""
		if cate == "flow" || cate == "dynamic_isp" {
			if v.Value > 0 {
				value := v.Value / 1024 / 1024 / 1024
				valueStr = fmt.Sprintf("%d %s", value, "GB")
			}
			if v.Flows > 0 {
				balanceStr = util.ItoS(v.Flows/1024/1024/1024) + " GB"
			}
		} else if cate == "unlimited" {
			if v.Value > 0 {
				value := v.Value / 86400
				valueStr = fmt.Sprintf("%d %s", value, "Day")
			}
			if v.Flows > nowTime {
				balance := v.Flows
				duration := balance - nowTime
				dayUnit := ""
				dayInfo := 0.00
				if duration > 86400 {
					dayInfo = math.Ceil(float64(duration) / 86400)
					dayUnit = "Days"
					if dayInfo == 1 {
						dayUnit = "Day"
					}
				} else {
					dayInfo = math.Ceil(float64(duration) / 3600)
					dayUnit = "Hours"
					if dayInfo == 1 {
						dayUnit = "Hour"
					}
				}
				balanceStr = fmt.Sprintf("%d %s", int(dayInfo), dayUnit)
			}
		} else {
			if v.Value > 0 {
				valueStr = fmt.Sprintf("%d %s", v.Value, "IPs")
			}
			if v.Balance > 0 {
				balanceStr = util.ItoS(v.Balance) + " IPs"
			}
		}

		info := models.ResUserUsageList{}
		info.Id = util.ItoS(v.Id)
		info.Email = v.Email
		info.Balance = balanceStr
		info.Value = valueStr
		info.Times = v.Number
		info.Collect = v.IsCollect
		info.Remark = v.Remark
		info.LastRedemptionTime = lastTimeStr
		lists = append(lists, info)
	}

	title := []string{"User email", "Current balance", "Cumulative redemption quantity", "Cumulative redemption times", "Last exchange time", "Notes"}

	var csvData [][]string
	csvData = append(csvData, title)
	for _, v := range lists {
		var info []string
		info = append(info, v.Email)
		info = append(info, v.Balance)
		info = append(info, v.Value)
		info = append(info, util.ItoS(v.Times))
		info = append(info, v.LastRedemptionTime)
		info = append(info, v.Remark)

		csvData = append(csvData, info)
	}

	err := DownloadCsv(c, "StatsUseRecord", csvData) //下载文件
	fmt.Println(err)
	return
}

// cdk生成列表下载
// @BasePath /api/v1
// @Summary cdk生成列表
// @Description cdk生成列表
// @Tags 个人中心 - 企业套餐
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} map[string]interface{} "cdk_lists：流量cdk列表（值为[]models.ResExchange{}对象）；to_lists：给他人充值列表（值为[]models.ResExchange{}对象）"
// @Router /center/cdk/ex/generate_list_download [post]
func GetGenerateListDownload(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	mode := strings.TrimSpace(c.DefaultPostForm("cate", "agent"))   // agent 代理商  balance-余额充值
	cateStr := strings.TrimSpace(c.DefaultPostForm("cate", "flow")) //isp  flow dynamic_isp unlimited
	timeType := c.DefaultPostForm("time_type", "0")                 // 0-生成时间 1-兑换时间
	startDate := c.DefaultPostForm("start_date", "")
	endDate := c.DefaultPostForm("end_date", "")
	cdkEmailStr := c.DefaultPostForm("cdk_email", "")
	statusStr := c.DefaultPostForm("status", "") //  0-all 1-未兑换 2-已兑换
	start := 0
	end := 0
	if startDate != "" {
		start = util.StoI(util.GetTimeStamp(startDate, "Y-m-d"))
	}
	if endDate != "" {
		end = util.StoI(util.GetTimeStamp(endDate, "Y-m-d"))
	}
	if mode == "" {
		mode = "agent"
	}
	cate := 2
	if cateStr == "flow" {
		cate = 3
	} else if cateStr == "isp" {
		cate = 2
	} else if cateStr == "dynamic_isp" { //动态流量套餐
		cate = 4
	} else if cateStr == "unlimited" {
		cate = 5
	} else if strings.Contains(cateStr, "static") {
		cate = 6
	}

	nowTime := util.GetNowInt()

	lists := []models.ResGenerateList{}
	cdkLists := models.GetGenerateListByCate(uid, start, end, cate, mode, timeType, cdkEmailStr, statusStr, cateStr)

	for _, v := range cdkLists {
		useTime := ""
		if v.UseTime > 0 {
			useTime = util.GetTimeStr(v.UseTime, "Y/m/d H:i:s")
		}
		balanceStr := ""
		valueStr := ""
		if v.Cate == 2 {
			if v.Balance > 0 {
				balanceStr = fmt.Sprintf("%d %s", v.Balance, "IPs")
			}
			if v.Value > 0 {
				valueStr = fmt.Sprintf("%d %s", v.Value, "IPs")
			}
		} else if v.Cate == 3 || v.Cate == 4 {
			if v.Balance > 0 {
				balanceStr = fmt.Sprintf("%d %s", v.Balance/1024/1024/1024, "GB")
			}
			if v.Value > 0 {
				value := v.Value / 1024 / 1024 / 1024
				valueStr = fmt.Sprintf("%d %s", value, "GB")
			}
		} else if v.Cate == 5 {
			if v.Balance > int64(nowTime) {
				balance := v.Balance
				duration := balance - int64(nowTime)
				dayUnit := ""
				dayInfo := 0.00
				if duration > 86400 {
					dayInfo = math.Ceil(float64(duration) / 86400)
					dayUnit = "Days"
					if dayInfo == 1 {
						dayUnit = "Day"
					}
				} else {
					dayInfo = math.Ceil(float64(duration) / 3600)
					dayUnit = "Hours"
					if dayInfo == 1 {
						dayUnit = "Hour"
					}
				}
				balanceStr = fmt.Sprintf("%d %s", int(dayInfo), dayUnit)
			}
			if v.Value > 0 {
				days := v.Value / 86400
				valueStr = fmt.Sprintf("%d %s", days, "Day")
			}
		}
		usageMode := 0
		if v.Name == "to_user" {
			usageMode = 1
		}

		info := models.ResGenerateList{}
		info.Id = util.ItoS(v.Id)
		info.CdkKey = v.Code
		info.Value = valueStr
		info.Status = v.Status
		info.UsageMode = usageMode
		info.GenerateTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i:s")
		info.RedemptionTime = useTime
		info.Balance = balanceStr
		info.Email = hideEmail(v.Email)
		info.Remark = v.GenerateRemark
		lists = append(lists, info)
	}

	title := []string{"Generation time", "Quantity", "CDKEY", "Email", "Usage modes", "Redemption Time", "Balance", "CDKEY Status", "Notes"}

	var csvData [][]string
	csvData = append(csvData, title)
	for _, v := range lists {
		usageMode := ""
		if v.UsageMode == 0 {
			usageMode = "CDKEY Redemption"
		} else {
			usageMode = "Top Up"
		}

		status := ""
		if v.Status == 1 {
			status = "Not redeemed"
		} else {
			status = "Redeemed"
		}

		var info []string
		info = append(info, v.GenerateTime)
		info = append(info, v.Value)
		info = append(info, v.CdkKey)
		info = append(info, v.Email)
		info = append(info, usageMode)
		info = append(info, v.RedemptionTime)
		info = append(info, v.Balance)
		info = append(info, status)
		info = append(info, v.Remark)

		csvData = append(csvData, info)
	}
	err := DownloadCsv(c, "GenerateListRecord", csvData) //下载文件
	fmt.Println(err)
	return
}

// @BasePath /api/v1
// @Summary CDK 统计收藏
// @Schemes
// @Description CDK	统计收藏
// @Tags 个人中心-CDK
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param sn formData string true "数据ID"
// @Param collect formData int true "是否收藏 1收藏 0取消收藏"
// @Produce json
// @Success 0 {array} map[string]interface{} "
// @Router /web/cdk/stats_collect [post]
func GetUserCdkCollect(c *gin.Context) {
	collect := com.StrTo(c.DefaultPostForm("collect", "0")).MustInt()

	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	idStr := strings.TrimSpace(c.DefaultPostForm("id", ""))
	aes_key := util.Md5(AesKey)
	idByte, err := util.AesDeCode(idStr, []byte(aes_key))
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_IP_HAS_NOT", nil)
		return
	}
	idStr = string(idByte)
	id := util.StoI(idStr)

	has, err := models.GetCdkUseStatsById(id)
	if err != nil || has.Id == 0 {
		JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil)
		return
	}
	if has.Uid != uid {
		JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil)
		return
	}
	if collect > 0 {
		params := map[string]interface{}{}
		params["is_collect"] = collect
		params["update_time"] = util.GetNowInt()
		err = models.EditCdkUseStatsById(has.Id, params)
	}

	res := map[string]interface{}{
		"collect": collect,
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", res)
	return
}

// @BasePath /api/v1
// @Summary CDK	统计备注
// @Schemes
// @Description CDK	统计备注
// @Tags 个人中心-CDK
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param sn formData string true "数据ID"
// @Param cate formData string true "类型 isp IP flow 流量 stats 统计"
// @Param remark formData string true "备注"
// @Produce json
// @Success 0 {array} map[string]interface{} "
// @Router /web/cdk/stats_remark [post]
func GetUserCdkRemark(c *gin.Context) {

	cate := strings.TrimSpace(c.DefaultPostForm("cate", "")) // 0-生成列表 1-兑换列表 -使用列表
	remark := strings.TrimSpace(c.DefaultPostForm("remark", ""))
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	idStr := strings.TrimSpace(c.DefaultPostForm("id", ""))
	aes_key := util.Md5(AesKey)
	idByte, err := util.AesDeCode(idStr, []byte(aes_key))
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_IP_HAS_NOT", nil)
		return
	}
	id := string(idByte)

	if cate == "0" {
		err, has := models.GetExchangeInfoWithId(id)
		if err != nil || has.Id == 0 {
			JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil)
			return
		}
		if has.Uid != uid {
			JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil)
			return
		}
		params := map[string]interface{}{}
		params["generate_remark"] = remark
		_ = models.EditExchangeById(has.Id, params)
	} else if cate == "1" {
		err, has := models.GetExchangeInfoWithId(id)
		if err != nil || has.Id == 0 {
			JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil)
			return
		}
		if has.BindUid != uid {
			JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil)
			return
		}
		params := map[string]interface{}{}
		params["redemption_remark"] = remark
		_ = models.EditExchangeById(has.Id, params)
	} else if cate == "2" {
		has, err := models.GetCdkUseStatsById(util.StoI(id))
		if err != nil || has.Id == 0 {
			JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil)
			return
		}
		if has.Uid != uid {
			JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil)
			return
		}
		params := map[string]interface{}{}
		params["remark"] = remark
		params["update_time"] = util.GetNowInt()
		err = models.EditCdkUseStatsById(has.Id, params)
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return
}
func hideEmail(email string) string {
	// 检查邮箱格式
	atIndex := strings.Index(email, "@")
	if atIndex == -1 || atIndex < 3 {
		// 返回原始邮箱（如果格式不对或前缀过短）
		return email
	}

	// 提取邮箱的用户名部分和域名部分
	username := email[:atIndex]
	domain := email[atIndex:]

	// 保留前两位和后两位，用 "*" 替换中间部分
	hiddenEmail := username[:2] + strings.Repeat("*", len(username)-4) + username[len(username)-2:] + domain

	return hiddenEmail
}
