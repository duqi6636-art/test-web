package controller

import (
	"api-360proxy/service/google"
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
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
	err, authInfo := models.GetUserGoogleAuthBy(gooKey)

	if err == nil && authInfo.ID != 0 {
		JsonReturn(c, e.ERROR, "__T_GOOGLE_AUTH_BIND", nil)
		return
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
	code := c.DefaultPostForm("code", "")
	if code == "" {
		JsonReturn(c, -1, "__T_CODE_ERROR", gin.H{})
		return
	}
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	username := user.Username
	gooKey := fmt.Sprintf(googleKey+"(%s)", username)
	err, authInfo := models.GetUserGoogleAuthBy(gooKey)
	if err != nil && authInfo.ID == 0 {
		JsonReturn(c, e.ERROR, "__T_GOOGLE_AUTH_NO", nil)
		return
	}

	ga := google.NewGoogleAuth()
	ret, err := ga.VerifyCode(authInfo.GoogleKey, code)
	if ret == true && err == nil {
		JsonReturn(c, e.SUCCESS, "__T_VERIFY_SUCCESS", nil)
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
