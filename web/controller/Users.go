package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	emailSender "api-360proxy/web/service/email"
	"fmt"
	"github.com/gin-gonic/gin"
	"math"
	"strings"
	"time"
)

// @BasePath /api/v1
// @Summary 自动登录
// @Description 自动登录
// @Tags 登陆注册相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {array} models.ResUser{}
// @Router /web/user/auto_login [post]
func LoginAuto(c *gin.Context) {
	session := c.DefaultPostForm("session", "")
	sessionInfo, err := models.GetSessionBySn(session)
	if err != nil {
		JsonReturn(c, e.SESSION_EXPIRED, "__T_SESSION_ERROR", nil)
		return
	}
	uid := sessionInfo.Uid

	if sessionInfo.LoginIp != c.ClientIP() {
		JsonReturn(c, e.ERROR, "__T_NOT_LOGIN", nil)
		return
	}
	err, user := models.GetUserById(uid)

	if err != nil {
		JsonReturn(c, e.ERROR, "__T_USER_INFO_ERROR", nil)
		return
	}
	if user.Status == 2 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_DISABLED", nil)
		return
	}
	if user.Status == 3 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_DISPUTE", nil)
		return
	}
	result := ResUserInfo(session, c.ClientIP(), user)
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
	return
}

// @BasePath /api/v1
// @Summary 获取用户信息
// @Description 获取用户信息
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {array} models.ResUser{}
// @Router /web/user/user_info [post]
func GetInfo(c *gin.Context) {
	session := c.DefaultPostForm("session", "")
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	result := ResUserInfo(session, c.ClientIP(), user)
	JsonReturn(c, 0, "__T_SUCCESS", result)
	return
}

// @BasePath /api/v1
// @Summary 根据邮箱获取用户信息
// @Description 根据邮箱获取用户信息
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param email formData string false "邮箱"
// @Param invite formData string false "邀请码"
// @Param package_id formData string false "套餐ID"
// @Produce json
// @Success 0 {object} map[string]interface{}
// @Router /web/user/email_info [post]
func GetInfoByEmail(c *gin.Context) {
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))
	invite_code := strings.TrimSpace(c.DefaultPostForm("invite", ""))
	package_id := strings.TrimSpace(c.DefaultPostForm("package_id", ""))

	result := map[string]interface{}{}
	is_pay := 0
	var err error
	user := models.Users{}
	if email != "" {
		// 获取用户信息
		err, user = models.GetUserByEmail(email)
	}

	packageInfo := models.GetSocksPackageInfoById(util.StoI(package_id))

	//专属特权  新版配置
	exInfo := models.ResExclusiveOffer{}

	is_ex := 0
	if packageInfo.PakType == "agent" {
		result["wallet"] = user.Wallet //钱包地址
		result["email"] = email
		result["is_pay"] = is_pay  //0未付费过  1已付费过
		result["is_ex"] = is_ex    //是否专属 1是
		result["ex_info"] = exInfo //专属信息
		JsonReturn(c, 0, "__T_SUCCESS", result)
		return
	}
	if err == nil && user.Id > 0 {
		if user.IsPay == "true" {
			is_pay = 1
		}
		err_i, pUser := models.GetUserInviterByUid(user.Id)
		if err_i == nil && pUser.ID > 0 {
			invite_code = strings.ToLower(pUser.InviterUsername)
		} else {
			invite_code = ""
		}
	} else {
		invite_code = strings.ToLower(invite_code)
	}
	if invite_code != "" {
		exclusiveInfo := models.GetExclusiveByCode(invite_code)
		if exclusiveInfo.Id > 0 {
			is_ex = 1
			exInfo.Percent = util.FtoS(exclusiveInfo.Ratio*100) + "%"
			exInfo.Ratio = exclusiveInfo.Ratio
			exInfo.Img = exclusiveInfo.Img
			exInfo.Code = exclusiveInfo.Code
			//专属折扣
			if exclusiveInfo.Money > 0 {
				if packageInfo.Price < exclusiveInfo.Money {
					exInfo.Discount = fmt.Sprintf("%.0f", (1-exclusiveInfo.DiscountLt)*100) + "%"
					exInfo.Money = math.Round(packageInfo.Price * (1 - exclusiveInfo.DiscountLt))
				} else {
					exInfo.Discount = fmt.Sprintf("%.0f", (1-exclusiveInfo.Discount)*100) + "%"
					exInfo.Money = math.Round(packageInfo.Price * (1 - exclusiveInfo.Discount))
				}
			}
		}
	}

	result["wallet"] = user.Wallet //钱包地址
	result["email"] = email
	result["is_pay"] = is_pay  //0未付费过  1已付费过
	result["is_ex"] = is_ex    //是否专属 1是
	result["ex_info"] = exInfo //专属信息

	JsonReturn(c, 0, "__T_SUCCESS", result)
	return
}

// @BasePath /api/v1
// @Summary 绑定邮箱
// @Description 绑定邮箱
// @Tags 登陆注册相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param email formData string true "邮箱"
// @Param password formData string true "密码"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/user/bind_email [post]
func BindEmail(c *gin.Context) {
	session := c.DefaultPostForm("session", "")
	email := c.DefaultPostForm("email", "")
	//code := c.DefaultPostForm("code", "")
	password := c.DefaultPostForm("password", "")
	//vtype := "bind"
	if session == "" {
		JsonReturn(c, e.ERROR, "__T_SESSION_EMPTY", nil)
		return
	}
	if email == "" {
		JsonReturn(c, -1, "__T_ACCOUNT_ERROR", gin.H{})
		return
	}
	//if code == "" {
	//	JsonReturn(c, -1, "__T_CODE_ERROR", gin.H{})
	//	return
	//}
	if password == "" {
		JsonReturn(c, -1, "__T_PASSWORD_EMPTY", gin.H{})
		return
	}
	if !util.CheckNewPwd(password) {
		JsonReturn(c, -1, "__T_PASSWORD_FORMAT", gin.H{})
		return
	}

	//isCheck := models.CheckVerifyCode(code, email,vtype)
	//if !isCheck {
	//	JsonReturn(c, -1, "__T_VERIFY_ERROR", gin.H{})
	//	return
	//}

	errs, uid := GetUIDbySession(session)
	if errs == false {
		JsonReturn(c, -1, "__T_SESSION_EXPIRE", gin.H{})
		return
	}
	err, user := models.GetUserById(uid)
	if err == nil && user.Email != "" {
		JsonReturn(c, -1, "__T_BOUND_EMAIL", nil)
		return
	}
	err1, newuser := models.GetUserByEmail(email)
	if err1 == nil && newuser.Id > 0 {
		JsonReturn(c, -1, "__T_ACCOUNT_EXIST", nil)
		return
	}

	user.Password = util.PassEncode(password, user.Username, 0)
	user.PlaintextPassword = password
	r := models.UpdateUserById(user.Id, &user)
	if !r {
		JsonReturn(c, -1, "bind email fail", nil)
		return
	}
	JsonReturn(c, 0, "__T_SUCCESS", nil)
	return
}

// @BasePath /api/v1
// @Summary 绑定钱包地址
// @Description 绑定钱包地址
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param wallet formData string false "钱包地址"
// @Param email formData string false "邮箱"
// @Param j_source formData string false "竞价平台"
// @Param j_code formData string false "竞价code"
// @Param j_domain formData string false "竞价域名"
// @Param p formData string false "终端"
// @Param origin formData string false "来源"
// @Param inviter_type formData string false "邀请类型 1链接  2填写推广码"
// @Param inviter_code formData string false "邀请码"
// @Produce json
// @Success 0 {object} map[string]interface{} "wallet：钱包地址"
// @Router /web/user/bind_wallet [post]
func BindWallet(c *gin.Context) {
	wallet := strings.TrimSpace(c.DefaultPostForm("wallet", ""))
	ip := c.ClientIP()
	params := GetParams(c)
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))
	jjSource := c.DefaultPostForm("j_source", "")                            // 竞价平台
	jjCode := c.DefaultPostForm("j_code", "")                                // 竞价code
	jjDomain := c.DefaultPostForm("j_domain", "")                            // 竞价域名
	p := c.DefaultPostForm("p", "")                                          // 终端
	origin := c.DefaultPostForm("origin", "")                                // ""
	inviter_type := strings.TrimSpace(c.DefaultPostForm("inviter_type", "")) // 邀请类型 1链接  2填写推广码
	inviter_code := strings.TrimSpace(c.DefaultPostForm("inviter_code", "")) // 邀请码
	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", nil)
		return
	}
	if !util.CheckEmail(email) {
		JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", nil)
		return
	}
	if wallet == "" {
		JsonReturn(c, -1, "__T_WALLET_ERROR", gin.H{})
		return
	}
	err, userHas := models.GetUserInfo(map[string]interface{}{"wallet": wallet, "status": 1})
	if err == nil && userHas.Id > 0 {
		JsonReturn(c, -1, "__T_WALLET_HAS_BIND", gin.H{})
		return
	}

	err, user := models.GetUserByEmail(email)
	uid := user.Id
	if err != nil || uid == 0 { //如果 未注册账号 就先注册账号
		username := GetUsername()
		params.DeviceOs = "web"
		platform := "web"
		switch p {
		case "1":
			platform = "web"
			break
		case "2":
			platform = "pc"
			break
		case "3":
			platform = "ios"
			break
		case "4":
			platform = "wap"
		case "5":
			platform = "android"
			break
		default:
			platform = "web"
		}

		// 邀请信息
		inviter_id := 0
		inviter_username := ""
		if inviter_code != "" {
			if inviter_type == "" { //邀请类型   1 链接  2邀请码
				inviter_type = "1"
			}
			// 判断是否是活动邀请码
			if strings.Contains(inviter_code, "act_") {
				err, pUser := models.GetUserActivityInviterByMap(map[string]interface{}{"inviter_code": inviter_code})
				if err != nil || pUser.Id == 0 {
					JsonReturn(c, e.ERROR, "Invitation code error", nil)
					return
				}
				inviter_id = pUser.Uid
				inviter_username = pUser.InviterCode
			} else {
				err, pUser := models.GetUserInviterByMap(map[string]interface{}{"inviter_code": inviter_code})
				if err != nil || pUser.ID == 0 {
					JsonReturn(c, e.ERROR, "Invitation code error", nil)
					return
				}
				inviter_id = pUser.Uid
				inviter_username = pUser.InviterCode
			}
		}
		password := util.RandStr("y", 8)
		regResult, msg, userInfo := DoRegister(ip, email, "", platform, password, params, username, origin, inviter_type, inviter_id, inviter_username, jjSource, jjCode, jjDomain, "", "", "")

		// 注册完成发送邮件
		res, msg_s := emailSender.SendEmail(email, password, ip)
		var sendResult = ""
		if res != true {
			sendResult = email + " 注册发送密码 发送邮件失败" + msg_s
		} else {
			sendResult = email + " 注册发送密码 发送邮件成功" + msg_s
		}
		fmt.Println("绑定创建用户", sendResult)
		if !regResult {
			JsonReturn(c, e.ERROR, msg, nil)
			return
		}
		uid = userInfo.Id
	}

	user.Wallet = wallet
	updateParams := make(map[string]interface{})
	updateParams["wallet"] = wallet

	r := models.EditUserByMap(map[string]interface{}{"id": uid}, updateParams)
	if r == nil {
		res := make(map[string]interface{})
		res["wallet"] = wallet
		AddWalletRecord(wallet, email, uid, 1, c.ClientIP())
		JsonReturn(c, 0, "__T_BIND_WALLET_SUCCESS", res)
		return
	}
	JsonReturn(c, -1, "__T_FAIL", nil)
	return
}

// @BasePath /api/v1
// @Summary 解绑钱包地址
// @Description 解绑钱包地址
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param email formData string false "邮箱"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/user/unbind_wallet [post]
func UnBindWallet(c *gin.Context) {
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))
	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", nil)
		return
	}
	if !util.CheckEmail(email) {
		JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", nil)
		return
	}

	err, user := models.GetUserByEmail(email)
	if err == nil && user.Id > 0 {
		updateParams := make(map[string]interface{})
		updateParams["wallet"] = ""
		AddWalletRecord("", email, user.Id, 2, c.ClientIP())
		r := models.EditUserByMap(map[string]interface{}{"id": user.Id}, updateParams)
		if r == nil {
			JsonReturn(c, 0, "__T_UNBIND_WALLET_SUCCESS", nil)
			return
		}
		JsonReturn(c, -1, "__T_FAIL", nil)
		return
	} else {
		JsonReturn(c, -1, "__T_ACCOUNT_ERROR", gin.H{})
		return
	}
}

// @BasePath /api/v1
// @Summary 修改 /重置密码
// @Description 修改 /重置密码
// @Tags 登陆注册相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param old_password formData string true "旧密码"
// @Param password formData string true "新密码"
// @Param re_password formData string true "确认密码"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/user/reset_pwd [post]
func ResetPass(c *gin.Context) {
	old_password := strings.TrimSpace(c.DefaultPostForm("old_password", ""))
	password := strings.TrimSpace(c.DefaultPostForm("password", ""))
	re_pass := strings.TrimSpace(c.DefaultPostForm("re_password", ""))

	if password == "" {
		JsonReturn(c, -1, "__T_PASSWORD_EMPTY", gin.H{})
		return
	}

	//if !util.CheckNewPwd(password) {
	//	JsonReturn(c, -1, "__T_NEW_PASSWORD_FORMAT_ERROR", gin.H{})
	//	return
	//}
	checkRes, checkMsg := util.CheckPwdNew(password) //密码验证
	if !checkRes {
		JsonReturn(c, -1, checkMsg, nil)
		return
	}

	if password != re_pass {
		JsonReturn(c, -1, "__T_RETRY_PASSWORD_NOT_MATCH", gin.H{})
		return
	}
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	if !CheckPwd(user.Password, old_password, user.Username) {
		JsonReturn(c, e.ERROR, "__T_OLD_PASSWORD_FAIL", nil)
		return
	}

	user.Password = util.PassEncode(password, user.Username, 0)
	user.PlaintextPassword = password
	r := models.UpdateUserById(user.Id, &user)

	//删除改用户所有的登录信息
	models.DeleteSession(map[string]interface{}{
		"uid": user.Id,
	})

	if !r {
		JsonReturn(c, -1, "find password fail", nil)
		return
	}
	JsonReturn(c, 0, "__T_SUCCESS", nil)
	return
}

// @BasePath /api/v1
// @Summary 用户购买记录
// @Description 用户购买记录
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param limit formData string false "每页数据量"
// @Param page formData string false "页数"
// @Produce json
// @Success 0 {array} map[string]interface{} "total:总记录数,total_page:总页数,lists:记录列表 "
// @Router /web/user/pay_list [post]
func UserOrderList(c *gin.Context) {
	limitStr := c.DefaultPostForm("limit", "12")
	pageStr := c.DefaultPostForm("page", "1")
	if limitStr == "" {
		limitStr = "10"
	}
	if pageStr == "" {
		pageStr = "1"
	}
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	limit := util.StoI(limitStr)
	page := util.StoI(pageStr)
	offset := (page - 1) * limit

	_, lists := models.GetSocksOrderPage(uid, offset, limit)
	total := models.GetOrderListCount(uid)
	if total > 0 {
		_, payPlatList := models.GetPayPlatConfList()
		resList := []models.ResSocksOrder{}
		for _, v := range lists {
			payPlat := "Other"
			for _, vp := range payPlatList {
				if v.PayPlat == vp.Id {
					payPlat = vp.ShowName
				}
			}
			pakType := ""
			char := "IPs"
			if v.PakType == "normal" {
				pakType = "BALANCE"
			} else if v.PakType == "static" {
				pakType = "Static Residential Proxies"
			} else if v.PakType == "flow" {
				pakType = "Residential Proxies"
				char = "GB"
			} else if v.PakType == "flow_agent" {
				pakType = "Residential Proxies (Business)"
				char = "GB"
			} else if v.PakType == "agent" {
				pakType = "Socks5 Proxies (Business)"
				char = "IPs"
			} else if v.PakType == "dynamic_isp" {
				pakType = "Dynamic ISP Proxies"
				char = "GB"
			} else {
				pakType = "Socks5 Proxies"
			}
			pakName := v.PakName
			if v.PakType == "static" {
				//msg := TextReturn(c, "__T_STATIC_AREA_"+strings.ToUpper(v.PakRegion))
				msg := strings.ToUpper(strings.Trim(v.PakRegion, ","))
				pakName = msg + "-" + v.PakName
			}
			info := models.ResSocksOrder{}
			info.PakId = v.PakId
			info.OrderId = v.OrderId
			info.Name = v.Name
			info.Value = v.Value
			info.Total = v.Value + v.Give + v.Gift + v.Reward
			info.Give = v.Give + v.Gift + v.Reward
			info.Char = char
			if v.PakType == "flow" || v.PakType == "flow_agent" {
				info.Value = v.Value / 1024 / 1024 / 1024
				info.Total = (v.Value + v.Give + v.Gift + v.Reward) / 1024 / 1024 / 1024
				info.Give = (v.Give + v.Gift + v.Reward) / 1024 / 1024 / 1024
			}
			info.TrueMoney = v.TrueMoney
			info.PayTime = util.GetTimeStr(v.PayTime, "Y-m-d H:i:s")
			info.PayPlat = payPlat
			info.PaySn = v.PaySn
			info.CreateTime = util.GetTimeStr(v.CreateTime, "Y-m-d H:i:s")
			info.Plat = v.Plat
			info.PakName = pakName
			info.PakType = pakType
			info.Invoice = v.Invoice
			resList = append(resList, info)
		}
		totalPage := int(math.Ceil(float64(total) / float64(limit)))

		resData := map[string]interface{}{}
		resData["lists"] = resList
		resData["total"] = total
		resData["total_page"] = totalPage
		JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resData)
		return
	} else {
		JsonReturn(c, e.ERROR, "__T_NO_DATA", gin.H{})
		return
	}
}

// @BasePath /api/v1
// @Summary 用户IP消耗记录
// @Description 用户IP消耗记录
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param limit formData string false "每页数据量"
// @Param page formData string false "页数"
// @Produce json
// @Success 0 {array} map[string]interface{} "total:总记录数,total_page:总页数,lists:记录列表 "
// @Router /web/user/used_ip_list [post]
func UsedIp(c *gin.Context) {
	limitStr := c.DefaultPostForm("limit", "10")
	pageStr := c.DefaultPostForm("page", "1")
	if limitStr == "" {
		limitStr = "10"
	}
	if pageStr == "" {
		pageStr = "1"
	}

	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	limit := util.StoI(limitStr)
	page := util.StoI(pageStr)
	offset := (page - 1) * limit

	infoLists := models.GetIpDeductionPage(uid, offset, limit)
	lists := []models.ResLogExtract{}
	for _, v := range infoLists {
		info := models.ResLogExtract{}
		info.UserName = v.UserName
		info.CreateTime = util.GetTimeStr(v.CreateTime, "Y/m/d H:i:s")
		lists = append(lists, info)
	}
	total := models.GetIpDeductionCount(uid)
	totalPage := int(math.Ceil(float64(total) / float64(limit)))
	result := map[string]interface{}{
		"total":      total,
		"total_page": totalPage,
		"lists":      lists,
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
	return
}

// 钱包地址 绑定解绑 记录
// status 1 绑定  2取消
func AddWalletRecord(wallet, email string, uid, status int, ip string) {
	nowTime := util.GetNowInt()
	err := models.AddUserWallet(models.UserWalletModel{
		Uid:        uid,
		Email:      email,
		Wallet:     wallet,
		Status:     status,
		Ip:         ip,
		CreateTime: nowTime,
	})
	fmt.Println("AddWalletRecord", err)
}

// @BasePath /api/v1
// @Summary 用户IP消耗日期和数量
// @Description 用户IP消耗日期和数量
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} map[string]interface{} "x_data : x轴数据, y_data: y轴数据, cate:分类名称"
// @Router /web/user/used_ip_data [post]
func GetUserUseIp(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	today := util.GetTodayTime()
	create := today - 10*86400
	list := models.GetIpDeductionByGroup(uid, create)
	x_data := []string{}
	for i := 0; i <= 10; i++ {
		day := create + i*86400
		dayStr := util.GetTimeStr(day, "Y-m-d")
		x_data = append(x_data, dayStr)
	}
	cateName := []string{
		"Number",
		"LongtimeUS",
		"LongtimeGlobal",
	}
	kvInfo := map[string]int{}
	cateInfo := map[string][]models.NumIpInfo{}
	for _, v := range list {
		if v.Cate == "" {
			v.Cate = "Number"
		}
		if v.Cate == "us" {
			v.Cate = "LongtimeUS"
		}
		if v.Cate == "all" {
			v.Cate = "LongtimeGlobal"
		}
		kvInfo[v.Today+"|"+v.Cate] = v.Num
		cateInfo[v.Cate] = append(cateInfo[v.Cate], v)
	}
	fmt.Println(kvInfo)
	y_data := map[string][]int{}
	for k, _ := range cateInfo {
		infoData := []int{}
		for _, vx := range x_data {
			str := vx + "|" + k
			//fmt.Println(str)
			info, ok := kvInfo[str]
			//fmt.Println(str,info)
			if !ok {
				info = 0
			}
			infoData = append(infoData, info)
		}
		//for _, vv := range vInfo {
		//	for _, vx := range x_data {
		//		str := vx + "|" + vv.Cate
		//		info, ok := kvInfo[str]
		//		if !ok {
		//			info = 0
		//		}
		//		infoData = append(infoData, info)
		//
		//	}
		//}
		y_data[k] = infoData
	}
	////fmt.Println(cateInfo)
	//y2_data := []interface{}{}
	//for _, vx := range x_data {
	//	for _, vc := range cateName {
	//		str := vx + "|" + vc
	//		info ,ok := kvInfo[str]
	//		if !ok{
	//			info = 0
	//		}
	//		y2_data = append(y2_data, info)
	//	}
	//}

	data := map[string]interface{}{}
	data["x_data"] = x_data
	data["y_data"] = y_data
	data["cate"] = cateName
	//data["cateInfo"] = cateInfo
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", data)
	return
}

// @BasePath /api/v1
// @Summary 用户IP消耗日期和数量
// @Description 用户IP消耗日期和数量
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param start_date formData string false "开始时间"
// @Param end_date formData string false "结束时间"
// @Param cate formData string false "类型 isp、static、flow"
// @Produce json
// @Success 0 {object} map[string]interface{} 	"x_data : x轴数据, y_data : y轴数据, cate :分类名称, unit :单位, max:最大值"
// @Router /web/user/used_data [post]
func GetUserUseIpNew(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	start_date := c.DefaultPostForm("start_date", "")
	end_date := c.DefaultPostForm("end_date", "")
	cate := c.DefaultPostForm("cate", "isp") // 类型 isp、static、flow

	var start, end int
	x_data := []string{}
	if start_date == "" || end_date == "" {
		today := util.GetTodayTime()
		create := today - 10*86400
		for i := 0; i <= 7; i++ {
			day := create + i*86400
			dayStr := util.GetTimeStr(day, "Y-m-d")
			x_data = append(x_data, dayStr)
		}
		start = create
		end = today
	} else {
		start = util.StoI(util.GetTimeStamp(start_date, "Y-m-d"))
		end = util.StoI(util.GetTimeStamp(end_date, "Y-m-d"))

		for i := 0; i <= (end-start)/86400; i++ {
			day := start + i*86400
			dayStr := util.GetTimeStr(day, "Y-m-d")
			x_data = append(x_data, dayStr)
		}
	}
	list := models.GetUserUsed(uid, cate, start, end)

	var unit string
	var cateName []string
	if cate == "isp" {
		unit = "IPs"
		cateName = []string{
			"Number",
		}
	} else if cate == "static" {
		unit = "IPs"
		cateName = []string{
			"LongtimeUS",
			"LongtimeGlobal",
		}
	} else if cate == "flow" {
		unit = "GB"
		cateName = []string{
			"Flow",
		}
	}

	kvInfo := map[string]int64{}
	cateInfo := map[string][]models.StUsedToday{}
	for _, vv := range cateName {
		cateInfo[vv] = []models.StUsedToday{}
		for _, v := range x_data {
			kvInfo[v+"|"+vv] = 0
		}
	}

	var max int64
	for _, v := range list {
		if v.Cate == "isp" {
			v.Cate = "Number"
		}
		if v.Cate == "us" {
			v.Cate = "LongtimeUS"
		}
		if v.Cate == "all" {
			v.Cate = "LongtimeGlobal"
		}
		if v.Cate == "custom" {
			v.Cate = "Flow"
		}
		kvInfo[util.GetTimeStr(v.Today, "Y-m-d")+"|"+v.Cate] = v.Num
		cateInfo[v.Cate] = append(cateInfo[v.Cate], v)

		if v.Num > max {
			max = v.Num
		}
	}
	if unit == "GB" {
		if max < (1024 * 1024 * 1024) {
			unit = "Mb"
		}
	}
	fmt.Println(cateInfo)
	fmt.Println(kvInfo)
	max = 0
	y_data := map[string][]int64{}
	for k, _ := range cateInfo {
		fmt.Println(k)
		infoData := []int64{}
		for _, vx := range x_data {
			fmt.Println(vx)
			if len(kvInfo) > 0 {
				str := vx + "|" + k
				info, ok := kvInfo[str]
				if !ok {
					info = 0
				}
				if unit == "GB" {
					info = int64(math.Ceil(float64(info) / 1024 / 1024 / 1024))
				} else {
					info = int64(math.Ceil(float64(info) / 1024 / 1024))
				}
				if info > max {
					max = info
				}
				infoData = append(infoData, info)
			} else {
				infoData = append(infoData, 0)
			}
		}
		y_data[k] = infoData
	}

	data := map[string]interface{}{}
	data["x_data"] = x_data
	data["y_data"] = y_data
	data["cate"] = cateName
	data["unit"] = unit
	data["max"] = max
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", data)
	return
}

// 查询是否存在主账户，不存在就添加
func GetAddUserAccount(userInfo models.Users) {

	// 查询该用户下面是否存在主账户 若不存在则添加账户信息
	accountInfo := models.UserAccount{}
	_, accountInfo = models.GetUserAccountMaster(userInfo.Id)

	account := ""
	password := ""
	if accountInfo.Id == 0 {
		account = userInfo.Username
		password = util.RandStr("r", 8)

		data := models.UserAccount{}
		data.Status = 1
		data.Remark = ""
		data.Uid = userInfo.Id
		data.Account = account
		data.Password = password
		data.Master = 1
		data.FlowUnit = "GB"
		data.CreateTime = int(time.Now().Unix())

		err, id := models.AddProxyAccount(data)
		fmt.Println("err_id", err, id)
	}
}
