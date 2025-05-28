package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
	"math"
	"strings"
	"time"
)

// @BasePath /api/v1
// @Summary 获取不限量的配置记录
// @Schemes
// @Description 获取不限量的配置记录
// @Tags 不限量
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {array} map[string]interface{} ""
// @Router /center/unlimited/server_record [post]
func GetUserUnlimitedLog(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	flowType := c.DefaultPostForm("flow_type", "2")
	uid := user.Id
	nowTime := util.GetNowInt()
	resLists := []models.ResUserUnlimitedModel{}
	if flowType == "2" {
		logs := models.ListPoolFlowDayByUidAll(uid)
		unlimitedCon := models.GetConfigVal("unlimited_concurrency_unit") //并发单位
		unlimitedBws := models.GetConfigVal("unlimited_bandwidth_unit")   //带宽单位
		if unlimitedCon == "" {
			unlimitedCon = "K"
		}
		if unlimitedBws == "" {
			unlimitedBws = "Mbps"
		}
		for _, log := range logs {
			status := 1 //默认状态为正常
			if log.ExpireTime < nowTime {
				status = 2 //已过期
			}

			exDate := ""
			if log.ExpireTime > 0 {
				exDate = util.GetTimeStr(log.ExpireTime, "d/m/Y H:i:s")
			}

			info := models.ResUserUnlimitedModel{}
			//info.Id 	     = log.Id
			info.ConfigNum = log.Config
			info.BandwidthNum = log.Bandwidth
			info.Config = fmt.Sprintf("%d %s", log.Config, unlimitedCon)
			info.Bandwidth = fmt.Sprintf("%d %s", log.Bandwidth, unlimitedBws)
			info.ExpireTime = exDate
			info.Ip = log.Ip
			info.Status = status

			resLists = append(resLists, info)
		}
	} else {
		logList := models.GetUserFlowDayPortByUid(uid)
		for _, log := range logList {
			exDate := ""
			if log.ExpiredTime > 0 {
				exDate = util.GetTimeStr(log.ExpiredTime, "d/m/Y H:i:s")
			}
			status := 1 //默认状态为正常
			if log.ExpiredTime < nowTime {
				status = 2 //已过期
			}
			info := models.ResUserUnlimitedModel{}
			info.ExpireTime = exDate
			info.Ip = log.Ip
			info.Port = log.Port
			info.Status = status
			resLists = append(resLists, info)
		}
	}
	JsonReturn(c, e.SUCCESS, "success", resLists)
	return
}

// 获取套餐信息->不限量信息数据
func GetUnlimitedPackage(c *gin.Context) {
	config := com.StrTo(c.DefaultPostForm("config", "0")).MustInt()
	bandwidth := com.StrTo(c.DefaultPostForm("bandwidth", "0")).MustInt()
	if config <= 0 {
		config = 200
	}
	if bandwidth <= 0 {
		config = 200
	}

	// 语言套餐信息配置
	lang := DealLanguageUrl(c)

	feeStr := models.GetConfigVal("hb_fee_ratio")
	if feeStr == "" {
		feeStr = "0.05"
	}
	feeRatio := util.StoF(feeStr)
	feeMoneyStr := models.GetConfigVal("hb_fee_max_money")
	if feeMoneyStr == "" {
		feeMoneyStr = "3000"
	}
	feeMoney := util.StoF(feeMoneyStr)

	pakType := "flow_day"
	packageList, ok := models.PackageListMap[pakType]
	if !ok {
		_, packageList = models.GetPackageListFlow(pakType, 0)
	}

	unlimitedConfigList := models.PackageUnlimitedMap //获取无限量配置列表

	configListMap := []int{}                                     //返回并发配置列表
	bandwidthListMap := []int{}                                  //返回带宽配置列表
	configMoneyMap := map[string]float64{}                       //配置信息
	bandwidthMoneyMap := map[string]float64{}                    //带宽信息
	unlimitedListArr := map[int][]models.PackageUnlimitedModel{} //不限量配置信息
	for _, v := range unlimitedConfigList {
		str := fmt.Sprintf("%d_%d", v.PackageId, v.Config)
		if v.Cate == "config" {
			configMoneyMap[str] = v.Money
			if !util.InArrayInt(v.Config, configListMap) {
				configListMap = append(configListMap, v.Config)
			}
		}
		if v.Cate == "bandwidth" {
			bandwidthMoneyMap[str] = v.Money
			if !util.InArrayInt(v.Config, bandwidthListMap) {
				bandwidthListMap = append(bandwidthListMap, v.Config)
			}
		}
		unlimitedListArr[v.PackageId] = append(unlimitedListArr[v.PackageId], v)
	}

	resInfo := []models.UnlimitedPackageListModel{}
	info := models.UnlimitedPackageListModel{}
	if len(packageList) > 0 {
		for _, vInfo := range packageList {
			// 文案配置
			infoDetail := models.PackageTextMap[lang+"_"+util.ItoS(vInfo.Id)]
			corner := vInfo.Corner
			corner2 := vInfo.ActTitle
			actDesc := vInfo.ActDesc
			labels := ""
			if infoDetail.Id > 0 {
				if infoDetail.Corner != "" {
					corner = infoDetail.Corner
				}
				if infoDetail.ActTitle != "" {
					corner2 = infoDetail.ActTitle
				}
				if infoDetail.ActDesc != "" {
					actDesc = infoDetail.ActDesc
				}
				if infoDetail.ActLabel != "" {
					labels = infoDetail.ActLabel
				}
			}
			price := vInfo.Price         //价格
			unit := vInfo.Unit           //单价
			showPrice := vInfo.ShowPrice //原价
			showUnit := vInfo.AllUnit    //原单价
			typeUnit := "Day"
			if vInfo.Day > 1 {
				typeUnit = "Days"
			}

			configMoney := configMoneyMap[fmt.Sprintf("%d_%d", vInfo.Id, config)]
			bandwidthMoney := bandwidthMoneyMap[fmt.Sprintf("%d_%d", vInfo.Id, bandwidth)]

			showPrice = vInfo.ShowPrice + configMoney + bandwidthMoney //计算套餐原单价展示
			oldUnitPrice := showPrice / float64(vInfo.Day)
			if oldUnitPrice > 0 {
				//originUnit = oPrice
				oStr := fmt.Sprintf("%.1f", math.Round(showPrice*10)/10)
				showUnit = util.StoF(oStr)
			}

			price = vInfo.Price + configMoney + bandwidthMoney //计算套餐单价展示

			unitPrice := price / float64(vInfo.Day)
			if unitPrice > 0 {
				//unit = math.Ceil(unitPrice)
				oStr := fmt.Sprintf("%.1f", math.Round(unitPrice*10)/10)
				unit = util.StoF(oStr)
			}

			value := vInfo.Value
			give := vInfo.Give
			gift := vInfo.Gift

			value = int64(value / 86400)
			give = int64(give / 86400)
			gift = int64(gift / 86400)
			subName := vInfo.SubName

			fee := 0.0
			if price < feeMoney { //如果套餐金额小于3000美元，则不收取手续费 来自新版 20240912需求
				fee = math.Floor(price*feeRatio*100) / 100
			} else {
				fee = 0
			}

			info.Id = vInfo.Id
			info.Code = vInfo.Code
			info.Name = vInfo.Name
			info.SubName = subName
			info.Price = price
			info.ShowPrice = showPrice
			info.Corner = corner
			info.ActTitle = corner2
			info.ActDesc = actDesc
			info.ActLabel = labels
			info.Default = vInfo.Default
			info.IsHot = vInfo.IsHot
			info.TypeUnit = typeUnit
			info.Unit = unit
			info.ShowUnit = showUnit
			info.Currency = "$"
			info.Value = int(value)
			info.Give = int(give)
			info.Gift = int(gift)
			info.Day = vInfo.Day
			info.Fee = fee
			info.Total = int(value) + int(give) + int(gift)

			unlimitedList, _ := unlimitedListArr[vInfo.Id]
			configList := []models.ResPackageUnlimited{}
			bandwidthList := []models.ResPackageUnlimited{}
			for _, v := range unlimitedList {
				configInfo := models.ResPackageUnlimited{}
				configInfo.Id = v.Id
				configInfo.Name = fmt.Sprintf("%d %s", v.Config, v.ConfigUnit)
				configInfo.Money = v.Money
				configInfo.Unit = v.Unit
				configInfo.Config = v.Config
				configInfo.ConfigUnit = v.ConfigUnit
				configInfo.Default = v.Default

				isDefault := 0
				if v.Cate == "config" {
					if v.Config == config {
						isDefault = 1
					}
					configInfo.Default = isDefault
					if v.Config > 0 {
						configList = append(configList, configInfo)
					}
				} else {
					if v.Config == bandwidth {
						isDefault = 1
					}
					configInfo.Default = isDefault
					if v.Config > 0 {
						bandwidthList = append(bandwidthList, configInfo)
					}
				}
			}
			info.ConfigList = configList
			info.BandwidthList = bandwidthList
			info.Default = vInfo.Default
			resInfo = append(resInfo, info)
		}
	}

	resData := map[string]interface{}{
		"config_list":    configListMap,
		"bandwidth_list": bandwidthListMap,
		"package_list":   resInfo,
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resData)
	return
}

// 设置不限量预警开关和邮件
func SettingEarlyWarning(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	status := com.StrTo(c.DefaultPostForm("status", "0")).MustInt()
	email := c.DefaultPostForm("email", "")
	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_ERROR", nil)
		return
	}
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	var uew = models.UnlimitedEarlyWarning{Uid: user.Id, Status: status, Email: email}
	uew.GetByUid()
	if uew.Id <= 0 {
		uew.Insert()
	} else {
		uew.Update()
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
}

// 获取不限量预警开关和邮件

func GetEarlyWarning(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	var uew = models.UnlimitedEarlyWarning{Uid: user.Id}
	uew.GetByUid()
	if uew.Id <= 0 {
		uew = models.UnlimitedEarlyWarning{Uid: user.Id, Status: 0, Email: user.Email}
		uew.Insert()
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", uew)
}

// 添加预警

func AddEarlyWarningDetail(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	status := com.StrTo(c.DefaultPostForm("status", "0")).MustInt()
	cpu := com.StrTo(c.DefaultPostForm("cpu", "0")).MustInt()
	memory := com.StrTo(c.DefaultPostForm("memory", "0")).MustInt()
	bandwidth := com.StrTo(c.DefaultPostForm("bandwidth", "0")).MustInt()
	concurrency := com.StrTo(c.DefaultPostForm("concurrency", "0")).MustInt()
	duration := com.StrTo(c.DefaultPostForm("duration", "0")).MustInt()
	instanceMap := c.PostFormMap("instance_data")
	if cpu <= 0 || memory <= 0 || bandwidth <= 0 || concurrency <= 0 || duration <= 0 {
		JsonReturn(c, e.ERROR, "__T_PARAMETERS_ERROR", nil)
		return
	}
	if len(instanceMap) <= 0 {
		JsonReturn(c, e.ERROR, "__T_PARAMETERS_INSTANCE_DATA_ERROR", nil)
		return
	}
	var uew = models.UnlimitedEarlyWarningDetail{Uid: user.Id}
	var now = time.Now().Unix()
	for insId, ip := range instanceMap {
		uew.InstanceId = insId
		uew.GetByUidAndInstanceId()
		if uew.Id > 0 {
			JsonReturn(c, e.ERROR, "__T_REPEAT_ADDITION_INSTANCE_ID", nil)
			return
		}
		uew = models.UnlimitedEarlyWarningDetail{
			Uid:         user.Id,
			Ip:          ip,
			Status:      status,
			InstanceId:  insId,
			Cpu:         cpu,
			Memory:      memory,
			Bandwidth:   bandwidth,
			Concurrency: concurrency,
			Duration:    duration,
			SendTime:    0,
			UpdateTime:  now,
			CreateTime:  now,
		}
		uew.Insert()
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
}

// 修改预警

func ChangeEarlyWarningDetail(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	status := com.StrTo(c.DefaultPostForm("status", "0")).MustInt()
	cpu := com.StrTo(c.DefaultPostForm("cpu", "0")).MustInt()
	memory := com.StrTo(c.DefaultPostForm("memory", "0")).MustInt()
	bandwidth := com.StrTo(c.DefaultPostForm("bandwidth", "0")).MustInt()
	concurrency := com.StrTo(c.DefaultPostForm("concurrency", "0")).MustInt()
	duration := com.StrTo(c.DefaultPostForm("duration", "0")).MustInt()
	id := com.StrTo(c.PostForm("id")).MustInt()
	if cpu <= 0 || memory <= 0 || bandwidth <= 0 || concurrency <= 0 || duration <= 0 {
		JsonReturn(c, e.ERROR, "__T_PARAMETERS_ERROR", nil)
		return
	}
	if id <= 0 {
		JsonReturn(c, e.ERROR, "__T_PARAMETERS_ID_ERROR", nil)
		return
	}
	now := time.Now().Unix()
	uew := models.UnlimitedEarlyWarningDetail{Id: id, Uid: user.Id}
	uew.GetByIdAndUId()
	if uew.Id <= 0 {
		JsonReturn(c, e.ERROR, "__T_PARAMETERS_ID_ERROR", nil)
		return
	}
	uew.Status = status
	uew.Cpu = cpu
	uew.Memory = memory
	uew.Bandwidth = bandwidth
	uew.Concurrency = concurrency
	uew.Duration = duration
	uew.SendTime = 0
	uew.UpdateTime = now
	uew.Update()
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
}

// 删除预警

func DelEarlyWarningDetail(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	id := com.StrTo(c.DefaultPostForm("id", "0")).MustInt()
	uew := models.UnlimitedEarlyWarningDetail{Id: id, Uid: user.Id}
	uew.Delete()
	JsonReturn(c, e.SUCCESS, "__SUCCESS", nil)
}

// 获取预警详情
func GetEarlyWarningDetailList(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uew := models.UnlimitedEarlyWarningDetail{Uid: user.Id}
	list := uew.GetAll()
	JsonReturn(c, e.SUCCESS, "__SUCCESS", list)
}

func GetUnlimitedPortDomain(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	numStr := strings.ToLower(c.DefaultPostForm("num", ""))
	expired := c.DefaultPostForm("expired", "")
	num := util.StoI(numStr)

	// 获取用户不限量端口
	userPortLogList := models.GetUserCanFlowDayPortByUid(user.Id, num, util.StoI(expired))
	hostArr := []ApiProxyJson{}
	for _, log := range userPortLogList {
		if log.ExpiredTime < util.GetNowInt() {
			continue
		}
		//端口 +1000为白名单端口
		jsonInfo := ApiProxyJson{
			Ip:   log.Ip,
			Port: log.Port + 1000,
		}
		hostArr = append(hostArr, jsonInfo)
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", hostArr)
	return
}
