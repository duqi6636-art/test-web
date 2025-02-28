package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"math"
	"strings"
	"time"
)

/* 新版账号密码*/
// 获取信息
// @BasePath /api/v1
// @Summary 获取流量信息
// @Description 获取流量信息
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} map[string]interface{} "flow：流量，flow_gb：流量GB，unit_gb：流量单位，flow_mb：流量MB，unit_mb：流量单位，flow_date：流量到期时间，flow_expire：流量到期状态，send_open：是否开启流量发送，send_flow：发送流量，send_unit：发送流量单位，day_expire：剩余天数，day：剩余天数，day_unit：剩余天数单位，day_use：是否能用"
// @Router /web/account/get_info [post]
func GetAccountInfo(c *gin.Context) {
	nowTime := util.GetNowInt()
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	// 获取流量信息
	flows := int64(0)
	flowDate := "--"
	flowExpire := 0
	flowStatus := 1
	userFlowInfo := models.GetUserFlowInfo(userInfo.Id)
	if userFlowInfo.ID != 0 {
		//if userFlowInfo.Flows > 0 {	// 这里注释掉，因为有些用户流量 允许用户的流量为负数 20250114 需求
		flows = userFlowInfo.Flows
		//}
		flowDate = util.GetTimeStr(userFlowInfo.ExpireTime, "d/m/Y")
		if userFlowInfo.ExpireTime < nowTime {
			flowExpire = 1
		}
		if userFlowInfo.Status == 2 {
			flowStatus = userFlowInfo.Status
		}
	}
	send_open := userFlowInfo.SendOpen
	send_flow := ""
	send_unit := "GB"
	//if userFlowInfo.SendFlows > 0 {
	flow := userFlowInfo.SendFlows
	send_unit = userFlowInfo.SendUnit
	if send_unit == "" {
		send_unit = "GB"
	}
	flowChar := int64(0)
	if send_unit == "GB" {
		flowChar = 1024 * 1024 * 1024
	} else {
		flowChar = 1024 * 1024
	}
	send_flow = util.ItoS(int(flow / flowChar)) //设置信息
	//}
	flowStr, flowUnit := DealFlowChar(flows, "GB")
	flowMbStr, flowMbUnit := DealFlowChar(flows, "MB")
	dayUnit := "Day"
	dayExpire := "--"
	day := 0       //剩余时间
	dayUse := 0    //是否能用
	dayStatus := 1 //状态
	flowDay := models.GetUserFlowDayByUid(userInfo.Id)
	if flowDay.Id > 0 {
		if flowDay.ExpireTime > nowTime {
			duration := flowDay.ExpireTime - nowTime
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

			dayExpire = util.GetTimeStr(flowDay.ExpireTime, "d/m/Y")
			dayUse = 1
		}
		if flowDay.Status == 2 {
			dayStatus = flowDay.Status
		}
	}

	// 获取动态Isp流量信息 -- start
	dynamicIspFlows := int64(0)
	dynamicIspDate := "--"
	dynamicIspExpire := 0
	userIspFlowInfo := models.GetUserDynamicIspInfo(userInfo.Id)
	dynamicIspStatus := 1
	if userIspFlowInfo.ID > 0 {
		if userIspFlowInfo.Flows > 0 {
			dynamicIspFlows = userIspFlowInfo.Flows
		}
		dynamicIspDate = util.GetTimeStr(userIspFlowInfo.ExpireTime, "d/m/Y")
		if userIspFlowInfo.ExpireTime < nowTime {
			dynamicIspExpire = 1
		}
		if userIspFlowInfo.Status == 2 {
			dynamicIspStatus = userIspFlowInfo.Status
		}
	}
	send_isp_open := userIspFlowInfo.SendOpen
	send_isp_flow := ""
	send_isp_unit := "GB"

	send_isp := userIspFlowInfo.SendFlows
	send_isp_unit = userFlowInfo.SendUnit
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
	dynamicIspInfo.Status = dynamicIspStatus
	dynamicIspInfo.SendOpen = send_isp_open
	dynamicIspInfo.SendUnit = send_isp_unit
	dynamicIspInfo.UnitGb = flowIspUnit
	dynamicIspInfo.UnitMb = flowIspMbUnit
	// -- 动态流量信息 -- end  //

	resData := map[string]interface{}{}
	resData["flow_status"] = flowStatus
	resData["flow"] = flows
	resData["flow_gb"] = flowStr
	resData["unit_gb"] = flowUnit
	resData["flow_mb"] = flowMbStr
	resData["unit_mb"] = flowMbUnit
	resData["flow_date"] = flowDate
	resData["flow_expire"] = flowExpire
	resData["send_open"] = send_open
	resData["send_flow"] = send_flow
	resData["send_unit"] = send_unit

	resData["day_expire"] = dayExpire
	resData["day"] = day
	resData["day_unit"] = dayUnit
	resData["day_use"] = dayUse             //是否能用
	resData["dynamic_isp"] = dynamicIspInfo //动态流量信息
	resData["day_status"] = dayStatus

	//sock5
	resData["sock5_num"] = userInfo.Balance
	resData["sock5_status"] = userInfo.IpStatus

	JsonReturn(c, 0, "__T_SUCCESS", resData)
	return
}

// @BasePath /api/v1
// PingExample godoc
// @Summary 获取用户流量信息
// @Schemes
// @Description 获取用户流量信息 返回余额信息
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {array} GetAccountInfoResponse{}
// @Router /center/flows/get_info [post]
func GetAccountInfoV2(c *gin.Context) {
	nowTime := util.GetNowInt()
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	// 获取流量信息 -- start
	flows := int64(0)
	flowDate := "--"
	flowExpire := 0
	userFlowInfo := models.GetUserFlowInfo(userInfo.Id)
	if userFlowInfo.ID > 0 {
		//if userFlowInfo.Flows > 0 { // 这里注释掉，因为有些用户流量 允许用户的流量为负数 20250114 需求
		flows = userFlowInfo.Flows
		//}
		flowDate = util.GetTimeStr(userFlowInfo.ExpireTime, "d/m/Y")
		if userFlowInfo.ExpireTime < nowTime {
			flowExpire = 1
		}
	}
	send_open := userFlowInfo.SendOpen
	send_flow := ""
	send_unit := "GB"
	//if userFlowInfo.SendFlows > 0 {
	flow := userFlowInfo.SendFlows
	send_unit = userFlowInfo.SendUnit
	if send_unit == "" {
		send_unit = "GB"
	}
	flowChar := int64(0)
	if send_unit == "GB" {
		flowChar = 1024 * 1024 * 1024
	} else {
		flowChar = 1024 * 1024
	}
	send_flow = util.ItoS(int(flow / flowChar)) //设置信息

	flowStr, flowUnit := DealFlowChar(flows, "GB")
	flowMbStr, flowMbUnit := DealFlowChar(flows, "MB")

	flowsInfo := GetFlowsStruct{}
	flowsInfo.Flow = flows
	flowsInfo.ExpireDate = flowDate
	flowsInfo.Expired = flowExpire
	flowsInfo.FlowGb = flowStr
	flowsInfo.FlowMb = flowMbStr
	flowsInfo.SendFlow = send_flow
	flowsInfo.Status = userFlowInfo.Status
	flowsInfo.SendOpen = send_open
	flowsInfo.SendUnit = send_unit
	flowsInfo.UnitGb = flowUnit
	flowsInfo.UnitMb = flowMbUnit
	// 流量信息 -- end  //

	// 不限量流量信息
	dayUnit := "Day"
	dayExpire := "--"
	day := 0    //剩余时间
	dayUse := 0 //是否能用
	flowDay := models.GetUserFlowDayByUid(userInfo.Id)
	if flowDay.Id > 0 {
		if flowDay.ExpireTime > nowTime {
			duration := flowDay.ExpireTime - nowTime
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

			dayExpire = util.GetTimeStr(flowDay.ExpireTime, "d/m/Y")
			dayUse = 1
		}
	}

	unlimitedInfo := GetUnlimitedStruct{}
	unlimitedInfo.Day = day
	unlimitedInfo.DayUnit = dayUnit
	unlimitedInfo.DayExpire = dayExpire
	unlimitedInfo.DayUse = dayUse            //是否能用
	unlimitedInfo.DayStatus = flowDay.Status //是否冻结

	// 获取动态Isp流量信息 -- start
	dynamicIspFlows := int64(0)
	dynamicIspDate := "--"
	dynamicIspExpire := 0
	userIspFlowInfo := models.GetUserDynamicIspInfo(userInfo.Id)
	if userIspFlowInfo.ID > 0 {
		if userIspFlowInfo.Flows > 0 {
			dynamicIspFlows = userIspFlowInfo.Flows
		}
		dynamicIspDate = util.GetTimeStr(userIspFlowInfo.ExpireTime, "d/m/Y")
		if userIspFlowInfo.ExpireTime < nowTime {
			dynamicIspExpire = 1
		}
	}
	send_isp_open := userIspFlowInfo.SendOpen
	send_isp_flow := ""
	send_isp_unit := "GB"

	send_isp := userIspFlowInfo.SendFlows
	send_isp_unit = userFlowInfo.SendUnit
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

	dynamicIspInfo := GetFlowsStruct{}
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
	// -- 动态流量信息 -- end  //

	resData := GetAccountInfoResponse{}
	resData.FlowDay = unlimitedInfo     //不限量信息  //新版分组展示
	resData.Flows = flowsInfo           //流量信息
	resData.DynamicIsp = dynamicIspInfo //动态流量信息

	// 获取用户ISP信息 -- start---------------------
	resData.Isp.Balance = userInfo.Balance
	// 获取用户ISP信息 -- end -------------------------

	// 获取用户代理商信息 -- start---------------------
	agentBalance := 0
	agentInfo := models.GetAgentBalanceByUid(userInfo.Id)
	if agentInfo.Id != 0 {
		if agentInfo.Balance < 0 {
			agentInfo.Balance = 0
		}
		agentBalance = agentInfo.Balance
	}
	resData.IspAgent.Balance = agentBalance
	//获取用户代理商信息 -- end---------------------

	// 获取静态信息 -- start-----------------------
	_, staticInfo := models.GetUserStaticIp(userInfo.Id)
	packageList := models.GetStaticPackageList()

	userBalance := map[int]int{}
	for _, vu := range staticInfo {
		if vu.PakRegion != "all" {
			balance, ok := userBalance[vu.PakId]
			if !ok {
				balance = 0
			}
			userBalance[vu.PakId] = vu.Balance + balance
		}
	}

	resInfo := []GetStaticResponse{}
	for _, vp := range packageList {
		info := GetStaticResponse{}
		info.Id = vp.Id
		ipNum, ok := userBalance[vp.Id]
		if !ok {
			ipNum = 0
		}
		info.PakName = vp.Name
		info.ExpireDay = vp.Value
		info.Balance = ipNum
		resInfo = append(resInfo, info)
	}

	resData.Static = resInfo

	// 获取静态信息 -- end-----------------------

	// 获取住宅代理 企业 -- start---------------------
	flowRedeem := "0"
	if userFlowInfo.ID != 0 {
		flowDate = util.GetTimeStr(userFlowInfo.ExpireTime, "d/m/Y")
		if userFlowInfo.ExpireTime < nowTime {
			flowExpire = 1
		}
		redeemFlow := userFlowInfo.BuyFlow
		if redeemFlow > 0 {
			redeemFlow = userFlowInfo.BuyFlow - userFlowInfo.CdkFlow
			if redeemFlow > flow { //如果买的 大于 剩余的 就展示剩余的
				redeemFlow = flow
			}
		}

		flowRedeem, _ = DealFlowChar(redeemFlow, "GB")
	}

	resData.FloWAgent = GetFowAgentResponse{
		Flow:       flowStr,
		FlowRedeem: flowRedeem,
		FlowUnit:   flowUnit,
		FlowMb:     flowMbStr,
		FlowMbUnit: flowMbUnit,
		FlowDate:   flowDate,
		FlowExpire: flowExpire,
	}
	// 获取住宅代理 企业 -- end---------------------

	// 获取用户余额记录 -- start---------------------
	balance := "0"
	balInfo := models.GetUserBalanceByUid(userInfo.Id)
	if balInfo.Id > 0 {
		balance = util.FtoS2(balInfo.Balance, 2)
	}
	resData.Balance = GetBalanceInfo{Balance: balance, Status: balInfo.Status}
	// 获取用户余额记录 -- end---------------------

	JsonReturn(c, 0, "__T_SUCCESS", resData)
	return
}

// 设置用户低于流量阀值 发送邮件
// @BasePath /api/v1
// @Summary 设置用户低于流量阀值 发送邮件
// @Description 设置用户低于流量阀值 发送邮件
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param open formData string true "是否开启发送邮件"
// @Param flow formData string true "流量阀值"
// @Param unit formData string true "流量阀值单位"
// @Produce json
// @Success 0 {object} map[string]interface{} "send_open：是否开启发送邮件，send_flow：发送流量阀值，send_unit：发送流量阀值单位"
// @Router /web/account/set_send [post]
func SetSendFlows(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	openStr := strings.TrimSpace(c.DefaultPostForm("open", "0"))
	flowStr := strings.TrimSpace(c.DefaultPostForm("flow", "0"))
	flowUnit := strings.TrimSpace(c.DefaultPostForm("unit", "GB"))

	flowUnit = strings.ToUpper(flowUnit)
	if flowUnit == "GB" || flowUnit == "MB" || flowUnit == "KB" {

	} else {
		JsonReturn(c, -1, "__T_PARAM_ERROR", gin.H{})
		return
	}
	if flowUnit == "" {
		flowUnit = "GB"
	}
	open := util.StoI(openStr)
	flow := int64(util.StoI(flowStr))
	flows := int64(0)
	if flowUnit == "GB" {
		flows = flow * 1024 * 1024 * 1024
	} else {
		flows = flow * 1024 * 1024
	}
	flowInfo := models.GetUserFlowInfo(userInfo.Id)
	if flowInfo.ID == 0 {
		JsonReturn(c, -1, "__T_NO_FLOW_INFO", gin.H{})
		return
	}
	upinfo := map[string]interface{}{}

	upinfo["send_open"] = open
	upinfo["send_flows"] = flows
	upinfo["send_unit"] = flowUnit
	if flowInfo.SendFlows <= flows {
		upinfo["send_has"] = 0
	}
	res := models.EditUserFlow(flowInfo.ID, upinfo)
	if res != nil {
		JsonReturn(c, -1, "__T_FAIL", gin.H{})
		return
	}

	resData := map[string]interface{}{}
	resData["send_open"] = open
	resData["send_flow"] = flowStr
	resData["send_unit"] = flowUnit

	JsonReturn(c, 0, "__T_SUCCESS", resData)
	return
}

// 获取用户流量记录 -折线或柱状图使用  -- 新
// @BasePath /api/v1
// @Summary 获取用户流量记录 -折线或柱状图使用  -- 新
// @Description 获取用户流量记录 -折线或柱状图使用  -- 新
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param start_date formData string true "开始日期"
// @Param end_date formData string true "结束日期"
// @Param url formData string true "站点地址  多选 ，以逗号分隔"
// @Param flow_unit formData string true "流量单位"
// @Param flow_type formData string true "流量类型 1：正常流量 2：无限流量"
// @Produce json
// @Success 0 {object} map[string]interface{} "x_data：日期数组，y_data：流量数组，cate：分类名称，unit：流量单位"
// @Router /web/account/chart_data [post]
func GetFlowStats(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	start_date := c.DefaultPostForm("start_date", "")
	end_date := c.DefaultPostForm("end_date", "")
	websiteStr := strings.TrimSpace(c.DefaultPostForm("url", ""))     // 站点地址  多选 ，以逗号分隔
	flow_unit := c.DefaultPostForm("flow_unit", "GB")                 // 站点地址
	country := c.DefaultPostForm("country", "")                       // 国家
	accountStr := strings.TrimSpace(c.DefaultPostForm("account", "")) // 子账号名称
	siteUrl := c.DefaultPostForm("site_url", "")                      // 访问地址
	flowUseType := c.DefaultPostForm("flow_use_type", "")             // 流量使用类型，仅flowtype =1时有效
	if flow_unit == "" {
		flow_unit = "GB"
	}
	flow_type := strings.TrimSpace(c.DefaultPostForm("flow_type", "1"))
	if flow_type == "" {
		flow_type = "1"
	}
	flowType := util.StoI(flow_type)
	if flowType == 0 {
		flowType = 1
	}
	var start, end int
	x_data := []string{}
	if start_date == "" || end_date == "" {
		today := util.GetTodayTime()
		create := today - 10*86400
		for i := 0; i <= 10; i++ {
			day := create + i*86400
			dayStr := util.GetTimeStr(day, "Y-m-d")
			x_data = append(x_data, dayStr)
		}
		start = create
		end = today
	} else {
		start = util.StoI(util.GetTimeStamp(start_date, "Y-m-d"))
		end = util.StoI(util.GetTimeStamp(end_date, "Y-m-d"))

		for i := 0; i <= (end-start)/86400; i++ {
			day := start + i*86400
			dayStr := util.GetTimeStr(day, "Y-m-d")
			x_data = append(x_data, dayStr)
		}
	}
	accountId := 0
	if accountStr != "" {
		err, userAccount := models.GetUserAccount(uid, accountStr)
		if err == nil && userAccount.Id > 0 {
			accountId = userAccount.Id
		}
	}
	websiteArr := strings.Split(websiteStr, ",")
	//list := models.GetUrlUsed(uid, website, start, end)
	list := []models.StAddressToday{}
	if flowType == 1 {
		//list = models.GetUrlUsed(uid, 0, start, end)
		list = models.GetFlowUsedStat(uid, accountId, start, end, country, siteUrl, flowUseType)
	} else if flowType == 2 {
		//list = models.GetUnlimitedUrlUsed(uid, 0, start, end)
		list = models.GetFlowDayUsedStat(uid, 0, start, end)
	} else if flowType == 3 {
		//list = models.GetDynamicUrlUsed(uid, 0, start, end)
		list = models.GetIspFlowUsedStat(uid, accountId, start, end, country, siteUrl)
	}

	var cateName []string
	cateName = []string{
		"Flow",
	}

	kvInfo := map[string]int64{}
	cateInfo := map[string][]models.StAddressToday{}
	for _, vv := range cateName {
		cateInfo[vv] = []models.StAddressToday{}
		for _, v := range x_data {
			kvInfo[v] = 0
		}
	}

	for _, v := range list {
		todays := util.GetTimeStr(v.Today, "Y-m-d")
		flows := int64(0)
		if websiteStr != "" {
			if util.InArrayString(v.Address, websiteArr) {
				flowNow, ok := kvInfo[todays]
				if !ok {
					flowNow = 0
				}
				flows = v.Flows + flowNow
				kvInfo[todays] = flows
			}
		} else {
			flowNow, ok := kvInfo[todays]
			if !ok {
				flowNow = 0
			}
			flows = v.Flows + flowNow
			kvInfo[todays] = flows
		}
		cateInfo["Flow"] = append(cateInfo["Flow"], v)
	}

	//fmt.Println(cateInfo)
	//fmt.Println(kvInfo)

	y_data := map[string][]float64{}
	for k, _ := range cateInfo {
		infoData := []float64{}
		for _, vx := range x_data {
			fmt.Println(vx)
			if len(kvInfo) > 0 {
				str := vx
				info, ok := kvInfo[str]
				infoStr := ""
				if !ok {
					info = 0
					infoStr = "0"
				}
				if flow_unit == "GB" {
					infoStr = fmt.Sprintf("%.2f", float64(info)/(float64(1024*1024*1024)))
				} else {
					infoStr = fmt.Sprintf("%.2f", float64(info)/(float64(1024*1024)))
				}
				infoFlow := util.StoF(infoStr)
				infoData = append(infoData, infoFlow)
			} else {
				infoData = append(infoData, 0)
			}
		}
		y_data[k] = infoData
	}

	data := map[string]interface{}{}
	data["x_data"] = x_data
	data["y_data"] = y_data
	data["cate"] = cateName
	data["unit"] = flow_unit
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", data)
	return
}

func GetFlowStatsCopy(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	start_date := c.DefaultPostForm("start_date", "")
	end_date := c.DefaultPostForm("end_date", "")
	websiteStr := strings.TrimSpace(c.DefaultPostForm("url", "")) // 站点地址  多选 ，以逗号分隔
	flow_unit := c.DefaultPostForm("flow_unit", "GB")             // 站点地址
	if flow_unit == "" {
		flow_unit = "GB"
	}

	var start, end int
	x_data := []string{}
	if start_date == "" || end_date == "" {
		today := util.GetTodayTime()
		create := today - 10*86400
		for i := 0; i <= 10; i++ {
			day := create + i*86400
			dayStr := util.GetTimeStr(day, "Y-m-d")
			x_data = append(x_data, dayStr)
		}
		start = create
		end = today
	} else {
		start = util.StoI(util.GetTimeStamp(start_date, "Y-m-d"))
		end = util.StoI(util.GetTimeStamp(end_date, "Y-m-d"))

		for i := 0; i <= (end-start)/86400; i++ {
			day := start + i*86400
			dayStr := util.GetTimeStr(day, "Y-m-d")
			x_data = append(x_data, dayStr)
		}
	}
	websiteArr := strings.Split(websiteStr, ",")
	//list := models.GetUrlUsed(uid, website, start, end)
	list := models.GetUrlUsed(uid, 0, start, end)
	list2 := models.GetUrlWhiteUsed(uid, 0, start, end)

	for _, vv := range list2 {
		list = append(list, vv)
	}

	var cateName []string
	cateName = []string{
		"Flow",
	}

	kvInfo := map[string]int64{}
	cateInfo := map[string][]models.StUrlToday{}
	for _, vv := range cateName {
		cateInfo[vv] = []models.StUrlToday{}
		for _, v := range x_data {
			kvInfo[v] = 0
		}
	}

	for _, v := range list {
		todays := util.GetTimeStr(v.Today, "Y-m-d")
		flows := int64(0)
		if websiteStr != "" {
			if util.InArrayString(v.Url, websiteArr) {
				flowNow, ok := kvInfo[todays]
				if !ok {
					flowNow = 0
				}
				flows = v.Flows + flowNow
				kvInfo[todays] = flows
			}
		} else {
			flowNow, ok := kvInfo[todays]
			if !ok {
				flowNow = 0
			}
			flows = v.Flows + flowNow
			kvInfo[todays] = flows
		}
		cateInfo["Flow"] = append(cateInfo["Flow"], v)
	}

	//fmt.Println(cateInfo)
	//fmt.Println(kvInfo)

	y_data := map[string][]float64{}
	for k, _ := range cateInfo {
		infoData := []float64{}
		for _, vx := range x_data {
			fmt.Println(vx)
			if len(kvInfo) > 0 {
				str := vx
				info, ok := kvInfo[str]
				infoStr := ""
				if !ok {
					info = 0
					infoStr = "0"
				}
				if flow_unit == "GB" {
					infoStr = fmt.Sprintf("%.2f", float64(info)/(float64(1024*1024*1024)))
				} else {
					infoStr = fmt.Sprintf("%.2f", float64(info)/(float64(1024*1024)))
				}
				infoFlow := util.StoF(infoStr)
				infoData = append(infoData, infoFlow)
			} else {
				infoData = append(infoData, 0)
			}
		}
		y_data[k] = infoData
	}

	data := map[string]interface{}{}
	data["x_data"] = x_data
	data["y_data"] = y_data
	data["cate"] = cateName
	data["unit"] = flow_unit
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", data)
	return
}

// 使用记录下载
// @BasePath /api/v1
// @Summary 使用记录下载
// @Description 使用记录下载
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param start_date formData string false "开始日期"
// @Param end_date formData string false "结束日期"
// @Param username formData string false "账户名称"
// @Param flow_unit formData string false "使用单位"
// @Param point formData string false "小数点"
// @Param flow_type formData string false "流量类型 1：正常流量 2：无限流量"
// @Produce octet-stream
// @Success 0 {object} string "csv文件"
// @Router /web/account/used_download [post]
func FlowStatsDownload(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	start_date := c.DefaultPostForm("start_date", "")
	end_date := c.DefaultPostForm("end_date", "")
	username := strings.TrimSpace(c.DefaultPostForm("username", "")) // 账户名称
	flow_unit := c.DefaultPostForm("flow_unit", "GB")                // 使用单位
	pointStr := c.DefaultPostForm("point", "0")                      // 小数点
	country := c.DefaultPostForm("country", "")                      // 国家
	siteUrl := c.DefaultPostForm("site_url", "")                     // 访问地址
	flowUseType := c.DefaultPostForm("flow_use_type", "")            // 流量使用类型，仅flowtype =1时有效
	if flow_unit == "" {
		flow_unit = "GB"
	}
	flow_type := strings.TrimSpace(c.DefaultPostForm("flow_type", "1"))
	if flow_type == "" {
		flow_type = "1"
	}
	flowType := util.StoI(flow_type)
	if flowType == 0 {
		flowType = 1
	}

	if pointStr == "" {
		pointStr = "0"
	}
	point := util.StoI(pointStr)

	var start, end int
	if start_date == "" || end_date == "" {
		today := util.GetTodayTime()
		create := today - 10*86400
		start = create
		end = today
	} else {
		start = util.StoI(util.GetTimeStamp(start_date, "Y-m-d"))
		end = util.StoI(util.GetTimeStamp(end_date, "Y-m-d"))
	}
	accountId := 0
	if username != "" {
		err, userAccount := models.GetUserAccount(uid, username)
		if err == nil && userAccount.Id > 0 {
			accountId = userAccount.Id
		}
	}

	flowChar := int64(1024 * 1024 * 1024)
	if flow_unit == "GB" {
		flowChar = 1024 * 1024 * 1024
	} else if flow_unit == "MB" {
		flowChar = 1024 * 1024
	} else if flow_unit == "KB" {
		flowChar = 1024
	} else if flow_unit == "Bytes" {
		flowChar = 1
	}

	title := []string{"Date", "Website", "Traffic"}

	if uid > 0 {
		csvData := [][]string{}
		csvData = append(csvData, title)
		lists := []models.StAddressToday{}

		if flowType == 1 {
			//list = models.GetUrlUsed(uid, 0, start, end)
			lists = models.GetFlowUsedStatDown(uid, accountId, start, end, country, siteUrl, flowUseType)
		} else if flowType == 2 {
			//list = models.GetUnlimitedUrlUsed(uid, 0, start, end)
			lists = models.GetFlowDayUsedStat(uid, 0, start, end)
		} else if flowType == 3 {
			//list = models.GetDynamicUrlUsed(uid, 0, start, end)
			lists = models.GetIspFlowUsedStatDown(uid, accountId, start, end, country, siteUrl)
		}
		for _, v := range lists {
			info := []string{}
			flowsStr := util.FtoS2(math.Round(float64(v.Flows)/float64(flowChar)), point)

			info = append(info, util.GetTimeStr(v.Today, "d-m-Y"))
			info = append(info, v.Address)
			info = append(info, flowsStr+" "+flow_unit)
			csvData = append(csvData, info)
		}

		err := DownloadCsv(c, "UsedRecord", csvData)
		fmt.Println(err)
	}
	return
}

// 获取统计域名
// @BasePath /api/v1
// @Summary 获取统计域名
// @Description 获取统计域名
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param url formData string true "域名"
// @Param flow_type formData string true "流量类型 1-流量 2-不限量 3-动态ISP"
// @Produce json
// @Success 0 {object} []models.StUrlLists{}
// @Router /web/account/url_list [post]
func GetUrlStats(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	url := strings.TrimSpace(c.DefaultPostForm("url", ""))

	uid := user.Id
	var start int
	today := util.GetTodayTime()
	start = today - 10*86400
	//list := models.GetUrlList(uid, start, url)

	flow_type := strings.TrimSpace(c.DefaultPostForm("flow_type", "1"))
	if flow_type == "" {
		flow_type = "1"
	}
	flowType := util.StoI(flow_type)
	if flowType == 0 {
		flowType = 1
	}
	lists := []models.StUrlLists{}
	if flowType == 1 {
		lists = models.GetUrlListStats(uid, start, url)
	} else if flowType == 2 {

	} else if flowType == 3 {
		lists = models.GetIspUrlListStats(uid, start, today+86400, url)
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", lists)
	return
}

// 获取长效Isp统计域名
// @BasePath /api/v1
// @Summary 获取统计域名
// @Description 获取统计域名
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param url formData string true "域名"
// @Param flow_type formData string true "流量类型 1-流量 2-不限量 3-动态ISP"
// @Produce json
// @Success 0 {object} []models.StUrlLists{}
// @Router /web/account/long_isp_url_list [post]
func GetLongIspUrlStats(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	startDate := c.DefaultPostForm("start_date", "")
	endDate := c.DefaultPostForm("end_date", "")
	var start, end int
	if startDate == "" || endDate == "" {
		today := util.GetTodayTime()
		create := today - 10*86400
		start = create
		end = today
	} else {
		start = util.StoI(util.GetTimeStamp(startDate, "Y-m-d"))
		end = util.StoI(util.GetTimeStamp(endDate, "Y-m-d"))
	}
	url := strings.TrimSpace(c.DefaultPostForm("url", ""))

	uid := user.Id
	list := models.GetIspUrlListStats(uid, start, end, url)
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", list)
	return
}
func GetIpCharData(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	start_date := c.DefaultPostForm("start_date", "")
	end_date := c.DefaultPostForm("end_date", "")

	var start, end int
	x_data := []string{}
	monthArr := []string{}
	if start_date == "" || end_date == "" {
		today := util.GetTodayTime()
		create := today - 10*86400
		for i := 0; i <= 10; i++ {
			day := create + i*86400
			dayStr := util.GetTimeStr(day, "Y-m-d")
			x_data = append(x_data, dayStr)
			date, _ := time.Parse("2006-01-02", dayStr)
			month := date.Format("200601")
			if !util.InArrayString(month, monthArr) {
				monthArr = append(monthArr, month)
			}
		}
		start = create
		end = today
	} else {
		start = util.StoI(util.GetTimeStamp(start_date, "Y-m-d"))
		end = util.StoI(util.GetTimeStamp(end_date, "Y-m-d"))

		for i := 0; i <= (end-start)/86400; i++ {
			day := start + i*86400
			dayStr := util.GetTimeStr(day, "Y-m-d")
			x_data = append(x_data, dayStr)
			date, _ := time.Parse("2006-01-02", dayStr)
			month := date.Format("200601")
			if !util.InArrayString(month, monthArr) {
				monthArr = append(monthArr, month)
			}
		}
	}
	list := models.GetIpCountByUidAndTime(uid, start, end, monthArr)
	fmt.Println(list)
	infoData := []float64{}
	for _, v := range x_data {
		res := true
		for _, vx := range list {
			if vx.Today == v {
				infoData = append(infoData, float64(vx.Num))
				res = false
			}
		}
		if res {
			infoData = append(infoData, 0)
		}
	}

	data := map[string]interface{}{}
	data["x_data"] = x_data
	data["y_data"] = infoData
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", data)
	return
}

func GetIpCharDataDownload(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	startDate := c.DefaultPostForm("start_date", "")
	endDate := c.DefaultPostForm("end_date", "")

	var start, end int
	xData := []string{}
	monthArr := []string{}
	if startDate == "" || endDate == "" {
		today := util.GetTodayTime()
		create := today - 10*86400
		for i := 0; i <= 10; i++ {
			day := create + i*86400
			dayStr := util.GetTimeStr(day, "Y-m-d")
			xData = append(xData, dayStr)
			date, _ := time.Parse("2006-01-02", dayStr)
			month := date.Format("200601")
			if !util.InArrayString(month, monthArr) {
				monthArr = append(monthArr, month)
			}
		}
		start = create
		end = today
	} else {
		start = util.StoI(util.GetTimeStamp(startDate, "Y-m-d"))
		end = util.StoI(util.GetTimeStamp(endDate, "Y-m-d"))

		for i := 0; i <= (end-start)/86400; i++ {
			day := start + i*86400
			dayStr := util.GetTimeStr(day, "Y-m-d")
			xData = append(xData, dayStr)
			date, _ := time.Parse("2006-01-02", dayStr)
			month := date.Format("200601")
			if !util.InArrayString(month, monthArr) {
				monthArr = append(monthArr, month)
			}
		}
	}
	list := models.GetTimeAndIpByUidAndDate(uid, start, end, monthArr)
	title := []string{"ExtractTime", "ExtractIP"}

	if uid > 0 {
		csvData := [][]string{}
		csvData = append(csvData, title)
		for _, v := range list {
			info := []string{}
			info = append(info, v.ExtractTime)
			info = append(info, v.IP)
			csvData = append(csvData, info)
		}
		err := DownloadCsv(c, "UsedRecord", csvData)
		fmt.Println(err)
	}
	return
}
