package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"slices"
	"strings"
)

// 静态 -ip详细信息
func GetStaticInfo(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	sn := c.DefaultPostForm("sn", "") //静态 IP列表
	if sn == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}
	idStr := util.MdDecode(sn, MdKey)
	id := util.StoI(idStr)
	_, ipInfo := models.GetStaticIpById(id)
	code := strings.ToLower(ipInfo.Country)
	err, balanceInfo := models.GetUserStaticIpByArea(uid, code)
	if err != nil || balanceInfo.Id == 0 {
		JsonReturn(c, -1, "__T_IP_BALANCE_LOW", nil)
		return
	} else {
		if balanceInfo.Balance < 1 {
			JsonReturn(c, -1, "__T_IP_BALANCE_LOW", nil)
			return
		}
	}
	JsonReturn(c, 0, "__T_SUCCESS", ipInfo)
	return
}

// 静态 IP使用列表
// @BasePath /api/v1
// @Summary 静态 IP使用列表
// @Description 静态 IP使用列表
// @Tags 个人中心 - 静态住宅代理
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param ip formData string false "ip /备注筛选"
// @Param status formData string false "状态"
// @Param field formData string false "排序字段"
// @Param sorter formData string false "排序  1降序"
// @Produce json
// @Success 0 {object} []models.ResIpStaticLogModel{} "成功"
// @Router /web/static/use_list [post]
func GetUsedStaticIpList(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	ip := strings.TrimSpace(c.DefaultPostForm("ip", ""))              //ip /备注筛选
	country := strings.TrimSpace(c.DefaultPostForm("country", ""))    //国家
	status := strings.TrimSpace(c.DefaultPostForm("status", ""))      //状态
	field := strings.TrimSpace(c.DefaultPostForm("field", "id"))      //	排序字段
	sorterType := strings.TrimSpace(c.DefaultPostForm("sorter", "1")) //排序  1降序
	if field == "" {
		field = "id"
	}
	if sorterType == "" {
		sorterType = "1"
	}
	sorter := " desc"
	if sorterType == "1" {
		sorter = " desc"
	} else {
		sorter = " asc"
	}
	orderBy := field + sorter
	logData := []models.ResIpStaticLogModel{}
	_, usedList := models.GetIpStaticIpBy(user.Id, ip, status, orderBy, country)

	// 新版IP资源状态
	ipStatusArr := []string{}
	for _, v := range usedList {
		if v.IsNew == 1 {
			ipStatusArr = append(ipStatusArr, v.Ip)
		}
	}
	offlineIps := []string{}
	if len(ipStatusArr) > 0 {
		stRes, stMsg, lists := StaticZtStatus(ipStatusArr)
		fmt.Println(stMsg)
		if stRes == true {
			for _, v := range lists {
				if v.Status == 3 {
					offlineIps = append(offlineIps, v.Ip)
				}
			}
		}
	}
	nowTime := util.GetNowInt()
	for _, v := range usedList {
		info := models.ResIpStaticLogModel{}
		is_expire := 1
		is_replace := 0
		if v.ExpireTime < nowTime {
			is_expire = 2
			if slices.Contains(offlineIps, v.Ip) {
				is_expire = 4 // 过期且已下线 不返回
			}
		} else {
			// 检查当前 IP 是否在 offlineIps 列表中
			if slices.Contains(offlineIps, v.Ip) {
				is_expire = 3 // 已下线
			}
			//十分钟内 没有替换过 IP 可以替换
			if nowTime-v.CreateTime <= 600 && v.Replaced == 0 {
				is_replace = 1
			}
		}
		info.Id = v.Id
		info.Ip = v.Ip
		info.Port = v.Port
		info.Country = v.Country
		info.State = v.State
		info.City = v.City
		info.Account = v.Account
		info.Password = v.Password
		info.Remark = v.Remark
		info.IsExpire = is_expire
		info.IsReplace = is_replace
		info.ExpireTime = util.GetTimeStr(v.ExpireTime, "d/m/Y")
		info.CreateTime = util.GetTimeStr(v.CreateTime, "d/m/Y")
		if v.Port > 0 {
			logData = append(logData, info)
		}
	}
	JsonReturn(c, 0, "__T_SUCCESS", logData)
	return
}

// 静态 IP使用列表
// @BasePath /api/v1
// @Summary 静态 IP使用列表
// @Description 静态 IP使用列表
// @Tags 个人中心 - 静态住宅代理
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param ip formData string false "ip /备注筛选"
// @Param status formData string false "状态"
// @Param field formData string false "排序字段"
// @Param sorter formData string false "排序  1降序"
// @Produce json
// @Success 0 {object} []models.ResIpStaticLogModel{} "成功"
// @Router /web/static/use_list [post]
func GetUsedStaticIpListBak(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	ip := strings.TrimSpace(c.DefaultPostForm("ip", ""))              //ip /备注筛选
	country := strings.TrimSpace(c.DefaultPostForm("country", ""))    //国家
	status := strings.TrimSpace(c.DefaultPostForm("status", ""))      //状态
	field := strings.TrimSpace(c.DefaultPostForm("field", "id"))      //	排序字段
	sorterType := strings.TrimSpace(c.DefaultPostForm("sorter", "1")) //排序  1降序
	if field == "" {
		field = "id"
	}
	if sorterType == "" {
		sorterType = "1"
	}
	sorter := " desc"
	if sorterType == "1" {
		sorter = " desc"
	} else {
		sorter = " asc"
	}
	orderBy := field + sorter
	logData := []models.ResIpStaticLogModel{}

	offlineIps := models.GetStaticOfflineIps() //下线IP列表
	offlineIpList := make([]string, len(offlineIps))
	for i, ipModel := range offlineIps {
		offlineIpList[i] = ipModel.Ip
	}

	_, usedList := models.GetIpStaticIpBy(user.Id, ip, status, orderBy, country)
	nowTime := util.GetNowInt()
	for _, v := range usedList {
		info := models.ResIpStaticLogModel{}
		is_expire := 1
		is_replace := 0
		if v.ExpireTime < nowTime {
			is_expire = 2
			if slices.Contains(offlineIpList, v.Ip) {
				is_expire = 4 // 过期且已下线 不返回
			}
		} else {
			// 检查当前 IP 是否在 offlineIps 列表中
			if slices.Contains(offlineIpList, v.Ip) {
				is_expire = 3 // 已下线
			}
			//十分钟内 没有替换过 IP 可以替换
			if nowTime-v.CreateTime <= 600 && v.Replaced == 0 {
				is_replace = 1
			}
		}
		info.Id = v.Id
		info.Ip = v.Ip
		info.Port = v.Port
		info.Country = v.Country
		info.State = v.State
		info.City = v.City
		info.Account = v.Account
		info.Password = v.Password
		info.Remark = v.Remark
		info.IsExpire = is_expire
		info.IsReplace = is_replace
		info.ExpireTime = util.GetTimeStr(v.ExpireTime, "d/m/Y")
		info.CreateTime = util.GetTimeStr(v.CreateTime, "d/m/Y")
		if v.Port > 0 {
			logData = append(logData, info)
		}
	}
	JsonReturn(c, 0, "__T_SUCCESS", logData)
	return
}

// 静态IP  提取记录
// @BasePath /api/v1
// @Summary 静态IP  提取记录
// @Description 静态IP  提取记录
// @Tags 个人中心 - 静态住宅代理
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param start_date formData string false "开始日期"
// @Param end_date formData string false "结束日期"
// @Produce json
// @Success 0 {object} []models.ResStaticRecordModel{} "成功"
// @Router /web/static/use_record [post]
func GetUsedStaticRecord(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	start_date := c.DefaultPostForm("start_date", "")
	end_date := c.DefaultPostForm("end_date", "")
	var start, end int
	if start_date != "" && end_date != "" {
		start = util.StoI(util.GetTimeStamp(start_date, "Y-m-d"))
		end = util.StoI(util.GetTimeStamp(end_date, "Y-m-d"))
	}

	logData := []models.ResStaticRecordModel{}
	_, usedList := models.GetUsedByLog(user.Id, start, end, 0)

	_, usedList2 := models.GetUsedByLog(user.Id, start, end, 1)
	for _, v := range usedList2 {
		usedList = append(usedList, v)
	}
	nowTime := util.GetNowInt()
	idNum := 0
	for _, v := range usedList {
		idNum = idNum + 1
		info := models.ResStaticRecordModel{}
		is_expire := 1
		if v.ExpireTime < nowTime {
			is_expire = 2
		}
		info.Id = idNum
		info.Ip = v.Ip
		info.Port = v.Port
		info.Country = v.Country
		info.State = v.State
		info.City = v.City
		info.Remark = v.Remark
		info.IsExpire = is_expire
		info.ExpireTime = util.GetTimeStr(v.ExpireTime, "d/m/Y")
		info.CreateTime = util.GetTimeStr(v.CreateTime, "d/m/Y")
		if v.Port > 0 {
			logData = append(logData, info)
		}
	}

	//sort.SliceStable(logData, func(i, j int) bool {
	//	return logData[i].Id > logData[j].Id
	//})
	JsonReturn(c, 0, "__T_SUCCESS", logData)
	return
}

// 使用记录下载
// @BasePath /api/v1
// @Summary 使用记录下载
// @Description 使用记录下载
// @Tags 个人中心 - 静态住宅代理
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param start_date formData string false "开始日期"
// @Param end_date formData string false "结束日期"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /web/static/use_download [post]
func UsedStaticRecordDownload(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	start_date := c.DefaultPostForm("start_date", "")
	end_date := c.DefaultPostForm("end_date", "")
	var start, end int
	if start_date != "" && end_date != "" {
		start = util.StoI(util.GetTimeStamp(start_date, "Y-m-d"))
		end = util.StoI(util.GetTimeStamp(end_date, "Y-m-d"))
	}
	uid := user.Id

	title := []string{"Serial Number", "Extract IP", "Country/Region", "State", "City", "Extraction Time"}

	if uid > 0 {
		csvData := [][]string{}
		csvData = append(csvData, title)

		numId := 0
		_, usedList := models.GetUsedByLog(uid, start, end, 0)

		_, usedList2 := models.GetUsedByLog(uid, start, end, 1)

		for _, v := range usedList {
			info := []string{}
			numId = numId + 1
			info = append(info, util.ItoS(numId))
			info = append(info, v.Ip)
			info = append(info, v.Country)
			info = append(info, v.State)
			info = append(info, v.City)
			info = append(info, util.GetTimeStr(v.CreateTime, "d-m-Y"))
			csvData = append(csvData, info)
		}

		for _, v := range usedList2 {
			info := []string{}
			numId = numId + 1
			info = append(info, util.ItoS(numId))
			info = append(info, v.Ip)
			info = append(info, v.Country)
			info = append(info, v.State)
			info = append(info, v.City)
			info = append(info, util.GetTimeStr(v.CreateTime, "d-m-Y"))
			csvData = append(csvData, info)
		}

		err := DownloadCsv(c, "UsedRecord", csvData)
		fmt.Println(err)
		//if err != nil {
		//	JsonReturn(c, e.ERROR, err.Error(), nil)
		//	return
		//}
	}
	return
}

// @BasePath /api/v1
// @Summary 更换静态IP
// @Schemes
// @Description 更换静态IP
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Param used_id formData string true "待更换的id"
// @Param sn formData string true "新的IPsn"
// @Param type formData string true "类型 // offline : 下线更换，online : 在线更换"
// @Produce json
// @Success 0 {array} map[string]interface{}
// @Router /center/static/change [post]
func ChangeStaticIp(c *gin.Context) {
	idStr := c.DefaultPostForm("used_id", "") //待操作的ID
	sn := c.DefaultPostForm("sn", "")
	typeStr := c.DefaultPostForm("type", "offline") //类型 // offline : 下线更换，online : 在线更换

	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	if idStr == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}
	used_id := util.StoI(idStr)
	err_l, ipLog := models.GetIpStaticIpById(used_id)
	if err_l != nil || ipLog.Id == 0 {
		JsonReturn(c, -1, "__T_STATIC_IP_USED_ERROR", nil)
		return
	}

	nowTime := util.GetNowInt()
	if ipLog.ExpireTime < nowTime {
		JsonReturn(c, -1, "__T_STATIC_IP_USED_ERROR", nil)
		return
	}
	if typeStr == "online" {
		if nowTime-ipLog.CreateTime > 600 || ipLog.Replaced == 1 {
			JsonReturn(c, -1, "__T_STATIC_IP_USED_ERROR", nil)
			return
		}
	}

	if sn == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR--IP", nil)
		return
	}

	ipStr := util.MdDecode(sn, MdKey)
	if ipStr == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR--IP", gin.H{"err": 01})
		return
	}
	err_check, existingLog := models.GetIpStaticIp(used_id, ipStr)
	if err_check == nil && existingLog.Id > 0 {
		JsonReturn(c, e.ERROR, "__T_IP_ALREADY_IN_USE", nil)
		return
	}
	regionSn := ""
	orderId := ""
	if ipLog.IsNew == 1 { //新资源中台
		// 释放资源
		relRes, relMsg := StaticZtRelease(userInfo.Id, ipLog.Ip)
		if relRes == false {
			JsonReturn(c, e.ERROR, relMsg, nil)
			return
		}
		regionSn = ipLog.Code
		orderId = ipLog.OrderId
	} else {
		// region ip地区
		regionInfo := models.GetStaticRegionBy(ipLog.Country, ipLog.State, ipLog.City)
		regionSn = regionInfo.RegionSn
		orderId = "9222" + util.GetOrderIds() //更换开通
	}

	//开通新资源
	durationTime := ipLog.ExpireTime - util.GetNowInt()
	if durationTime < 0 {
		durationTime = 0
	}
	durationTime = durationTime + 2*24*3600
	openRes, openMsg := StaticZtOpen(userInfo.Id, durationTime, ipStr, orderId, regionSn)
	if openRes == false {
		JsonReturn(c, e.ERROR, openMsg, gin.H{"err": 01})
		return
	}

	models.AddDelStaticLog(c.ClientIP(), ipLog) //记录日志

	data := map[string]interface{}{}
	data["ip"] = ipStr
	data["port"] = ipLog.Port
	data["state"] = ipLog.State
	data["city"] = ipLog.City
	data["country"] = ipLog.Country
	data["code"] = ipLog.Code
	data["status"] = 1
	if typeStr == "offline" {
		data["expire_time"] = ipLog.ExpireTime + 2*24*3600 // 下线更换，续费两天
	} else {
		//十分钟内 没有替换过 IP 可以替换
		if nowTime-ipLog.CreateTime <= 600 && ipLog.Replaced == 0 {
			data["expire_time"] = ipLog.ExpireTime
			data["replaced"] = 1 // 在线更换，标记为已更换
		}
	}
	err := models.SetIpStaticIp(ipLog.Id, data)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_IP_CHANGE_FAILED", nil)
		return
	}
	JsonReturn(c, e.SUCCESS, "__T_IP_CHANGE_SUCCESS", gin.H{})
	return
}

func ChangeStaticIpBak(c *gin.Context) {
	idStr := c.DefaultPostForm("used_id", "") //待操作的ID
	sn := c.DefaultPostForm("sn", "")
	typeStr := c.DefaultPostForm("type", "offline") //类型 // offline : 下线更换，online : 在线更换

	resCode, msg, _ := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	if idStr == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}
	used_id := util.StoI(idStr)
	err_l, ipLog := models.GetIpStaticIpById(used_id)
	if err_l != nil || ipLog.Id == 0 {
		JsonReturn(c, -1, "__T_STATIC_IP_USED_ERROR", nil)
		return
	}

	nowTime := util.GetNowInt()
	if ipLog.ExpireTime < nowTime {
		JsonReturn(c, -1, "__T_STATIC_IP_USED_ERROR", nil)
		return
	}
	if typeStr == "online" {
		if nowTime-ipLog.CreateTime > 600 || ipLog.Replaced == 1 {
			JsonReturn(c, -1, "__T_STATIC_IP_USED_ERROR", nil)
			return
		}
	}

	if sn != "" {
		idStr := util.MdDecode(sn, MdKey)
		id := util.StoI(idStr)
		_, ipInfo := models.GetStaticIpById(id)

		models.AddDelStaticLog(c.ClientIP(), ipLog) //记录日志

		data := map[string]interface{}{}
		data["ip"] = ipInfo.Ip
		data["port"] = ipInfo.Port
		data["state"] = ipInfo.State
		data["city"] = ipInfo.City
		data["country"] = ipInfo.Country
		data["code"] = ipInfo.Code
		if typeStr == "offline" {
			data["expire_time"] = ipLog.ExpireTime + 2*24*3600 // 下线更换，续费两天
		} else {
			//十分钟内 没有替换过 IP 可以替换
			if nowTime-ipLog.CreateTime <= 600 && ipLog.Replaced == 0 {
				data["expire_time"] = ipLog.ExpireTime
				data["replaced"] = 1 // 在线更换，标记为已更换
			}
		}
		err := models.SetIpStaticIp(ipLog.Id, data)
		if err != nil {
			JsonReturn(c, e.ERROR, "__T_IP_CHANGE_FAILED", nil)
			return
		}
		JsonReturn(c, e.SUCCESS, "__T_IP_CHANGE_SUCCESS", gin.H{})
		return
	}

}

// @BasePath /api/v1
// @Summary 检测ip是否可以替换
// @Schemes
// @Description 检测ip是否可以替换
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param ip formData string true "IP地址"
// @Produce json
// @Success 0 {array} map[string]interface{}
// @Router /web/static/check_replace [post]
func CheckReplace(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	nowTime := util.GetNowInt()
	ip := c.DefaultPostForm("ip", "")
	is_replace := 0
	if ip != "" {
		// 提取记录
		err_l, ipLog := models.GetIpStaticIp(uid, ip)
		if err_l == nil && ipLog.Id > 0 {
			//十分钟内 没有替换过 IP 可以替换
			if nowTime-ipLog.CreateTime <= 600 && ipLog.Replaced == 0 {
				is_replace = 1
			}
		}
	}
	data := map[string]interface{}{}
	data["is_replace"] = is_replace
	JsonReturn(c, 0, "__T_SUCCESS", data)
}

func CheckKyc(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	needKyc := false

	// 检查是否需要实名认证（上线后注册的用户）
	if models.CheckUserNeedKyc(user.CreateTime) {
		kycStatus := models.CheckUserKycStatus(uid)
		if kycStatus != 1 { // 未实名认证
			needKyc = true
		}
	}

	JsonReturn(c, 0, "__T_SUCCESS", needKyc)
}
