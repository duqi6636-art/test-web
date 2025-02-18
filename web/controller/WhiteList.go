package controller

import (
	"api-360proxy/pkg/ipdat"
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"strings"
)

// 获取白名单列表信息
// @BasePath /api/v1
// @Summary 获取白名单列表信息
// @Description 获取白名单列表信息
// @Tags 个人中心 - 白名单相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param search formData string false "搜索关键字"
// @Param account_id formData string false "账号ID"
// @Param flow_type formData string false "流量类型"
// @Produce json
// @Success 0 {object} []models.ResUserWhitelistIp{}
// @Router /web/white/lists [post]
func IpWhitelists(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	resList := []models.ResUserWhitelistIp{}
	search := strings.TrimSpace(c.DefaultPostForm("search", ""))
	accountIdStr := strings.TrimSpace(c.DefaultPostForm("account_id", "0"))
	flow_type := strings.TrimSpace(c.DefaultPostForm("flow_type", "1"))
	if flow_type == "" {
		flow_type = "1"
	}
	flowType := util.StoI(flow_type)
	if flowType == 0 {
		flowType = 1
	}
	if accountIdStr == "" {
		accountIdStr = "0"
	}
	if uid > 0 {
		regionList := models.GetAllCountryV2("")
		regionMap := map[string]models.ExtractCountry{}

		for _, val := range regionList {
			regionMap[val.Country] = val
		}

		accountId := util.StoI(accountIdStr)
		lists := models.GetWhitelistIpsByUid(uid, accountId, flowType, search)
		fmt.Println(lists)
		for _, val := range lists {
			minutes := "Random"
			cate := 2
			if val.Minutes > 0 {
				minutes = util.ItoS(val.Minutes) + " mins"
				cate = 1
			}
			configStr := ""
			if val.Country != "" {
				configStr = configStr + "region-" + val.Country
			}
			if val.City != "" {
				configStr = configStr + "-city-" + val.City
			}
			if val.Minutes > 0 {
				sessid := util.RandStr("r", 8)
				if configStr == "" {
					configStr = configStr + "sessid-" + sessid + "-sessTime-" + util.ItoS(val.Minutes)
				} else {
					configStr = configStr + "-sessid-" + sessid + "-sessTime-" + util.ItoS(val.Minutes)
				}
			}
			regionInfo := regionMap[val.Country]
			info := models.ResUserWhitelistIp{}
			info.Id = val.Id + 10000
			info.WhitelistIp = val.WhitelistIp
			info.Country = val.Country
			info.Name = regionInfo.Name
			info.Img = regionInfo.Img
			info.City = val.City
			info.State = val.State
			info.Asn = val.Asn
			info.Hostname = val.Hostname
			info.Cate = cate
			info.FlowType = val.FlowType
			info.Minute = val.Minutes
			info.Minutes = minutes
			info.Configs = configStr
			info.Remark = val.Remark
			info.CreateTime = util.GetTimeStr(val.CreateTime, "d/m/Y H:i")
			resList = append(resList, info)
		}
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resList)
	return

}

// 添加 白名单IP
// @BasePath /api/v1
// @Summary 添加 白名单IP
// @Description 添加 白名单IP
// @Tags 个人中心 - 白名单相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param ip formData string true "IP地址"
// @Param flow_type formData string false "流量类型"
// @Param country formData string false "国家"
// @Param city formData string false "城市"
// @Param cate formData string false "类型 1 sticky ip  2 random IP"
// @Param minute formData string false "持续时间"
// @Param remark formData string false "备注"
// @Param account_id formData string false "账号ID"
// @Param hostname formData string false "主机名"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/white/add [post]
func AddWhitelist(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	ip := c.DefaultPostForm("ip", "")
	if ip == "" || !IsPublicIP(net.ParseIP(ip)) {
		JsonReturn(c, e.ERROR, "__T_IP_NOT_FORMAT", nil)
		return
	}

	ipInfo, _ := ipdat.IPDat.GetIpInfo(ip)
	countryCode := strings.ToUpper(ipInfo.CountryCode)
	if countryCode == "CN" || countryCode == "内网" || countryCode == "保留" || countryCode == "" {
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

	has, err := models.GetUserWhitelistIpByIp(ip, flowType)
	if err == nil && has.Id > 0 {
		JsonReturn(c, e.ERROR, "__T_IP_HAS_USED", nil)
		return
	}

	country := strings.TrimSpace(c.DefaultPostForm("country", ""))
	city := strings.TrimSpace(c.DefaultPostForm("city", ""))
	cate := strings.TrimSpace(c.DefaultPostForm("cate", "1")) //类型 1 sticky ip  2 random IP
	minuteStr := strings.TrimSpace(c.DefaultPostForm("minute", ""))
	remark := strings.TrimSpace(c.DefaultPostForm("remark", ""))
	accountIdStr := strings.TrimSpace(c.DefaultPostForm("account_id", "0"))
	hostname := strings.TrimSpace(c.DefaultPostForm("hostname", ""))
	if accountIdStr == "" {
		accountIdStr = "0"
	}
	minute := util.StoI(minuteStr)
	if cate == "1" {
		if minuteStr == "" || minute == 0 {
			JsonReturn(c, e.ERROR, "__T_IP_STICKY_TIME_ERROR", nil)
			return
		}
	} else {
		minute = 0
	}

	if strings.ToLower(country) == "global" || strings.ToLower(country) == "random" {
		country = ""
		city = ""
	}
	if strings.ToLower(city) == "global" || strings.ToLower(city) == "random" {
		city = ""
	}
	accountId := util.StoI(accountIdStr)
	//totalList := models.GetWhitelistIpsByUid(uid,accountId,"")
	//total := len(totalList)
	//if total >= 100 {
	//	JsonReturn(c, e.ERROR, "__T_IP_HAS_MORE_LIMIT", nil)
	//	return
	//}

	addInfo := models.CmUserWhitelistIp{}
	addInfo.Uid = uid
	addInfo.AccountId = accountId
	addInfo.Username = user.Username
	addInfo.WhitelistIp = ip
	addInfo.Country = strings.ToUpper(country)
	addInfo.City = city
	//addInfo.Cate        = cate
	addInfo.FlowType = flowType
	addInfo.Hostname = hostname
	addInfo.Minutes = minute
	addInfo.Status = 1
	addInfo.Remark = remark
	addInfo.Ip = c.ClientIP()
	addInfo.CreateTime = util.GetNowInt()
	err = models.AddUserWhitelistIp(addInfo)
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return

}

// 获取详细信息
// @BasePath /api/v1
// @Summary 获取详细信息
// @Description 获取详细信息
// @Tags 个人中心 - 白名单相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param id formData string true "白名单ID"
// @Produce json
// @Success 0 {object} map[string]interface{} "name：名称；img：图片"
// @Router /web/white/detail [post]
func GetWhitelist(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	idStr := c.DefaultPostForm("id", "")
	id := util.StoI(idStr) - 10000
	has, err := models.GetUserWhitelistIpById(id)
	if err != nil || has.Id == 0 {
		JsonReturn(c, e.ERROR, "__T_IP_HAS_NOT", nil)
		return
	}
	if has.Uid != uid {
		JsonReturn(c, e.ERROR, "__T_IP_INFO_ERROR", nil)
		return
	}
	city := has.City
	if city == "" {
		city = "Random"
	}
	cate := 2
	if has.Minutes > 0 {
		cate = 1
	}
	regionInfo := models.GetByCountry(has.Country)
	info := models.ResUserWhitelistIp{}
	info.Id = has.Id + 10000
	info.WhitelistIp = has.WhitelistIp
	info.Country = has.Country
	info.State = has.State
	info.Asn = has.Asn
	info.City = city
	info.Hostname = has.Hostname
	info.Cate = cate
	info.Minutes = util.ItoS(has.Minutes)
	info.FlowType = has.FlowType
	info.Remark = has.Remark
	info.CreateTime = util.GetTimeStr(has.CreateTime, "d/m/Y H:i")
	infoByte, _ := json.Marshal(info)
	resInfo := map[string]interface{}{}
	json.Unmarshal(infoByte, &resInfo)
	resInfo["name"] = regionInfo.Name
	resInfo["img"] = regionInfo.Img

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resInfo)
	return
}

// 编辑
// @BasePath /api/v1
// @Summary 编辑白名单
// @Description 编辑白名单
// @Tags 个人中心 - 白名单相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param ip formData string true "IP地址"
// @Param id formData string true "白名单ID"
// @Param country formData string false "国家"
// @Param city formData string false "城市"
// @Param cate formData string false "类型 1 sticky ip  2 random IP"
// @Param minute formData string false "持续时间"
// @Param remark formData string false "备注"
// @Param hostname formData string false "主机名"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/white/edit [post]
func EditWhitelist(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	ip := c.DefaultPostForm("ip", "")
	idStr := c.DefaultPostForm("id", "")
	if ip == "" || !IsPublicIP(net.ParseIP(ip)) {
		JsonReturn(c, e.ERROR, "__T_IP_NOT_FORMAT", nil)
		return
	}
	id := util.StoI(idStr) - 10000
	has, err := models.GetUserWhitelistIpById(id)
	if err != nil || has.Id == 0 {
		JsonReturn(c, e.ERROR, "__T_IP_HAS_NOT", nil)
		return
	}
	if has.Uid != uid {
		JsonReturn(c, e.ERROR, "__T_IP_INFO_ERROR", nil)
		return
	}

	hasIp, errIp := models.GetUserWhitelistIpByIp(ip, has.FlowType)
	if errIp == nil && hasIp.Id > 0 && hasIp.Id != has.Id {
		JsonReturn(c, e.ERROR, "__T_IP_HAS_USED", nil)
		return
	}

	country := strings.TrimSpace(c.DefaultPostForm("country", ""))
	state := strings.TrimSpace(c.DefaultPostForm("state", ""))
	asn := strings.TrimSpace(c.DefaultPostForm("asn", ""))
	city := strings.TrimSpace(c.DefaultPostForm("city", ""))
	cate := strings.TrimSpace(c.DefaultPostForm("cate", "1")) //类型 1 sticky ip  2 random IP
	minuteStr := strings.TrimSpace(c.DefaultPostForm("minute", ""))
	remark := strings.TrimSpace(c.DefaultPostForm("remark", ""))
	hostname := strings.TrimSpace(c.DefaultPostForm("hostname", ""))
	minute := util.StoI(minuteStr)
	if cate == "1" {
		if minuteStr == "" || minute == 0 {
			JsonReturn(c, e.ERROR, "__T_IP_STICKY_TIME_ERROR", nil)
			return
		}
	} else {
		minute = 0
	}
	if strings.ToLower(country) == "global" || strings.ToLower(country) == "random" {
		country = ""
		state = ""
		city = ""
	}
	if strings.ToLower(state) == "global" || strings.ToLower(state) == "random" {
		state = ""
		city = ""
	}
	if strings.ToLower(city) == "global" || strings.ToLower(city) == "random" {
		city = ""
	}
	if strings.ToLower(asn) == "global" || strings.ToLower(asn) == "random" {
		asn = ""
	}

	params := map[string]interface{}{}
	params["whitelist_ip"] = ip
	params["country"] = country
	params["state"] = state
	params["city"] = city
	params["asv"] = asn
	params["cate"] = cate
	params["hostname"] = hostname
	params["minutes"] = minute
	params["remark"] = remark
	params["update_time"] = util.GetNowInt()
	err = models.EditUserWhitelistIp(has.Id, params)

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return
}

// 删除
// @BasePath /api/v1
// @Summary 删除白名单
// @Description 删除白名单
// @Tags 个人中心 - 白名单相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param id formData string true "白名单ID"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/white/delete [post]
func DelWhitelist(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	idStr := c.DefaultPostForm("id", "")
	id := util.StoI(idStr) - 10000
	has, err := models.GetUserWhitelistIpById(id)
	if err != nil || has.Id == 0 {
		JsonReturn(c, e.ERROR, "__T_IP_HAS_NOT", nil)
		return
	}
	if has.Uid != uid {
		JsonReturn(c, e.ERROR, "__T_IP_INFO_ERROR", nil)
		return
	}
	err1 := models.DeleteUserWhitelist(id)
	if err1 != nil {
		JsonReturn(c, e.ERROR, err1.Error(), nil)
		return
	}
	//params := map[string]interface{}{}
	//params["status"] = -1
	//params["update_time"] = util.GetNowInt()
	//err = models.EditUserWhitelistIp(has.Id,params)

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return
}

// 白名单下载
// @BasePath /api/v1
// @Summary 白名单下载
// @Description 白名单下载
// @Tags 个人中心 - 白名单相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param search formData string false "搜索关键字"
// @Param account_id formData string false "账号ID"
// @Param flow_type formData string false "流量类型"
// @Produce octet-stream
// @Success 0 {object} interface{}
// @Router /web/white/download [post]
func WhitelistDownload(c *gin.Context) {
	title := []string{"Number", "IP", "Config", "Hostname-port"}
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id
	search := strings.TrimSpace(c.DefaultPostForm("search", ""))
	accountIdStr := strings.TrimSpace(c.DefaultPostForm("account_id", "0"))
	flow_type := strings.TrimSpace(c.DefaultPostForm("flow_type", "1"))
	if flow_type == "" {
		flow_type = "1"
	}
	flowType := util.StoI(flow_type)
	if flowType == 0 {
		flowType = 1
	}
	if accountIdStr == "" {
		accountIdStr = "0"
	}

	if uid > 0 {
		csvData := [][]string{}
		csvData = append(csvData, title)
		accountId := util.StoI(accountIdStr)
		lists := models.GetWhitelistIpsByUid(uid, accountId, flowType, search)

		for k, v := range lists {
			linshi := []string{}
			configStr := ""
			if v.Country != "" {
				configStr = configStr + "region-" + v.Country
			}
			if v.City != "" {
				configStr = configStr + "-city-" + v.City
			}
			if v.Minutes > 0 {
				sessid := util.RandStr("r", 8)
				//configStr = configStr + "-sessid-"+sessid+"-sessTime-" + util.ItoS(v.Minutes)
				if configStr == "" {
					configStr = configStr + "sessid-" + sessid + "-sessTime-" + util.ItoS(v.Minutes)
				} else {
					configStr = configStr + "-sessid-" + sessid + "-sessTime-" + util.ItoS(v.Minutes)
				}
			}
			linshi = append(linshi, util.ItoS(k+1))
			linshi = append(linshi, v.WhitelistIp)
			linshi = append(linshi, configStr)
			linshi = append(linshi, v.Hostname)

			csvData = append(csvData, linshi)
		}

		err := DownloadCsv(c, "Whitelists", csvData)
		fmt.Println(err)
		//if err != nil {
		//	JsonReturn(c, e.ERROR, err.Error(), nil)
		//	return
		//}
	}

	return
}

// 是否是公网IP
func IsPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch true {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}
