package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
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
	_, usedList := models.GetIpStaticIpBy(user.Id, ip, status, orderBy)
	nowTime := util.GetNowInt()
	for _, v := range usedList {
		info := models.ResIpStaticLogModel{}
		is_expire := 1
		if v.ExpireTime < nowTime {
			is_expire = 2
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
