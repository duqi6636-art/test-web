package controller

import (
	"api-360proxy/service/google"
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

var googleKey = "360Proxy"

// google 验证器
// @BasePath /api/v1
// @Summary 创建信息
// @Description 创建信息
// @Tags google验证器
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} map[string]interface{} "account：用户标识，secret ：密钥，url：二维码地址，qrcode：二维码地址带参数"
// @Router /web/auth/auth_info [post]
func GetGoogleAuth(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	username := user.Username
	gooKey := fmt.Sprintf(googleKey+"(%s)", username)

	err, authInfo := models.GetUserGoogleAuthBy(gooKey)

	if err == nil && authInfo.ID != 0 {
		JsonReturn(c, e.ERROR, "__T_GOOGLE_AUTH_BIND", nil)
		return
	}
	gAuth := google.NewGoogleAuth()
	secret := gAuth.GetSecret()
	url := gAuth.GetQrcode(gooKey, secret)

	data := map[string]interface{}{}
	data["account"] = gooKey
	data["secret"] = secret
	data["url"] = url
	data["qrcode"] = strings.TrimRight(models.GetConfigV("API_DOMAIN_URL"), "/") + "/qrcode?data=" + url
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", data)
	return
}

// 绑定信息
// @BasePath /api/v1
// @Summary 绑定信息
// @Description 绑定信息
// @Tags google验证器
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param code formData string true "验证码"
// @Param secret formData string true "密钥"
// @Produce json
// @Success 0 {object} interface{} "success：绑定成功"
// @Router /web/auth/bing_auth [post]
func VerifyCodeBind(c *gin.Context) {
	code := c.DefaultPostForm("code", "")
	secret := c.DefaultPostForm("secret", "")
	if code == "" {
		JsonReturn(c, -1, "__T_CODE_ERROR", gin.H{})
		return
	}
	if secret == "" {
		JsonReturn(c, -1, "Key error", gin.H{})
		return
	}
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	username := user.Username
	gooKey := fmt.Sprintf(googleKey+"(%s)", username)
	err, authInfo := models.GetUserAuthByUsername(gooKey, "google_auth")

	if err == nil && authInfo.ID != 0 {
		JsonReturn(c, e.ERROR, "__T_GOOGLE_AUTH_BIND", nil)
		return
	}

	isOpen := 1
	err, authInfo2 := models.GetUserAuthByUid(user.Id, "email")
	if err == nil && authInfo2.ID != 0 && authInfo2.IsOpen == 1 {
		isOpen = 0
	}

	gAuth := google.NewGoogleAuth()
	result, _err := gAuth.VerifyCode(secret, code)
	if _err != nil || result == false {
		JsonReturn(c, e.ERROR, "__T_GOOGLE_AUTH_FAIL", nil)
		return
	}
	authMap := models.UserGoogleAuth{}
	authMap.Username = gooKey
	authMap.GoogleKey = secret
	authMap.Uid = user.Id
	authMap.Cate = "google_auth"
	authMap.IsOpen = isOpen
	authMap.Create_time = util.GetNowInt()
	err = models.CreateUserGoogleAuth(authMap)
	if err != nil {
		JsonReturn(c, e.ERROR, "error", nil)
		return
	}
	JsonReturn(c, e.SUCCESS, "__T_GOOGLE_AUTH_SUCCESS", nil)
	return
}

// 验证信息
// @BasePath /api/v1
// @Summary 验证信息
// @Description 验证信息
// @Tags google验证器
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param code formData string true "验证码"
// @Produce json
// @Success 0 {object} interface{} "success：验证成功"
// @Router /web/auth/verify_code [post]
func VerifyCode(c *gin.Context) {
	username := c.DefaultPostForm("username", "")
	code := c.DefaultPostForm("code", "")
	if code == "" {
		JsonReturn(c, -1, "__T_CODE_ERROR", gin.H{})
		return
	}
	platform := c.DefaultPostForm("platform", "web") // 终端
	if platform == "" {
		platform = "web"
	}

	err, user := models.GetUserByUsername(username)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_SESSION_ERROR", nil)
		return
	}
	if user.Status == 2 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_DISABLED", nil)
		return
	}
	if user.Status == 3 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_DISABLED", nil)
		return
	}

	gooKey := fmt.Sprintf(googleKey+"(%s)", username)
	err, authInfo := models.GetUserAuthByUsername(gooKey, "google_auth")
	if err != nil && authInfo.ID == 0 {
		JsonReturn(c, e.ERROR, "__T_GOOGLE_AUTH_NO", nil)
		return
	}
	ip := c.ClientIP()
	sessionInfo, err := models.GetSessionByUsername(username, ip, platform)
	if err != nil || sessionInfo.ID == 0 {
		JsonReturn(c, e.ERROR, "__T_SESSION_ERROR", nil)
		return
	}
	session := sessionInfo.SessionId
	nowTime := util.GetNowInt()
	ipInfo := GetIpInfo(ip)
	ga := google.NewGoogleAuth()
	ret, err := ga.VerifyCode(authInfo.GoogleKey, code)
	if ret == true && err == nil {
		//操作登录设备记录
		errD, hasD := models.GetLoginDeviceBy(user.Id, ip)
		if errD != nil || hasD.ID == 0 {
			addDevice := models.LoginDevices{}
			addDevice.Cate = "google"
			addDevice.Uid = user.Id
			addDevice.Username = user.Username
			addDevice.Email = user.Email
			addDevice.Device = "cherry-" + util.RandStr("r", 16)
			addDevice.DeviceNo = ip
			addDevice.Platform = "web"
			addDevice.Ip = ip
			addDevice.Trust = nowTime //
			addDevice.Country = ipInfo.Country
			addDevice.State = ipInfo.Province
			addDevice.City = ipInfo.City
			addDevice.Session = session
			addDevice.UpdateTime = nowTime
			addDevice.CreateTime = nowTime
			models.AddLoginDevice(addDevice)
		} else {
			up := map[string]interface{}{
				"update_time": nowTime,
				"trust":       nowTime,
			}
			models.EditLoginDeviceInfo(hasD.ID, up)
		}
		data := ResUserInfo(session, ip, user)
		JsonReturn(c, e.SUCCESS, "__T_VERIFY_SUCCESS", data)
		return
	}
	JsonReturn(c, e.ERROR, "__T_GOOGLE_AUTH_FAIL", nil)
	return
}

// 解绑信息
// @BasePath /api/v1
// @Summary 解绑信息
// @Description 解绑信息
// @Tags google验证器
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} interface{} "success：解绑成功"
// @Router /web/auth/unbind_auth [post]
func VerifyCodeUnBind(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	err := models.DeleteUserGoogleAuth(uid)
	if err != nil {
		JsonReturn(c, e.ERROR, "error", nil)
		return
	}
	JsonReturn(c, e.SUCCESS, "__T_UNBIND_SUCCESS", nil)
	return
}

// 切换登录认证方式开关（email 或 google），并保持两者互斥：开启一种则关闭另一种

func SetOpen(c *gin.Context) {
	// 参数绑定和验证
	var params struct {
		IsOpen string `form:"is_open" binding:"required,oneof=on off 1 0"`
		Cate   string `form:"cate" binding:"required,oneof=email google_auth"`
	}

	if err := c.ShouldBind(&params); err != nil {
		JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil)
		return
	}

	cate := params.Cate
	open := params.IsOpen

	// 用户身份校验
	resCode, msg, user := DealUser(c)
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	// 获取用户认证记录
	authList, err := models.GetAuthByUid(user.Id)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_FAIL", nil)
		return
	}

	if len(authList) == 0 {
		JsonReturn(c, e.ERROR, "__T_AUTH_UNBIND", nil)
		return
	}

	openInfo := models.UserGoogleAuth{}
	closeInfo := models.UserGoogleAuth{}
	for _, val := range authList {
		if val.Cate == cate {
			openInfo = val
		} else {
			closeInfo = val
		}
	}

	if openInfo.ID == 0 {
		JsonReturn(c, e.ERROR, "__T_AUTH_UNBIND", nil)
		return
	}

	upInfo := map[string]interface{}{}
	shutInfo := map[string]interface{}{}
	msgs := "__T_SUCCESS"
	if open == "on" || open == "1" {
		upInfo["is_open"] = 1
		shutInfo["is_open"] = 0
		if closeInfo.ID != 0 {
			models.EditAuthById(closeInfo.ID, shutInfo)
		}
		msgs = "__T_SUCCESS_OPEN"
	} else {
		upInfo["is_open"] = 0
		msgs = "__T_SUCCESS_CLOSE"
	}
	models.EditAuthById(openInfo.ID, upInfo)

	emailAuth := models.UserLoginAuth{}
	googleInfoAuth := models.UserLoginAuth{}
	resList, _ := models.GetAuthByUid(user.Id)
	for _, val := range resList {
		lc := strings.ToLower(val.Cate)
		if lc == "email" {
			emailAuth = models.UserLoginAuth{IsOpen: val.IsOpen, Cate: lc, Info: val.Username}
		} else {
			googleInfoAuth = models.UserLoginAuth{IsOpen: val.IsOpen, Cate: "google_auth", Info: val.Username}
		}
	}
	result := map[string]models.UserLoginAuth{
		"email_auth":  emailAuth,
		"google_auth": googleInfoAuth,
	}

	JsonReturn(c, e.SUCCESS, msgs, result)
	return
}

func BindEmailAuth(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	code := c.DefaultPostForm("code", "")
	email := c.DefaultPostForm("email", "")
	if code == "" {
		JsonReturn(c, -1, "__T_CODE_ERROR", gin.H{})
		return
	}
	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", map[string]string{})
		return
	}
	if !util.CheckEmail(email) {
		JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", map[string]string{})
		return
	}

	vtype := "bind_email"
	errC, codeInfo := models.CheckVerifyByCode(code, email, vtype)
	codeId := codeInfo.ID
	if errC != nil || codeId == 0 {
		JsonReturn(c, -1, "__T_VERIFY_ERROR", map[string]string{"class_id": "code"})
		return
	}
	if codeInfo.Ip != c.ClientIP() {
		JsonReturn(c, -1, "__T_VERIFY_ERROR", map[string]string{"class_id": "code"})
		return
	}

	err, authInfo := models.GetUserAuthByUid(uid, "email")
	if err == nil && authInfo.ID != 0 {
		JsonReturn(c, e.ERROR, "__T_EMAIL_AUTH_BIND", nil)
		return
	}

	isOpen := 1
	err, authInfo2 := models.GetUserAuthByUid(uid, "google_auth")
	if err == nil && authInfo2.ID != 0 && authInfo2.IsOpen == 1 {
		isOpen = 0
	}

	authMap := models.UserGoogleAuth{}
	authMap.Username = email
	authMap.GoogleKey = email
	authMap.Uid = user.Id
	authMap.Cate = "email"
	authMap.IsOpen = isOpen
	authMap.Create_time = util.GetNowInt()
	err = models.CreateUserGoogleAuth(authMap)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_EMAIL_AUTH_FAIL", nil)
		return
	}

	if codeId > 0 {
		// 验证码验证完成后销毁
		cRes := models.UpdateCodeStatus(map[string]interface{}{"id": codeId})
		fmt.Println(cRes)
	}
	JsonReturn(c, e.SUCCESS, "__T_EMAIL_AUTH_SUCCESS", nil)
	return
}

func UnBindEmailAuth(c *gin.Context) {
	code := c.DefaultPostForm("code", "")
	email := c.DefaultPostForm("email", "")

	if code == "" {
		JsonReturn(c, -1, "__T_CODE_ERROR", gin.H{})
		return
	}
	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", map[string]string{})
		return
	}
	if !util.CheckEmail(email) {
		JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", map[string]string{})
		return
	}
	vtype := "unbind_email"
	errC, codeInfo := models.CheckVerifyByCode(code, email, vtype)
	codeId := codeInfo.ID
	if errC != nil || codeId == 0 {
		JsonReturn(c, -1, "__T_VERIFY_ERROR", map[string]string{"class_id": "code"})
		return
	}
	if codeInfo.Ip != c.ClientIP() {
		JsonReturn(c, -1, "__T_VERIFY_ERROR", map[string]string{"class_id": "code"})
		return
	}
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	err := models.DeleteUserGoogleAuthByCate(uid, "email")
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_UNBIND_FAIL", nil)
		return
	}
	if codeId > 0 {
		// 验证码验证完成后销毁
		cRes := models.UpdateCodeStatus(map[string]interface{}{"id": codeId})
		fmt.Println(cRes)
	}
	JsonReturn(c, e.SUCCESS, "__T_UNBIND_SUCCESS", nil)
	return
}

func VerifyEmailCode(c *gin.Context) {
	username := c.DefaultPostForm("username", "")
	code := c.DefaultPostForm("code", "")
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))
	platform := c.DefaultPostForm("platform", "web") // 终端
	if platform == "" {
		platform = "web"
	}

	err, user := models.GetUserByUsername(username)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_SESSION_ERROR", nil)
		return
	}
	if user.Status == 2 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_DISABLED", nil)
		return
	}
	if user.Status == 3 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_DISABLED", nil)
		return
	}

	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", map[string]string{"class_id": "username"})
		return
	}
	if code == "" {
		JsonReturn(c, -1, "__T_VERIFY_EMPTY", map[string]string{"class_id": "code"})
		return
	}

	err, authInfo := models.GetUserAuthByUid(user.Id, "email")
	if err != nil && authInfo.ID == 0 {
		JsonReturn(c, e.ERROR, "__T_GOOGLE_AUTH_NO", nil)
		return
	}

	vtype := "check_login"
	// 验证 验证码
	errC, codeInfo := models.CheckVerifyByCode(code, email, vtype)
	codeId := codeInfo.ID
	if errC != nil || codeId == 0 {
		JsonReturn(c, -1, "__T_VERIFY_ERROR", map[string]string{"class_id": "code"})
		return
	}
	if codeInfo.Ip != c.ClientIP() {
		JsonReturn(c, -1, "__T_VERIFY_ERROR", map[string]string{"class_id": "code"})
		return
	}

	ip := c.ClientIP()
	sessionInfo, err := models.GetSessionByUsername(username, ip, platform)
	if err != nil || sessionInfo.ID == 0 {
		JsonReturn(c, e.ERROR, "__T_SESSION_ERROR", nil)
		return
	}
	session := sessionInfo.SessionId
	nowTime := util.GetNowInt()
	ipInfo := GetIpInfo(ip)
	if codeId > 0 {
		//操作登录设备记录
		errD, hasD := models.GetLoginDeviceBy(user.Id, ip)
		if errD != nil || hasD.ID == 0 {
			addDevice := models.LoginDevices{}
			addDevice.Cate = "email"
			addDevice.Uid = user.Id
			addDevice.Username = user.Username
			addDevice.Email = user.Email
			addDevice.Device = "cherry-" + util.RandStr("r", 16)
			addDevice.DeviceNo = ip
			addDevice.Platform = "web"
			addDevice.Ip = ip
			addDevice.Trust = nowTime //
			addDevice.Country = ipInfo.Country
			addDevice.State = ipInfo.Province
			addDevice.City = ipInfo.City
			addDevice.Session = session
			addDevice.UpdateTime = nowTime
			addDevice.CreateTime = nowTime
			models.AddLoginDevice(addDevice)
		} else {
			up := map[string]interface{}{
				"update_time": nowTime,
				"trust":       nowTime,
			}
			models.EditLoginDeviceInfo(hasD.ID, up)
		}
		// 验证码验证完成后销毁
		cRes := models.UpdateCodeStatus(map[string]interface{}{"id": codeId})
		fmt.Println(cRes)

		data := ResUserInfo(session, ip, user)
		JsonReturn(c, e.SUCCESS, "__T_VERIFY_SUCCESS", data)
		return
	}
	JsonReturn(c, e.ERROR, "__T_FAIL", nil)
	return
}

func GetUserAuthInfo(c *gin.Context) {
	resCode, msg, user := DealUser(c)
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	result := map[string]models.UserLoginAuth{
		"email_auth":  {},
		"google_auth": {},
	}
	authList, err := models.GetAuthByUid(user.Id)
	if err == nil {
		if len(authList) == 0 {
			JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
			return
		}
		for _, val := range authList {
			cate := strings.ToLower(val.Cate)
			switch cate {
			case "email":
				result["email_auth"] = models.UserLoginAuth{IsOpen: val.IsOpen, Cate: cate, Info: val.Username}
			case "google_auth":
				result["google_auth"] = models.UserLoginAuth{IsOpen: val.IsOpen, Cate: cate, Info: val.Username}
			default:
				models.AddLog(models.LogModel{Code: "user_auth_unknown_cate", Text: fmt.Sprintf("uid=%d cate=%s", user.Id, val.Cate), CreateTime: util.GetTimeStr(util.GetNowInt(), "Y-m-d H:i:s")})
			}
		}
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
	return
}

func SafeDevice(c *gin.Context) {
	lang := strings.ToLower(c.DefaultPostForm("lang", "en")) //语言
	resCode, msg, user := DealUser(c)                        //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	session := c.DefaultPostForm("session", "")
	listsInfo := models.ListLoginDevices(user.Id)
	resList := []models.ResLoginDevices{}
	for _, v := range listsInfo {
		device := v.DeviceNo
		if device == "" {
			device = v.Ip
		}
		online := 0
		if v.Session == session {
			online = 1
		}
		trust := 0
		if v.Trust > 0 {
			trust = 1
		}
		loginTime := v.CreateTime
		if v.UpdateTime > loginTime {
			loginTime = v.UpdateTime
		}
		info := models.ResLoginDevices{}
		info.Id = v.ID
		info.Uid = v.Uid
		info.Cate = v.Cate
		info.Device = v.Device
		info.Platform = v.Platform
		info.Ip = v.Ip
		info.Country = v.Country
		info.State = v.State
		info.City = v.City
		info.Online = online
		info.Trust = trust
		info.CreateTime = util.GetTimeHISByLang(loginTime, lang)
		resList = append(resList, info)
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resList)
	return
}

func DelDevice(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	idStr := strings.TrimSpace(c.DefaultPostForm("id", "")) // 用户名
	if idStr == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}
	id := util.StoI(idStr)
	if id == 0 {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}
	deviceInfo := models.GetLoginDeviceById(id)
	if deviceInfo.ID == 0 || deviceInfo.Uid != user.Id {
		JsonReturn(c, -1, "__T_NO_DATA", nil)
		return
	}
	data := map[string]interface{}{
		"status":      -1,
		"update_time": util.GetNowInt(),
	}
	err := models.EditLoginDeviceInfo(id, data)
	fmt.Println(err)

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", gin.H{})
	return
}
