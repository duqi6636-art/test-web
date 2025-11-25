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
	"time"
)

// 获取账号主账户
// @BasePath /api/v1
// @Summary 获取账号主账户
// @Description 获取账号主账户
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} map[string]interface{} "account：账号；password：密码"
// @Router /web/user/long_isp/get_main_user_account [post]
func GetLongIspMainAccount(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	data := models.LongIspUserAccount{}
	data.Status = 1
	data.Remark = ""
	// 查询该用户下面是否存在相同账户 若存在则修改账户信息
	accountInfo := models.LongIspUserAccount{}
	_, accountInfo = models.GetLongIspUserAccountMaster(userInfo.Id)

	account := ""
	password := ""
	if accountInfo.Id == 0 {
		//判断是否有账号和该主账号重复，有的就随机加个字母
		_, HasAccount := models.GetLongIspUserAccountNeqId(0, userInfo.Username)
		if HasAccount.Id > 0 {
			account = userInfo.Username + util.RandStr("s", 2)
		} else {
			account = userInfo.Username
		}

		account = userInfo.Username
		password = util.RandStr("r", 8)
		data.Uid = userInfo.Id
		data.Account = account
		data.Password = password
		data.Master = 1
		data.FlowUnit = "GB"
		data.CreateTime = int(time.Now().Unix())
		err, id := models.AddLongIspProxyAccount(data)
		fmt.Println("err_id", err, id)

	} else {
		account = accountInfo.Account
		password = accountInfo.Password
	}
	resMap := map[string]interface{}{}
	resMap["account"] = account
	resMap["password"] = password
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resMap)
	return
}

// 修改账号主账户用户名密码
// @BasePath /api/v1
// @Summary 修改账号主账户用户名密码
// @Description 修改账号主账户用户名密码
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param username formData string true "账户"
// @Param password formData string true "密码"
// @Produce json
// @Success 0 {object} map[string]interface{} "account：账号；password：密码"
// @Router /web/user/long_isp/set_main_pass [post]
func SetLongIspMainAccount(c *gin.Context) {
	session := c.DefaultPostForm("session", "")                      // session
	username := strings.TrimSpace(c.DefaultPostForm("username", "")) // 账户
	password := strings.TrimSpace(c.DefaultPostForm("password", "")) // 密码

	if session == "" {
		JsonReturn(c, -1, "__T_SESSION_ERROR", gin.H{})
		return
	}

	if password == "" {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_PASSWORD_ERROR", nil)
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
	//账户密码不能一样
	if username == password {
		JsonReturn(c, e.ERROR, "__T_USERNAME_PASSWORD_SAME", nil)
		return
	}
	errs, uid := GetUIDbySession(session)
	if errs == false {
		JsonReturn(c, -1, "__T_SESSION_EXPIRE", gin.H{})
		return
	}
	err, userInfo := models.GetUserById(uid)
	if err != nil || userInfo.Id == 0 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_ERROR", nil)
		return
	}

	// 查询该用户下面是否存在账户 若存在则修改账户信息
	accountInfo := models.LongIspUserAccount{}
	_, accountInfo = models.GetLongIspUserAccountMaster(userInfo.Id)

	if accountInfo.Id > 0 {

		_, hasAccount := models.GetLongIspUserAccountNeqId(accountInfo.Id, username)
		if hasAccount.Id != 0 {
			JsonReturn(c, e.ERROR, "__T_ACCOUNT_USERNAME_EXIST", nil)
			return
		}
		//写入历史密码
		if accountInfo.Password != password {
			// 不能使用历史密码
			historyPasswordArr := models.GetHistoryPasswordArr(uid)
			fmt.Println("historyPasswordArr", historyPasswordArr)
			if util.InArray(password, historyPasswordArr) {
				JsonReturn(c, e.ERROR, "__T_PASSWORD_HISTORY_ERROR", nil)
				return
			}
			models.AddHistoryPassword(uid, accountInfo.Password, accountInfo.Id)
		}

		data := map[string]interface{}{}
		data["update_time"] = int(time.Now().Unix())
		data["account"] = username
		data["password"] = password

		errs = models.UpdateLongIspUserAccountById(accountInfo.Id, data)
	}

	resMap := map[string]interface{}{}
	resMap["account"] = username
	resMap["password"] = password
	JsonReturn(c, e.SUCCESS, "__T_EDIT_SUCCESS", resMap)
	return
}

// 获取长效Isp子账户列表
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
// @Router /web/account/long_isp/sub_account_lists [post]
func GetLongIspUserChildAccountList(c *gin.Context) {
	username := strings.TrimSpace(c.DefaultPostForm("username", "")) // 用户名
	status := com.StrTo(c.DefaultPostForm("status", "10")).MustInt()
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id
	var accountLists []models.LongIspUserAccount
	_, accountLists = models.GetLongIspUserAccountList(uid, username)

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
		info.CreateTime = Time2DateEn(v.CreateTime)
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

// 添加/编辑 长效isp流量账号子账户
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
// @Router /web/account/long_isp/add_edit [post]
func SaveLongIspFlowChildAccount(c *gin.Context) {

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
		JsonReturn(c, e.ERROR, "__T_FLOW_NUMBER", nil)
		return
	}
	limitFlows := int64(util.StoI(flowStr))
	if limitFlows <= 0 {
		JsonReturn(c, e.ERROR, "__T_FLOW_NUMBER", nil)
		return
	}
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
	accountInfo := models.LongIspUserAccount{}
	if accountId > 0 {

		accountInfo, _ = models.GetLongIspUserAccountById(accountId)
		if accountInfo.Uid != uid {
			JsonReturn(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
			return
		}
		var hasAccount models.LongIspUserAccount
		_, hasAccount = models.GetLongIspUserAccountNeqId(accountId, username)

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
		_, hasAccount := models.GetLongIspUserAccount(0, username)
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
	flowInfo := models.GetUserDynamicIspInfo(uid)
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
			kcFlow = limitFlows - accountInfo.LimitFlow                                                                                                                           //流量差额
			flows = accountInfo.Flows + kcFlow                                                                                                                                    //操作后流量
			err := models.CreateLogLongIspFlowsAccount(uid, accountInfo.Id, accountInfo.Flows, accountInfo.LimitFlow, flows, limitFlows, accountInfo.Account, ip, "edit", "user") // 加个 流量变动日志  存 变动前的数据 剩余 ，和 配置额度   和变动后的配置额度    时间，IP
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

		models.UpdateLongIspUserAccountById(accountInfo.Id, upMap) //更新子账号信息

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
		data := models.LongIspUserAccount{}
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

		err, accId := models.AddLongIspProxyAccount(data)
		if err != nil {
			// 账户添加失败：发送产品侧预警（规则驱动模板+回退），并返回失败
			runtime := map[string]any{
				"username":  userInfo.Username,
				"childUser": username,
				"error":     err.Error(),
			}
			fallbackTpl := fmt.Sprintf("预警：【cherry】用户【%s】事件【认证账户添加】状态【失败】 信息：添加子账户失败，子账号：%s，错误：%s", userInfo.Username, username, err.Error())
			models.SendProductAlertWithRule("child_add_failed", runtime, fallbackTpl)
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

	resP := models.RedisLPUSH("list_account_long_isp_flow", string(listStr))
	fmt.Println(resP)

	JsonReturn(c, e.SUCCESS, "__T_EDIT_SUCCESS", nil)
	return
}

// 删除长效Isp流量子账号
// @BasePath /api/v1
// @Summary 删除 流量子账号
// @Description 删除 流量子账号
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param account_id formData string true "子账户ID"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/account/long_isp/sub_account_del [post]
func DelLongIspUserChildAccount(c *gin.Context) {
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
	accountInfo, _ := models.GetLongIspUserAccountById(accountId)
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
	resP := models.RedisLPUSH("list_account_long_isp_flow", string(listStr))
	fmt.Println(resP)

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return
}

// 长效Isp子账号启用  / 禁用
// @BasePath /api/v1
// @Summary 账号启用  / 禁用
// @Description 账号启用  / 禁用
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param account_id formData string true "子账户ID"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/account/long_isp/sub_account_enable_disable [post]
func LongIspChildAccountEnableOrDisable(c *gin.Context) {
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

	accountInfo, _ := models.GetLongIspUserAccountById(accountId)
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

	var b bool

	b = models.UpdateLongIspUserAccountById(accountInfo.Id, upMap)
	if !b {
		JsonReturn(c, e.ERROR, "__T_FAIL", nil)
		return
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return
}

// 长效Isp设置用户低于流量阀值 发送邮件
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
// @Router /web/account/long_isp/set_send [post]
func LongIspSetSendFlows(c *gin.Context) {
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
	flowInfo := models.GetUserDynamicIspInfo(userInfo.Id)
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
	res := models.EditUserDynamicIspByUid(flowInfo.Uid, upinfo)
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

// 长效Isp账号信息-判断是否可用
// @BasePath /api/v1
// @Summary 账号信息
// @Description 账号信息
// @Tags 个人中心 - 流量帐密子账号
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param account_id formData string true "帐密ID"
// @Produce json
// @Success 0 {object} map[string]interface{} "account_id:帐密ID,account:用户名,password:密码,limit_flow:流量限制,flow:已用流量,flow_unit:流量单位,is_use:能否使用生成命令"
// @Router /web/account/long_isp/detail [post]
func LongIspUserFlowAccountDetail(c *gin.Context) {
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

	accountInfo, _ := models.GetLongIspUserAccountById(accountId)

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
