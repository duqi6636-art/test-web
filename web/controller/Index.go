package controller

import (
	"api-360proxy/pkg/ipdat"
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"bytes"
	"fmt"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/gin-gonic/gin"
	"image/png"
	"strconv"
	"strings"
	"time"
)

func Index(c *gin.Context) {
	c.String(200, "ok")
	return
}
func Index404(c *gin.Context) {
	c.String(404, "404")
	return
}

// @BasePath /api/v1
// @Summary 检查是否 中国大陆
// @Description 检查是否 中国大陆
// @Tags 检测相关
// @Accept x-www-form-urlencoded
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/check_is_cn [post]
func CheckIsCn(c *gin.Context) {
	ip := c.ClientIP()
	ipInfo, _ := ipdat.IPDat.GetIpInfo(ip)
	//ipInfo := GetIpInfo(ip)
	res := map[string]interface{}{}
	res["is_cn"] = 0
	if ipInfo.CountryCode == "CN" || ipInfo.CountryCode == "内网" || ipInfo.CountryCode == "保留" || ipInfo.CountryCode == "" {
		res["is_cn"] = 1
	}
	res["timestamp"] = util.GetNowInt()
	res["hour_timestamp"] = util.GetTodayHour()
	JsonReturn(c, 0, "__T_SUCCESS", res)
	return
}

// 生成二维码接口
func Qrcode(c *gin.Context) {
	data := c.DefaultQuery("data", "")
	h := c.DefaultQuery("h", "200")

	height, _ := strconv.Atoi(h)
	qrCode, _ := qr.Encode(data, qr.M, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, height, height)

	var b bytes.Buffer
	png.Encode(&b, qrCode)
	c.Header("content-type", "image/png")
	c.Writer.WriteString(string(b.Bytes()))
	return
}

// @BasePath /api/v1
// @Summary 关键词IP追踪统计
// @Description 关键词IP追踪统计
// @Tags 检测相关
// @Accept x-www-form-urlencoded
// @Param source formData string false "平台标识"
// @Param code formData string false "关键词"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/ip_source [post]
func IpSource(c *gin.Context) {
	source := strings.TrimSpace(c.DefaultPostForm("source", ""))
	code := strings.TrimSpace(c.DefaultPostForm("code", ""))

	ip := c.ClientIP()
	if code != "" && source == "" {
		source = "channel"
	}
	if source != "" && code != "" {
		nowTime := util.GetNowInt()
		keyword := source + "-" + code
		//err,has := models.GetIpSourceBy(keyword)

		var res error
		info := models.IpSourceModel{}
		info.CreateTime = nowTime
		info.Ip = ip
		info.Source = source
		info.Code = code
		info.Keyword = keyword
		res = models.AddIpSource(info)
		//if err == nil && has.Id > 0 {
		//	res = models.EditIpSource(has.Id,&info)
		//}else{
		//	info.Source = source
		//	info.Code = code
		//	info.Keyword = keyword
		//	res = models.AddIpSource(info)
		//}
		if res != nil {
			JsonReturn(c, -1, "__T_FAIL", nil)
			return
		}
	}

	JsonReturn(c, 0, "__T_SUCCESS", nil)
	return
}

// 生成用户域名
func CreateUserDomain(userInfo models.Users) bool {
	nowTime := int(time.Now().Unix())
	// 查询用户是否存在记录
	domainInfo, err := models.GetUserDomain(userInfo.Id)
	if err == nil && len(domainInfo) > 0 {
		return false
	}
	// 查询负载均衡机器列表
	serverList := models.GetServerList()
	if len(serverList) == 0 {
		return false
	}
	serverIpList := map[string][]string{}
	for _, v := range serverList {
		serverIpList[v.Area] = append(serverIpList[v.Area], v.Ip)
	}

	// 查询Dns解析数据
	dnsInfo := models.GetDnsInfo(models.GetConfigVal("dns_domain"))
	if dnsInfo.Id == 0 {
		return false
	}

	// 创建用户专属域名
	str_md5 := util.Md5(userInfo.Username + userInfo.RegIp)
	domain := str_md5[0:12]
	//sg.360proxy.com
	//us.360proxy.com
	//de.360proxy.com
	domainList := map[string]string{"sg": "Singapore", "us": "United States", "de": "Europe"}
	for k, v := range domainList {
		area := k
		userDomain := domain + k + "." + models.GetConfigVal("dns_domain")
		err = models.CreateDomain(models.AddCmUserDomain{
			Uid:        userInfo.Id,
			Username:   userInfo.Username,
			Domain:     userDomain,
			Title:      v,
			CreateTime: nowTime,
			Status:     1,
		})
		if err != nil {
			return false
		}

		random := userInfo.Id % 3
		ip := serverIpList[area][random]

		// 写入dns解析
		dnsErr := models.AddDnsInfo(models.Records{
			DomainId: dnsInfo.DomainId,
			Name:     userDomain,
			Type:     "A",
			Content:  ip,
			Ttl:      60,
		})
		if dnsErr != nil {
			return false
		}
	}

	return true
}

//// 获取用户域名列表
//func GetUserDomain(c *gin.Context) {
//	resCode, msg, user := DealUser(c) //处理用户信息
//	if resCode != e.SUCCESS {
//		JsonReturn(c, resCode, msg, nil)
//		return
//	}
//
//	// 查询用户域名列表
//	domainList, err := models.GetUserDomain(user.Id)
//	ispList := []map[string]interface{}{}
//	flowList := []map[string]interface{}{}
//	if len(domainList) > 0 && err == nil {
//		for _, v := range domainList {
//			isp := map[string]interface{}{}
//			isp["title"] = fmt.Sprintf("(%s) %s:3500", v.Title, v.Domain)
//			isp["value"] = v.Domain + ":3500"
//			ispList = append(ispList, isp)
//
//			flow := map[string]interface{}{}
//			flow["title"] = fmt.Sprintf("(%s) %s:3600 ", v.Title, v.Domain)
//			flow["value"] = v.Domain + ":3600"
//			flowList = append(flowList, flow)
//		}
//
//		JsonReturn(c, e.SUCCESS, "__T_SUCCESS", map[string]interface{}{
//			"isp":  ispList,
//			"flow": flowList,
//		})
//		return
//	} else {
//		domainArr := map[string]string{"sg": "Singapore", "us": "United States", "de": "Europe"}
//		dns_domain := models.GetConfigVal("dns_domain")
//
//		for k, v := range domainArr {
//			ispList = append(ispList, map[string]interface{}{
//				"title": "(" + v + ") " + k + "." + dns_domain + ":3500",
//				"value": k + "." + dns_domain + ":3500",
//			})
//			flowList = append(flowList, map[string]interface{}{
//				"title": "(" + v + ") " + k + "." + dns_domain + ":3600",
//				"value": k + "." + dns_domain + ":3600",
//			})
//		}
//
//		JsonReturn(c, e.SUCCESS, "__T_SUCCESS", map[string]interface{}{
//			"isp":  ispList,
//			"flow": flowList,
//		})
//		return
//	}
//}

// 获取用户域名列表
// @BasePath /api/v1
// @Summary 获取用户域名列表
// @Description 获取用户域名列表
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} map[string]interface{} "isp：isp域名列表；flow：流量域名列表；whitelist：白名单域名列表；flow_day：流量域名流量列表；flow_day_white：白名单域名流量列表"
// @Router /web/user/get_domain [post]
func GetUserDomain(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	// 查询用户分类表
	flowCate := models.GetUserFlowCate(user.Id)
	ispCate := models.GetUserIspCate(user.Id)
	ispCate = (ispCate % 5) + 1

	// 查询用户域名列表
	flowDomainList := models.GetDnsFlowDomainList(flowCate)
	ispDomainList := models.GetDnsFlowDomainList(flowCate)

	// 默认域名列表
	domainConfList := models.GetDnsDomainList(1)

	ispList := []map[string]interface{}{}
	flowList := []map[string]interface{}{}
	whiteList := []map[string]interface{}{}
	if len(flowDomainList) > 0 {
		for _, v := range flowDomainList {
			// 流量
			flow := map[string]interface{}{}
			flow["title"] = fmt.Sprintf("(%s) %s:3600", v.Country, v.Domain)
			flow["value"] = v.Domain + ":3600"
			flowList = append(flowList, flow)

			// 白名单
			white := map[string]interface{}{}
			white["title"] = fmt.Sprintf("(%s) %s:3650", v.Country, v.Domain)
			white["value"] = v.Domain + ":3650"
			whiteList = append(whiteList, white)
		}
	} else {
		for _, v := range domainConfList {
			// 流量
			flow := map[string]interface{}{}
			flow["title"] = fmt.Sprintf("(%s) %s:3600", v.Country, v.Domain)
			flow["value"] = v.Domain + ":3600"
			flowList = append(flowList, flow)

			// 白名单
			white := map[string]interface{}{}
			white["title"] = fmt.Sprintf("(%s) %s:3650", v.Country, v.Domain)
			white["value"] = v.Domain + ":3650"
			whiteList = append(whiteList, white)
		}
	}

	if len(ispDomainList) > 0 {
		for _, v := range ispDomainList {
			isp := map[string]interface{}{}
			isp["title"] = fmt.Sprintf("(%s) %s:3500", v.Country, v.Domain)
			isp["value"] = v.Domain + ":3500"
			ispList = append(ispList, isp)
		}
	} else {
		for _, v := range domainConfList {
			isp := map[string]interface{}{}
			isp["title"] = fmt.Sprintf("(%s) %s:3500", v.Country, v.Domain)
			isp["value"] = v.Domain + ":3500"
			ispList = append(ispList, isp)
		}
	}

	flowDayPoolList := models.ListPoolFlowDayByUid(user.Id)
	flowDayList := []map[string]interface{}{}
	flowDayWhiteList := []map[string]interface{}{}
	if len(flowDayPoolList) > 0 {
		userFlowDay := models.GetUserFlowDayByUid(user.Id)
		userArr := map[string]int{}
		for _, val := range userFlowDay {
			userArr[val.Hostname] = val.Status
		}
		for _, val := range flowDayPoolList {
			flowDay := map[string]interface{}{}
			flowDayWhite := map[string]interface{}{}
			host := val.Ip + ":" + util.ItoS(val.Port)
			host2 := val.Ip + ":" + util.ItoS(val.Port2)
			title := host
			title2 := host2
			if val.Country != "" {
				title = fmt.Sprintf("(%s)%s", val.Country, host)
				title2 = fmt.Sprintf("(%s)%s", val.Country, host2)
			}
			status, ok := userArr[val.Ip]
			if !ok {
				status = 0
			}
			flowDay["title"] = title
			flowDay["ip"] = val.Ip
			flowDay["port"] = val.Port
			flowDay["value"] = host
			flowDay["status"] = status

			flowDayWhite["title"] = title2
			flowDayWhite["value"] = host2
			flowDayWhite["status"] = status
			flowDayWhite["ip"] = val.Ip
			flowDayWhite["port"] = val.Port2
			flowDayList = append(flowDayList, flowDay)
			flowDayWhiteList = append(flowDayWhiteList, flowDayWhite)
		}
	} else {
		flowDay := map[string]interface{}{}
		flowDayWhite := map[string]interface{}{}
		host := "hostname:port"
		flowDay["title"] = host
		flowDay["value"] = host
		flowDay["ip"] = "hostname"
		flowDay["port"] = "port"
		flowDay["status"] = 0
		flowDayWhite["title"] = host
		flowDayWhite["value"] = host
		flowDayWhite["ip"] = "hostname"
		flowDayWhite["port"] = "port"
		flowDayWhite["status"] = 0
		flowDayList = append(flowDayList, flowDay)
		flowDayWhiteList = append(flowDayWhiteList, flowDayWhite)
	}

	unlimitedPortList := models.GetUserUnlimitedPortByUid(user.Id)
	unlimitedPortListResult := []map[string]interface{}{}
	for _, port := range unlimitedPortList {
		unlimitedPort := map[string]interface{}{}
		unlimitedPort["ip"] = port.Ip
		unlimitedPort["port"] = port.Port
		unlimitedPort["region"] = port.Region
		unlimitedPort["expired"] = port.ExpiredTime
		unlimitedPort["minute"] = port.Minute
		unlimitedPort["expire_time"] = util.GetTimeStr(port.ExpiredTime, "Y.m.d H:i:s")
		unlimitedPort["label"] = port.Ip + "(" + util.GetTimeStr(port.ExpiredTime, "Y.m.d H:i:s") + ")"
		unlimitedPortListResult = append(unlimitedPortListResult, unlimitedPort)
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", map[string]interface{}{
		"isp":            ispList,
		"flow":           flowList,
		"whitelist":      whiteList,
		"flow_day":       flowDayList,
		"flow_day_white": flowDayWhiteList,
		"unlimited_port": unlimitedPortListResult,
	})
	return
}
