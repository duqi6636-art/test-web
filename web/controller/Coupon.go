package controller

import (
	"api-360proxy/pkg/util"
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"math"
	"sort"
	"strings"
)

// 获取优惠券信息
// @BasePath /api/v1
// @Summary 获取优惠券信息
// @Description 获取优惠券信息
// @Tags 优惠券相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object}  models.ResCouponInfo{} "优惠券信息"
// @Router /web/coupon/coupon [post]
func GetCouponOne(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	err, list := models.GetCouponList(uid, "")
	nowTime := util.GetNowInt()
	couponList := []models.ResCouponInfo{}

	if err == nil && len(list) > 0 {
		for _, v := range list {
			info := models.ResCouponInfo{}
			useTime := ""
			if v.UseTime > 0 {
				useTime = util.GetTimeStr(v.UseTime, "Y/m/d H:i")
			}
			value := ""
			if v.Type == "1" {
				value = util.FtoS(v.Value)
			}
			if v.Type == "2" {
				value = util.FtoS(v.Value)
			}
			if v.Type == "3" {
				value = util.FtoS(v.Value * 100)
			}
			if v.Type == "4" {
				value = util.FtoS(v.Value * 100)
			}
			expTime := "forever"
			if v.Expire > 0 {
				expTime = util.GetTimeStr(v.Expire, "Y/m/d H:i")
			}
			var is_use = 0
			if v.Status == 1 && (v.Expire == 0 || v.Expire >= nowTime) {
				is_use = 1
			}
			mealArr := strings.Split(v.Meals, ",")

			info.Id = v.Id
			info.Code = v.Code
			info.Name = v.Name
			info.Status = v.Status
			info.UseTime = useTime
			info.Title = v.Title
			info.Type = v.Type
			info.Value = value
			info.MealArr = mealArr
			info.Expire = expTime
			info.CreateTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i")
			if is_use == 1 {
				couponList = append(couponList, info)
			}
		}
		//排序
		sort.SliceStable(couponList, func(i, j int) bool {
			return couponList[i].Id > couponList[j].Id
		})
	}
	resInfo := models.ResCouponInfo{}
	if len(couponList) > 0 {
		_, user := models.GetUserById(uid)
		resInfo = couponList[0]
		if user.IsPay == "true" {
			//idStr := strings.Join(resInfo.MealArr,",")
			packageList := models.GetPackageListByPIds(resInfo.MealArr)
			if len(packageList) > 0 {
				mealArr := resInfo.MealArr
				mealOldArr := map[string]string{}
				//mealOldArr := []string{}
				for _, v := range packageList {
					id := util.ItoS(v.Id)
					pid := util.ItoS(v.Pid)
					mealOldArr[pid] = id
				}
				resMealArr := []string{}
				for _, vv := range mealArr {
					idInfo, has := mealOldArr[vv]
					if has {
						resMealArr = append(resMealArr, idInfo)
					} else {
						resMealArr = append(resMealArr, vv)
					}
				}
				resInfo.MealArr = resMealArr
			}
		}
		JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resInfo)
		return
	}

	JsonReturn(c, e.ERROR, "__T_NO_DATA", gin.H{})
	return
}

// 获取优惠券列表
// @BasePath /api/v1
// @Summary 获取优惠券列表
// @Description 获取优惠券列表
// @Tags 优惠券相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param useful formData string true "是否只显示有效的优惠券"
// @Produce json
// @Success 0 {object} map[string]interface{} "ip_list：已使用IP券列表（值为[]models.ResCouponList{}对象），money_list：已使用金额券列表（值为[]models.ResCouponList{}对象）"
// @Router /web/coupon/get_list [post]
func GetCouponList(c *gin.Context) {
	sessionId := c.DefaultPostForm("session", "")
	useful := c.DefaultPostForm("useful", "1")
	if sessionId == "" {
		JsonReturn(c, e.SESSION_EXPIRED, "__T_SESSION_EMPTY", nil)
		return
	}
	res, uid := GetUIDbySession(sessionId)
	if !res {
		JsonReturn(c, e.SESSION_EXPIRED, "__T_SESSION_ERROR", nil)
		return
	}
	err, list := models.GetCouponList(uid, "")
	nowTime := util.GetNowInt()
	ipList := []models.ResCouponList{}
	moneyList := []models.ResCouponList{}
	mealLists := models.GetPackageList()
	packageArr := map[int]string{}
	for _, v := range mealLists {
		packageArr[v.Id] = v.Name
	}
	if err == nil && len(list) > 0 {
		for _, v := range list {
			info := models.ResCouponList{}
			useTime := ""
			if v.UseTime > 0 {
				useTime = util.GetTimeStr(v.UseTime, "Y/m/d H:i")
			}
			value := ""
			couType := ""
			typeName := v.Type
			if v.Type == "1" {
				value = "$" + util.FtoS(v.Value)
				couType = "money"
				typeName = "money"
			}
			if v.Type == "2" {
				value = util.FtoS(v.Value) + " IPs"
				couType = "ip"
				typeName = "ip"
			}
			if v.Type == "3" {
				value = util.FtoS(v.Value*100) + "%"
				couType = "money"
				typeName = "discount"
			}
			if v.Type == "4" {
				value = util.FtoS(v.Value*100) + "%" + " IPs"
				couType = "ip"
				typeName = "discount"
			}
			expTime := "forever"
			if v.Expire > 0 {
				expTime = util.GetTimeStr(v.Expire, "Y/m/d H:i")
			}
			var is_use = 0
			if v.Status == 1 && (v.Expire == 0 || v.Expire >= nowTime) {
				is_use = 1
			}
			mealArr := strings.Split(v.Meals, ",")
			meal := ""
			for _, vv := range mealArr {
				mId := util.StoI(vv)
				mName := packageArr[mId]
				meal = meal + mName + ","
			}
			meal = strings.TrimRight(meal, ",")
			info.Id = v.Id
			info.Code = v.Code
			info.Name = v.Name
			info.Status = v.Status
			info.UseTime = useTime
			info.Title = v.Title
			info.Type = typeName
			info.Value = value
			info.Meals = meal
			info.Expire = expTime
			info.CreateTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i")
			if useful == "1" && is_use == 1 {
				if couType == "ip" {
					ipList = append(ipList, info)
				}
				if couType == "money" {
					moneyList = append(moneyList, info)
				}
			}
			if useful == "2" && is_use == 0 {
				if couType == "ip" {
					ipList = append(ipList, info)
				}
				if couType == "money" {
					moneyList = append(moneyList, info)
				}
			}
		}
	}
	resList := map[string]interface{}{}
	resList["ip_list"] = ipList
	resList["money_list"] = moneyList
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resList)
	return
}

// 检测用户券信息和发放券
func CouponInfoAuto(userInfo models.Users, cate string) {
	is_pay := "no_pay"
	if cate == "agent" {
		is_pay = "agent"
	} else {
		if userInfo.IsPay == "true" {
			is_pay = "payed"
		}
	}
	couponConfList := models.GetCouponListByCate(0, cate)           //获取不同类型券配置
	couponUserList := models.GetCouponListByCate(userInfo.Id, cate) //获取用户已领取的不同类型券
	nowTime := util.GetNowInt()
	if len(couponConfList) > 0 {
		hasCidArr := []int{}
		for _, userV := range couponUserList {
			hasCidArr = append(hasCidArr, userV.Cid)
		}
		for _, confV := range couponConfList {
			codeStr := GetUuid()
			codeArr := strings.Split(codeStr, "-")
			lens := len(codeArr)
			code := codeArr[0] + "-" + codeArr[lens-2] + "-" + codeArr[lens-1]
			info := confV
			info.Id = 0
			info.Code = code
			info.Expire = nowTime + confV.ExpiryDay*86400
			info.BindUid = userInfo.Id
			info.BindUsername = userInfo.Username
			info.Cate = confV.Cate
			info.CreateTime = nowTime
			// 用户没有领取过 才会继续往下执行
			if !util.InArrayInt(confV.Cid, hasCidArr) {
				// 如果是已付费用户 且 配置了已付费用户券
				if is_pay == "payed" && (confV.UserType == "all" || confV.UserType == "payed") {
					err := models.AddCoupon(info)
					fmt.Println("payed:", err)
				}
				// 如果是未付费用户 且 配置了未付费用户券
				if is_pay == "no_pay" && (confV.UserType == "all" || confV.UserType == "no_pay") {
					err := models.AddCoupon(info)
					fmt.Println("no_pay:", err)
				}
				// 如果是代理商用户 且 配置了代理商券
				if is_pay == "agent" && (confV.UserType == "all" || confV.UserType == "agent") {
					err := models.AddCoupon(info)
					fmt.Println("agent:", err)
				}
			}
		}
	}
}

// 检测新用户券信息和发放券
func NewCouponInfoAuto(userInfo models.Users, cate string) {

	couponConfList := models.GetCouponListByCate(0, cate)           //获取不同类型券配置
	couponUserList := models.GetCouponListByCate(userInfo.Id, cate) //获取用户已领取的不同类型券
	// 获取优惠券列表-判断是否已下掉
	nowTime := util.GetNowInt()
	if len(couponConfList) > 0 {
		hasCidArr := []int{}
		for _, userV := range couponUserList {
			hasCidArr = append(hasCidArr, userV.Cid)
		}
		for _, confV := range couponConfList {
			codeStr := GetUuid()
			codeArr := strings.Split(codeStr, "-")
			lens := len(codeArr)
			code := codeArr[0] + "-" + codeArr[lens-2] + "-" + codeArr[lens-1]
			info := confV
			info.Id = 0
			info.Code = code
			info.Expire = nowTime + confV.ExpiryDay*86400
			info.BindUid = userInfo.Id
			info.BindUsername = userInfo.Username
			info.Cate = confV.Cate
			info.CreateTime = nowTime
			// 用户没有领取过 才会继续往下执行
			if !util.InArrayInt(confV.Cid, hasCidArr) {

				// 注册时间>24小时
				if nowTime-userInfo.CreateTime > 24*60*60 && confV.UserType == "no_pay" {

					err := models.AddCoupon(info)
					fmt.Println("no_pay:", err)
				}
			}
		}
	}
}

// 获取优惠券列表
// @BasePath /api/v1
// @Summary 获取优惠券列表
// @Description 获取优惠券列表
// @Tags 优惠券相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} map[string]interface{} "all_list：全部列表,used_list：已使用列表,expired_list：已过期列表"
// @Router /web/coupon/card_holder [post]
func CardHolder(c *gin.Context) {
	sessionId := c.DefaultPostForm("session", "")
	if sessionId == "" {
		JsonReturn(c, e.SESSION_EXPIRED, "__T_SESSION_EMPTY", nil)
		return
	}
	res, uid := GetUIDbySession(sessionId)
	if !res {
		JsonReturn(c, e.SESSION_EXPIRED, "__T_SESSION_ERROR", nil)
		return
	}
	// 查询用户信息
	err, userInfo := models.GetUserById(uid)
	if err != nil || userInfo.Id == 0 {
		JsonReturn(c, e.SESSION_EXPIRED, "__T_SESSION_ERROR", nil)
		return
	}

	// 获取用户拥有的所有优惠券
	err, list := models.GetCardList(uid, "")
	nowTime := util.GetNowInt()
	// 数据处理
	mealLists := models.GetPackageList()
	packageArr := map[int]string{}
	packageMoneyArr := map[int]string{}
	for _, v := range mealLists {
		packageArr[v.Id] = v.Name
		if userInfo.IsPay == "true" {
			packageMoneyArr[v.Id] = fmt.Sprintf("%s_%f", v.PakType, v.Price)
		} else {
			if v.Pid == 0 {
				packageMoneyArr[v.Id] = fmt.Sprintf("%s_%f", v.PakType, v.Price)
			}
		}
	}
	allList := []models.ResCouponList{}     // 全部未使用的列表
	usedList := []models.ResCouponList{}    // 已使用的列表
	expiredList := []models.ResCouponList{} // 已过期的列表
	if err == nil && len(list) > 0 {
		for _, v := range list {
			info := models.ResCouponList{}
			value := ""
			unit := ""
			typeName := v.Type
			if v.Type == "1" {
				value = fmt.Sprintf("%d", int(v.Value))
				typeName = "money"
				unit = "$"
			}
			if v.Type == "2" {
				value = fmt.Sprintf("%d", int(v.Value))
				typeName = "ip"
				unit = "IPs"
			}
			if v.Type == "3" {
				couponRatio := 1 - v.Value
				if couponRatio <= 0.9 {
					couponRatio = math.Ceil(couponRatio*10) * 10
				} else {
					couponRatio = couponRatio * 100
				}
				value = fmt.Sprintf("%d", int(100-couponRatio)) + "%"
				typeName = "money"
				unit = "$"
			}
			if v.Type == "4" {
				value = fmt.Sprintf("%d", int(v.Value*100)) + "%"
				typeName = "ip"
				unit = "IPs"
			}
			expTime := "forever"
			if v.Expire > 0 {
				expTime = util.GetTimeStr(v.Expire, "Y-m-d")
			}
			mealArr := strings.Split(v.Meals, ",")
			var meal []string
			pakId := ""
			pakType := "isp"
			pakMoney := 0.00
			for _, vv := range mealArr {
				mId := util.StoI(vv)
				mName := packageArr[mId]
				meal = append(meal, mName)

				// 判断应该选中的套餐id（选中金额最大套餐）
				if s, ok := packageMoneyArr[mId]; ok {
					arr := strings.Split(s, "_")
					pakType = arr[0]
					money := util.StoF(arr[1])
					if money > pakMoney {
						pakMoney = money
						pakId = fmt.Sprintf("%d", mId)
						if pakType == "flow" {
							unit = "GB"
						}
					}
				}
			}
			meals := fmt.Sprintf("Applicable to %s package", strings.Join(meal, ","))
			info.Id = v.Id
			info.Code = v.Code
			info.Name = v.Name
			info.Status = v.Status
			info.Title = v.Title
			info.Type = typeName
			if typeName == "money" {
				unit = "$"
			}
			info.Value = value
			info.Meals = meals
			info.PakId = pakId
			info.PakType = pakType
			info.ExpireTime = v.Expire
			info.Expire = expTime
			info.Cron = v.Cron
			info.Unit = unit
			info.CreateTime = util.GetTimeStr(v.CreateTime, "Y-m-d")

			if v.Status == 1 && v.Expire > nowTime {
				allList = append(allList, info)
			}

			if v.Status == 2 {
				usedList = append(usedList, info)
			}

			if v.Expire <= nowTime {
				expiredList = append(expiredList, info)
			}

		}
	}
	resList := map[string]interface{}{}
	resList["all_list"] = []string{}
	resList["used_list"] = []string{}
	resList["expired_list"] = []string{}
	// 排序
	sort.SliceStable(allList, func(i, j int) bool {
		return allList[i].ExpireTime < allList[j].ExpireTime
	})
	if len(allList) > 0 {
		resList["all_list"] = allList
	}
	if len(usedList) > 0 {
		resList["used_list"] = usedList
	}
	if len(expiredList) > 0 {
		resList["expired_list"] = expiredList
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resList)
	return
}

// 兑换优惠券
// @BasePath /api/v1
// @Summary 兑换优惠券
// @Description 兑换优惠券
// @Tags 优惠券相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param code formData string true "兑换码"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /web/coupon/redeem_coupons [post]
func RedeemCoupons(c *gin.Context) {
	sessionId := c.DefaultPostForm("session", "")
	code := c.DefaultPostForm("code", "") // 兑换码
	if sessionId == "" {
		JsonReturn(c, e.SESSION_EXPIRED, "__T_SESSION_EMPTY", nil)
		return
	}
	res, uid := GetUIDbySession(sessionId)
	if !res {
		JsonReturn(c, e.SESSION_EXPIRED, "__T_SESSION_ERROR", nil)
		return
	}
	// 查询用户信息
	err, userInfo := models.GetUserById(uid)
	if err != nil || userInfo.Id == 0 {
		JsonReturn(c, e.SESSION_EXPIRED, "__T_SESSION_ERROR", nil)
		return
	}
	nowTime := util.GetNowInt()
	is_pay := "no_pay"
	if userInfo.IsPay == "true" {
		is_pay = "payed"
	}

	// 查询优惠券
	err, couponInfo := models.GetCoupon(code)
	if err != nil || couponInfo.Id == 0 || couponInfo.Status != 1 {
		JsonReturn(c, -1, "__T_COUPON_ERROR", nil)
		return
	}

	no_used := 0
	has := 0
	if couponInfo.UseCycle >= 1 { // 此平台下限制用户使用次数
		_, couponUse := models.GetCouponByUsePlatform(uid, couponInfo.UseCycle, couponInfo.Platform)
		has = len(couponUse)
		used := 0
		for _, v := range couponUse {
			if v.Status == 1 {
				no_used = no_used + 1
			} else {
				used = used + 1
			}
		}
		if used >= couponInfo.UseNumber {
			JsonReturn(c, -1, "__T_COUPON_USED_LIMIT", nil)
			return
		}
	}

	// 同一分组下只可兑换一次cdk
	if couponInfo.GroupId > 0 {
		cdkInfo, _ := models.GetCdkByGroupId(uid, couponInfo.GroupId)
		if cdkInfo.Id != 0 && couponInfo.Cid != cdkInfo.Cid {
			JsonReturn(c, -1, "__T_COUPON_USED_LIMIT", nil)
			return
		}
	}

	if couponInfo.UserType != "all" {
		if couponInfo.UserType == "agent" {
			if is_pay != "payed" {
				JsonReturn(c, -1, "__T_COUPON_COMPARE", nil)
				return
			}
		} else {
			if is_pay != couponInfo.UserType {
				JsonReturn(c, -1, "__T_COUPON_COMPARE", nil)
				return
			}
		}
	}

	//已过期
	if couponInfo.ExpiryDay > 1000000000 && couponInfo.Expire < nowTime {
		JsonReturn(c, -1, "__T_COUPON_EXPIRED", nil)
		return
	}
	if couponInfo.Expire > 0 && couponInfo.Expire < nowTime {
		JsonReturn(c, -1, "__T_COUPON_EXPIRED", nil)
		return
	}
	//已被使用
	if couponInfo.Status == 2 {
		JsonReturn(c, -1, "__T_COUPON_USED", nil)
		return
	}

	var expire = 0
	if couponInfo.Expire > 0 && couponInfo.ExpiryDay > 0 {
		expire = nowTime + couponInfo.ExpiryDay*86400
		if expire > couponInfo.Expire {
			expire = couponInfo.Expire
		}
	} else if couponInfo.Expire > 0 && couponInfo.ExpiryDay == 0 {
		expire = couponInfo.Expire
	} else if couponInfo.Expire == 0 && couponInfo.ExpiryDay > 0 {
		expire = nowTime + couponInfo.ExpiryDay*86400
	} else {
		expire = 0
	}

	if couponInfo.UseType == 2 {
		if no_used == 0 && has == 0 {
			info := couponInfo
			info.Id = 0
			info.Expire = expire
			info.BindUid = userInfo.Id
			info.BindUsername = userInfo.Username
			info.CreateTime = nowTime
			info.Cron = couponInfo.Cron
			err_1 := models.AddCoupon(info)
			fmt.Println(err_1)
		} else {
			JsonReturn(c, -1, "__T_COUPON_USED_LIMIT", nil)
			return
		}
	} else {
		if couponInfo.BindUid > 0 && couponInfo.BindUid != uid {
			JsonReturn(c, -1, "__T_COUPON_ERROR", nil)
			return
		}
		if couponInfo.BindUid == 0 {
			couponParam := map[string]interface{}{}
			couponParam["expire"] = expire
			couponParam["bind_uid"] = userInfo.Id
			couponParam["bind_username"] = userInfo.Username
			couponParam["create_time"] = nowTime
			editCou := models.EditCouponBind(couponInfo.Code, uid, couponParam)
			fmt.Println(editCou)
		}
	}

	JsonReturn(c, 0, "ok", nil)
	return

}

type ResCouponList struct {
	Id         int    `json:"id"`
	Type       string `json:"type"`         // money：优惠券 ip：赠送IP/流量
	Code       string `json:"code"`         // 优惠券码
	Title      string `json:"title"`        // 标题
	Meals      string `json:"meals"`        // 标题文案描述
	Money      string `json:"money"`        // 返回值信息 根据不同的类型判断返回
	MoneyStr   string `json:"money_str"`    // 返回值信息 根据不同的类型判断返回
	MoneyRatio string `json:"money_ratio"`  // - money比例
	Ip         string `json:"ip"`           // 返回值信息 根据不同的类型判断返回
	IpStr      string `json:"ip_str"`       // 返回值信息 根据不同的类型判断返回
	IpRatio    string `json:"ip_ratio"`     // + IP比例
	ExpireTime int    `json:"expire_time"`  // 过期时间
	Expire     string `json:"expire"`       // 过期时间
	PayType    string `json:"pay_type"`     // 支付类型
	PayTypeArr []int  `json:"pay_type_arr"` // 支付类型
}

// 获取优惠券下拉列表接口
// @BasePath /api/v1
// @Summary 获取优惠券下拉列表接口
// @Description 获取优惠券下拉列表接口
// @Tags 优惠券相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param package_id formData string true "套餐id"
// @Produce json
// @Success 0 {object} []ResCouponList{} "成功"
// @Router /web/coupon/use_coupon [post]
func GetCouponListByPakId(c *gin.Context) {
	sessionId := c.DefaultPostForm("session", "")
	packageIdStr := c.DefaultPostForm("package_id", "")
	if sessionId == "" {
		JsonReturn(c, -1, "__T_SESSION_EMPTY", nil)
		return
	}
	res, uid := GetUIDbySession(sessionId)
	if !res {
		JsonReturn(c, -1, "__T_SESSION_ERROR", nil)
		return
	}
	nowTime := util.GetNowInt()            // 当前时间
	_, userInfo := models.GetUserById(uid) // 用户信息
	// 用户类型和券类型不匹配
	isPay := "no_pay"
	if userInfo.IsPay == "true" {
		isPay = "payed"
	}

	packageId := util.StoI(packageIdStr)
	packageInfo := models.GetSocksPackageInfoById(packageId) //套餐信息
	packageId = packageInfo.Id
	unit := ""
	if packageInfo.PakType == "flow" {
		unit = "GB"
	} else {
		unit = "IPs"
	}

	//// 获取所有套餐列表
	mealLists := models.GetPackageList()
	packageArr := map[int]string{}
	for _, v := range mealLists {
		packageArr[v.Id] = v.Name
	}

	// 查询所有符合条件的优惠券
	err, list := models.GetCouponListByPakId(uid, packageId, "")

	resList := []ResCouponList{} // 优惠券列表
	if err == nil && len(list) > 0 {
		for _, v := range list {
			info := ResCouponList{}
			title := ""
			ipNums := ""
			ipRatio := ""
			money := ""
			moneyRatio := ""
			couType := v.Type
			if v.Type == "1" {
				title = "- $" + util.FtoS(v.Value)
				moneyRatio = "$" + util.FtoS(v.Value)
				money = util.FtoS(v.Value)
				couType = "money"
			}
			if v.Type == "2" {
				title = util.FtoS(v.Value) + " " + unit
				ipRatio = util.FtoS(v.Value) + " " + unit
				ipNums = util.FtoS(v.Value)
				couType = "ip"
			}
			if v.Type == "3" {
				value := 1 - v.Value
				jianMoney := packageInfo.Price * v.Value
				flowNumStr := util.ItoS(int(packageInfo.Value / 1024 / 1024 / 1024))
				flowNum := util.StoF(flowNumStr)
				if value <= 0.9 {
					jianMoney = (packageInfo.Unit - (packageInfo.Unit * value)) * flowNum
					value = math.Ceil(value*10) * 10
				} else {
					value = value * 100
					jianMoney = (packageInfo.Unit - math.Ceil(packageInfo.Unit*value/10)/10) * flowNum
				}

				title = util.FtoS2(100-value, 0) + "%"
				moneyRatio = util.FtoS2(100-value, 0) + "%"
				money = fmt.Sprintf("%.1f", jianMoney)

				//money = fmt.Sprintf("%.0f", math.Ceil(packageInfo.Price*v.Value))

				couType = "money"
			}
			if v.Type == "4" {
				title = util.FtoS(v.Value*100) + "% " + unit
				ipRatio = util.FtoS(v.Value*100) + "% " + unit
				if packageInfo.PakType == "flow" {
					ipNums = util.ItoS(int(math.Ceil(float64(packageInfo.Value)*v.Value)) / 1024 / 1024 / 1024)
				} else {
					ipNums = util.ItoS(int(math.Ceil(float64(packageInfo.Value) * v.Value)))
				}
				couType = "ip"
			}
			// 优惠券过期时间处理
			expTime := "forever"
			var expire = 0
			if v.Expire > 0 && v.ExpiryDay > 0 {
				expire = nowTime + v.ExpiryDay*86400
				if expire > v.Expire {
					expire = v.Expire
				}
			} else if v.Expire > 0 && v.ExpiryDay == 0 {
				expire = v.Expire
			} else if v.Expire == 0 && v.ExpiryDay > 0 {
				expire = nowTime + v.ExpiryDay*86400
			} else {
				expire = 0
			}
			if expire > 0 {
				expTime = util.GetTimeStr(v.Expire, "Y-m-d")
			}
			// 判断此优惠券是否可以使用逻辑
			var isUse = 0
			if v.Status == 1 && (expire == 0 || expire >= nowTime) {
				if v.UserType != "all" {
					if v.UserType == "agent" {
						if isPay == "payed" {
							isUse = 1
						} else {
							isUse = 0
						}
					} else if isPay == v.UserType {
						isUse = 1
					} else {
						isUse = 0
					}
				} else {
					isUse = 1
				}
			}

			meal := []string{}
			mealArr := strings.Split(v.Meals, ",")
			for _, vv := range mealArr {
				mId := util.StoI(vv)
				mName := packageArr[mId]
				meal = append(meal, mName)
			}
			meals := fmt.Sprintf("Applicable to %s package", strings.Join(meal, ","))
			info.Id = v.Id
			info.Code = v.Code
			info.Title = title
			info.Type = couType
			info.Money = money
			info.MoneyRatio = moneyRatio
			info.Ip = ipNums
			info.IpRatio = ipRatio
			info.Meals = meals
			info.Expire = expTime
			info.ExpireTime = v.Expire
			info.IpStr = ipNums + unit
			info.MoneyStr = "$" + money

			if isUse == 1 {
				if couType == "ip" || couType == "money" {
					resList = append(resList, info)
				}
			}
		}

		//排序
		sort.SliceStable(resList, func(i, j int) bool {
			return resList[i].ExpireTime < resList[j].ExpireTime
		})
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resList)
	return
}

// 获取我的优惠券列表
// @BasePath /api/v1
// @Summary 获取优惠券列表
// @Description 获取优惠券列表
// @Tags 优惠券相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} map[string]interface{} "all_list：全部列表,used_list：已使用列表,expired_list：已过期列表"
// @Router /web/coupon/my_coupons [post]
func GetMyCoupons(c *gin.Context) {

	lang := strings.ToLower(c.DefaultPostForm("lang", "en"))
	if lang == "" {
		lang = "en"
	}
	sessionId := c.DefaultPostForm("session", "")
	if sessionId == "" {
		JsonReturn(c, e.SESSION_EXPIRED, "__T_SESSION_EMPTY", nil)
		return
	}
	res, uid := GetUIDbySession(sessionId)
	if !res {
		JsonReturn(c, e.SESSION_EXPIRED, "__T_SESSION_ERROR", nil)
		return
	}
	// 查询用户信息
	err, userInfo := models.GetUserById(uid)
	if err != nil || userInfo.Id == 0 {
		JsonReturn(c, e.SESSION_EXPIRED, "__T_SESSION_ERROR", nil)
		return
	}

	// 获取用户拥有的所有优惠券
	err, list := models.GetCardList(uid, "")
	nowTime := util.GetNowInt()
	// 数据处理
	mealLists := models.GetPackageList()
	packageArr := map[int]string{}
	packageMoneyArr := map[int]string{}
	for _, v := range mealLists {
		packageArr[v.Id] = v.Name
		if userInfo.IsPay == "true" {
			packageMoneyArr[v.Id] = fmt.Sprintf("%s_%f", v.PakType, v.Price)
		} else {
			if v.Pid == 0 {
				packageMoneyArr[v.Id] = fmt.Sprintf("%s_%f", v.PakType, v.Price)
			}
		}
	}
	var unusedList []models.ResCouponList  // 全部未使用的列表
	var usedList []models.ResCouponList    // 已使用的列表
	var expiredList []models.ResCouponList // 已过期的列表
	if err == nil && len(list) > 0 {
		for _, v := range list {
			info := models.ResCouponList{}
			value := ""
			unit := ""
			typeName := v.Type
			if v.Type == "1" {
				value = fmt.Sprintf("%d", int(v.Value))
				typeName = "money"
				unit = "$"
			}
			if v.Type == "2" {
				value = fmt.Sprintf("%d", int(v.Value))
				typeName = "ip"
				unit = "IPs"
			}
			if v.Type == "3" {
				couponRatio := 1 - v.Value
				if couponRatio <= 0.9 {
					couponRatio = math.Ceil(couponRatio*10) * 10
				} else {
					couponRatio = couponRatio * 100
				}
				value = fmt.Sprintf("%d", int(100-couponRatio)) + "%"
				typeName = "money"
				unit = "$"
			}
			if v.Type == "4" {
				value = fmt.Sprintf("%d", int(v.Value*100)) + "%"
				typeName = "ip"
				unit = "IPs"
			}
			expTime := "forever"
			if v.Expire > 0 {
				expTime = util.GetTimeStr(v.Expire, "Y-m-d")
			}
			mealArr := strings.Split(v.Meals, ",")
			var meal []string
			pakId := ""
			pakType := "isp"
			pakMoney := 0.00
			for _, vv := range mealArr {
				mId := util.StoI(vv)
				mName := packageArr[mId]
				meal = append(meal, mName)

				// 判断应该选中的套餐id（选中金额最大套餐）
				if s, ok := packageMoneyArr[mId]; ok {
					arr := strings.Split(s, "_")
					pakType = arr[0]
					money := util.StoF(arr[1])
					if money > pakMoney {
						pakMoney = money
						pakId = fmt.Sprintf("%d", mId)
						if pakType == "flow" {
							unit = "GB"
						}
					}
				}
			}
			meals := fmt.Sprintf("Applicable to %s package", strings.Join(meal, ","))
			if v.Condition != "" {
				meals = v.Condition
			}
			cron := v.Cron

			info.Id = v.Id
			info.Code = v.Code
			info.Name = v.Name
			info.Status = v.Status
			info.Title = v.Title
			info.Type = typeName
			if typeName == "money" {
				unit = "$"
			}
			info.Value = value
			info.Meals = meals
			info.PayType = v.PayType
			info.PakId = pakId
			info.PakType = pakType
			info.ExpireTime = v.Expire
			info.Expire = expTime
			info.Cron = cron
			info.Unit = unit
			info.CreateTime = util.GetTimeStr(v.CreateTime, "Y-m-d")

			if v.Status == 1 && v.Expire > nowTime {
				unusedList = append(unusedList, info)
			}

			if v.Status == 2 {
				usedList = append(usedList, info)
			}

			if v.Expire <= nowTime {
				expiredList = append(expiredList, info)
			}

		}
	}
	resList := map[string]interface{}{}
	resList["unused_list"] = []string{}
	resList["used_list"] = []string{}
	resList["expired_list"] = []string{}
	// 排序
	sort.SliceStable(unusedList, func(i, j int) bool {
		return unusedList[i].ExpireTime < unusedList[j].ExpireTime
	})
	if len(unusedList) > 0 {
		resList["unused_list"] = unusedList
	}
	if len(usedList) > 0 {
		resList["used_list"] = usedList
	}
	if len(expiredList) > 0 {
		resList["expired_list"] = expiredList
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resList)
	return
}

// 获取优惠券下拉列表接口
// @BasePath /api/v1
// @Summary 获取优惠券下拉列表接口
// @Description 获取优惠券下拉列表接口
// @Tags 优惠券相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param package_id formData string true "套餐id"
// @Produce json
// @Success 0 {object} []ResCouponList{} "成功"
// @Router /web/coupon/my_coupons_list [post]
func GetMyCouponListByPakId(c *gin.Context) {
	sessionId := c.DefaultPostForm("session", "")
	packageIdStr := c.DefaultPostForm("package_id", "")
	if sessionId == "" {
		JsonReturn(c, -1, "__T_SESSION_EMPTY", nil)
		return
	}
	res, uid := GetUIDbySession(sessionId)
	if !res {
		JsonReturn(c, -1, "__T_SESSION_ERROR", nil)
		return
	}
	nowTime := util.GetNowInt()            // 当前时间
	_, userInfo := models.GetUserById(uid) // 用户信息
	// 用户类型和券类型不匹配
	is_pay := "no_pay"
	if userInfo.IsPay == "true" {
		is_pay = "payed"
	}

	packageId := util.StoI(packageIdStr)
	packageInfo := models.GetSocksPackageInfoById(packageId) //套餐信息
	packageId = packageInfo.Id
	unit := ""
	if packageInfo.PakType == "flow" {
		unit = "GB"
	} else {
		unit = "IPs"
	}

	// 获取所有套餐列表
	mealLists := models.GetPackageList()
	packageArr := map[int]string{}
	for _, v := range mealLists {
		packageArr[v.Id] = v.Name
	}

	// 查询所有符合条件的优惠券
	err, list := models.GetCouponListByPakId(uid, packageId, "")

	resList := []ResCouponList{} // 优惠券列表
	if err == nil && len(list) > 0 {
		for _, v := range list {
			info := ResCouponList{}
			title := ""
			ipNums := ""
			ipRatio := ""
			money := ""
			moneyRatio := ""
			couType := v.Type
			if v.Type == "1" {
				title = "- $" + util.FtoS(v.Value)
				moneyRatio = "$" + util.FtoS(v.Value)
				money = util.FtoS(v.Value)
				couType = "money"
			}
			if v.Type == "2" {
				title = util.FtoS(v.Value) + " " + unit
				ipRatio = util.FtoS(v.Value) + " " + unit
				ipNums = util.FtoS(v.Value)
				couType = "ip"
			}
			if v.Type == "3" {
				value := 1 - v.Value
				jianMoney := packageInfo.Price * v.Value
				flowNumStr := util.ItoS(int(packageInfo.Value / 1024 / 1024 / 1024))
				flowNum := util.StoF(flowNumStr)
				if value <= 0.9 {
					jianMoney = (packageInfo.Unit - (packageInfo.Unit * value)) * flowNum
					value = math.Ceil(value*10) * 10
				} else {
					value = value * 100
					jianMoney = (packageInfo.Unit - math.Ceil(packageInfo.Unit*value/10)/10) * flowNum
				}

				title = util.FtoS2(100-value, 0) + "%"
				moneyRatio = util.FtoS2(100-value, 0) + "%"
				money = fmt.Sprintf("%.1f", jianMoney)

				//money = fmt.Sprintf("%.0f", math.Ceil(packageInfo.Price*v.Value))

				couType = "money"
			}
			if v.Type == "4" {
				title = util.FtoS(v.Value*100) + "% " + unit
				ipRatio = util.FtoS(v.Value*100) + "% " + unit
				if packageInfo.PakType == "flow" {
					ipNums = util.ItoS(int(math.Ceil(float64(packageInfo.Value)*v.Value)) / 1024 / 1024 / 1024)
				} else {
					ipNums = util.ItoS(int(math.Ceil(float64(packageInfo.Value) * v.Value)))
				}
				couType = "ip"
			}
			// 优惠券过期时间处理
			expTime := "forever"
			var expire = 0
			if v.Expire > 0 && v.ExpiryDay > 0 {
				expire = nowTime + v.ExpiryDay*86400
				if expire > v.Expire {
					expire = v.Expire
				}
			} else if v.Expire > 0 && v.ExpiryDay == 0 {
				expire = v.Expire
			} else if v.Expire == 0 && v.ExpiryDay > 0 {
				expire = nowTime + v.ExpiryDay*86400
			} else {
				expire = 0
			}
			if expire > 0 {
				expTime = util.GetTimeStr(v.Expire, "Y-m-d")
			}
			// 判断此优惠券是否可以使用逻辑
			var is_use = 0
			if v.Status == 1 && (expire == 0 || expire >= nowTime) {
				if v.UserType != "all" {
					if v.UserType == "agent" {
						if is_pay == "payed" {
							is_use = 1
						} else {
							is_use = 0
						}
					} else if is_pay == v.UserType {
						is_use = 1
					} else {
						is_use = 0
					}
				} else {
					is_use = 1
				}
			}

			meal := []string{}
			mealArr := strings.Split(v.Meals, ",")
			for _, vv := range mealArr {
				mId := util.StoI(vv)
				mName := packageArr[mId]
				meal = append(meal, mName)
			}
			meals := fmt.Sprintf("Applicable to %s package", strings.Join(meal, ","))
			if v.Condition != "" {
				meals = v.Condition
			}
			payTypeArr := []int{}
			if v.PayType != "" {
				strArr := strings.Split(v.PayType, ",")
				for _, str := range strArr {
					payTypeArr = append(payTypeArr, util.StoI(str))
				}
			}
			info.Id = v.Id
			info.Code = v.Code
			info.Title = title
			info.Type = couType
			info.Money = money
			info.MoneyRatio = moneyRatio
			info.Ip = ipNums
			info.IpRatio = ipRatio
			info.Meals = meals
			info.Expire = expTime
			info.ExpireTime = v.Expire
			info.IpStr = ipNums + unit
			info.MoneyStr = "$" + money
			info.PayType = v.PayType
			info.PayTypeArr = payTypeArr

			if is_use == 1 {
				resList = append(resList, info)
				//if couType == "ip" || couType == "money" {
				//}
			}
		}

		//排序
		sort.SliceStable(resList, func(i, j int) bool {
			return resList[i].ExpireTime < resList[j].ExpireTime
		})
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resList)
	return
}

// @BasePath /api/v1
// @Summary 获取优惠券弹窗
// @Schemes
// @Description 获取优惠券弹窗
// @Tags 优惠券
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {array} map[string]interface{}  "show: 是否显示, data: 弹窗数据"
// @Router /web/coupon/popup [post]
func GetCouponPopup(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, e.SUCCESS, msg, nil)
		return
	}
	resInfo := map[string]interface{}{}
	resInfo["show"] = 0
	if userInfo.IsPay == "true" { // 付费后不弹
		JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resInfo)
		return
	}
	uid := userInfo.Id
	has := models.GetClickLog(uid) // 今天弹过不弹
	if has.Id > 0 {
		JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resInfo)
		return
	}
	_, count := models.GetCouponListByUserType(uid, "no_pay")
	if count > 0 {
		resInfo["show"] = 1
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resInfo)
	return
}

// @BasePath /api/v1
// @Summary 优惠券弹窗点击
// @Schemes
// @Description 获取优惠券弹窗
// @Tags 优惠券
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {array} map[string]interface{}  "show: 是否显示, data: 弹窗数据"
// @Router /web/coupon/popup_click [post]
func ClickCouponPopup(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		if resCode == e.SESSION_EXPIRED {
			JsonReturn(c, e.ERROR, msg, nil)
			return
		}
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id
	_ = models.AddPopupClickLog(uid, userInfo.Username, userInfo.Email, "web", c.ClientIP(), util.Lingchen())
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", gin.H{})
	return
}
