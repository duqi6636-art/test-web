package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

// 获取 国家列表
// @BasePath /api/v1
// @Summary 获取 国家列表
// @Description 获取 国家列表
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param country formData string false "国家"
// @Param limit formData string false "限制数量"
// @Produce json
// @Success 0 {object} map[string][]models.ExtractCountry{} "popular：热门国家；other：其他国家"
// @Router /web/user/get_country [post]
func GetCountry(c *gin.Context) {
	country := strings.TrimSpace(c.DefaultPostForm("country", ""))
	limitStr := strings.TrimSpace(c.DefaultPostForm("limit", "0"))
	limit := util.StoI(limitStr)
	// 获取所有国家数据
	var countryList []models.ExtractCountry
	var otherList []models.ExtractCountry
	countryInfoList := models.GetAllCountry(limit, country)
	for _, v := range countryInfoList {
		if country == "" {
			if v.Img != "" {
				countryList = append(countryList, v)
			} else {
				otherList = append(otherList, v)
			}
		} else {
			if v.Country != "" {
				if v.Img != "" {
					countryList = append(countryList, v)
				} else {
					otherList = append(otherList, v)
				}
			}
		}
	}
	res := map[string][]models.ExtractCountry{}
	res["popular"] = countryList // 热门国家
	res["other"] = otherList     // 其他国家

	JsonReturn(c, 0, "__T_SUCCESS", res)
	return
}

// @BasePath /api/v1
// @Summary 获取大洲/省
// @Schemes
// @Description 获取大洲/省
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param country formData string false "国家标识筛选"
// @Produce json
// @Success 0 {array} models.ExtractProvince{}
// @Router /web/user/get_state [post]
func GetState(c *gin.Context) {
	country := strings.TrimSpace(c.DefaultPostForm("country", "")) // 国家标识
	country = strings.ToUpper(country)
	cityRandom := models.ExtractProvince{}
	cityRandom.Id = 0
	cityRandom.Name = "Random"
	cityRandom.Code = ""
	cityRandom.Status = 1
	resLists := []models.ExtractProvince{}
	resLists = append(resLists, cityRandom)
	// 获取所有洲省数据
	if country != "" {
		if country == "ALL" {
			country = ""
		}
		stateLists := models.GetStateByCountry(country, "")
		for _, v := range stateLists {
			resLists = append(resLists, v)
		}
	}
	JsonReturn(c, 0, "__T_SUCCESS", resLists)
	return
}

// 获取城市列表
// @BasePath /api/v1
// @Summary 获取城市列表
// @Description 获取城市列表
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param country formData string false "国家标识"
// @Param city formData string false "城市名称"
// @Produce json
// @Success 0 {object} []models.ExtractCity{} "城市列表"
// @Router /web/user/get_city [post]
func GetCity(c *gin.Context) {
	country := c.DefaultPostForm("country", "")              // 国家标识
	state := c.DefaultPostForm("state", "")                  // 州省标识
	city := strings.TrimSpace(c.DefaultPostForm("city", "")) // 城市名称
	country = strings.ToUpper(country)
	state = strings.ToLower(state)
	city = strings.ToLower(city)

	cityRandom := models.ExtractCity{}
	cityRandom.Id = 0
	cityRandom.Name = "Random"
	cityRandom.Code = ""
	cityRandom.Status = 1
	cityRandom.State = ""
	cityRandom.Country = ""

	resLists := []models.ExtractCity{}
	resLists = append(resLists, cityRandom)
	// 获取城市数据
	if country != "" || state != "" {
		cityLists := models.GetCityByCountry(country, state, city)
		for _, v := range cityLists {
			resLists = append(resLists, v)
		}
	}
	JsonReturn(c, 0, "__T_SUCCESS", resLists)
	return
}

// @BasePath /api/v1
// @Summary 获取ISP列表
// @Schemes
// @Description 获取ISP列表
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param country formData string false "国家标识筛选"
// @Produce json
// @Success 0 {array} models.ExtractIsp{}
// @Router /web/user/get_isp [post]
func GetCountryIsp(c *gin.Context) {
	country := strings.TrimSpace(c.DefaultPostForm("country", "")) // 国家标识

	redisConn := models.RedisCountryCityPort.Get()
	defer redisConn.Close()
	redisKey := fmt.Sprintf("country-isp-%s", country)
	listStr, _ := redis.String(redisConn.Do("GET", redisKey))
	var resData []map[string]interface{}
	if len(listStr) > 0 {
		json.Unmarshal([]byte(listStr), &resData)
		JsonReturn(c, 0, "__T_SUCCESS", resData)
		return
	}
	country = strings.ToUpper(country)
	// 获取所有洲省数据
	resLists := []models.ExtractIsp{}
	if country != "" {
		if country == "ALL" {
			country = ""
		}
		resLists = models.GetIspCountry(country)
	}
	// 存储到redis
	res, _ := json.Marshal(resLists)
	listStr = string(res)
	redisConn.Do("SETEX", redisKey, 1*60*10, listStr)
	JsonReturn(c, 0, "__T_SUCCESS", resLists)
	return
}

// 获取 国家列表 --- 不分页
// @BasePath /api/v1
// @Summary 获取 国家列表 --- 不分页
// @Description 获取 国家列表 --- 不分页
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param country formData string false "国家"
// @Produce json
// @Success 0 {object} []models.ExtractCountry{}
// @Router /web/user/get_country_V2 [post]
func GetCountryV2(c *gin.Context) {
	country := strings.TrimSpace(c.DefaultPostForm("country", ""))
	// 获取所有国家数据
	countryInfoList := models.GetAllCountryV2(country)

	JsonReturn(c, 0, "__T_SUCCESS", countryInfoList)
	return
}

// 添加 账号子账户
// @BasePath /api/v1
// @Summary 获取账号子账户
// @Description 获取账号子账户
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param flow_type formData string true "类型  1流量  3 动态ISP"
// @Produce json
// @Success 0 {object} map[string]interface{} "account：账号；password：密码"
// @Router /web/user/user_account [post]
func AddUserAccount(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	data := models.UserAccount{}
	data.Status = 1
	data.Remark = ""
	// 查询该用户下面是否存在相同账户 若存在则修改账户信息
	accountInfo := models.UserAccount{}
	_, accountInfo = models.GetUserAccountMaster(userInfo.Id)

	account := ""
	password := ""
	if accountInfo.Id == 0 {
		account = userInfo.Username
		password = util.RandStr("r", 8)
		data.Uid = userInfo.Id
		data.Account = account
		data.Password = password
		data.Master = 1
		data.FlowUnit = "GB"
		data.CreateTime = int(time.Now().Unix())

		err, id := models.AddProxyAccount(data)
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

// 修改账号子账户密码
// @BasePath /api/v1
// @Summary 修改账号子账户密码
// @Description 修改账号子账户密码
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param username formData string true "账户"
// @Param password formData string true "密码"
// @Param flow_type formData string true "类型  1流量  3 动态ISP"
// @Produce json
// @Success 0 {object} map[string]interface{} "account：账号；password：密码"
// @Router /web/user/set_pass [post]
func SetUserAccount(c *gin.Context) {
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
	accountInfo := models.UserAccount{}
	_, accountInfo = models.GetUserAccountMaster(userInfo.Id)

	if accountInfo.Id > 0 {
		//写入历史密码
		_, hasAccount := models.GetUserAccountNeqId(accountInfo.Id, username)
		if hasAccount.Id != 0 {
			JsonReturn(c, e.ERROR, "__T_ACCOUNT_USERNAME_EXIST", nil)
			return
		}
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

		errs = models.UpdateUserAccountById(accountInfo.Id, data)

	}

	resMap := map[string]interface{}{}
	resMap["account"] = username
	resMap["password"] = password
	JsonReturn(c, e.SUCCESS, "__T_EDIT_SUCCESS", resMap)
	return
}

// 获取 国家列表 --- 不分页（不限量套餐使用）
func GetFlowDayCountry(c *gin.Context) {
	country := strings.TrimSpace(c.DefaultPostForm("country", ""))
	// 获取所有国家数据
	countryInfoList := models.GetAllFlowDayCountry(country)

	JsonReturn(c, 0, "__T_SUCCESS", countryInfoList)
	return
}

// @BasePath /api/v1
// @Summary 获取国家端口列表
// @Schemes
// @Description 获取国家带城市端口列表
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户session"
// @Param type formData string true "类型 flow，flow_day，flow_day_white，isp，whitelist"
// @Produce json
// @Success 0 {array} models.ResExtractCountryCity{}
// @Router /web/user/country_domain_list [post]
func GetCountryDomainList(c *gin.Context) {
	mType := strings.TrimSpace(c.DefaultPostForm("type", ""))
	//redisConn := models.RedisCountryCityPort.Get()
	//defer redisConn.Close()
	//redisKey := fmt.Sprintf("countryCityPort_list:type-%s", mType)
	//listStr, _ := redis.String(redisConn.Do("GET", redisKey))
	//var resData []map[string]interface{}
	//if len(listStr) > 0 {
	//	json.Unmarshal([]byte(listStr), &resData)
	//	JsonReturn(c, 0, "__T_SUCCESS", resData)
	//	return
	//}

	// 获取所有国家列表
	allCountryList := models.GetAllCountryV2("")
	// 获取所有国家端口列表
	countryPortList := models.GetPortsByCountry()
	// 获取默认国家端口
	defaultCountryPort := models.GetCountryPortByCountry("US")
	// 初始化返回的结果
	var countryCityPortList []models.ResExtractCountryCity

	for _, v := range allCountryList {
		info := models.ResExtractCountryCity{}
		info.Name = v.Name
		info.Country = v.Country
		info.Img = v.Img
		info.Keyword = v.Keyword
		info.Sort = v.Sort
		// 获取国家下面端口
		var ports []map[string]string
		for _, port := range countryPortList {
			if port.Country == v.Country {
				// 端口后缀
				portSuffix := ":"
				if mType == "whitelist" {
					portSuffix = portSuffix + "3650"
				} else {
					portSuffix = portSuffix + "3600"
				}
				ports = transformPorts([]string{port.Port1, port.Port2, port.Port3}, portSuffix)
				break
			}
		}
		if len(ports) == 0 {
			// 端口后缀
			portSuffix := ":"
			if mType == "whitelist" {
				portSuffix = portSuffix + "3650"
			} else {
				portSuffix = portSuffix + "3600"
			}
			ports = transformPorts([]string{defaultCountryPort.Port1, defaultCountryPort.Port2, defaultCountryPort.Port3}, portSuffix)
		}
		info.Ports = ports

		// 将数据添加到数组
		countryCityPortList = append(countryCityPortList, info)
	}

	//// 存储到redis
	//res, _ := json.Marshal(countryCityPortList)
	//listStr = string(res)
	//redisConn.Do("SETEX", redisKey, 1*60*60, listStr)
	JsonReturn(c, 0, "__T_SUCCESS", countryCityPortList)
	return
}

// @BasePath /api/v1
// @Summary 获取长效国家城市域名列表
// @Schemes
// @Description 获取国家带城市端口列表
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Produce json
// @Success 0 {array} models.ResExtractCountryCity{}
// @Router /web/user/get_city_port [post]
func GetLongIspCountryCityPort(c *gin.Context) {

	// 获取所有国家端口列表
	countryPortList := models.GetLongIspPortsByCountry()
	// 初始化返回的结果
	var countryCityPortList []models.ResExtractCountryCity
	// 获取国家下面端口
	var ports []map[string]string
	for _, port := range countryPortList {
		// 端口后缀
		info := models.ResExtractCountryCity{}
		info.Name = port.Name
		info.Country = port.Country
		info.Keyword = port.Keyword
		info.Sort = port.Sort
		info.Img = port.Img
		portSuffix := ":"
		portSuffix = portSuffix + "3680"
		ports = transformPorts([]string{port.Port1, port.Port2, port.Port3}, portSuffix)
		info.Ports = ports
		// 将数据添加到数组
		countryCityPortList = append(countryCityPortList, info)
	}
	JsonReturn(c, 0, "__T_SUCCESS", countryCityPortList)
	return
}

func transformPorts(ports []string, portSuffix string) []map[string]string {
	var transformed []map[string]string

	for _, port := range ports {
		parts := strings.SplitN(port, ")", 2)
		if len(parts) == 2 {
			title := port
			value := parts[1]
			transformed = append(transformed, map[string]string{
				"title": title + portSuffix,
				"value": strings.TrimSpace(value + portSuffix),
			})
		}
	}

	return transformed
}

// / 类型 flow，flow_day，flow_day_white，isp，whitelist
func GetDefaultCountryPort(uid int, agentType string) []map[string]string {

	// 查询用户分类表
	flowCate := models.GetUserFlowCate(uid)
	ispCate := models.GetUserIspCate(uid)
	flowDayPool := models.GetPoolFlowDayByUid(uid)
	ispCate = (ispCate % 5) + 1

	// 查询用户域名列表
	flowDomainList := models.GetDnsFlowDomainList(flowCate)
	ispDomainList := models.GetDnsFlowDomainList(flowCate)

	// 默认域名列表
	domainConfList := models.GetDnsDomainList(1)

	var ispList []map[string]string
	var flowList []map[string]string
	var whiteList []map[string]string
	if len(flowDomainList) > 0 {
		for _, v := range flowDomainList {
			// 流量
			flow := map[string]string{}
			flow["title"] = fmt.Sprintf("(%s) %s:3600", v.Country, v.Domain)
			flow["value"] = v.Domain + ":3600"
			flowList = append(flowList, flow)

			// 白名单
			white := map[string]string{}
			white["title"] = fmt.Sprintf("(%s) %s:3650", v.Country, v.Domain)
			white["value"] = v.Domain + ":3650"
			whiteList = append(whiteList, white)
		}
	} else {
		for _, v := range domainConfList {
			// 流量
			flow := map[string]string{}
			flow["title"] = fmt.Sprintf("(%s) %s:3600", v.Country, v.Domain)
			flow["value"] = v.Domain + ":3600"
			flowList = append(flowList, flow)

			// 白名单
			white := map[string]string{}
			white["title"] = fmt.Sprintf("(%s) %s:3650", v.Country, v.Domain)
			white["value"] = v.Domain + ":3650"
			whiteList = append(whiteList, white)
		}
	}

	if len(ispDomainList) > 0 {
		for _, v := range ispDomainList {
			isp := map[string]string{}
			isp["title"] = fmt.Sprintf("(%s) %s:3500", v.Country, v.Domain)
			isp["value"] = v.Domain + ":3500"
			ispList = append(ispList, isp)
		}
	} else {
		for _, v := range domainConfList {
			isp := map[string]string{}
			isp["title"] = fmt.Sprintf("(%s) %s:3500", v.Country, v.Domain)
			isp["value"] = v.Domain + ":3500"
			ispList = append(ispList, isp)
		}
	}

	flowDay := map[string]interface{}{}
	flowWhiteDay := map[string]interface{}{}
	if flowDayPool.Id > 0 {
		host := flowDayPool.Ip + ":" + util.ItoS(flowDayPool.Port)
		host2 := flowDayPool.Ip + ":" + util.ItoS(flowDayPool.Port2)
		title := host
		title2 := host2
		if flowDayPool.Country != "" {
			title = fmt.Sprintf("(%s)%s", flowDayPool.Country, host)
			title2 = fmt.Sprintf("(%s)%s", flowDayPool.Country, host2)
		}
		flowDay["title"] = title
		flowDay["value"] = host

		flowWhiteDay["title"] = title2
		flowWhiteDay["value"] = host2

	} else {
		host := "hostname:port"
		flowDay["title"] = host
		flowDay["value"] = host
		flowWhiteDay["title"] = host
		flowWhiteDay["value"] = host
	}

	if agentType == "whitelist" {

		return whiteList
	} else {

		return flowList
	}

}
