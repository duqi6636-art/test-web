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

// 余额充值相关接口

// @BasePath /api/v1
// @Summary 获取基础配置信息
// @Schemes
// @Description 获取基础配置信息
// @Tags 余额充值
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {array} map[string]interface{} ""
// @Router /center/balance/config [post]
func BalanceConfigInfo(c *gin.Context) {
	lang := strings.ToLower(c.DefaultPostForm("lang", "en")) //语言
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	fmt.Println("uid:", uid)

	_, packageList := models.GetPackageListFlow("balance", 0)
	pakList := []models.ResBalancePackage{}
	pakMore := models.ResBalancePackage{}
	for _, v := range packageList {
		info := models.ResBalancePackage{
			Id:      v.Id,
			Name:    v.Name,
			Price:   v.Price,
			Default: v.Default,
			Cate:    "",
		}
		if v.Code == "balance" {
			info.Price = 0
			info.Cate = "other"
			pakMore = info
		} else {
			pakList = append(pakList, info)
		}
	}
	//获取配置信息
	//configList := models.GetBalanceConfigList("all", 0, 0) //获取默认配置信息

	//获取配置信息
	configList := models.GetBalanceConfigList("agent",uid,0)
	if len(configList) == 0 {
		userLevelInfo := models.GetUserMemberByUid(uid) //获取用户等级信息
		if userLevelInfo.Id > 0 {
			configList = models.GetBalanceConfigList("level",0,userLevelInfo.LevelID) //获取等级配置信息
		}
	}
	if len(configList) == 0 {
		configList = models.GetBalanceConfigList("all",0,0) //获取默认配置信息
	}

	aes_key := util.Md5(AesKey)
	config := []models.ResConfBalanceConfigModel{}
	for _, v := range configList {
		idStr, err := util.AesEnCode([]byte(util.ItoS(v.Id)), []byte(aes_key))
		if err != nil {
			idStr = util.ItoS(v.Id)
		}
		info := models.ResConfBalanceConfigModel{}
		info.Id = idStr
		info.Name = v.Name
		info.Cate = v.Cate
		info.Unit = v.Unit
		info.Price = v.Price
		info.PriceOrigin = v.PriceOrigin
		info.Min = v.Min
		info.Max = v.Max
		info.Num = v.Num
		config = append(config, info)
	}

	//静态国家列表
	countryList := models.GetStaticCountryByLang(lang)
	paypal_client_id := strings.Trim(models.GetConfigVal("PAYPAL_CLIENTID"), " ") //paypalID
	resData := map[string]interface{}{
		"config":           config,
		"package_list":     pakList,
		"package_more":     pakMore,
		"paypal_client_id": paypal_client_id,
		"country":          countryList,
	}
	JsonReturn(c, 0, "__T_SUCCESS", resData)
	return
}

// @BasePath /api/v1
// @Summary 获取余额记录
// @Schemes
// @Description 获取余额记录
// @Tags 余额充值
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {array} map[string]interface{} ""
// @Router /center/balance/records [post]
func BalanceRecord(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	records := models.GetUserBalanceLogBy(uid, 2, "", -1, 0, 0)
	lists := []models.ResUserBalanceLogModel{}
	if len(records) > 0 {

		configList := models.GetBalanceConfigList("all", 0, 0)
		configMap := map[string]models.ConfBalanceConfigModel{}
		for _, v := range configList {
			configMap[v.Cate] = v
		}

		for _, v := range records {
			confInfo, ok := configMap[v.Cate]
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

			info := models.ResUserBalanceLogModel{}
			info.Id = v.Id
			info.Money = fmt.Sprintf("$%.2f", v.Money)
			info.Name = name
			info.Cate = v.Cate
			info.Value = value
			info.Unit = unit
			info.Number = v.Number
			info.CreateTime = util.GetTimeStr(v.CreateTime, "d-m-Y H:i")
			lists = append(lists, info)
		}
	}
	JsonReturn(c, 0, "__T_SUCCESS", lists)
	return
}

// @BasePath /api/v1
// @Summary CDK生成兑换码-IP ----批量
// @Schemes
// @Description CDK生成兑换码-IP ----批量
// @Tags 余额充值
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param id formData string true "兑换比例ID"
// @Param number formData string true "数量  (ip数量，流量数 ，等)"
// @Param quantity formData string true "生成数量"
// @Param country formData string false "静态的时候使用"
// @Produce json
// @Success 0 {array} map[string]interface{} "cdkey:兑换码, balance:余额"
// @Router /center/cdk/balance_batch [post]
func BalanceGenerateBatchCdk(c *gin.Context) {
	time.Sleep(1)
	idStr := strings.TrimSpace(c.DefaultPostForm("id", ""))
	value_str := strings.TrimSpace(c.DefaultPostForm("number", ""))
	count_str := strings.TrimSpace(c.DefaultPostForm("quantity", ""))

	aes_key := util.Md5(AesKey)
	idByte, err := util.AesDeCode(idStr, []byte(aes_key))
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_CONFIG_NOT_EXIST", nil)
		return
	}
	idStr = string(idByte)
	id := util.StoI(idStr)
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	value0 := util.StoI(value_str)
	if value0 == 0 {
		JsonReturn(c, e.ERROR, "__T_NUMBER_ERROR", nil)
		return
	}
	count := util.StoI(count_str)
	if count <= 0 {
		JsonReturn(c, e.ERROR, "__T_COUNT_NUMBER_ERROR", nil)
		return
	}
	uid := user.Id
	//获取配置信息
	confInfo := models.GetBalanceConfigById(id)

	if confInfo.Id == 0 {
		JsonReturn(c, e.ERROR, "__T_CONFIG_NOT_EXIST", nil)
		return
	}
	country := strings.TrimSpace(c.DefaultPostForm("country", ""))
	cate := confInfo.Cate
	// 静态参数验证
	if strings.Contains(confInfo.Cate, "static") {
		if country == "" {
			JsonReturn(c, e.ERROR, "__T_EX_STATIC_AREA_ERROR", nil)
			return
		}
		areaInfo := models.GetStaticCountryByCountry(country)
		if areaInfo.Code == "" {
			JsonReturn(c, e.ERROR, "__T_EX_STATIC_AREA_ERROR", nil)
			return
		}
		cate = "static"
	}
	minNum, maxNum, numNum := confInfo.Min, confInfo.Max, confInfo.Num

	minNum = util.IfInt(minNum <= 0, minNum, 1) //单次最少数量
	numNum = util.IfInt(numNum <= 0, numNum, 1) //生成数量

	if count > numNum {
		JsonReturn(c, e.ERROR, "__T_COUNT_NUMBER_BIG_ERROR", nil)
		return
	}

	if value0 < minNum {
		if maxNum > 0 && value0 > maxNum {
			JsonReturn(c, e.ERROR, "__T_EX_NUMBER_MAX", nil)
			return
		}
		JsonReturn(c, e.ERROR, "__T_BALANCE_NUMBER_MIN_"+strings.ToUpper(cate), nil)
		return
	}
	price := confInfo.Price
	balance := 0.0
	balInfo := models.GetUserBalanceByUid(uid)
	if balInfo.Id != 0 {
		if balInfo.Balance < 0 {
			balInfo.Balance = 0
		}
		balance = balInfo.Balance
	}
	if balInfo.Status == 2 {
		JsonReturn(c, e.ERROR, "__T_PACKAGE_FORBIDDEN", nil)
		return
	}
	value := int64(value0) * confInfo.Value
	// 最终价格 = 单价 * 单个数量 * 生成CDK数量
	totalRequiredBalance := price * float64(value0) * float64(count)
	if totalRequiredBalance > balance {
		JsonReturn(c, e.ERROR, "__T_BALANCE_LOW", nil)
		return
	}
	onePrice := price * float64(value0)

	var codes []string
	nowTime := util.GetNowInt()
	for i := 0; i < count; i++ {

		str := util.ItoS(user.Id) + "_" + util.GetNowTimeStr() + "_" + util.RandStr("r", 8)
		cdkStr := util.Md5(str)
		code := strings.ToUpper(cdkStr[0:6] + cdkStr[20:])
		if cate == "static" {
			code = code + "-" + strings.ToUpper(country)
		}
		// 异步处理生成 cdk   ----start
		info := models.PushCdkey{}
		info.Mode = "balance"
		info.Cate = confInfo.Cate
		info.Country = strings.ToLower(country)
		info.Cdkey = code
		info.Number = value
		info.Need = onePrice //单个cdk价格
		info.CdkType = "cdk"
		info.Uid = user.Id
		info.BindUsername = user.Username
		info.BindEmail = user.Email
		info.Ip = c.ClientIP()
		info.CreateTime = nowTime
		listStr, _ := json.Marshal(info)
		resP := models.RedisLPUSH("list_balance_cdk", string(listStr))
		fmt.Println(resP)
		// 异步处理生成 cdk   ----end
		codes = append(codes, code)
	}

	// 写入记录
	eee := models.AddUserBalanceLog(uid, 2, totalRequiredBalance, balance, confInfo.Cate, value, count, -1, nowTime,country)
	fmt.Println("add user balance log", eee)

	balanceNew := balance - totalRequiredBalance
	//info := map[string]interface{}{"balance": balanceNew}
	//err := models.EditUserBalanceByUid(uid, info)
	//fmt.Println("update user balance", err)

	data := map[string]interface{}{}
	data["cdkey_list"] = codes
	data["balance"] = fmt.Sprintf("%.2f", balanceNew)
	JsonReturn(c, e.SUCCESS, "__T_EX_GENERATE_OK", data)
	return
}

// @BasePath /api/v1
// @Summary 自用充值到用户余额
// @Schemes
// @Description 自用充值到用户余额
// @Tags 余额充值
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param id formData string true "兑换比例ID"
// @Param value formData string true "数量  (ip数量，流量数 ，等)"
// @Param country formData string false "静态的时候使用"
// @Produce json
// @Success 0 {array} map[string]interface{} ""
// @Router /center/cdk/balance_self_use [post]
func SelfBalanceCdk(c *gin.Context) {
	idStr := strings.TrimSpace(c.DefaultPostForm("id", ""))
	aes_key := util.Md5(AesKey)
	idByte, err := util.AesDeCode(idStr, []byte(aes_key))
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_CONFIG_NOT_EXIST", nil)
		return
	}
	idStr = string(idByte)
	id := util.StoI(idStr)
	value_str := strings.TrimSpace(c.DefaultPostForm("value", ""))

	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	value0 := util.StoI(value_str)
	if value0 == 0 {
		JsonReturn(c, e.ERROR, "__T_NUMBER_ERROR", nil)
		return
	}

	uid := user.Id

	//today := util.GetTodayTime()
	//selfList := models.GetCdkListByUid(uid, uid, "balance", "agent_self", today)
	//numSelf := len(selfList)
	//confTodayNumStr := models.GetConfigVal("CONFIG_CDK_SELF_NUM_DAY")
	//tNum := util.StoI(confTodayNumStr)
	//if tNum == 0{
	//	tNum = 3
	//}
	//if numSelf >= tNum {
	//	JsonReturn(c, e.ERROR, "__T_AGENT_SELF_LIMIT", nil)
	//	return
	//}

	//获取配置信息
	confInfo := models.GetBalanceConfigById(id)
	if confInfo.Id == 0 {
		JsonReturn(c, e.ERROR, "__T_CONFIG_NOT_EXIST", nil)
		return
	}
	country := strings.TrimSpace(c.DefaultPostForm("country", ""))
    cate := confInfo.Cate
	// 静态参数验证
	if strings.Contains(confInfo.Cate, "static") {
		if country == "" {
			JsonReturn(c, e.ERROR, "__T_EX_STATIC_AREA_ERROR", nil)
			return
		}
		areaInfo := models.GetStaticCountryByCountry(country)
		if areaInfo.Code == "" {
			JsonReturn(c, e.ERROR, "__T_EX_STATIC_AREA_ERROR", nil)
			return
		}
		cate = "static"
	}
	minNum, maxNum := confInfo.Min, confInfo.Max

	minNum = util.IfInt(minNum <= 0, minNum, 1)   //单次最少数量
	maxNum = util.IfInt(maxNum <= 0, maxNum, 100) //单次最多数量

	if value0 < minNum {
		if maxNum > 0 && value0 > maxNum {
			JsonReturn(c, e.ERROR, "__T_EX_NUMBER_MAX", nil)
			return
		}
		JsonReturn(c, e.ERROR, "__T_BALANCE_NUMBER_MIN_"+strings.ToUpper(cate), nil)
		return
	}

	price := confInfo.Price
	balance := 0.0
	balInfo := models.GetUserBalanceByUid(uid)
	if balInfo.Id != 0 {
		if balInfo.Balance < 0 {
			balInfo.Balance = 0
		}
		balance = balInfo.Balance
	}
	if balInfo.Status == 2 {
		JsonReturn(c, e.ERROR, "__T_PACKAGE_FORBIDDEN", nil)
		return
	}
	value := int64(value0) * confInfo.Value
	if cate == "static" {
		value = int64(value0)
	}
	// 最终价格 = 单价 * 单个数量 * 生成CDK数量
	totalRequiredBalance := price * float64(value0)
	if totalRequiredBalance > balance {
		JsonReturn(c, e.ERROR, "__T_BALANCE_LOW", nil)
		return
	}

	nowTime := util.GetNowInt()
	str := util.ItoS(user.Id) + "_" + util.GetNowTimeStr() + "_" + util.RandStr("r", 8)
	cdkStr := util.Md5(str)
	code := strings.ToUpper(cdkStr[0:6] + cdkStr[20:])
	if cate == "static" {
		code = code + "-"+ strings.ToUpper(country)
	}
	// 异步处理生成 cdk   ----start
	info := models.PushCdkey{}
	info.Mode = "balance"
	info.Cate = confInfo.Cate
	info.Country = strings.ToLower(country)
	info.Cdkey = code
	info.Number = value
	info.CdkType = "self"
	info.Need = totalRequiredBalance
	info.Uid = user.Id
	info.BindUid = user.Id
	info.BindUsername = user.Username
	info.BindEmail = user.Email
	info.Ip = c.ClientIP()
	info.CreateTime = nowTime
	listStr, _ := json.Marshal(info)
	resP := models.RedisLPUSH("list_balance_cdk", string(listStr))
	fmt.Println(resP)
	// 异步处理生成 cdk   ----end

	balanceNew := balance - totalRequiredBalance
	data := map[string]interface{}{}
	data["cdkey"] = code
	data["balance"] = fmt.Sprintf("%.2f", balanceNew)
	if resP == nil {
		// 写入记录
		eee := models.AddUserBalanceLog(uid, 3, totalRequiredBalance, balance, confInfo.Cate, value, 1, -1, nowTime,country)
		fmt.Println("add user balance log", eee)

		JsonReturn(c, 0, "__T_EX_SUCCESS", data)
		return
	}
	JsonReturn(c, e.ERROR, "error", nil)
	return
}

// @BasePath /api/v1
// @Summary 换轮转ISP流量
// @Schemes
// @Description 换轮转ISP流量
// @Tags 个人中心-CDK
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param cdkey formData string true "CDK码"
// @Produce json
// @Success 0 {array} map[string]interface{} "user_balance:用户剩余IP数, user_flow:用户流量数"
// @Router /center/cdk/ex_dynamic [post]
func ExchangeDynamicCdk(c *gin.Context) {
	time.Sleep(1)
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	cdkey := strings.TrimSpace(c.DefaultPostForm("code", ""))
	if cdkey == "" {
		JsonReturn(c, -1, "__T_EX_EMPTY", nil)
		return
	}

	nowTime := util.GetNowInt()
	//res := false
	var err error

	err, exInfo := models.GetExchangeInfo(cdkey)
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
	if exInfo.Cate != 4 {
		JsonReturn(c, -1, "__T_EX_CODE_ERROR", nil)
		return
	}

	cate := "dynamic_isp"

	// 异步处理兑换 cdk   ----start
	info := models.PushCdkey{}
	info.Mode = exInfo.Mode
	info.Cate = cate
	info.Cdkey = cdkey
	info.Number = exInfo.Value
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

	dynamicIspFlows := int64(0)
	dynamicIspDate := "--"
	dynamicIspExpire := 0
	userIspFlowInfo := models.GetUserDynamicIspInfo(uid)
	if userIspFlowInfo.ID > 0 {
		if userIspFlowInfo.Flows > 0 {
			dynamicIspFlows = userIspFlowInfo.Flows + exInfo.Value
		}
		expTime := nowTime + 86400*30 //默认30天
		dynamicIspDate = util.GetTimeStr(expTime, "d/m/Y")
	}
	send_isp_open := userIspFlowInfo.SendOpen
	send_isp_flow := ""
	send_isp_unit := "GB"

	send_isp := userIspFlowInfo.SendFlows
	send_isp_unit = userIspFlowInfo.SendUnit
	if send_isp_unit == "" {
		send_isp_unit = "GB"
	}
	flowIspChar := int64(0)
	if send_isp_unit == "GB" {
		flowIspChar = 1024 * 1024 * 1024
	} else {
		flowIspChar = 1024 * 1024
	}
	send_isp_flow = util.ItoS(int(send_isp / flowIspChar)) //设置信息

	flowIspStr, flowIspUnit := DealFlowChar(dynamicIspFlows, "GB")
	flowIspMbStr, flowIspMbUnit := DealFlowChar(dynamicIspFlows, "MB")

	dynamicIspInfo := models.GetFlowsStruct{}
	dynamicIspInfo.Flow = dynamicIspFlows
	dynamicIspInfo.ExpireDate = dynamicIspDate
	dynamicIspInfo.Expired = dynamicIspExpire
	dynamicIspInfo.FlowGb = flowIspStr
	dynamicIspInfo.FlowMb = flowIspMbStr
	dynamicIspInfo.SendFlow = send_isp_flow
	dynamicIspInfo.Status = userIspFlowInfo.Status
	dynamicIspInfo.SendOpen = send_isp_open
	dynamicIspInfo.SendUnit = send_isp_unit
	dynamicIspInfo.UnitGb = flowIspUnit
	dynamicIspInfo.UnitMb = flowIspMbUnit

	if resP == nil {
		JsonReturn(c, 0, "__T_EX_SUCCESS", dynamicIspInfo)
		return
	}
	JsonReturn(c, -1, "__T_EX_ERROR--!", nil)
	return
}

// @BasePath /api/v1
// @Summary 兑换不限量
// @Schemes
// @Description 兑换不限量
// @Tags 个人中心-CDK
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param cdkey formData string true "CDK码"
// @Produce json
// @Success 0 {array} GetUnlimitedStruct{} ""
// @Router /center/cdk/ex_unlimited [post]
func ExchangeUnlimitedCdk(c *gin.Context) {
	time.Sleep(1)
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	lang := strings.ToLower(c.DefaultPostForm("lang", "en")) //语言
	cdkey := strings.TrimSpace(c.DefaultPostForm("code", ""))
	if cdkey == "" {
		JsonReturn(c, -1, "__T_EX_EMPTY", nil)
		return
	}

	nowTime := util.GetNowInt()
	//res := false
	var err error

	err, exInfo := models.GetExchangeInfo(cdkey)
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
	if exInfo.Cate != 4 {
		JsonReturn(c, -1, "__T_EX_CODE_ERROR", nil)
		return
	}

	cate := "unlimited"

	// 异步处理兑换 cdk   ----start
	info := models.PushCdkey{}
	info.Mode = exInfo.Mode
	info.Cate = cate
	info.Cdkey = cdkey
	info.Number = exInfo.Value
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

	flowDay := models.GetUserFlowDayByUid(uid)
	// 不限量流量信息
	dayUnit := "Day"
	dayExpire := "--"
	day := 0    //剩余时间
	dayUse := 0 //是否能用
	if flowDay.Id > 0 {
		expTime := flowDay.ExpireTime + int(exInfo.Value)
		if flowDay.ExpireTime > nowTime {
			expTime = flowDay.ExpireTime + int(exInfo.Value)
		} else {
			expTime = nowTime + int(exInfo.Value)
		}
		duration := expTime - nowTime
		dayUnit = ""
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
		day = int(dayInfo)

		dayExpire = util.GetTimeByLang(flowDay.ExpireTime, lang)
		dayUse = 1
	}

	unlimitedInfo := GetUnlimitedStruct{}
	unlimitedInfo.Day = day
	unlimitedInfo.DayUnit = dayUnit
	unlimitedInfo.DayExpire = dayExpire
	unlimitedInfo.DayUse = dayUse            //是否能用
	unlimitedInfo.DayStatus = flowDay.Status //是否冻结

	if resP == nil {
		JsonReturn(c, 0, "__T_EX_SUCCESS", unlimitedInfo)
		return
	}
	JsonReturn(c, -1, "__T_EX_ERROR--!", nil)
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
// @Router /center/balance/stats_use [post]
func GetUserBalanceCdkStats(c *gin.Context) {
	mode := strings.TrimSpace(c.DefaultPostForm("mode", "balance")) // agent 代理商  balance-余额充值
	cate := strings.TrimSpace(c.DefaultPostForm("cate", "flow"))    //isp  flow dynamic_isp unlimited
	startDate := c.DefaultPostForm("start_date", "")
	endDate := c.DefaultPostForm("end_date", "")
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))
	collect := com.StrTo(c.DefaultPostForm("collect", "0")).MustInt()
	sortType := strings.TrimSpace(c.DefaultPostForm("sort_type", "")) // 0-余额 1-数量 2-次数 3-最后兑换时间
	sort := strings.TrimSpace(c.DefaultPostForm("sort", ""))          // 排序升降 0-升序 1-降序
	if mode == "" {
		mode = "balance"
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
	//if cate == "" {
	//	cate = "flow"
	//}
	cate = "" //20241127 查询全部类型  后期需要按照类型筛选时 注释掉此行，并把上面的注释去掉

	configList := models.GetBalanceConfigList("all", 0, 0)
	configMap := map[string]models.ConfBalanceConfigModel{}
	for _, v := range configList {
		configMap[v.Cate] = v
	}
	//nowTime := util.GetNowInt()
	cdkLists := models.GetCdkUseStats(uid, mode, cate, email, collect, start, end, sortType, sort)
	lists := []models.ResUserUsageList{}
	aes_key := util.Md5(AesKey)
	for _, v := range cdkLists {
		lastTimeStr := ""
		if v.LastTime > 0 {
			lastTimeStr = util.GetTimeStr(v.LastTime, "Y/m/d H:i:s")
		}
		confInfo, ok := configMap[v.Cate]
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
		idStr, err := util.AesEnCode([]byte(util.ItoS(v.Id)), []byte(aes_key))
		if err != nil {
			idStr = util.ItoS(v.Id)
		}
		info := models.ResUserUsageList{}
		info.Id = idStr
		info.Email = hideEmail(v.Email)
		info.ExchangeType = name
		info.Balance = ""
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
// @Summary CDK	获取用户CDK统计使用列表
// @Schemes
// @Description CDK	获取用户CDK使用列表
// @Tags 个人中心-CDK
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param id formData string true "统计记录ID"
// @Produce json
// @Success 0 {array} []StatsUserExchangeList{}"
// @Router /center/balance/stats_detail [post]
func GetUserBalanceCdkStatsDetail(c *gin.Context) {
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
	configList := models.GetBalanceConfigList("all", 0, 0)
	configMap := map[string]models.ConfBalanceConfigModel{}
	for _, v := range configList {
		configMap[v.Cate] = v
	}
	packageList := models.GetStaticPackageList()
	packArr := map[int]int{}
	for _, v := range packageList {
		packArr[v.Value] = v.Id
	}

	nowTime := util.GetNowInt()
	res := []StatsUserExchangeList{}
	listInfo := models.GetCdkUseStatsByUser(uid, has.BindUid, "", has.Mode, 0, 0)
	for _, v := range listInfo {
		confInfo, ok := configMap[v.Cate]
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
		valueStr := fmt.Sprintf("%d %s", value,unit)
		balanceStr := ""
		if v.Cate == "isp" {
			_,userInfo := models.GetUserById(v.BindUid)
			balanceStr = fmt.Sprintf("%d %s", userInfo.Balance, unit)
		}
		if v.Cate == "flow" {
			userInfo := models.GetUserFlowInfo(v.BindUid)
			balance := userInfo.Flows
			if balance > 0 && valueUnit > 0 {
				balance = balance / valueUnit
			}
			balanceStr = fmt.Sprintf("%d %s", balance, unit)
		}
		if v.Cate == "dynamic_isp" {
			userInfo := models.GetUserDynamicIspInfo(v.BindUid)
			balance := userInfo.Flows
			if balance > 0 && valueUnit > 0 {
				balance = balance / valueUnit
			}
			balanceStr = fmt.Sprintf("%d %s", balance, unit)
		}
		if v.Cate == "unlimited" {
			userInfo := models.GetUserFlowDay(v.BindUid)
			balance := userInfo.ExpireTime
			if balance > nowTime {
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
			}else {
				balanceStr = "0 Day"
			}
		}
		cateStr := v.Cate
		if strings.Contains(cateStr,"static") {
			exArr := strings.Split(cateStr, "-")
			exDay := util.StoI(exArr[1])
			if  exDay == 0 {
				exDay = 7
			}
			pakId := packArr[exDay]
			if pakId > 0 {
				pakId = packageList[0].Id
			}
			userInfo,_ := models.GetUserStaticByPak(v.BindUid, pakId)
			balanceStr = fmt.Sprintf("%d %s", userInfo.Balance, unit)
		}


		info := StatsUserExchangeList{}
		info.Id = v.Id
		info.ExchangeType = name
		info.Balance = balanceStr
		info.Value = valueStr
		info.Times = v.Number
		info.LastTime = util.GetTimeStr(v.LastTime, "Y/m/d H:i:s")
		res = append(res, info)
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", res)
	return
}

type StatsUserExchangeList struct {
	Id           int    `json:"id"`
	ExchangeType string `json:"exchange_type"` // 兑换类型
	Balance      string `json:"balance"`       //用户余额
	Value        string `json:"value"`         // 累计兑换数量
	Times        int    `json:"times"`         // 累计兑换次数
	LastTime     string `json:"last_time"`     // 最后兑换时间
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
// @Router /center/balance/stats_download [post]
func GetUserBalanceCdkStatsDownload(c *gin.Context) {
	mode := strings.TrimSpace(c.DefaultPostForm("mode", "agent")) // agent 代理商  balance-余额充值
	cate := strings.TrimSpace(c.DefaultPostForm("cate", "flow"))  //isp  flow dynamic_isp unlimited
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))
	collect := com.StrTo(c.DefaultPostForm("collect", "0")).MustInt()
	start_date := c.DefaultPostForm("start_date", "")
	end_date := c.DefaultPostForm("end_date", "")
	sortType := strings.TrimSpace(c.DefaultPostForm("sort_type", "")) // 0-余额 1-数量 2-次数 3-最后兑换时间
	sort := strings.TrimSpace(c.DefaultPostForm("sort", ""))          // 排序升降 0-升序 1-降序
	if mode == "" {
		mode = "balance"
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
	//if cate == "" {
	//	cate = "flow"
	//}
	cate = "" //20241127 查询全部类型  后期需要按照类型筛选时 注释掉此行，并把上面的注释去掉
	configList := models.GetBalanceConfigList("all", 0, 0)
	configMap := map[string]models.ConfBalanceConfigModel{}
	for _, v := range configList {
		configMap[v.Cate] = v
	}
	//nowTime := util.GetNowInt()
	cdkLists := models.GetCdkUseStats(uid, mode, cate, email, collect, start, end, sortType, sort)
	lists := []models.ResUserUsageList{}
	for _, v := range cdkLists {
		lastTimeStr := ""
		if v.LastTime > 0 {
			lastTimeStr = util.GetTimeStr(v.LastTime, "Y/m/d H:i:s")
		}
		confInfo, ok := configMap[v.Cate]
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

		info := models.ResUserUsageList{}
		info.Id = util.ItoS(v.Id)
		info.Email = hideEmail(v.Email)
		info.ExchangeType = name
		info.Balance = ""
		info.Value = valueStr
		info.Times = v.Number
		info.Collect = v.IsCollect
		info.Remark = v.Remark
		info.LastRedemptionTime = lastTimeStr

		lists = append(lists, info)
	}

	title := []string{"User email", "Cumulative redemption quantity", "Cumulative redemption times", "Last exchange time", "Notes"}

	var csvData [][]string
	csvData = append(csvData, title)
	for _, v := range lists {
		var info []string
		info = append(info, v.Email)
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
