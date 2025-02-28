package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/setting"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"strings"
)

// @BasePath /api/v1
// @Summary 获取基础配置信息
// @Schemes
// @Description 获取基础配置信息
// @Tags 个人中心-Api提取
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {array} map[string]interface{} "extract 提取接口地址，add 添加接口地址，del 删除接口地址，lists 白名单列表接口地址，has_api_ip 是否有API IP，当前IP是否在白名单中，ip 当前IP的白名单信息"
// @Router /center/flow_api/info [post]
func FlowApiInfo(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	// IP信息
	ip := c.ClientIP()
	hasIp := 0
	has, err := models.GetFlowApiWhiteByUidIp(uid, ip, 1)
	if err == nil && has.Id > 0 {
		hasIp = 1
	}
	domain := models.GetConfigVal("API_DOMAIN_URL") //
	apiUrl := strings.TrimRight(domain, "/") + "/api/extract_ip"
	//addUrl := strings.TrimRight(domain, "/") + "/api/add_ip"
	//delUrl := strings.TrimRight(domain, "/") + "/api/del_ip"
	//listsUrl := strings.TrimRight(domain, "/") + "/api/lists_ip"
	//aes_key := util.Md5(AesKey)
	//userStr, err := util.AesEnCode([]byte(user.Username), []byte(aes_key))
	//userKey := util.Md5(userStr)
	//params := "?user=" + user.Username + "&user_key=" + userKey
	//params2 := params + "&ip="
	resData := map[string]interface{}{
		"has_api_ip": hasIp,
		"ip":         ip,
		"extract":    apiUrl,
		//"add":        addUrl + params2,
		//"del":        delUrl + params2,
		//"lists":      listsUrl + params,
	}
	JsonReturn(c, 0, "__T_SUCCESS", resData)
	return
}

// @BasePath /api/v1
// @Summary 获取白名单列表信息
// @Schemes
// @Description 获取白名单列表信息
// @Tags 个人中心-Api提取
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param search formData string true "搜索关键字"
// @Param flow_type formData string true "类型 1普通流量 2无限制流量"
// @Param status formData string true "状态 on/off"
// @Param start_date formData string true "开始时间 2002-01-01"
// @Param end_date formData string true "结束时间 2002-01-01"
// @Produce json
// @Success 0 {array} models.ResUserWhitelistApi{}
// @Router /center/flow_api/lists [post]
func FlowApiWhitelist(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	resList := []models.ResUserWhitelistApi{}

	search := strings.TrimSpace(c.DefaultPostForm("search", ""))
	flow_type := strings.TrimSpace(c.DefaultPostForm("flow_type", "1"))
	lang := strings.ToLower(c.DefaultPostForm("lang", "en"))      //语言
	statusStr := strings.ToLower(c.DefaultPostForm("status", "")) //状态
	startDate := c.DefaultPostForm("start_date", "")
	endDate := c.DefaultPostForm("end_date", "")

	start := 0
	end := 0
	if startDate != "" {
		start = util.StoI(util.GetTimeStamp(startDate, "Y-m-d"))
	}
	if endDate != "" {
		end = util.StoI(util.GetTimeStamp(endDate, "Y-m-d")) + 86399
	}

	flowType := util.StoI(flow_type)
	if flowType == 0 {
		flowType = 1
	}
	if statusStr != "" {
		if statusStr == "on" {
			statusStr = "1"
		} else {
			statusStr = "2"
		}
	}
	status := util.StoI(statusStr)

	aes_key := util.Md5(AesKey)
	if uid > 0 {
		lists := models.GetFlowApiWhiteByUid(uid, flowType, search, status, start, end)
		for _, val := range lists {
			info := models.ResUserWhitelistApi{}
			idStr, err := util.AesEnCode([]byte(util.ItoS(val.Id)), []byte(aes_key))
			if err != nil {
				idStr = util.ItoS(val.Id)
			}
			statusR := "off"
			if val.Status == 1 {
				statusR = "on"
			}
			info.Id = idStr
			info.WhitelistIp = val.WhitelistIp
			info.Remark = val.Remark
			info.FlowType = val.FlowType
			info.Status = statusR
			info.CreateTime = util.GetTimeHIByLang(val.CreateTime, lang)
			resList = append(resList, info)
		}
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resList)
	return

}

// @BasePath /api/v1
// @Summary 下载白名单列表信息
// @Schemes
// @Description 下载白名单列表信息
// @Tags 个人中心-Api提取
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param search formData string true "搜索关键字"
// @Param flow_type formData string true "类型 1普通流量 2无限制流量"
// @Param status formData string true "状态 on/off"
// @Param start_date formData string true "开始时间 2002-01-01"
// @Param end_date formData string true "结束时间 2002-01-01"
// @Produce json
// @Success 0 {array} models.ResUserWhitelistApi{}
// @Router /center/flow_api/download [post]
func FlowApiWhitelistDownload(c *gin.Context) {
	title := []string{"IP", "SupportType", "AddTime", "Remarks"}
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	search := strings.TrimSpace(c.DefaultPostForm("search", ""))
	flow_type := strings.TrimSpace(c.DefaultPostForm("flow_type", "1"))
	statusStr := strings.ToLower(c.DefaultPostForm("status", "")) //状态
	startDate := c.DefaultPostForm("start_date", "")
	endDate := c.DefaultPostForm("end_date", "")

	start := 0
	end := 0
	if startDate != "" {
		start = util.StoI(util.GetTimeStamp(startDate, "Y-m-d"))
	}
	if endDate != "" {
		end = util.StoI(util.GetTimeStamp(endDate, "Y-m-d")) + 86399
	}

	flowType := util.StoI(flow_type)
	if flowType == 0 {
		flowType = 1
	}
	if statusStr != "" {
		if statusStr == "on" {
			statusStr = "1"
		} else {
			statusStr = "2"
		}
	}
	status := util.StoI(statusStr)

	if uid > 0 {
		csvData := [][]string{}
		csvData = append(csvData, title)
		lists := models.GetFlowApiWhiteByUid(uid, flowType, search, status, start, end)
		for _, val := range lists {
			linshi := []string{}
			linshi = append(linshi, val.WhitelistIp)
			linshi = append(linshi, "Country")
			linshi = append(linshi, util.GetTimeStr(val.CreateTime, "d/m/Y H:i"))
			linshi = append(linshi, val.Remark)

			csvData = append(csvData, linshi)
		}
		err := DownloadCsv(c, "Whitelists", csvData)
		fmt.Println(err)
	}
	return
}

// @BasePath /api/v1
// @Summary 添加白名单信息
// @Schemes
// @Description 添加白名单信息
// @Tags 个人中心-Api提取
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param ip formData string true "白名单IP"
// @Param flow_type formData string true "类型 1普通流量 2无限制流量"
// @Param remark formData string false "备注"
// @Produce json
// @Success 0 {array} map[string]interface{}
// @Router /center/flow_api/add [post]
func AddFlowApiWhite(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	country := strings.TrimSpace(c.DefaultPostForm("country", ""))
	if strings.ToLower(country) == "global" || strings.ToLower(country) == "random" {
		country = ""
	}
	ip := c.DefaultPostForm("ip", "")
	if ip == "" || !IsPublicIP(net.ParseIP(ip)) {
		JsonReturn(c, e.ERROR, "__T_IP_NOT_FORMAT", nil)
		return
	}
	ipInfo := GetIpInfo(ip)
	if ipInfo.CountryCode == "CN" || ipInfo.CountryCode == "内网" || ipInfo.CountryCode == "保留" || ipInfo.CountryCode == "" {
		JsonReturn(c, e.ERROR, "__T_UNSERVICED_AREA", nil)
		return
	}
	flow_type := strings.TrimSpace(c.DefaultPostForm("flow_type", "1"))
	if flow_type == "" {
		flow_type = "1"
	}
	flowType := util.StoI(flow_type)
	if flowType == 0 {
		flowType = 1
	}

	has, err := models.GetFlowApiWhiteByIp(ip, flowType, 0)
	if err == nil && has.Id > 0 {
		JsonReturn(c, e.ERROR, "__T_IP_HAS_USED", nil)
		return
	}

	remark := strings.TrimSpace(c.DefaultPostForm("remark", ""))

	totalList := models.GetFlowApiWhiteByUid(uid, flowType, "", 0, 0, 0)
	total := len(totalList)
	if total >= 100 {
		JsonReturn(c, e.ERROR, "__T_IP_HAS_MORE_LIMIT", nil)
		return
	}

	addInfo := models.MdUserWhitelistApi{}
	addInfo.Uid = uid
	addInfo.Username = user.Username
	addInfo.Country = country
	addInfo.WhitelistIp = ip
	addInfo.Status = 1
	addInfo.FlowType = flowType
	addInfo.Remark = remark
	addInfo.Ip = c.ClientIP()
	addInfo.Cate = 1
	addInfo.CreateTime = util.GetNowInt()
	err = models.AddFlowApiWhite(addInfo)
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return

}

// @BasePath /api/v1
// @Summary 白名单IP启用禁用
// @Schemes
// @Description 白名单IP启用禁用
// @Tags 个人中心-Api提取
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param id formData string true "白名单IP对映的ID"
// @Param status formData string true "状态  on启用  off禁用"
// @Produce json
// @Success 0 {array} map[string]interface{}
// @Router /center/flow_api/set_white [post]
func SetFlowApiWhite(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	idStr := c.DefaultPostForm("id", "")
	statusStr := strings.TrimSpace(c.DefaultPostForm("status", ""))

	aes_key := util.Md5(AesKey)
	idByte, err := util.AesDeCode(idStr, []byte(aes_key))
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_IP_HAS_NOT", nil)
		return
	}
	idStr = string(idByte)
	id := util.StoI(idStr)
	has, err := models.GetFlowApiWhiteById(id)
	if err != nil || has.Id == 0 {
		JsonReturn(c, e.ERROR, "__T_IP_HAS_NOT", nil)
		return
	}
	if has.Uid != uid {
		JsonReturn(c, e.ERROR, "__T_IP_INFO_ERROR", nil)
		return
	}
	status := 0
	if statusStr == "on" {
		status = 1
	} else {
		status = 2
	}

	params := map[string]interface{}{}
	params["status"] = status
	params["update_time"] = util.GetNowInt()
	err = models.EdiFlowApiWhiteById(has.Id, params)

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return
}

// @BasePath /api/v1
// @Summary 删除白名单IP详细信息
// @Schemes
// @Description 删除白名单IP详细信息
// @Tags 个人中心-Api提取
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param id formData string true "白名单IP对映的ID"
// @Produce json
// @Success 0 {array} map[string]interface{}
// @Router /center/flow_api/delete [post]
func DelFlowApiWhite(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	idStr := c.DefaultPostForm("id", "")
	aes_key := util.Md5(AesKey)
	idByte, err := util.AesDeCode(idStr, []byte(aes_key))
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_IP_HAS_NOT", nil)
		return
	}
	idStr = string(idByte)

	id := util.StoI(idStr)
	has, err := models.GetFlowApiWhiteById(id)
	if err != nil || has.Id == 0 {
		JsonReturn(c, e.ERROR, "__T_IP_HAS_NOT", nil)
		return
	}
	if has.Uid != uid {
		JsonReturn(c, e.ERROR, "__T_IP_INFO_ERROR", nil)
		return
	}

	params := map[string]interface{}{}
	params["status"] = -1
	params["update_time"] = util.GetNowInt()
	err = models.EdiFlowApiWhiteById(has.Id, params)

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return
}

// @BasePath /api/v1
// @Summary 提取IP信息
// @Schemes
// @Description 提取IP信息
// @Tags 个人中心-Api提取
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param num formData string true "提取数量"
// @Produce json
// @Success 0 {array} map[string]interface{}
// @Router /api/extract_ip [get]
func ExtractIp(c *gin.Context) {
	numStr := c.DefaultQuery("num", "1")
	num := util.StoI(numStr)
	if num <= 0 {
		JsonReturnShow(c, e.ERROR, "__T_NUMBER_ERROR", nil)
		return
	}
	if num > 500 {
		num = 500
	}
	country := strings.ToLower(c.DefaultQuery("regions", ""))
	protocol := strings.ToLower(c.DefaultQuery("protocol", ""))
	typeStr := strings.ToLower(c.DefaultQuery("type", ""))
	lt := strings.ToLower(c.DefaultQuery("lt", ""))
	st := strings.ToLower(c.DefaultQuery("st", ""))
	cate := strings.TrimSpace(c.DefaultPostForm("cate", "1")) //类型 1 sticky ip  2 random IP
	if protocol == "" {
		protocol = "http"
	}
	if protocol != "http" && protocol != "socks5" {
		JsonReturnShow(c, e.ERROR, "__T_PARAM_ERROR-- protocol", nil)
		return
	}
	if typeStr != "txt" && typeStr != "json" {
		JsonReturnShow(c, e.ERROR, "__T_PARAM_ERROR-- type", nil)
		return
	}
	ip := c.ClientIP()
	has, err := models.GetFlowApiWhiteByIp(ip, 1, 0)
	if err != nil || has.Id == 0 {
		JsonReturnShow(c, e.ERROR, "__T_IP_HAS_NOT", nil)
		return
	}
	if has.Status != 1 {
		JsonReturnShow(c, e.ERROR, "__T_IP_INFO_ERROR", nil)
		return
	}
	nowTime := util.GetNowInt()
	userFlowInfo := models.GetUserFlowInfo(has.Uid)
	if userFlowInfo.Flows <= 0 {
		JsonReturnShow(c, e.ERROR, "__T_NO_FLOW_INFO", nil)
		return
	}
	if userFlowInfo.Status != 1 {
		JsonReturn(c, -1, "__T_NO_FLOW_INFO", gin.H{})
		return
	}
	if userFlowInfo.ExpireTime < nowTime {
		JsonReturn(c, -1, "__T_FLOW_EXPIRED", gin.H{})
		return
	}

	ltStr := getBreakLine(lt, st)
	hostStrArr := []string{}
	hostArr := []ApiProxyJson{}
	uniqueNumbers := make(map[int]bool)
	portArr := []int{}
	var info models.ApiProxyClientInfo
	if country == "" || country == "global" {
		info, err = models.GetApiProxyAll(cate)
	} else {
		info, err = models.GetApiProxyClientByArea(country, cate)
	}
	if err != nil || info.Id == 0 {
		JsonReturnShow(c, e.ERROR, "__T_PARAM_ERROR-- info", nil)
		return
	}
	minPort := info.StartPort
	maxPort := minPort + info.PortNumb
	if num > info.PortNumb {
		num = info.PortNumb
	}
	portNum := minPort
	for len(portArr) < num {
		if !uniqueNumbers[portNum] {
			uniqueNumbers[portNum] = true
			portArr = append(portArr, portNum)
		}
		portNum++
		if portNum > maxPort {
			break
		}
	}
	area := "all"
	if info.Area != "" {
		area = info.Area
	}
	state := "as"
	if strings.Contains(info.Tag, "nasa") {
		state = "na"
	} else if strings.Contains(info.Tag, "eu") {
		state = "eu"
	} else if strings.Contains(info.Tag, "asa") {
		state = "as"
	}
	domain := area + "-" + state + "." + setting.AppConfig.FlowApiUrl
	for _, port := range portArr {
		hostStr := domain + ":" + util.ItoS(port)
		hostStrArr = append(hostStrArr, hostStr)

		jsonInfo := ApiProxyJson{
			Ip:   domain,
			Port: port,
		}
		hostArr = append(hostArr, jsonInfo)
	}

	models.AddLogApiUseInfo(has.Uid, num, has.WhitelistIp, country, protocol, typeStr, lt, st)
	hostList := strings.Join(hostStrArr, ltStr)
	if typeStr == "txt" {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusOK, hostList)
		return
	}
	JsonReturnShow(c, e.SUCCESS, "__T_SUCCESS", hostArr)
	return
}

// 获取列表
func FlowApiListsWhite(c *gin.Context) {
	user := c.DefaultQuery("user", "")
	user_key := c.DefaultQuery("user_key", "")
	if user == "" || user_key == "" {
		JsonReturnShow(c, e.ERROR, "__T_PARAM_ERROR", nil)
		return
	}

	aes_key := util.Md5(AesKey)
	userStr, err := util.AesEnCode([]byte(user), []byte(aes_key))
	userKey := util.Md5(userStr)
	if err != nil {
		JsonReturnShow(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
		return
	}
	if userKey != user_key {
		JsonReturnShow(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
		return
	}

	err, userInfo := models.GetUserByUsername(user)
	if err != nil || userInfo.Id == 0 {
		JsonReturnShow(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
		return
	}
	ip := c.ClientIP()
	ipInfo := GetIpInfo(ip)
	if ipInfo.CountryCode == "CN" || ipInfo.CountryCode == "内网" || ipInfo.CountryCode == "保留" || ipInfo.CountryCode == "" {
		JsonReturnShow(c, e.ERROR, "__T_UNSERVICED_AREA", nil)
		return
	}

	uid := userInfo.Id
	flowType := 1

	totalList := models.GetFlowApiWhiteByUid(uid, flowType, "", 0, 0, 0)
	resList := []models.ResFlowApiWhiteApi{}
	for _, val := range totalList {
		info := models.ResFlowApiWhiteApi{}
		info.Ip = val.WhitelistIp
		info.Remark = val.Remark
		resList = append(resList, info)
	}
	JsonReturnShow(c, e.SUCCESS, "__T_SUCCESS", resList)
	return

}

// 添加
func FlowApiAddWhite(c *gin.Context) {
	user := c.DefaultQuery("user", "")
	user_key := c.DefaultQuery("user_key", "")
	ip := c.DefaultQuery("ip", "")
	if user == "" || user_key == "" {
		JsonReturnShow(c, e.ERROR, "__T_PARAM_ERROR", nil)
		return
	}
	if ip == "" || !IsPublicIP(net.ParseIP(ip)) {
		JsonReturnShow(c, e.ERROR, "__T_IP_NOT_FORMAT", nil)
		return
	}

	aes_key := util.Md5(AesKey)
	userStr, err := util.AesEnCode([]byte(user), []byte(aes_key))
	userKey := util.Md5(userStr)

	if err != nil {
		JsonReturnShow(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
		return
	}
	if userKey != user_key {
		JsonReturnShow(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
		return
	}

	err, userInfo := models.GetUserByUsername(user)
	if err != nil || userInfo.Id == 0 {
		JsonReturnShow(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
		return
	}

	ipInfo := GetIpInfo(ip)
	if ipInfo.CountryCode == "CN" || ipInfo.CountryCode == "内网" || ipInfo.CountryCode == "保留" || ipInfo.CountryCode == "" {
		JsonReturnShow(c, e.ERROR, "__T_UNSERVICED_AREA", nil)
		return
	}

	uid := userInfo.Id
	flowType := 1

	has, err := models.GetFlowApiWhiteByIp(ip, flowType, 0)
	if err == nil && has.Id > 0 {
		JsonReturnShow(c, e.ERROR, "__T_IP_HAS_USED", nil)
		return
	}

	totalList := models.GetFlowApiWhiteByUid(uid, flowType, "", 0, 0, 0)
	total := len(totalList)
	if total >= 100 {
		JsonReturnShow(c, e.ERROR, "__T_IP_HAS_MORE_LIMIT", nil)
		return
	}

	addInfo := models.MdUserWhitelistApi{}
	addInfo.Uid = uid
	addInfo.Username = userInfo.Username
	addInfo.WhitelistIp = ip
	addInfo.Status = 1
	addInfo.FlowType = flowType
	addInfo.Remark = ""
	addInfo.Cate = 2
	addInfo.Ip = c.ClientIP()
	addInfo.CreateTime = util.GetNowInt()
	err = models.AddFlowApiWhite(addInfo)
	resInfo := map[string]interface{}{
		"whitelist_ip": ip,
		"request_ip":   c.ClientIP(),
	}
	JsonReturnShow(c, e.SUCCESS, "__T_SUCCESS", resInfo)
	return

}

// 删除
func FlowApiDelWhite(c *gin.Context) {
	user := c.DefaultQuery("user", "")
	user_key := c.DefaultQuery("user_key", "")
	ip := c.DefaultQuery("ip", "")
	if user == "" || user_key == "" {
		JsonReturnShow(c, e.ERROR, "__T_PARAM_ERROR", nil)
		return
	}
	if ip == "" || !IsPublicIP(net.ParseIP(ip)) {
		JsonReturnShow(c, e.ERROR, "__T_IP_NOT_FORMAT", nil)
		return
	}

	aes_key := util.Md5(AesKey)
	userStr, err := util.AesEnCode([]byte(user), []byte(aes_key))
	userKey := util.Md5(userStr)

	if err != nil {
		JsonReturnShow(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
		return
	}
	if userKey != user_key {
		JsonReturnShow(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
		return
	}

	err, userInfo := models.GetUserByUsername(user)
	if err != nil || userInfo.Id == 0 {
		JsonReturnShow(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
		return
	}

	uid := userInfo.Id
	flowType := 1

	has, err := models.GetFlowApiWhiteByUidIp(uid, ip, flowType)
	if err != nil || has.Id == 0 {
		JsonReturnShow(c, e.ERROR, "__T_IP_HAS_NOT", nil)
		return
	}

	params := map[string]interface{}{}
	params["status"] = -1
	params["update_time"] = util.GetNowInt()
	err = models.EdiFlowApiWhiteById(has.Id, params)
	JsonReturnShow(c, e.SUCCESS, "__T_SUCCESS", nil)
	return

}

type ApiProxyJson struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

// 获取换行符
func getBreakLine(lt string, st string) string {
	switch lt {
	case "1":
		return "\r\n"
	case "2":
		return "</br>"
	case "3":
		return "\r"
	case "4":
		return "\n"
	case "5":
		return "\t"
	case "6":
		return st
	default:
		return "\r\n"
	}
}

// 验证IP是否在白名单中
func ExistWhiteList(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	// IP信息
	ip := c.ClientIP()
	countryHasIp := 0
	cityHasIp := 0
	countryHas, err := models.GetFlowApiWhiteByUidIp(uid, ip, 1)
	if err == nil && countryHas.Id > 0 {
		countryHasIp = 1
	}

	cityHas, err := models.GetWhiteByUidIp(uid, ip, 1)
	if err == nil && cityHas.Id > 0 {
		countryHasIp = 1
	}
	data := map[string]interface{}{
		"ip":             ip,
		"country_has_ip": countryHasIp,
		"city_has_ip":    cityHasIp,
	}
	JsonReturn(c, 0, "__T_SUCCESS", data)
	return
}

// @BasePath /api/v1
// @Summary 白名单IP-api生成域名
// @Schemes
// @Description 白名单IP-api生成域名
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param country formData string false "国家"
// @Param cate formData string false 类型 1 sticky ip 2 random IP
// @Param num formData string false "数量"
// @Produce json
// @Router /center/white/api_domain [post]
func ApiDomain(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	country := strings.ToLower(c.DefaultPostForm("country", ""))
	numStr := strings.ToLower(c.DefaultPostForm("num", ""))
	cate := strings.TrimSpace(c.DefaultPostForm("cate", "1")) //类型 1 sticky ip  2 random IP

	num := util.StoI(numStr)
	hostStrArr := []string{}
	hostArr := []ApiProxyJson{}
	uniqueNumbers := make(map[int]bool)
	portArr := []int{}
	var info models.ApiProxyClientInfo
	var err error

	if country == "" || country == "global" {
		info, err = models.GetApiProxyAll(cate)
	} else {
		info, err = models.GetApiProxyClientByArea(country, cate)
	}
	if err != nil || info.Id == 0 {
		JsonReturnShow(c, e.ERROR, "__T_PARAM_ERROR-- info", nil)
		return
	}

	minPort := info.StartPort
	maxPort := minPort + info.PortNumb
	if num > info.PortNumb && cate == "2" {
		num = info.PortNumb
	}
	portNum := minPort
	for len(portArr) < num {
		if cate == "2" {
			portArr = append(portArr, minPort)
			if len(portArr) == num {
				break
			}
		} else {
			if !uniqueNumbers[portNum] {
				uniqueNumbers[portNum] = true
				portArr = append(portArr, portNum)
			}
			portNum++
			if portNum > maxPort {
				break
			}
		}

	}

	area := "all"
	if info.Area != "" {
		area = info.Area
	}
	state := "as"
	if strings.Contains(info.Tag, "nasa") {
		state = "na"
	} else if strings.Contains(info.Tag, "eu") {
		state = "eu"
	} else if strings.Contains(info.Tag, "asa") {
		state = "as"
	}
	domain := area + "-" + state + "." + setting.AppConfig.FlowApiUrl
	for _, port := range portArr {
		hostStr := domain + ":" + util.ItoS(port)
		hostStrArr = append(hostStrArr, hostStr)

		jsonInfo := ApiProxyJson{
			Ip:   domain,
			Port: port,
		}
		hostArr = append(hostArr, jsonInfo)
	}
	go models.AddLogApiUseInfo(user.Id, num, c.ClientIP(), country, "", "", "", "")
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", hostArr)
	return
}
