package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
	"strings"
)

// 获取所有账户列表
// @BasePath /api/v1
// @Summary 获取所有账户列表
// @Description 获取所有账户列表
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param username formData string true "用户名"
// @Param status formData string true "状态"
// @Produce json
// @Success 0 {object} map[string]interface{} "total:总数,enabled:启用,disabled:禁用,warning:流量警告,lists:列表（值为[]models.ResUserAccount{}模型）"
// @Router /web/account/all_lists [post]
func GetUserAccountAllList(c *gin.Context) {
	username := strings.TrimSpace(c.DefaultPostForm("username", "")) // 用户名
	statusStr := c.DefaultPostForm("status", "")
	startDate := c.DefaultPostForm("start_date", "")
	endDate := c.DefaultPostForm("end_date", "")
	if statusStr == "" {
		statusStr = "10"
	}
	status := com.StrTo(statusStr).MustInt()

	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id

	start := 0
	end := 0
	if startDate != "" {
		start = util.StoI(util.GetTimeStamp(startDate, "Y-m-d"))
	}
	if endDate != "" {
		end = util.StoI(util.GetTimeStamp(endDate, "Y-m-d")) + 86399
	}

	_, accountLists := models.GetUserAccountAllList(uid, username, start, end)

	total := len(accountLists)
	enabled := 0
	disabled := 0
	warning := 0

	data := []models.ResUserAccount{}
	for _, v := range accountLists {
		flowChar := int64(1024 * 1024 * 1024)
		if v.FlowUnit == "MB" {
			flowChar = 1024 * 1024
		}
		limitFlowB := v.LimitFlow
		useFlowB := limitFlowB - v.Flows
		if useFlowB < 0 {
			useFlowB = limitFlowB
		}

		useFlow := fmt.Sprintf("%.2f", float64(useFlowB)/float64(flowChar)) //已使用列表
		limitFlow := int(limitFlowB / flowChar)
		if v.Status == 1 {
			enabled = enabled + 1
		} else {
			disabled = disabled + 1
		}
		if v.Flows == 0 {
			warning = warning + 1
		}
		percent := fmt.Sprintf("%.2f", float64(useFlowB*100)/float64(limitFlowB))

		info := models.ResUserAccount{}
		info.Id = v.Id
		info.Account = v.Account
		info.Password = v.Password
		info.LimitFlow = util.ItoS(limitFlow)
		info.UseFlow = useFlow + v.FlowUnit
		info.FlowUnit = v.FlowUnit
		info.Flows = v.Flows
		info.Master = v.Master
		info.Status = v.Status
		info.Remark = v.Remark
		info.Percent = percent
		info.CreateTime = util.GetTimeStr(v.CreateTime, "d-m-Y")
		if status == 10 {
			data = append(data, info)
		} else {
			if status == v.Status {
				data = append(data, info)
			}
		}
	}
	resData := map[string]interface{}{}
	resData["total"] = total
	resData["enabled"] = enabled
	resData["disabled"] = disabled
	resData["warning"] = warning
	resData["lists"] = data
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resData)
	return
}

// 获取所有账户列表
// @BasePath /api/v1
// @Summary 获取所有账户列表
// @Description 获取所有账户列表
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param username formData string true "用户名"
// @Param status formData string true "状态"
// @Produce json
// @Success 0 {object} map[string]interface{} "total:总数,enabled:启用,disabled:禁用,warning:流量警告,lists:列表（值为[]models.ResUserAccount{}模型）"
// @Router /web/account/all_lists [post]
func GetUserAccountAllListDownload(c *gin.Context) {
	username := strings.TrimSpace(c.DefaultPostForm("username", "")) // 用户名
	statusStr := c.DefaultPostForm("status", "")
	startDate := c.DefaultPostForm("start_date", "")
	endDate := c.DefaultPostForm("end_date", "")
	if statusStr == "" {
		statusStr = "10"
	}
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id

	start := 0
	end := 0
	if startDate != "" {
		start = util.StoI(util.GetTimeStamp(startDate, "Y-m-d"))
	}
	if endDate != "" {
		end = util.StoI(util.GetTimeStamp(endDate, "Y-m-d")) + 86399
	}
	_, accountLists := models.GetUserAccountAllList(uid, username, start, end)
	csvData := [][]string{}
	title := []string{"Username", "Password", "Traffic Used", "Traffic Limit", "Remark", "Create Time"}
	csvData = append(csvData, title)
	for _, v := range accountLists {
		flowChar := int64(1024 * 1024 * 1024)
		if v.FlowUnit == "MB" {
			flowChar = 1024 * 1024
		}
		limitFlowB := v.LimitFlow
		useFlowB := limitFlowB - v.Flows
		if useFlowB < 0 {
			useFlowB = limitFlowB
		}
		useFlow := fmt.Sprintf("%.2f", float64(useFlowB)/float64(flowChar)) //已使用列表
		limitFlow := int(limitFlowB / flowChar)
		info := []string{}
		info = append(info, v.Account)
		info = append(info, v.Password)
		info = append(info, useFlow+v.FlowUnit)
		info = append(info, util.ItoS(limitFlow)+v.FlowUnit)
		info = append(info, v.Remark)
		info = append(info, util.GetTimeStr(v.CreateTime, "d-m-Y"))
		csvData = append(csvData, info)
	}
	err := DownloadCsv(c, "Account Information", csvData)
	fmt.Println(err)
	return
}

// 获取所有当前可用账户列表
// @BasePath /api/v1
// @Summary 获取所有账户列表
// @Description 获取所有账户列表
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} map[string]interface{} "
// @Router /web/account/lists_available [post]
func GetUserAccountListAvailable(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id
	_, accountLists := models.GetUserAvailableAccount(uid)

	data := []models.UserAccountPass{}
	for _, v := range accountLists {
		info := models.UserAccountPass{}
		if v.Master == 1 {
			flowsInfo := models.GetUserFlowInfo(v.Uid)
			info.Flows = flowsInfo.Flows
		} else {
			info.Flows = v.Flows
		}
		info.AccountId = v.Id
		info.Account = v.Account
		info.Password = v.Password
		data = append(data, info)
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", data)
	return
}

// 获取子账户列表
// @BasePath /api/v1
// @Summary 获取子账户列表
// @Description 获取子账户列表
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param username formData string true "用户名"
// @Param status formData string true "状态"
// @Param flow_type formData string true "类型  1流量  3 动态ISP"
// @Produce json
// @Success 0 {object} map[string]interface{} "total:总数,enabled:启用,disabled:禁用,warning:流量警告,lists:列表（值为[]models.ResUserAccount{}模型）"
// @Router /web/account/lists [post]
func GetUserAccountList(c *gin.Context) {
	username := strings.TrimSpace(c.DefaultPostForm("username", "")) // 用户名
	status := com.StrTo(c.DefaultPostForm("status", "10")).MustInt()
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id

	var accountLists []models.UserAccount
	_, accountLists = models.GetUserAccountList(uid, username)

	total := len(accountLists)
	enabled := 0
	disabled := 0
	warning := 0

	var data []models.ResUserAccount
	for _, v := range accountLists {
		flowChar := int64(1024 * 1024 * 1024)
		if v.FlowUnit == "MB" {
			flowChar = 1024 * 1024
		}
		limitFlowB := v.LimitFlow
		useFlowB := limitFlowB - v.Flows
		if useFlowB < 0 {
			useFlowB = limitFlowB
		}

		useFlow := fmt.Sprintf("%.2f", float64(useFlowB)/float64(flowChar)) //已使用列表
		limitFlow := int(limitFlowB / flowChar)
		if v.Status == 1 {
			enabled = enabled + 1
		} else {
			disabled = disabled + 1
		}
		if v.Flows == 0 {
			warning = warning + 1
		}
		percent := fmt.Sprintf("%.2f", float64(useFlowB*100)/float64(limitFlowB))

		info := models.ResUserAccount{}
		info.Id = v.Id
		info.Account = v.Account
		info.Password = v.Password
		info.LimitFlow = util.ItoS(limitFlow)
		info.UseFlow = useFlow + v.FlowUnit
		info.FlowUnit = v.FlowUnit
		info.Flows = v.Flows
		info.Master = v.Master
		info.Status = v.Status
		info.Remark = v.Remark
		info.Percent = percent
		info.CreateTime = util.GetTimeStr(v.CreateTime, "d-m-Y")
		if status == 10 {
			data = append(data, info)
		} else {
			if status == v.Status {
				data = append(data, info)
			}
		}
	}
	resData := map[string]interface{}{}
	resData["total"] = total
	resData["enabled"] = enabled
	resData["disabled"] = disabled
	resData["warning"] = warning
	resData["lists"] = data
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resData)
	return
}

// 添加/编辑 流量账号子账户
// @BasePath /api/v1
// @Summary 添加/编辑 流量账号子账户
// @Description 添加/编辑 流量账号子账户
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param account_id formData string true "帐密ID"
// @Param username formData string true "用户名"
// @Param password formData string true "密码"
// @Param remark formData string true "备注"
// @Param flow formData string true "流量限制"
// @Param flow_unit formData string true "流量单位"
// @Param status formData string true "状态"
// @Param flow_type formData string true "类型  1流量  3 动态ISP"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/account/add_edit [post]
func AddUserFlowAccount(c *gin.Context) {
	accountId := com.StrTo(c.DefaultPostForm("account_id", "0")).MustInt() // 帐密ID
	username := strings.TrimSpace(c.DefaultPostForm("username", ""))       // 用户名
	password := strings.TrimSpace(c.DefaultPostForm("password", ""))       // 密码
	remark := strings.TrimSpace(c.DefaultPostForm("remark", ""))           // 备注
	flowStr := strings.TrimSpace(c.DefaultPostForm("flow", ""))            // 流量限制
	flowUnit := strings.TrimSpace(c.DefaultPostForm("flow_unit", "GB"))    // 流量单位
	status := com.StrTo(c.DefaultPostForm("status", "1")).MustInt()

	if status != 1 && status != 0 {
		JsonReturn(c, e.ERROR, "__T_STATUS_ERROR", nil)
		return

	}

	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	ip := c.ClientIP()
	uid := userInfo.Id
	nowTime := util.GetNowInt()

	if username == "" {
		JsonReturn(c, e.ERROR, "__T_USERNAME_ERROR", nil)
		return
	}
	if password == "" {
		JsonReturn(c, e.ERROR, "__T_PASSWORD_ERROR", nil)
		return
	}
	//账户密码不能一样
	if username == password {
		JsonReturn(c, e.ERROR, "__T_USERNAME_PASSWORD_SAME", nil)
		return
	}
	if flowStr == "" {
		JsonReturn(c, e.ERROR, "__T_FLOW_NUMBER_ERROR", nil)
		return
	}
	limitFlows := int64(util.StoI(flowStr))

	if flowUnit == "" {
		flowUnit = "GB"
	}
	if !util.CheckUserAccount(username) {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_USERNAME_ERROR", nil)
		return
	}
	if !util.CheckUserPassword(password) {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_PASSWORD_ERROR", nil)
		return
	}
	accountInfo := models.UserAccount{}
	if accountId > 0 {
		accountInfo, _ = models.GetUserAccountById(accountId)
		if limitFlows <= 0 && accountInfo.Master != 1 { //主账号可以不分配流量 20250319
			JsonReturn(c, e.ERROR, "__T_FLOW_NUMBER_ERROR", nil)
			return
		}
		if accountInfo.Uid != uid {
			JsonReturn(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
			return
		}
		var hasAccount models.UserAccount
		_, hasAccount = models.GetUserAccountNeqId(accountId, username)

		if hasAccount.Id != 0 {
			JsonReturn(c, e.ERROR, "__T_ACCOUNT_USERNAME_EXIST", nil)
			return
		}
		// 不能使用历史密码
		if accountInfo.Password != password {
			historyPasswordArr := models.GetHistoryPasswordArr(uid)
			fmt.Println("historyPasswordArr", historyPasswordArr)
			if util.InArray(password, historyPasswordArr) {
				JsonReturn(c, e.ERROR, "__T_PASSWORD_HISTORY_ERROR", nil)
				return
			}
		}
	} else {
		if limitFlows <= 0 {
			JsonReturn(c, e.ERROR, "__T_FLOW_NUMBER_ERROR", nil)
			return
		}
		_, hasAccount := models.GetUserAccount(0, username)
		if hasAccount.Id != 0 {
			JsonReturn(c, e.ERROR, "__T_ACCOUNT_USERNAME_EXIST", nil)
			return
		}
		// 不能使用历史密码
		historyPasswordArr := models.GetHistoryPasswordArr(uid)
		if util.InArray(password, historyPasswordArr) {
			JsonReturn(c, e.ERROR, "__T_PASSWORD_HISTORY_ERROR", nil)
			return
		}
	}
	flowInfo := models.GetUserFlowInfo(uid)
	if flowInfo.ID == 0 {
		JsonReturn(c, -1, "__T_NO_FLOW_INFO", gin.H{})
		return
	}
	if flowInfo.ExpireTime < nowTime {
		JsonReturn(c, -1, "__T_FLOW_EXPIRED", gin.H{})
		return
	}
	flowChar := int64(1024 * 1024 * 1024)
	if flowUnit == "MB" {
		flowChar = 1024 * 1024
	}

	limitFlows = limitFlows * flowChar
	if accountId > 0 {
		if limitFlows != accountInfo.LimitFlow {
			//改小了
			if accountInfo.LimitFlow > limitFlows {
				// 子账号使用流量不得大于新分配流量值
				if (accountInfo.LimitFlow - accountInfo.Flows) > limitFlows {
					JsonReturn(c, -1, "_T_USED_FLOW_MORE_THAN_LIMIT", gin.H{}) //用户流量不足
					return
				}
			} else {
				// 如果主账号剩余流量小于新分配的增值
				if flowInfo.Flows < (limitFlows - accountInfo.LimitFlow) {
					JsonReturn(c, -1, "__T_FLOW_NO_ENOUGH", gin.H{}) //用户流量不足
					return
				}
			}
		}
	} else {
		if flowInfo.Flows < limitFlows {
			JsonReturn(c, -1, "__T_FLOW_NO_ENOUGH", gin.H{}) //用户流量不足
			return
		}
	}

	kcFlow := limitFlows // 扣除的主账号流量    // 增加的子账户流量
	flows := limitFlows  // 流量余额
	//异步处理信息
	dealInfo := models.PushAccount{}
	// 查询该用户下面是否存在相同账户 若存在则修改账户信息
	if accountId > 0 {
		if limitFlows != accountInfo.LimitFlow {
			kcFlow = limitFlows - accountInfo.LimitFlow                                                                                                                    //流量差额
			flows = accountInfo.Flows + kcFlow                                                                                                                             //操作后流量
			err := models.CreateLogFlowsAccount(uid, accountInfo.Id, accountInfo.Flows, accountInfo.LimitFlow, flows, limitFlows, accountInfo.Account, ip, "edit", "user") // 加个 流量变动日志  存 变动前的数据 剩余 ，和 配置额度   和变动后的配置额度    时间，IP
			fmt.Println(err)
		} else {
			kcFlow = 0
		}

		upMap := map[string]interface{}{}
		upMap["account"] = username
		upMap["flow_unit"] = flowUnit
		upMap["password"] = password
		upMap["status"] = status
		upMap["remark"] = remark
		upMap["expire_time"] = flowInfo.ExpireTime

		models.UpdateUserAccountById(accountInfo.Id, upMap) //更新子账号信息

		// 异步处理流量
		dealInfo.Cate = "edit"
		dealInfo.Uid = uid
		dealInfo.AccountId = accountInfo.Id
		dealInfo.Flows = kcFlow //待操作的数据信息
		dealInfo.LimitFlow = limitFlows
		dealInfo.FlowUnit = flowUnit
		dealInfo.Ip = ip
		dealInfo.ExpireTime = flowInfo.ExpireTime
		dealInfo.CreateTime = nowTime
		//写入历史密码
		if accountInfo.Password != password {
			models.AddHistoryPassword(uid, accountInfo.Password, accountInfo.Id)
		}
	} else {
		data := models.UserAccount{}
		data.Status = 1
		data.Remark = remark
		data.Uid = uid
		data.Account = username
		data.Password = password
		data.Master = 0
		data.Status = status
		data.FlowUnit = flowUnit
		data.CreateTime = nowTime
		data.ExpireTime = flowInfo.ExpireTime

		err, accId := models.AddProxyAccount(data)
		if err != nil {
			// 账户添加失败：发送产品侧预警（规则驱动模板+回退），并返回失败
			runtime := map[string]any{
				"username":  userInfo.Username,
				"childUser": username,
				"error":     err.Error(),
			}
			fallbackTpl := fmt.Sprintf("预警：【cherry】用户【%s】事件【认证账户添加】状态【失败】 信息：添加子账户失败，子账号：%s，错误：%s", userInfo.Username, username, err.Error())
			SendProductAlertWithRule("child_add_failed", runtime, fallbackTpl)
			JsonReturn(c, e.ERROR, "__T_ACCOUNT_USERNAME_ERROR", nil)
			return
		}
		fmt.Println(err, accId)
		dealInfo.AccountId = accId

		// 异步处理流量
		dealInfo.Cate = "add"
		dealInfo.Uid = uid
		dealInfo.Flows = flows
		dealInfo.LimitFlow = flows
		dealInfo.FlowUnit = flowUnit
		dealInfo.Ip = ip
		dealInfo.ExpireTime = flowInfo.ExpireTime
		dealInfo.CreateTime = nowTime
	}
	listStr, _ := json.Marshal(dealInfo)

	resP := models.RedisLPUSH("list_account_flow", string(listStr))
	fmt.Println(resP)

	JsonReturn(c, e.SUCCESS, "__T_EDIT_SUCCESS", nil)
	return
}

// 添加/编辑 流量账号子账户
// @BasePath /api/v1
// @Summary 修改账号名称及密码
// @Description 修改账号名称及密码
// @Tags 个人中心 - 修改账号名称及密码
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param account_id formData string true "帐密ID"
// @Param username formData string true "用户名"
// @Param password formData string true "密码"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/account/set_pass [post]
func SetUserAccountPass(c *gin.Context) {
	accountId := com.StrTo(c.DefaultPostForm("account_id", "0")).MustInt() // 帐密ID
	username := strings.TrimSpace(c.DefaultPostForm("username", ""))       // 用户名
	password := strings.TrimSpace(c.DefaultPostForm("password", ""))       // 密码

	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id
	if username == "" {
		JsonReturn(c, e.ERROR, "__T_USERNAME_ERROR", nil)
		return
	}
	if password == "" {
		JsonReturn(c, e.ERROR, "__T_PASSWORD_ERROR", nil)
		return
	}
	//账户密码不能一样
	if username == password {
		JsonReturn(c, e.ERROR, "__T_USERNAME_PASSWORD_SAME", nil)
		return
	}
	if !util.CheckUserAccount(username) {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_USERNAME_ERROR", nil)
		return
	}
	if !util.CheckUserPassword(password) {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_PASSWORD_ERROR", nil)
		return
	}
	accountInfo := models.UserAccount{}
	if accountId > 0 {
		accountInfo, _ = models.GetUserAccountById(accountId)
		if accountInfo.Uid != uid {
			JsonReturn(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
			return
		}
		var hasAccount models.UserAccount
		_, hasAccount = models.GetUserAccountNeqId(accountId, username)

		if hasAccount.Id != 0 {
			JsonReturn(c, e.ERROR, "__T_ACCOUNT_USERNAME_EXIST", nil)
			return
		}
		// 不能使用历史密码
		if accountInfo.Password != password {
			historyPasswordArr := models.GetHistoryPasswordArr(uid)
			fmt.Println("historyPasswordArr", historyPasswordArr)
			if util.InArray(password, historyPasswordArr) {
				JsonReturn(c, e.ERROR, "__T_PASSWORD_HISTORY_ERROR", nil)
				return
			}
		}
		upMap := map[string]interface{}{}
		upMap["account"] = username
		upMap["password"] = password

		models.UpdateUserAccountById(accountInfo.Id, upMap) //更新子账号信息
		//写入历史密码
		if accountInfo.Password != password {
			models.AddHistoryPassword(uid, accountInfo.Password, accountInfo.Id)
		}
		JsonReturn(c, e.SUCCESS, "__T_EDIT_SUCCESS", nil)
		return
	}
	JsonReturn(c, e.ERROR, "__T_FAIL", nil)
	return
}

// 账号信息
// @BasePath /api/v1
// @Summary 账号信息
// @Description 账号信息
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param account_id formData string true "帐密ID"
// @Produce json
// @Success 0 {object} map[string]interface{} "account_id:帐密ID,account:用户名,password:密码,limit_flow:流量限制,flow:已用流量,flow_unit:流量单位,is_use:能否使用生成命令"
// @Router /web/account/detail [post]
func UserFlowAccountDetail(c *gin.Context) {
	session := c.DefaultPostForm("session", "")                            // session
	accountId := com.StrTo(c.DefaultPostForm("account_id", "0")).MustInt() // 帐密ID
	if session == "" {
		JsonReturn(c, -1, "__T_SESSION_ERROR", gin.H{})
		return
	}
	errs, uid := GetUIDbySession(session)
	if errs == false || uid == 0 {
		JsonReturn(c, -1, "__T_SESSION_EXPIRE", gin.H{})
		return
	}
	if accountId == 0 {
		JsonReturn(c, -1, "__T_PARAM_ERROR--Account", gin.H{})
		return
	}

	accountInfo, _ := models.GetUserAccountById(accountId)

	if accountInfo.Id == 0 || accountInfo.Uid != uid {
		JsonReturn(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
		return
	}
	isUse := 0
	nowTime := util.GetNowInt()
	flowInfo := models.GetUserFlowInfo(uid)
	if accountInfo.Master == 1 {
		if flowInfo.ExpireTime > nowTime && flowInfo.Flows > 0 {
			isUse = 1
		}
	} else {
		if accountInfo.Flows > 0 && accountInfo.ExpireTime > nowTime {
			isUse = 1
		}
	}

	data := map[string]interface{}{}
	data["account_id"] = accountId
	data["account"] = accountInfo.Account
	data["password"] = accountInfo.Password
	data["limit_flow"] = accountInfo.LimitFlow
	data["flow"] = accountInfo.Flows
	data["flow_unit"] = accountInfo.FlowUnit
	data["is_use"] = isUse //能否使用生成命令
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", data)
	return
}

// 删除 流量子账号
// @BasePath /api/v1
// @Summary 删除 流量子账号
// @Description 删除 流量子账号
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param account_id formData string true "子账户ID"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/account/del [post]
func DelUserAccount(c *gin.Context) {
	session := c.DefaultPostForm("session", "")                            // session
	accountId := com.StrTo(c.DefaultPostForm("account_id", "0")).MustInt() // 帐密ID
	if session == "" {
		JsonReturn(c, -1, "__T_SESSION_ERROR", gin.H{})
		return
	}

	errs, uid := GetUIDbySession(session)
	if errs == false || uid == 0 {
		JsonReturn(c, -1, "__T_SESSION_EXPIRE", gin.H{})
		return
	}
	if accountId == 0 {
		JsonReturn(c, -1, "__T_PARAM_ERROR--Account Info", gin.H{})
		return
	}
	accountInfo, _ := models.GetUserAccountById(accountId)
	if accountInfo.Uid != uid {
		JsonReturn(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
		return
	}
	if accountInfo.Master != 0 {
		JsonReturn(c, -1, "__T_MASTER_ACCOUNT_NO", gin.H{})
		return
	}

	dealInfo := models.PushAccount{}
	// 异步处理流量
	dealInfo.Cate = "del"
	dealInfo.Uid = uid
	dealInfo.AccountId = accountInfo.Id
	dealInfo.Flows = accountInfo.Flows
	dealInfo.LimitFlow = accountInfo.LimitFlow
	dealInfo.FlowUnit = accountInfo.FlowUnit
	dealInfo.ExpireTime = accountInfo.ExpireTime
	dealInfo.CreateTime = util.GetNowInt()
	listStr, _ := json.Marshal(dealInfo)
	resP := models.RedisLPUSH("list_account_flow", string(listStr))
	fmt.Println(resP)

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return
}

// 账号启用  / 禁用
// @BasePath /api/v1
// @Summary 账号启用  / 禁用
// @Description 账号启用  / 禁用
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param account_id formData string true "子账户ID"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/account/enable_disable [post]
func AccountEnableOrDisable(c *gin.Context) {
	session := c.DefaultPostForm("session", "")                            // session
	accountId := com.StrTo(c.DefaultPostForm("account_id", "0")).MustInt() // 帐密ID
	if session == "" {
		JsonReturn(c, -1, "__T_SESSION_ERROR", gin.H{})
		return
	}

	errs, uid := GetUIDbySession(session)
	if errs == false || uid == 0 {
		JsonReturn(c, -1, "__T_SESSION_EXPIRE", gin.H{})
		return
	}
	if accountId == 0 {
		JsonReturn(c, -1, "__T_PARAM_ERROR--Account Info", gin.H{})
		return
	}

	accountInfo, _ := models.GetUserAccountById(accountId)
	if accountInfo.Uid != uid {
		JsonReturn(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
		return
	}
	status := accountInfo.Status
	if status == 1 {
		status = 0
	} else if status == 0 {
		status = 1
	}
	upMap := map[string]interface{}{}
	if accountInfo.Master != 0 {
		JsonReturn(c, -1, "__T_MASTER_ACCOUNT_NO", gin.H{})
		return
	}
	upMap["status"] = status
	upMap["update_time"] = util.GetNowInt()

	b := models.UpdateUserAccountById(accountInfo.Id, upMap)

	if !b {
		JsonReturn(c, e.ERROR, "__T_FAIL", nil)
		return
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return
}
