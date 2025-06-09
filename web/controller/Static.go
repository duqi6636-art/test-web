package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
	"sort"
	"strings"
)

// 静态 IP购买列表
func GetStaticIpList11(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	_, staticInfo := models.GetUserStaticIp(uid)
	packageList := models.GetStaticPackageList()
	resInfo := map[string]interface{}{}
	usInfo := []models.ResUserStaticIp{}
	allInfo := []models.ResUserStaticIp{}
	allStaticIps := 0

	for _, v := range staticInfo {
		info := models.ResUserStaticIp{}
		info.Id = v.PakId
		info.ExpireDay = v.ExpireDay
		balance := v.Balance
		if v.PakRegion == "us" {
			info.Balance = balance
			usInfo = append(usInfo, info)
		} else {
			info.Balance = balance
			allInfo = append(allInfo, info)
		}
		allStaticIps = allStaticIps + balance
	}

	resUs := []models.ResUserStaticIp{}
	resAll := []models.ResUserStaticIp{}
	for _, vp := range packageList {
		pakName := vp.Name
		info := models.ResUserStaticIp{}
		info.Id = vp.Id
		info.ExpireDay = vp.Value
		info.PakName = pakName
		usBalance := 0
		for _, v := range usInfo {
			if v.Id == vp.Id {
				usBalance = v.Balance
			}
		}
		info.Balance = usBalance
		resUs = append(resUs, info)

		allBalance := 0
		for _, v := range allInfo {
			if v.ExpireDay == vp.Value {
				allBalance = allBalance + v.Balance
			}

			//if v.Id == vp.Id {
			//	allBalance = v.Balance
			//}
		}
		info.Balance = allBalance
		resAll = append(resAll, info)
	}

	resInfo["us"] = resUs
	resInfo["all"] = resAll
	resInfo["all_ips"] = allStaticIps
	JsonReturn(c, 0, "__T_SUCCESS", resInfo)
	return
}

// 用户余额记录
// @BasePath /api/v1
// @Summary 用户余额记录
// @Description 用户余额记录
// @Tags 个人中心 - 静态住宅代理
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} []models.ResUserStaticIp{} "成功"
// @Router /web/static/ip_list [post]
func GetUserStaticIp(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	_, staticInfo := models.GetUserStaticIp(uid)
	packageList := models.GetStaticPackageList()

	userBalance := map[int]int{}
	status := 1
	for _, vu := range staticInfo {
		if vu.PakRegion != "all" {
			balance, ok := userBalance[vu.PakId]
			if !ok {
				balance = 0
			}
			userBalance[vu.PakId] = vu.Balance + balance
		}
		if vu.Status == 2 {
			status = 2
		}
	}

	resInfo := []models.ResUserStaticIp{}
	for _, vp := range packageList {
		info := models.ResUserStaticIp{}
		info.Id = vp.Id
		ipNum, ok := userBalance[vp.Id]
		if !ok {
			ipNum = 0
		}
		info.PakName = vp.Name
		info.ExpireDay = vp.Value
		info.Balance = ipNum
		info.Status = status
		resInfo = append(resInfo, info)
	}

	JsonReturn(c, 0, "__T_SUCCESS", resInfo)
	return
}

// 获取静态IP数量
// @BasePath /api/v1
// @Summary 获取静态IP数量
// @Description 获取静态IP数量
// @Tags 套餐页
// @Accept x-www-form-urlencoded
// @Produce json
// @Success 0 {object} []models.StaticIpCountryModel{}
// @Router /web/package/static_num [post]
func GetStaticRegionNum(c *gin.Context) {
	regionLists := models.GetStaticIpCountry()
	ipLists := models.GetStaticIpPool()

	ipNums := map[string]int{}
	for _, val := range ipLists {
		country := strings.ToLower(val.Country)
		ipNum, ok := ipNums[country]
		if !ok {
			ipNum = 0
		}
		newNum := ipNum + 1
		ipNums[country] = newNum
	}

	resRegion := []models.StaticIpCountryModel{}
	for _, v := range regionLists {
		country := strings.ToLower(v.Country)

		ipNum, ok := ipNums[country]
		if ok && ipNum > 0 {
			info := v
			info.IpNumber = ipNum + 200
			resRegion = append(resRegion, info)
		}
	}

	JsonReturn(c, 0, "__T_SUCCESS", resRegion)
	return
}

// 静态直连相关
var MdKey = "a8b1c1J9Q2K2"

// 静态国家/地区 州/省 城市
// @BasePath /api/v1
// @Summary 静态国家/地区 州/省 城市
// @Description 静态国家/地区 州/省 城市
// @Tags 个人中心 - 静态住宅代理
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {array} []models.ResUserStaticIpRegion{} "成功"
// @Router /web/static/region [post]
func GetRegion(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	_, staticInfo := models.GetUserStaticList(uid)
	packageList := models.GetStaticPackageList()

	buyNum := 0

	countryLists := []models.StaticIpCountryModel{}
	allInfo := models.StaticIpCountryModel{}
	allInfo.Name = "All"
	allInfo.Code = "All"
	allInfo.Country = "All"
	allInfo.IpNumber = 0
	countryList := models.GetStaticIpCountry()   //地区列表
	countryLists = append(countryLists, allInfo) //地区列表- 加个默认
	for _, v := range countryList {
		countryLists = append(countryLists, v)
	}

	countryList = append(countryList, allInfo)
	regionList := models.GetStaticRegion() //地区列表
	stateMapString := map[string]int{}
	cityArr := map[string][]models.StaticCityModel{}
	cityInfo := models.StaticCityModel{}
	for _, v := range regionList {
		cityStr := strings.ToUpper(v.Country) + "|||" + v.State
		_, ok := stateMapString[cityStr]

		if !ok {
			stateMapString[cityStr] = v.Sort
		}
		cityInfo.Code = v.City
		cityInfo.Name = v.City
		cityInfo.Sort = v.Sort
		cityArr[cityStr] = append(cityArr[cityStr], cityInfo)
	}
	stateArr := map[string][]models.StaticStateModel{}
	for k, v := range stateMapString {
		cityList := cityArr[k]
		//排序
		sort.SliceStable(cityList, func(i, j int) bool {
			return cityList[i].Sort > cityList[j].Sort
		})
		strArr := strings.Split(k, "|||")
		country := strArr[0]
		state := strArr[1]
		stateInfo := models.StaticStateModel{}
		stateInfo.Code = state
		stateInfo.Name = state
		stateInfo.Sort = v
		stateInfo.CityList = cityList
		stateArr[country] = append(stateArr[country], stateInfo)
	}

	countryAllList := []models.ResStaticRegion{}
	resMap := map[string]models.ResStaticRegion{}
	for _, v := range countryLists {
		country := strings.ToUpper(v.Country)
		stateInfo, ok := stateArr[country]
		if !ok {
			stateInfo = []models.StaticStateModel{}
		}
		//排序
		sort.SliceStable(stateInfo, func(i, j int) bool {
			return stateInfo[i].Sort > stateInfo[j].Sort
		})
		resInfo := models.ResStaticRegion{}
		//resInfo.Country = country
		resInfo.Code = country
		resInfo.Name = country
		resInfo.IpNumber = v.IpNumber
		resInfo.StatList = stateInfo
		resMap[country] = resInfo
		countryAllList = append(countryAllList, resInfo)
	}

	userBalance := map[int]int{}
	userStatic := map[int][]models.ResStaticRegion{}
	for _, vu := range staticInfo {
		if vu.PakRegion != "all" {
			balance, ok := userBalance[vu.PakId]
			if !ok {
				balance = 0
			}
			userBalance[vu.PakId] = vu.Balance + balance
			cList, oks := resMap[strings.ToUpper(vu.PakRegion)]
			cList.Balance = vu.Balance
			if oks {
				userStatic[vu.PakId] = append(userStatic[vu.PakId], cList)
			}
			buyNum = 1
		}
	}

	resInfo := []models.ResUserStaticIpRegion{}
	for _, vp := range packageList {
		info := models.ResUserStaticIpRegion{}
		info.Id = vp.Id
		ipNum, ok := userBalance[vp.Id]
		if !ok {
			ipNum = 0
		}
		userCountry := []models.ResStaticRegion{}
		if buyNum == 0 {
			for _, v := range countryAllList {
				v.Name = fmt.Sprintf("%s  %s (%d)", v.Code, "Remaining IPs", ipNum)
				userCountry = append(userCountry, v)
			}
		} else {
			one := resMap["ALL"]
			one.Name = fmt.Sprintf("%s  %s (%d)", one.Code, "Remaining IPs", ipNum)
			one.Balance = ipNum
			userCountry = append(userCountry, one)
			userHas := userStatic[vp.Id]
			for _, v := range userHas {
				v.Name = fmt.Sprintf("%s  %s (%d)", v.Code, "Remaining IPs", v.Balance)
				userCountry = append(userCountry, v)
			}
		}
		info.PakName = vp.Name
		info.ExpireDay = vp.Value
		info.Balance = ipNum
		info.CountryList = userCountry
		resInfo = append(resInfo, info)
	}

	JsonReturn(c, 0, "__T_SUCCESS", resInfo)
	return
}

// 静态 IP池列表
// @BasePath /api/v1
// @Summary 静态 IP池列表
// @Description 静态 IP池列表
// @Tags 个人中心 - 静态住宅代理
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录信息"
// @Param country formData string false "国家"
// @Param state formData string false "洲/省"
// @Param city formData string false "城市"
// @Produce json
// @Success 0 {array} []models.StaticIpInfo{} "成功"
// @Router /web/static/pools [post]
func StaticIpList(c *gin.Context) {
	session := c.DefaultPostForm("session", "")                    // 用户名
	country := strings.TrimSpace(c.DefaultPostForm("country", "")) //国家
	state := strings.TrimSpace(c.DefaultPostForm("state", ""))     //州/省
	city := strings.TrimSpace(c.DefaultPostForm("city", ""))       //城市
	usedStr := ""
	//uid := 0
	if session != "" {
		errs, uid := GetUIDbySession(session)
		if errs == false {
			JsonReturn(c, -1, "__T_SESSION_EXPIRE", gin.H{})
			return
		}
		err, usedList := models.GetIpStaticIpByUid(uid)
		if err == nil && len(usedList) > 0 {
			for _, v := range usedList {
				usedStr = usedStr + "," + v.Ip
			}
			usedStr = usedStr + ","
		}
	}

	country = strings.ToLower(country) //国家
	state = strings.ToLower(state)     //州/省
	city = strings.ToLower(city)       //城市

	staticIpData := []models.StaticIpInfo{}
	lists := models.GetStaticIpPoolRand(country, state, city)
	for _, v := range lists {
		if usedStr != "" && strings.Contains(usedStr, ","+v.Ip+",") {
			continue
		}
		ipArr := strings.Split(v.Ip, ".")
		ip := ipArr[0] + "." + ipArr[1] + ".***.***"
		info := models.StaticIpInfo{}
		info.Ip = ip
		info.Ping = util.GetRandomInt(30, 100)
		info.Sn = util.MdEncode(util.ItoS(v.Id), MdKey)
		info.Code = v.Code
		info.Country = v.Country
		info.State = v.State
		info.City = v.City
		info.Port = util.ItoS(v.Port)
		staticIpData = append(staticIpData, info)
		num := 50

		if len(staticIpData) >= num {
			break
		}
	}
	JsonReturn(c, 0, "__T_SUCCESS", staticIpData)
	return
}

// 提取扣费IP
// @BasePath /api/v1
// @Summary 提取扣费IP
// @Description 提取扣费IP
// @Tags 个人中心 - 静态住宅代理
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param ip formData string true "IP地址"
// @Param sn formData string true "SN码"
// @Param static_id formData string true "套餐ID"
// @Produce json
// @Success 0 {object} []models.ResUserStaticIp{} "成功"
// @Router /web/static/use [post]
func UseStatic(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	_, staticInfo := models.GetUserStaticIp(uid)
	for _, vu := range staticInfo {
		if vu.Status == 2 {
			JsonReturn(c, -1, "__T_PACKAGE_FORBIDDEN", gin.H{})
			return
		}
	}

	nowTime := util.GetNowInt()
	ip := c.DefaultPostForm("ip", "")
	if ip != "" {
		// 提取记录
		err_l, ipLog := models.GetIpStaticIp(uid, ip)
		if err_l == nil && ipLog.Id > 0 {
			if ipLog.ExpireTime < nowTime {
				JsonReturn(c, -1, "__T_IP_EXPIRED", nil)
				return
			}
			JsonReturn(c, 0, "__T_SUCCESS", nil)
			return
		}
	}
	sn := c.DefaultPostForm("sn", "")
	staticIdStr := c.DefaultPostForm("static_id", "") //长效套餐ID
	if staticIdStr == "" {
		JsonReturn(c, -1, "__T_IP_BALANCE_LOW", nil)
		return
	}
	static_id := util.StoI(staticIdStr)
	if static_id == 0 {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}
	if sn != "" {
		idStr := util.MdDecode(sn, MdKey)
		id := util.StoI(idStr)
		_, ipInfo := models.GetStaticIpById(id)
		useIP := ipInfo.Ip
		err_l, ipLog := models.GetIpStaticIp(uid, useIP)
		if err_l == nil && ipLog.Id > 0 {
			JsonReturn(c, -1, "__T_STATIC_IP_HAS_USED", nil) //已提取，请勿重复提取
			return
		}
		code := strings.ToLower(ipInfo.Country)
		err, balanceInfo := models.GetUserStaticByPakRegion(uid, static_id, code)
		if err != nil || balanceInfo.Id == 0 {
			JsonReturn(c, -1, "__T_IP_BALANCE_LOW", nil)
			return
		} else {
			if balanceInfo.Balance < 1 {
				JsonReturn(c, -1, "__T_IP_BALANCE_LOW", nil)
				return
			}
		}

		// 开始扣费
		err1 := models.StaticKf(code, c.ClientIP(), ipInfo, user, balanceInfo)
		if err1 == nil {
			_, staticInfo := models.GetUserStaticIp(uid) //用户购买记录
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

			resInfo := []models.ResUserStaticIp{}
			for _, vp := range packageList {
				info := models.ResUserStaticIp{}
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
			JsonReturn(c, 0, "__T_SUCCESS", resInfo)
			return
		}
	}
	JsonReturn(c, -1, "__T_FAIL", nil)
	return
}

// 批量提取静态

func BatchUseStatic(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	_, staticInfo := models.GetUserStaticIp(uid)
	for _, vu := range staticInfo {
		if vu.Status == 2 {
			JsonReturn(c, -1, "__T_PACKAGE_FORBIDDEN", gin.H{})
			return
		}
	}
	sn := c.DefaultPostForm("sn_list", "")
	var snList = make([]string, 0)
	err := json.Unmarshal([]byte(sn), &snList)
	if len(snList) <= 0 || err != nil {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}
	staticId := com.StrTo(c.DefaultPostForm("static_id", "0")).MustInt() //长效套餐ID
	if staticId <= 0 {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}
	resInfo := make(map[int]models.ResUserStaticIp)
	_, staticInfoList := models.GetUserStaticIp(uid) //用户购买记录
	packageList := models.GetStaticPackageList()
	userBalance := map[int]int{}
	for _, vu := range staticInfoList {
		if vu.PakRegion != "all" {
			balance, ok := userBalance[vu.PakId]
			if !ok {
				balance = 0
			}
			userBalance[vu.PakId] = vu.Balance + balance
		}
	}
	for _, vp := range packageList {
		info := models.ResUserStaticIp{}
		info.Id = vp.Id
		ipNum, ok := userBalance[vp.Id]
		if !ok {
			ipNum = 0
		}
		info.PakName = vp.Name
		info.ExpireDay = vp.Value
		info.Balance = ipNum
		resInfo[info.Id] = info
	}
	for _, val := range snList {
		id := util.StoI(util.MdDecode(val, MdKey))
		_, ipInfo := models.GetStaticIpById(id)
		useIP := ipInfo.Ip
		err_l, ipLog := models.GetIpStaticIp(uid, useIP)
		if err_l == nil && ipLog.Id > 0 {
			continue
		}
		code := strings.ToLower(ipInfo.Country)
		err, balanceInfo := models.GetUserStaticByPakRegion(uid, staticId, code)
		if err != nil || balanceInfo.Id == 0 {
			continue
		} else {
			if balanceInfo.Balance < 1 {
				continue
			}
		}

		// 开始扣费
		err1 := models.StaticKf(code, c.ClientIP(), ipInfo, user, balanceInfo)
		if err1 == nil {

			userBalance := map[int]int{}
			for _, vu := range staticInfoList {
				if vu.PakRegion != "all" {
					balance, ok := userBalance[vu.PakId]
					if !ok {
						balance = 0
					}
					userBalance[vu.PakId] = vu.Balance + balance
				}
			}

			for _, vp := range packageList {
				info := models.ResUserStaticIp{}
				info.Id = vp.Id
				ipNum, ok := userBalance[vp.Id]
				if !ok {
					ipNum = 0
				}
				info.PakName = vp.Name
				info.ExpireDay = vp.Value
				info.Balance = ipNum
				resInfo[info.Id] = info
			}
		}
	}
	var resList = make([]models.ResUserStaticIp, 0)
	for _, val := range resInfo {
		resList = append(resList, val)
	}
	JsonReturn(c, 0, "__T_SUCCESS", resList)
	return
}

// 修改账号子账户密码
// @BasePath /api/v1
// @Summary 修改账号子账户密码
// @Description 修改账号子账户密码
// @Tags 个人中心 - 静态住宅代理
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param id formData string true "待修改的ID"
// @Param password formData string true "密码"
// @Param account formData string true "用户名"
// @Param remark formData string false "备注"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /web/static/set_account [post]
func SetIPAccount(c *gin.Context) {
	password := strings.TrimSpace(c.DefaultPostForm("password", "")) // 密码
	account := strings.TrimSpace(c.DefaultPostForm("account", ""))   // 用户名
	remark := strings.TrimSpace(c.DefaultPostForm("remark", ""))     // 用户名
	resCode, msg, userInfo := DealUser(c)                            //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	if account == "" {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_USERNAME_ERROR", nil)
		return
	}
	if password == "" {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_PASSWORD_ERROR", nil)
		return
	}
	if !util.CheckUserAccount(account) {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_USERNAME_ERROR", nil)
		return
	}
	if !util.CheckUserPassword(password) {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_PASSWORD_ERROR", nil)
		return
	}

	idStr := c.DefaultPostForm("id", "") //待操作的ID
	if idStr == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}
	id := util.StoI(idStr)
	err_l, ipLog := models.GetIpStaticIpById(id)
	if err_l != nil || ipLog.Id == 0 {
		JsonReturn(c, -1, "__T_STATIC_IP_USED_ERROR", nil)
		return
	}
	if ipLog.Uid != userInfo.Id {
		JsonReturn(c, -1, "__T_STATIC_IP_USED_ERROR", nil)
		return
	}
	data := map[string]interface{}{}
	data["account"] = account
	data["password"] = password
	data["remark"] = remark
	err1 := models.SetIpStaticIp(ipLog.Id, data)
	fmt.Println(err1)
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", gin.H{})
	return
}

type ResUserStaticIp struct {
	Id         int    `json:"id"`
	PakName    string `json:"pak_name"`    // 套餐类型
	Balance    int    `json:"balance"`     // 剩余IP
	ExpireDay  int    `json:"expire_day"`  // 过期天数
	ExpireTime string `json:"expire_time"` // 过期时间
}

// 续费 前信息
// @BasePath /api/v1
// @Summary 续费 前信息
// @Description 续费 前信息
// @Tags 个人中心 - 静态住宅代理
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param ip formData string true "待续费的IP"
// @Produce json
// @Success 0 {object} map[string]interface{} "region：国家地区，ip：ip地址，balance：用户剩余IP数量，lists：用户剩余静态住宅代理列表（值为[]ResUserStaticIp{}模型）"
// @Router /web/static/check_info [post]
func BeforeRecharge(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	ip := c.DefaultPostForm("ip", "") //待续费的ID
	if ip == "" {
		JsonReturn(c, -1, "IP info Error", nil)
		return
	}
	resList := []ResUserStaticIp{}
	uid := user.Id
	// 提取记录
	err_l, ipLog := models.GetIpStaticIp(uid, ip)
	if err_l != nil || ipLog.Id == 0 {
		JsonReturn(c, -1, "__T_STATIC_IP_USED_ERROR", nil)
		return
	}

	_, ipInfo := models.GetStaticIpByIp(ipLog.Ip)
	if ipInfo.Id == 0 || ipInfo.Status != 1 {
		JsonReturn(c, -1, "__T_IP_OFFLINE", nil)
		return
	}
	//if ipInfo.Uid > 0 && ipInfo.Uid != uid {
	//	JsonReturn(c, -1, "__T_IP_HAS_USED", nil)
	//	return
	//}

	err, balanceList := models.GetUserStaticIpByRegion(uid, strings.ToLower(ipLog.Country))
	if err != nil || len(balanceList) == 0 {
		resInfo := map[string]interface{}{
			"region":  ipLog.Country,
			"ip":      ipLog.Ip,
			"balance": 0,
			"lists":   resList,
		}

		JsonReturn(c, 0, "success", resInfo)
		return
	}
	nowTime := util.GetNowInt()
	expire := ipLog.ExpireTime
	if expire < nowTime {
		expire = nowTime
	}
	balance := 0
	for _, v := range balanceList {
		balance = balance + v.Balance
		expireTime := expire + v.ExpireDay*86400
		info := ResUserStaticIp{}
		info.Id = v.Id
		info.PakName = util.ItoS(v.ExpireDay) + " Day"
		info.Balance = v.Balance
		info.ExpireDay = v.ExpireDay
		info.ExpireTime = util.GetTimeStr(expireTime, "d-m-Y")
		resList = append(resList, info)
	}
	resInfo := map[string]interface{}{
		"region":  ipLog.Country,
		"ip":      ipLog.Ip,
		"balance": balance,
		"lists":   resList,
	}

	JsonReturn(c, 0, "success", resInfo)
	return

}

type ResBatchBeforeRecharge struct {
	Id         int    `json:"id"`
	PakName    string `json:"pak_name"`    // 套餐类型
	Country    string `json:"country"`     // 国家
	Ip         string `json:"ip"`          // ip
	RechargeId string `json:"recharge_id"` // 续费需要的id
	ExpireDay  int    `json:"expire_day"`  // 过期天数
	ExpireTime string `json:"expire_time"` // 过期时间
}

func BatchBeforeRecharge(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	ips := c.DefaultPostForm("ips", "") //待续费的ID
	var ipsList = make([]string, 0)
	err := json.Unmarshal([]byte(ips), &ipsList)
	if len(ipsList) <= 0 || err != nil {
		JsonReturn(c, -1, "IP info Error", nil)
		return
	}
	resMap := make(map[string][]ResBatchBeforeRecharge)
	resInfoMap := make(map[string]ResBatchBeforeRecharge)
	var balanceMap = make(map[string]map[string]int)
	uid := user.Id
	var expireDayList = []int{7, 30}
	//var resInfoMap =
	for _, ip := range ipsList {
		err_l, ipLog := models.GetIpStaticIp(uid, ip)
		if err_l != nil || ipLog.Id == 0 {
			continue
		}
		_, ipInfo := models.GetStaticIpByIp(ipLog.Ip)
		if ipInfo.Id == 0 || ipInfo.Status != 1 {
			continue
		}
		_, balanceList := models.GetUserStaticIpByRegion(uid, strings.ToLower(ipLog.Country))
		nowTime := util.GetNowInt()
		expire := ipLog.ExpireTime
		if expire < nowTime {
			expire = nowTime
		}
		balance := 0
		var resInfo = ResBatchBeforeRecharge{
			Country: ipLog.Country,
			Ip:      ipLog.Ip,
		}
		for _, expireDay := range expireDayList {
			resInfo.ExpireTime = util.GetTimeStr(expire+expireDay*86400, "d-m-Y")
			resInfo.RechargeId = fmt.Sprintf("%v:%v", 0, ipLog.Id)
			resInfo.ExpireDay = expireDay
			resInfo.PakName = util.ItoS(expireDay) + " Day"
			key := fmt.Sprintf("%v_%v", ip, expireDay)
			resInfoMap[key] = resInfo
			balanceMap[resInfo.PakName] = map[string]int{ipLog.Country: balance}
		}

		for _, v := range balanceList {

			balance = balance + v.Balance

			expireTime := expire + v.ExpireDay*86400
			resInfo.Id = v.Id
			resInfo.PakName = util.ItoS(v.ExpireDay) + " Day"
			resInfo.ExpireDay = v.ExpireDay
			resInfo.ExpireTime = util.GetTimeStr(expireTime, "d-m-Y")
			resInfo.RechargeId = fmt.Sprintf("%v:%v", v.Id, ipLog.Id)
			//if val, ok := resMap[resInfo.PakName]; ok {
			//	resMap[resInfo.PakName] = append(val, resInfo)
			//} else {
			//	resMap[resInfo.PakName] = append(val, resInfo)
			//}
			key := fmt.Sprintf("%v_%v", ip, v.ExpireDay)
			resInfoMap[key] = resInfo
			if val, ok := balanceMap[resInfo.PakName]; ok {
				if v, ok := val[ipLog.Country]; ok {
					val[ipLog.Country] = v + balance
				} else {
					val[ipLog.Country] = balance
				}
			} else {
				balanceMap[resInfo.PakName] = map[string]int{ipLog.Country: balance}
			}
		}
	}

	for _, resInfo := range resInfoMap {
		if val, ok := resMap[resInfo.PakName]; ok {
			resMap[resInfo.PakName] = append(val, resInfo)
		} else {
			resMap[resInfo.PakName] = append(val, resInfo)
		}
	}

	JsonReturn(c, 0, "success", map[string]interface{}{"balance": balanceMap, "res": resMap})
	return
}

// 续费
// @BasePath /api/v1
// @Summary 续费
// @Description 续费
// @Tags 个人中心 - 静态住宅代理
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param id formData string true "待续费的ID"
// @Param static_id formData string true "购买的长效套餐ID"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /web/static/recharge [post]
func IpRecharge(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	staticIdStr := c.DefaultPostForm("static_id", "") //购买的长效套餐ID
	idStr := c.DefaultPostForm("id", "")              //待续费的ID
	if idStr == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}
	id := util.StoI(idStr)

	if staticIdStr == "" {
		JsonReturn(c, -1, "__T_RECHARGE_IP_ERROR", nil)
		return
	}
	static_id := util.StoI(staticIdStr)
	if static_id == 0 {
		JsonReturn(c, -1, "__T_RECHARGE_IP_ERROR", nil)
		return
	}

	uid := user.Id
	//nowTime := util.GetNowInt()
	// 提取记录
	err_l, ipLog := models.GetIpStaticIpById(id)
	if err_l != nil || ipLog.Id == 0 {
		JsonReturn(c, -1, "__T_STATIC_IP_USED_ERROR", nil)
		return
	}
	_, ipInfo := models.GetStaticIpByIp(ipLog.Ip)
	if ipInfo.Id == 0 || ipInfo.Status != 1 {
		JsonReturn(c, -1, "__T_IP_OFFLINE", nil)
		return
	}
	// 处理IP异常的情况，续费的IP不判断是否已被使用过的IP   20250122
	//if ipInfo.Uid > 0 && ipInfo.Uid != uid {
	//	JsonReturn(c, -1, "__T_IP_HAS_USED", nil)
	//	return
	//}

	err, balanceInfo := models.GetUserStaticIpById(uid, static_id)
	if err != nil || balanceInfo.Id == 0 {
		JsonReturn(c, 2, "__T_IP_BALANCE_LOW", nil)
		return
	} else {
		if balanceInfo.Balance < 1 {
			JsonReturn(c, 2, "__T_IP_BALANCE_LOW", nil)
			return
		}
	}

	// 开始扣费
	err1 := models.Recharge(c.ClientIP(), ipLog, balanceInfo)
	if err1 == nil {
		JsonReturn(c, 0, "__T_SUCCESS", nil)
		return
	}

	JsonReturn(c, -1, "__T_FAIL", nil)
	return
}

func BatchIpRecharge(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	//staticId := util.StoI(c.DefaultPostForm("static_id", "0")) //购买的长效套餐ID
	ids := c.DefaultPostForm("ids", "") //待续费的ID 格式：staticId:id,staticId1:id1,staticId2:id2

	if len(ids) <= 0 {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}
	list := strings.Split(ids, ",")
	uid := user.Id
	var count int
	for _, val := range list {
		idList := strings.Split(val, ":")
		if len(idList) < 2 {
			continue
		}
		staticId := util.StoI(idList[0])
		id := util.StoI(idList[1])
		if staticId == 0 {
			continue
		}
		// 提取记录
		err_l, ipLog := models.GetIpStaticIpById(id)
		if err_l != nil || ipLog.Id == 0 {
			continue
		}
		_, ipInfo := models.GetStaticIpByIp(ipLog.Ip)
		if ipInfo.Id == 0 || ipInfo.Status != 1 {
			continue
		}
		// 处理IP异常的情况，续费的IP不判断是否已被使用过的IP   20250122
		//if ipInfo.Uid > 0 && ipInfo.Uid != uid {
		//	JsonReturn(c, -1, "__T_IP_HAS_USED", nil)
		//	return
		//}

		err, balanceInfo := models.GetUserStaticIpById(uid, staticId)
		if err != nil || balanceInfo.Id == 0 {
			continue
		} else {
			if balanceInfo.Balance < 1 {
				continue
			}
		}
		// 开始扣费
		_ = models.Recharge(c.ClientIP(), ipLog, balanceInfo)
		count++
	}
	if count > 0 {
		JsonReturn(c, 0, "__T_SUCCESS", nil)
	} else {
		JsonReturn(c, e.ERROR, "__T_IP_BALANCE_LOW", nil)
	}

}

// 删除
// @BasePath /api/v1
// @Summary 删除
// @Description 删除
// @Tags 个人中心 - 静态住宅代理
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param id formData string true "待操作的ID"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /web/static/del [post]
func DelStatic(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	idStr := c.DefaultPostForm("id", "") //待操作的ID
	if idStr == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}
	id := util.StoI(idStr)
	err_l, ipLog := models.GetIpStaticIpById(id)
	if err_l != nil || ipLog.Id == 0 {
		JsonReturn(c, -1, "__T_STATIC_IP_USED_ERROR", nil)
		return
	}
	if ipLog.Uid != user.Id {
		JsonReturn(c, -1, "__T_STATIC_IP_USED_ERROR", nil)
		return
	}
	nowTime := util.GetNowInt()
	if ipLog.ExpireTime > nowTime {
		JsonReturn(c, -1, "__T_NO_DEL", nil)
		return
	}

	err1 := models.DelStaticLog(c.ClientIP(), ipLog)
	if err1 == nil {
		JsonReturn(c, 0, "__T_SUCCESS", nil)
		return
	}

	JsonReturn(c, -1, "__T_FAIL", nil)
	return
}
