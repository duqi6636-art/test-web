package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
)

// 获取有效期流量套餐
// @BasePath /api/v1
// @Summary 获取有效期流量套餐
// @Description 获取有效期流量套餐
// @Tags 套餐页
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录信息"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {array} map[string]interface{} "fresh：最后更新时间，flow：流量套餐列表（值为：map[string][]models.ResIpPackageFlow{}模型），flow_agent：代理流量套餐列表（值为：map[string][]models.ResIpPackageFlow{}模型）"
// @Router /web/package/flow [post]
func GetPackageFlow(c *gin.Context) {
	sessionId := c.DefaultPostForm("session", "")
	lang := strings.ToLower(c.DefaultPostForm("lang", "en"))
	if lang == "" {
		lang = "en"
	}
	if lang == "zh-tw" || lang == "zh" || lang == "tw" || lang == "zh-cn" || lang == "cn" {
		lang = "zh-tw"
	}

	isPay := 0
	uid := 0
	if sessionId != "" {
		_, uid = GetUIDbySession(sessionId)
		_, user := models.GetUserById(uid)
		if user.IsPay == "true" {
			isPay = 1
		}
	}
	resList, newTime := dealPackage("flow", lang, isPay, uid)
	resAgentList, newTime2 := dealFlowAgentPackage(lang, isPay, uid)
	if newTime < newTime2 {
		newTime = newTime2
	}
	resInfo := map[string]interface{}{}
	resInfo["fresh"] = newTime
	resInfo["flow"] = resList
	resInfo["flow_agent"] = resAgentList

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resInfo)
	return
}

// 获取新用户5G流量套餐列表
func GetPackageNewFlowList(c *gin.Context) {
	_, packageList := models.GetNewPackageFlowList()
	//for i, cmPackage := range packageList {
	//
	//	if cmPackage.Id == 104 {
	//		packageList[i].Name = "5GB"
	//		break
	//	}
	//}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", packageList)
	return
}

// 获取万圣节活动套餐
func GetHalloweenActivityPackages(c *gin.Context) {
	lang := strings.ToLower(c.DefaultPostForm("lang", "en"))
	if lang == "" {
		lang = "en"
	}
	if lang == "zh-tw" || lang == "zh" || lang == "tw" || lang == "zh-cn" || lang == "cn" {
		lang = "zh-tw"
	}

	resList := dealHalloweenActivityPackage(lang)

	resInfo := map[string]interface{}{}
	resInfo["flow"] = resList

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resInfo)
	return
}

func GetHalloweenEnabled(c *gin.Context) {
	sessionId := c.DefaultPostForm("session", "")
	packageIdStr := c.DefaultPostForm("packageId", "")
	var user models.Users
	uid := 0
	if sessionId != "" {
		_, uid = GetUIDbySession(sessionId)
		_, user = models.GetUserById(uid)
	}

	/// 获取套餐信息
	packageId := util.StoI(packageIdStr)
	packageInfo := models.GetPackageInfoById(packageId)

	/// 只能购买一次套餐限制
	if user.IsPay == "true" { // 已付费用户

		if packageInfo.UseType == 1 { // 已付费用户不能购买新用户套餐

			JsonReturn(c, e.ERROR, "__T_ONLY_NEW_USER", nil)
			return
		} else if packageInfo.BuyTimes == 1 { // 活动套餐限制

			// 判断是否购买过此套餐
			count := models.GetOrderCountWith(uid, packageId)
			if count > 0 {
				JsonReturn(c, e.ERROR, "__T_ONLY_BUY_ONE", nil)
				return
			}
		}
	} else { // 未付费用户

		if packageInfo.UseType == 2 { // 未付费用户不能购买老用户套餐

			JsonReturn(c, e.ERROR, "__T_ONLY_OLD_USER", nil)
			return
		}
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", packageInfo)

}

// @BasePath /api/v1
// @Summary 获取动态ISP套餐
// @Description 获取动态ISP套餐
// @Tags 支付-套餐
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录信息"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {array} map[string]interface{} "fresh：最后更新时间，dynamic_isp：套餐列表（值为：map[string][]models.ResIpPackageFlow{}模型"
// @Router /center/package/dynamic_isp [post]
func GetDynamicISPPackageList(c *gin.Context) {
	sessionId := c.DefaultPostForm("session", "")
	is_pay := 0
	if sessionId != "" {
		_, uid := GetUIDbySession(sessionId)
		_, user := models.GetUserById(uid)
		if user.IsPay == "true" {
			is_pay = 1
		}
	}
	resList := getFlowLists(c, "dynamic_isp", is_pay)
	resInfo := map[string]interface{}{}
	resInfo["dynamic_isp"] = resList
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resInfo)
	return
}

// @BasePath /api/v1
// @Summary 获取不限量流量套餐列表
// @Description 获取不限量流量套餐列表
// @Tags 支付-套餐
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录信息"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {array} map[string]interface{} "fresh：最后更新时间，flow_day：套餐列表（值为：map[string][]models.ResIpPackageFlow{}模型"
// @Router /center/package/flow_day [post]
func GetFlowDayPackageList(c *gin.Context) {
	config := com.StrTo(c.DefaultPostForm("config", "0")).MustInt()
	bandwidth := com.StrTo(c.DefaultPostForm("bandwidth", "0")).MustInt()
	if config <= 0  {
		config = 200
	}
	if bandwidth <= 0  {
		bandwidth = 200
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

	configListMap := []int{} //返回并发配置列表
	bandwidthListMap := []int{} //返回带宽配置列表
	configMoneyMap := map[string]float64{} //配置信息
	bandwidthMoneyMap := map[string]float64{} //带宽信息
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

			price := vInfo.Price		//价格
			unit := vInfo.Unit			//单价
			showPrice := vInfo.ShowPrice  //原价
			showUnit := vInfo.AllUnit	//原单价

			typeUnit := "Day"
			if vInfo.Day > 1 {
				typeUnit = "Days"
			}

			configMoney := configMoneyMap[fmt.Sprintf("%d_%d", vInfo.Id, config)]
			bandwidthMoney := bandwidthMoneyMap[fmt.Sprintf("%d_%d", vInfo.Id, bandwidth)]

			showPrice = vInfo.ShowPrice + configMoney + bandwidthMoney  //计算套餐原单价展示
			oldUnitPrice := showPrice / float64(vInfo.Day)
			if oldUnitPrice > 0 {
				//originUnit = oPrice
				oStr := fmt.Sprintf("%.1f", math.Round(showPrice * 10) / 10)
				showUnit = util.StoF(oStr)
			}
			fmt.Println("showPrice:", showPrice)
			fmt.Println("configMoney:", configMoney)
			fmt.Println("bandwidthMoney:", bandwidthMoney)

			price = vInfo.Price + configMoney + bandwidthMoney  //计算套餐单价展示
			unitPrice := price / float64(vInfo.Day)
			fmt.Println("price:", price)
			if unitPrice > 0 {
				//unit = math.Ceil(unitPrice)
				oStr := fmt.Sprintf("%.1f", math.Round(unitPrice * 10) / 10)
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
			if vInfo.Price < feeMoney { //如果套餐金额小于3000美元，则不收取手续费 来自新版 20240912需求
				fee = math.Floor(vInfo.Price*feeRatio*100) / 100
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

			unlimitedList,_ := unlimitedListArr[vInfo.Id]
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
		"config_list":  configListMap,
		"bandwidth_list":  bandwidthListMap,
		"flow_day": resInfo,
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resData)
	return
}

// @BasePath /api/v1
// @Summary 获取ip套餐列表
// @Description 获取ip套餐列表
// @Tags 支付-套餐
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录信息"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {array} map[string]interface{} "isp：套餐列表（值为：map[string][]models.ResSocksIpPackage{}模型 agent：套餐列表（值为：map[string][]models.ResSocksIpPackage{}模型"
// @Router /center/package/socks5 [post]
func GetSocks5PackageList(c *gin.Context) {
	sessionId := c.DefaultPostForm("session", "")
	is_pay := 0
	uid := 0
	if sessionId != "" {
		_, uid = GetUIDbySession(sessionId)
		_, user := models.GetUserById(uid)
		if user.IsPay == "true" {
			is_pay = 1
		}
	}
	resList := dealPackageList(c, "isp", is_pay, uid)
	resAgentList := getFlowLists(c, "agent", is_pay)
	resInfo := map[string]interface{}{}
	resInfo["isp"] = resList
	resInfo["agent"] = resAgentList

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resInfo)
	return
}

// @BasePath /api/v1
// @Summary 获取静态套餐列表
// @Description 获取静态套餐列表
// @Tags 支付-套餐
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录信息"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {array} map[string]interface{} "long_package：套餐列表（值为：map[string][]models.ResIpPackageLong{}模型 long_area：地区列表（值为：map[string][]models.ResPackageAreaInfo{}模型"
// @Router /center/package/static [post]
func GetStaticPackage(c *gin.Context) {
	resCode, msg, _ := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	lang := strings.ToLower(c.DefaultPostForm("lang", "en"))

	allPackageList, err := models.AllPackageList()
	PackageListMap := map[string][]models.CmPackage{}
	PayedPackageListMap := map[string][]models.CmPackage{}
	PackageTextMap := map[string]models.CmPackageInfo{}
	PackageAreaTextMap := map[string]models.CmPackageInfo{}
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range allPackageList {
		str := v.PakType
		if v.Pid == 0 {
			PackageListMap[str] = append(PackageListMap[str], v)
		} else {
			PayedPackageListMap[str] = append(PayedPackageListMap[str], v)
		}
	}
	// 套餐文案 // 语言套餐信息配置
	errInfo, packageInfoList := models.GetPackageInfoList()
	if errInfo != nil {
		log.Fatal(errInfo)
	}
	for _, v := range packageInfoList {
		str := v.Lang + "_" + util.ItoS(v.PackageId)
		PackageTextMap[str] = v
		strArea := v.Lang + "_" + util.ItoS(v.PackageId) + "_" + util.ItoS(v.AreaId)
		PackageAreaTextMap[strArea] = v
	}

	pak_type := "static"
	packageList, _ := PackageListMap[pak_type]

	// 套餐列表
	list := []models.ResIpPackageLong{}
	defaultId := 0
	for _, v := range packageList {
		// 文案配置
		infoDetail := PackageTextMap[lang+"_"+util.ItoS(v.Id)]
		corner := v.Corner
		if infoDetail.Id > 0 {
			if infoDetail.Corner != "" {
				corner = infoDetail.Corner
			}
		}
		info := models.ResIpPackageLong{}
		info.Id = v.Id
		info.Pid = v.Pid
		info.Code = v.Code
		info.Name = v.Name
		info.Number = v.Day
		info.Corner = corner
		info.Default = v.Default
		info.Unit = v.Unit
		info.Price = v.Price
		info.IsHot = v.IsHot
		list = append(list, info)
		if v.Default == "true" {
			defaultId = v.Id
		}
	}

	// 查询地区价格
	packagePackList := models.GetPackageAreaList()

	packageAreaList := []models.CmPackageAreaInfo{}
	areaPriceMap := map[string]models.PackageContinentPrice{}
	for _, v := range packagePackList {
		pakId := util.ItoS(v.PackageId)
		pInfo := models.PackageContinentPrice{}
		pInfo.Price = v.Money
		pInfo.Unit = v.Unit
		pInfo.PackageId = v.PackageId
		str := pakId + "__" + v.Country
		areaPriceMap[str] = pInfo
		//默认地区的价格
		if v.PackageId == defaultId {
			packageAreaList = append(packageAreaList, v)
		}
	}
	//剩余IP数
	countryIpNum := map[string]int{}
	countryIpList := models.GetStaticCountryList()
	for _, v := range countryIpList {
		countryIpNum[v.Code] = v.IpNumber
	}

	countryList := models.GetStaticCountryListByLang(lang)
	countryNameMap := map[string]string{}
	for _, v := range countryList {
		countryNameMap[v.Code] = v.Name
	}
	defaultPackList := []models.ResPackageAreaInfo{}
	for _, v := range packageAreaList {
		// 文案配置
		infoDetail := PackageAreaTextMap[lang+"_"+util.ItoS(v.PackageId)+"_"+util.ItoS(v.Id)]
		corner := ""
		if infoDetail.Id > 0 {
			if infoDetail.Corner != "" {
				corner = infoDetail.Corner
			}
		}
		priceArr := []models.PackageContinentPrice{}
		for _, vp := range packageList {
			str := util.ItoS(vp.Id) + "__" + v.Country
			priceInfo, p_ok := areaPriceMap[str]
			if !p_ok {
				priceInfo = models.PackageContinentPrice{}
			}
			priceArr = append(priceArr, priceInfo)
		}
		countryName, ok := countryNameMap[strings.ToLower(v.Country)]
		if !ok {
			countryName = v.CountryName
		}
		ipNumber, okn := countryIpNum[strings.ToLower(v.Country)]
		if !okn {
			ipNumber = 199
		}
		info := models.ResPackageAreaInfo{}
		info.Id = v.Id
		info.PackageId = v.PackageId
		info.Area = v.Area
		info.CountryName = countryName
		info.Country = strings.ToUpper(v.Country)
		info.CountryImg = v.CountryImg
		info.Money = v.Money
		info.Unit = v.Unit
		info.Default = v.Default
		info.IsHot = v.IsHot
		info.Sort = v.Sort
		info.IpNumber = ipNumber
		info.Corner = corner
		info.PriceArr = priceArr
		defaultPackList = append(defaultPackList, info)
	}

	data := make(map[string]interface{})
	data["long_package"] = list
	data["long_area"] = defaultPackList
	JsonReturn(c, 0, "__T_SUCCESS", data)
	return
}

// 处理套餐 ISP
func dealPackageList(c *gin.Context, pak_type string, isOld, uid int) []models.ResSocksIpPackage {
	err, packageList := models.GetPackageListFlow(pak_type, isOld)
	lang := DealLanguageUrl(c)
	langUrl := "/" + lang + "/"
	if lang == "en" || lang == "" {
		langUrl = "/"
	}

	// 获取用户拥有的所有优惠券
	_, availableCoupons := models.GetAvailableCouponListByUid(uid)
	list := []models.ResSocksIpPackage{}
	num := len(packageList)
	key := 0
	if err == nil && num > 0 {
		key = num - 1
		for _, v := range packageList {
			// 文案配置
			infoDetail := models.PackageTextMap[lang+"_"+util.ItoS(v.Id)]
			corner := v.Corner
			corner2 := v.ActTitle
			actDesc := v.ActDesc
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

			info := models.ResSocksIpPackage{}
			info.Id = v.Id
			info.Pid = v.Pid
			info.Code = v.Code
			info.Name = v.Name
			info.SubName = v.SubName //副标题
			info.Number = v.Value
			info.Gift = v.Gift
			info.Give = v.Give
			info.Discount = v.Discount
			info.Price = v.Price
			info.Default = v.Default
			info.IsHot = v.IsHot
			info.Corner = corner
			info.ActTitle = corner2
			info.ActLabel = labels
			info.ActDesc = actDesc
			info.Unit = v.Unit
			info.AllUnit = v.AllUnit
			info.Currency = v.Currency
			info.TotalNum = num
			info.TotalKey = key
			info.Lang = langUrl
			/// 如果有优惠券替换优惠券角标
			for _, coupon := range availableCoupons {

				if strings.Contains(coupon.Meals, strconv.Itoa(info.Id)) {
					cron := coupon.Cron
					info.CouponLabel = cron
					break
				}
			}

			list = append(list, info)
		}
	}
	return list
}

// 处理套餐   动态流量   不限量  ISP代理商
func getFlowLists(c *gin.Context, pak_type string, isOld int) []models.ResIpPackageFlow {
	packageList, ok := models.PackageListMap[pak_type]
	if isOld == 1 {
		payedPackageList, ok := models.PayedPackageListMap[pak_type]
		if ok && len(payedPackageList) > 0 {
			packageList = payedPackageList
		}
	}
	// 语言套餐信息配置
	lang := DealLanguageUrl(c)

	list := []models.ResIpPackageFlow{}
	num := len(packageList)
	if ok && num > 0 {
		contentConf := models.GetConfigVal("FlowConfigText")
		for _, v := range packageList {
			value := int64(v.Value / 1024 / 1024 / 1024)
			info := models.ResIpPackageFlow{}
			info.Id = v.Id
			info.Pid = v.Pid
			info.Code = v.Code
			info.Name = v.Name
			info.SubName = v.SubName //副标题
			info.Gift = int64(v.Gift / 1024 / 1024 / 1024)
			info.Give = int64(v.Give / 1024 / 1024 / 1024)
			info.Number = v.Day
			info.Value = value
			info.Total = value + int64(v.Give/1024/1024/1024)
			info.Price = v.Price
			info.Unit = v.Unit
			info.AllUnit = v.AllUnit
			info.Default = v.Default
			info.Currency = v.Currency
			info.TotalNum = num
			corner := v.Corner
			corner2 := v.ActTitle
			actDesc := v.ActDesc
			labels := ""
			// 文案配置
			//infoDetail := packInfo[v.Id]
			infoDetail := models.PackageTextMap[lang+"_"+util.ItoS(v.Id)]
			content := contentConf
			if infoDetail.Id > 0 {
				if infoDetail.Content != "" {
					content = infoDetail.Content
				}
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
			textArr := util.Split(content, "|")
			info.Corner = corner
			info.TextArr = textArr
			info.ActTitle = corner2
			info.ActDesc = actDesc
			info.ActLabel = labels

			list = append(list, info)
		}
	}
	return list
}

// 处理语言信息
func DealLanguageUrl(c *gin.Context) string {
	resLang := "en"
	lang := c.DefaultPostForm("lang", "")
	if lang == "zh-tw" || lang == "zh" || lang == "tw" || lang == "zh-cn" || lang == "cn" {
		lang = "zh-tw"
	}
	confLang := models.GetConfigVal("WEB_LANGUAGE")
	confLangMap := util.Split(confLang, ",")
	if lang != "" && util.InArrayString(lang, confLangMap) {
		resLang = lang
	}
	return resLang
}

// @BasePath /api/v1
// @Summary 获取ip套餐列表
// @Description 获取ip套餐列表
// @Tags 支付-套餐
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录信息"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {array} map[string]interface{} "isp：套餐列表（值为：map[string][]models.ResSocksIpPackage{}模型 agent：套餐列表（值为：map[string][]models.ResSocksIpPackage{}模型"
// @Router /center/package/low_price [post]
func GetLowPrice(c *gin.Context) {
	price := map[string]float64{}
	price["flow_person"] = models.GetLowPrice("flow")
	price["flow_business"] = models.GetLowPrice("flow_agent")
	price["isp_person"] = models.GetLowPrice("isp")
	price["isp_business"] = models.GetLowPrice("agent")
	price["dynamic_isp"] = models.GetLowPrice("dynamic_isp")
	price["flow_day"] = models.GetLowPrice("flow_day")
	price["static_7day"] = models.GetStaticLowPrice(7)
	price["static_30day"] = models.GetStaticLowPrice(30)
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", price)
	return
}

// 处理套餐 流量
func dealPackage(pakType, lang string, isOld int, uid int) (map[string][]models.ResIpPackageFlow, int) {
	// 获取用户拥有的所有优惠券
	_, availableCoupons := models.GetAvailableCouponListByUid(uid)
	// 获取套餐的多语言文案
	contentConf := models.GetConfigVal("FlowConfigText") //配置默认文案
	PackageTextMap := map[string]models.CmPackageInfo{}
	_, packageInfoList := models.GetPackageInfoList()
	for _, v := range packageInfoList {
		str := v.Lang + "_" + util.ItoS(v.PackageId)
		PackageTextMap[str] = v
	}
	newTime := 0
	_, packageList := models.GetPackageListFlow(pakType, 0)

	packAct := map[int]models.ResIpPackageFlow{} //老用户
	if isOld > 0 {
		_, packageList2 := models.GetPackageListFlow(pakType, 1)
		num := len(packageList2)
		if num > 0 {
			for _, v := range packageList2 {
				infoOld := models.ResIpPackageFlow{}
				if newTime < v.UpdateTime {
					newTime = v.UpdateTime
				}
				// 文案配置
				infoDetail := PackageTextMap[lang+"_"+util.ItoS(v.Id)]
				content := contentConf
				corner := v.Corner
				actTitle := v.ActTitle
				actImg := v.ActImg
				if infoDetail.Id > 0 {
					if infoDetail.Corner != "" {
						corner = infoDetail.Corner
					}
					if infoDetail.ActTitle != "" {
						actTitle = infoDetail.ActTitle
					}
					if infoDetail.ActImg != "" {
						actImg = infoDetail.ActImg
					}
					if infoDetail.Content != "" {
						content = infoDetail.Content
					}
				}
				textArr := util.Split(content, "|")
				value := int64(v.Value / 1024 / 1024 / 1024)
				infoOld.Id = v.Id
				infoOld.Code = v.Code
				infoOld.Name = v.Name
				infoOld.SubName = v.SubName //副标题
				infoOld.Gift = int64(v.Gift / 1024 / 1024 / 1024)
				infoOld.Give = int64(v.Give / 1024 / 1024 / 1024)
				infoOld.Number = v.Day
				infoOld.Value = value
				infoOld.Total = value
				infoOld.Price = v.Price
				infoOld.ShowPrice = v.ShowPrice
				infoOld.Unit = v.Unit
				infoOld.AllUnit = v.AllUnit
				infoOld.Corner = corner
				infoOld.ActTitle = actTitle
				infoOld.ActImg = actImg
				infoOld.TextArr = textArr
				infoOld.Default = v.Default
				infoOld.IsHot = v.IsHot
				infoOld.Currency = v.Currency
				infoOld.Sort = v.Sort
				infoOld.UseType = v.UseType
				infoOld.Alias = v.Alias
				//staticUrl := models.GetConfigVal("STATIC_DOMAIN_URL") //官网静态文件地址
				//infoOld.StaticUrl = strings.TrimRight(staticUrl, "/")
				infoOld.GiftUnit = "GB"

				/// 如果有优惠券则展示优惠券标签
				if uid > 0 {
					for _, coupon := range availableCoupons {
						if strings.Contains(coupon.Meals, strconv.Itoa(infoOld.Id)) {
							infoOld.CouponLabel = coupon.Cron
							break
						}
					}
				}

				packAct[v.Pid] = infoOld
			}
		}
	}

	orderNumList := models.GetOrderCountWithUid(uid)
	orderNumMap := map[int]int{}
	for _, v := range orderNumList {
		orderNumMap[v.PakId] = v.Count
	}
	allPackage := models.ResIpPackageFlow{}
	ResPackage := map[string][]models.ResIpPackageFlow{}
	for _, v := range packageList {
		info := models.ResIpPackageFlow{}
		if newTime < v.UpdateTime {
			newTime = v.UpdateTime
		}
		// 文案配置
		infoDetail := PackageTextMap[lang+"_"+util.ItoS(v.Id)]
		content := contentConf
		corner := v.Corner
		actTitle := v.ActTitle
		actImg := v.ActImg
		if infoDetail.Id > 0 {
			if infoDetail.Corner != "" {
				corner = infoDetail.Corner
			}
			if infoDetail.ActTitle != "" {
				actTitle = infoDetail.ActTitle
			}
			if infoDetail.ActImg != "" {
				actImg = infoDetail.ActImg
			}
			if infoDetail.Content != "" {
				content = infoDetail.Content
			}
		}
		textArr := util.Split(content, "|")
		dayStr := util.ItoS(v.Day)
		value := int64(v.Value / 1024 / 1024 / 1024)
		info.Id = v.Id
		info.Code = v.Code
		info.Name = v.Name
		info.SubName = v.SubName //副标题
		info.Gift = int64(v.Gift / 1024 / 1024 / 1024)
		info.Give = int64(v.Give / 1024 / 1024 / 1024)
		info.Number = v.Day
		info.Value = value
		info.Total = value
		info.Price = v.Price
		info.ShowPrice = v.ShowPrice
		info.Unit = v.Unit
		info.AllUnit = v.AllUnit
		info.Corner = corner
		info.ActTitle = actTitle
		info.ActImg = actImg
		info.TextArr = textArr
		info.Default = v.Default
		info.IsHot = v.IsHot
		info.Currency = v.Currency
		info.Sort = v.Sort
		info.UseType = v.UseType
		info.Alias = v.Alias
		//staticUrl := models.GetConfigVal("STATIC_DOMAIN_URL") //官网静态文件地址
		//info.StaticUrl = strings.TrimRight(staticUrl, "/")
		info.GiftUnit = "GB"

		/// 如果有优惠券替换优惠券角标
		if uid > 0 {
			for _, coupon := range availableCoupons {
				if strings.Contains(coupon.Meals, strconv.Itoa(info.Id)) {
					cron := coupon.Cron
					info.CouponLabel = cron
					break
				}
			}
		}

		if isOld == 1 { //老用户的信息
			infoOld, pOk := packAct[v.Id]
			if pOk {
				info = infoOld
			}
		}

		if v.IsAll == 1 {
			//allPackage = info
			if info.UseType == 1 && isOld == 0 {
				allPackage = info
			}
			if info.UseType != 1 && isOld == 1 {
				allPackage = info
			}
		}

		if v.Day == 30 {
			orderNum, ok := orderNumMap[info.Id]
			if !ok {
				orderNum = 0
			}
			/// 6G,7G,8G,9G,10G的套餐不加入套餐列表
			if info.Code == "flow_6" || info.Code == "flow_7" || info.Code == "flow_8" || info.Code == "flow_9" || info.Code == "flow_10" {
				continue
			}
			if info.Code == "flow_5" { //5G的套餐根是否新用户套餐判断
				if info.UseType == 1 && isOld == 0 {
					ResPackage[dayStr] = append(ResPackage[dayStr], info)
				}
				if info.UseType != 1 && isOld == 1 {
					ResPackage[dayStr] = append(ResPackage[dayStr], info)
				}
			} else if info.Code == "flow_60" && info.UseType == 2 { //60G的老用户套餐
				// 判断是否购买过这个套餐
				count := orderNum // 需要优化 20241128
				if count <= 0 {
					if isOld == 1 {
						ResPackage[dayStr] = append(ResPackage[dayStr], info)
					}
				}
			} else if info.Code == "flow_60" { //60G的用户套餐
				// 判断是否购买过这个套餐
				count := orderNum
				if isOld == 0 || count > 0 {
					ResPackage[dayStr] = append(ResPackage[dayStr], info)
				}
			} else if info.Code == "flow_agent_1000" && info.UseType == 2 { //1000G的老用户套餐
				// 判断是否购买过这个套餐
				count := orderNum
				if count <= 0 {
					if isOld == 1 {
						ResPackage[dayStr] = append(ResPackage[dayStr], info)
					}
				}
			} else if info.Code == "flow_agent_1000" { //1000G的老用户套餐
				// 判断是否购买过这个套餐 114是1000G的老用户套餐id
				count := orderNum
				if isOld == 0 || count > 0 {
					ResPackage[dayStr] = append(ResPackage[dayStr], info)
				}
			} else {
				ResPackage[dayStr] = append(ResPackage[dayStr], info)
			}
		} else {
			ResPackage[dayStr] = append(ResPackage[dayStr], info)
		}
	}

	//listStr, _ := json.Marshal(allPackage)
	//fmt.Println("--------",string(listStr))
	//listStr, _ := json.Marshal(ResPackage)
	//fmt.Println("--------",string(listStr))

	resLists := map[string][]models.ResIpPackageFlow{}
	for k, v := range ResPackage {
		if k != util.ItoS(allPackage.Number) && pakType != "flow_agent" {
			v = append(v, allPackage)
		}
		resLists[k] = v
	}
	for _, v := range resLists {
		// 排序
		sort.SliceStable(v, func(i, j int) bool {
			return v[i].Sort > v[j].Sort
		})
	}

	return resLists, newTime

}

// 处理代理商套餐
func dealFlowAgentPackage(lang string, isOld int, uid int) (map[string][]models.ResIpPackageFlow, int) {
	//pakType := "flow_agent"
	// 获取用户拥有的所有优惠券
	_, availableCoupons := models.GetAvailableCouponListByUid(uid)

	// 获取套餐的多语言文案
	contentConf := models.GetConfigVal("FlowConfigText") //配置默认文案
	PackageTextMap := map[string]models.CmPackageInfo{}
	_, packageInfoList := models.GetPackageInfoList()
	for _, v := range packageInfoList {
		str := v.Lang + "_" + util.ItoS(v.PackageId)
		PackageTextMap[str] = v
	}
	newTime := 0
	_, packageList := models.GetPackageListFlowAgent(30, 0)

	packAct := map[int]models.ResIpPackageFlow{} //老用户
	if isOld > 0 {
		_, packageList2 := models.GetPackageListFlowAgent(30, 1)
		num := len(packageList2)
		if num > 0 {
			for _, v := range packageList2 {
				infoOld := models.ResIpPackageFlow{}
				if newTime < v.UpdateTime {
					newTime = v.UpdateTime
				}
				// 文案配置
				infoDetail := PackageTextMap[lang+"_"+util.ItoS(v.Id)]
				content := contentConf
				corner := v.Corner
				actTitle := v.ActTitle
				actImg := v.ActImg
				if infoDetail.Id > 0 {
					if infoDetail.Corner != "" {
						corner = infoDetail.Corner
					}
					if infoDetail.ActTitle != "" {
						actTitle = infoDetail.ActTitle
					}
					if infoDetail.ActImg != "" {
						actImg = infoDetail.ActImg
					}
					if infoDetail.Content != "" {
						content = infoDetail.Content
					}
				}
				textArr := util.Split(content, "|")
				value := int64(v.Value / 1024 / 1024 / 1024)
				infoOld.Id = v.Id
				infoOld.Code = v.Code
				infoOld.Name = v.Name
				infoOld.SubName = v.SubName //副标题
				infoOld.Gift = int64(v.Gift / 1024 / 1024 / 1024)
				infoOld.Give = int64(v.Give / 1024 / 1024 / 1024)
				infoOld.Number = v.Day
				infoOld.Value = value
				infoOld.Total = value
				infoOld.Price = v.Price
				infoOld.ShowPrice = v.ShowPrice
				infoOld.Unit = v.Unit
				infoOld.AllUnit = v.AllUnit
				infoOld.Corner = corner
				infoOld.ActTitle = actTitle
				infoOld.ActImg = actImg
				infoOld.TextArr = textArr
				infoOld.Default = v.Default
				infoOld.IsHot = v.IsHot
				infoOld.Currency = v.Currency
				infoOld.Sort = v.Sort
				infoOld.UseType = v.UseType
				infoOld.Alias = v.Alias
				//staticUrl := models.GetConfigVal("STATIC_DOMAIN_URL") //官网静态文件地址
				//infoOld.StaticUrl = strings.TrimRight(staticUrl, "/")
				infoOld.GiftUnit = "GB"

				/// 如果有优惠券替换优惠券角标
				if uid > 0 {
					for _, coupon := range availableCoupons {
						if strings.Contains(coupon.Meals, strconv.Itoa(infoOld.Id)) {
							infoOld.CouponLabel = coupon.Cron
							break
						}
					}
				}
				packAct[v.Pid] = infoOld
			}
		}
	}

	orderNumList := models.GetOrderCountWithUid(uid) //20241211 优化  有原来的循环查表GetOrderCountWith 修改为查map
	orderNumMap := map[int]int{}
	for _, v := range orderNumList {
		orderNumMap[v.PakId] = v.Count
	}
	//allPackage := models.ResIpPackageFlow{}
	ResPackage := map[string][]models.ResIpPackageFlow{}
	for _, v := range packageList {
		info := models.ResIpPackageFlow{}
		if newTime < v.UpdateTime {
			newTime = v.UpdateTime
		}
		// 文案配置
		infoDetail := PackageTextMap[lang+"_"+util.ItoS(v.Id)]
		content := contentConf
		corner := v.Corner
		actTitle := v.ActTitle
		actImg := v.ActImg
		if infoDetail.Id > 0 {
			if infoDetail.Corner != "" {
				corner = infoDetail.Corner
			}
			if infoDetail.ActTitle != "" {
				actTitle = infoDetail.ActTitle
			}
			if infoDetail.ActImg != "" {
				actImg = infoDetail.ActImg
			}
			if infoDetail.Content != "" {
				content = infoDetail.Content
			}
		}
		textArr := util.Split(content, "|")
		dayStr := util.ItoS(v.Day)
		value := int64(v.Value / 1024 / 1024 / 1024)
		info.Id = v.Id
		info.Code = v.Code
		info.Name = v.Name
		info.SubName = v.SubName //副标题
		info.Gift = int64(v.Gift / 1024 / 1024 / 1024)
		info.Give = int64(v.Give / 1024 / 1024 / 1024)
		info.Number = v.Day
		info.Value = value
		info.Total = value
		info.Price = v.Price
		info.ShowPrice = v.ShowPrice
		info.Unit = v.Unit
		info.AllUnit = v.AllUnit
		info.Corner = corner
		info.ActTitle = actTitle
		info.ActImg = actImg
		info.TextArr = textArr
		info.Default = v.Default
		info.IsHot = v.IsHot
		info.Currency = v.Currency
		info.Sort = v.Sort
		info.UseType = v.UseType
		info.Alias = v.Alias
		//staticUrl := models.GetConfigVal("STATIC_DOMAIN_URL") //官网静态文件地址
		//info.StaticUrl = strings.TrimRight(staticUrl, "/")
		info.GiftUnit = "GB"

		/// 如果有优惠券替换优惠券角标
		if uid > 0 {
			for _, coupon := range availableCoupons {
				if strings.Contains(coupon.Meals, strconv.Itoa(info.Id)) {
					info.CouponLabel = coupon.Cron
					break
				}
			}
		}

		if isOld == 1 { //老用户的信息
			infoOld, pOk := packAct[v.Id]
			if pOk {
				info = infoOld
			}
		}

		//if v.IsAll == 1 {
		//	allPackage = info
		//}

		if v.Day == 30 {
			orderNum, ok := orderNumMap[info.Id]
			if !ok {
				orderNum = 0
			}

			/// 6G,7G,8G,9G,10G的套餐不加入套餐列表
			if info.Code == "flow_6" || info.Code == "flow_7" || info.Code == "flow_8" || info.Code == "flow_9" || info.Code == "flow_10" {
				continue
			}
			if info.Code == "flow_5" { //5G的套餐根是否新用户套餐判断
				if info.UseType == 1 && isOld == 0 {
					ResPackage[dayStr] = append(ResPackage[dayStr], info)
				}
				if info.UseType != 1 && isOld == 1 {
					ResPackage[dayStr] = append(ResPackage[dayStr], info)
				}
			} else if info.Code == "flow_60" && info.UseType == 2 { //60G的老用户套餐
				// 判断是否购买过这个套餐
				count := orderNum
				if count <= 0 {
					if isOld == 1 {
						ResPackage[dayStr] = append(ResPackage[dayStr], info)
					}
				}
			} else if info.Code == "flow_60" { //60G的用户套餐
				// 判断是否购买过这个套餐
				count := orderNum
				if isOld == 0 || count > 0 {
					ResPackage[dayStr] = append(ResPackage[dayStr], info)
				}
			} else if info.Code == "flow_agent_1000" && info.UseType == 2 { //1000G的老用户套餐
				// 判断是否购买过这个套餐
				count := orderNum
				if count <= 0 {
					if isOld == 1 {
						ResPackage[dayStr] = append(ResPackage[dayStr], info)
					}
				}
			} else if info.Code == "flow_agent_1000" { //1000G的老用户套餐
				// 判断是否购买过这个套餐 114是1000G的老用户套餐id
				count := orderNum
				if isOld == 0 || count > 0 {
					ResPackage[dayStr] = append(ResPackage[dayStr], info)
				}
			} else {
				ResPackage[dayStr] = append(ResPackage[dayStr], info)
			}
		} else {
			ResPackage[dayStr] = append(ResPackage[dayStr], info)
		}
	}
	ResPackage["60"] = ResPackage["30"]
	ResPackage["90"] = ResPackage["30"]
	ResPackage["120"] = ResPackage["30"]
	ResPackage["150"] = ResPackage["30"]
	ResPackage["180"] = ResPackage["30"]
	resLists := map[string][]models.ResIpPackageFlow{}
	for k, v := range ResPackage {
		//if k != util.ItoS(allPackage.Number) && pakType != "flow_agent" {
		//	v = append(v, allPackage)
		//}
		resLists[k] = v
	}
	for _, v := range resLists {
		// 排序
		sort.SliceStable(v, func(i, j int) bool {
			return v[i].Sort > v[j].Sort
		})
	}
	return resLists, newTime
}

func dealHalloweenActivityPackage(lang string) []models.ResIpPackageFlow {
	// 获取套餐的多语言文案
	contentConf := models.GetConfigVal("FlowConfigText") //配置默认文案
	PackageTextMap := map[string]models.CmPackageInfo{}
	_, packageInfoList := models.GetPackageInfoList()
	for _, v := range packageInfoList {
		str := v.Lang + "_" + util.ItoS(v.PackageId)
		PackageTextMap[str] = v
	}
	// 104-新用户fow_5, 113-老用户60G, 114-老用户1000G, 84-不限量-周
	ids := []string{"104", "113", "114", "84"}

	_, packageList := models.GetPackageListWith(ids)

	var ResPackage []models.ResIpPackageFlow
	for _, v := range packageList {

		info := models.ResIpPackageFlow{}
		// 文案配置
		infoDetail := PackageTextMap[lang+"_"+util.ItoS(v.Id)]
		content := contentConf
		corner := v.Corner
		actTitle := v.ActTitle
		actImg := v.ActImg
		if infoDetail.Id > 0 {
			if infoDetail.Corner != "" {
				corner = infoDetail.Corner
			}
			if infoDetail.ActTitle != "" {
				actTitle = infoDetail.ActTitle
			}
			if infoDetail.ActImg != "" {
				actImg = infoDetail.ActImg
			}
			if infoDetail.Content != "" {
				content = infoDetail.Content
			}
		}
		textArr := util.Split(content, "|")
		value := int64(v.Value / 1024 / 1024 / 1024)
		info.Id = v.Id
		info.Code = v.Code
		info.Name = v.Name
		info.SubName = v.SubName //副标题
		info.Gift = int64(v.Gift / 1024 / 1024 / 1024)
		info.Give = int64(v.Give / 1024 / 1024 / 1024)
		info.Number = v.Day
		info.Value = value
		info.Total = value
		info.Price = v.Price
		info.ShowPrice = v.ShowPrice
		info.Unit = v.Unit
		info.AllUnit = v.AllUnit
		info.Corner = corner
		info.ActTitle = actTitle
		info.ActImg = actImg
		info.TextArr = textArr
		info.Default = v.Default
		info.IsHot = v.IsHot
		info.Currency = v.Currency
		info.Sort = v.Sort
		info.Alias = v.Alias
		//staticUrl := models.GetConfigVal("STATIC_DOMAIN_URL") //官网静态文件地址
		//info.StaticUrl = strings.TrimRight(staticUrl, "/")
		info.GiftUnit = "GB"

		ResPackage = append(ResPackage, info)

	}

	return ResPackage

}
