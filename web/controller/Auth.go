package controller

import (
	"api-360proxy/pkg/ipdat"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	emailSender "api-360proxy/web/service/email"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common/regions"
	"github.com/unknwon/com"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type ResultInfo struct {
	Retcode  int    `json:"retcode"`
	Retmsg   string `json:"retmsg"`
	Response struct {
		CaptchaCode int    `json:"CaptchaCode"`
		CaptchaMsg  string `json:"CaptchaMsg"`
	}
}

// @BasePath /api/v1
// @Summary 忘记密码发送邮件前的验证
// @Description 忘记密码发送邮件前的验证
// @Tags 登陆注册相关
// @Accept x-www-form-urlencoded
// @Param ticket query string false "验证码票据"
// @Param randstr query string false "随机数"
// @Param email query string true "邮箱"
// @Param type query string true "验证码类型"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /web/auth/sign [post]
func GetAuthSign(c *gin.Context) {
	signParam := GetParams(c)
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))
	code_type := c.DefaultPostForm("type", "reg")
	if email == "" {
		JsonReturn(c, -1, "__T_EMAIL_IS_MUST", nil)
		return
	}

	if !util.CheckEmail(email) {
		JsonReturn(c, -1, "__T_EMAIL_FORMAT_ERROR", nil)
		return
	}
	if code_type == "reg" || code_type == "bind" || code_type == "login" || code_type == "find" {

	} else {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
	}
	// 验证 -- start
	captchaSwitch := strings.TrimSpace(models.GetConfigVal("CaptchaRegisterSwitch")) // 滑块验证注册开关 由原来的开关变成 类型配置  1 滑块验证  2 google人机验证
	if captchaSwitch == "1" {                                                        // 滑块验证
		ticket := c.DefaultPostForm("ticket", "")
		randStr := c.DefaultPostForm("randstr", "")
		if ticket == "" || randStr == "" {
			JsonReturn(c, -1, "__T_CAPTCHA_FAIL", nil)
			return
		}

		res, Msg := CaptchaHandle(c, ticket, randStr)
		if !res {
			JsonReturn(c, -1, Msg, nil)
			return
		}
	}

	if captchaSwitch == "2" { //google 人机验证
		googleResponse := c.DefaultPostForm("google_robot_response", "")
		if googleResponse == "" {
			JsonReturn(c, -1, "__T_CAPTCHA_FAIL", nil)
			return
		}
		authRes, authMsg := CheckGoogleRecaptcha(c, googleResponse)
		if authRes == false {
			JsonReturn(c, -1, authMsg, nil)
			return
		}
	}
	// 验证 -- end

	err, user := models.GetUserByEmail(email)
	// 注册/绑定  判断用户是否存在
	if code_type == "reg" || code_type == "bind" {
		if err == nil && user.Id > 0 {
			JsonReturn(c, -1, "__T_ACCOUNT_EXIST", map[string]string{"class_id": "username"})
			return
		}
	}

	// 登录 找回密码  判断用户是否存在
	if code_type == "login" || code_type == "find" {
		if err != nil {
			JsonReturn(c, -1, "__T_USER_NOT_EXIST", map[string]string{"class_id": "username"})
			return
		}
	}

	ip := c.ClientIP()
	ipInfo, _ := ipdat.IPDat.GetIpInfo(ip)
	default_mail := models.GetConfigVal("default_email")
	r, msg := emailSender.SendEmailCode(email, code_type, signParam, ip, ipInfo.CountryCode, default_mail)

	if r {
		JsonReturn(c, 0, msg, gin.H{})
		return
	} else {
		JsonReturn(c, -1, msg, gin.H{})
		return
	}
}

// 处理滑块 验证
func CaptchaHandle(c *gin.Context, ticket, randStr string) (bool, string) {
	if ticket == "" || randStr == "" {
		return false, "__T_CAPTCHA_FAIL"
	}
	secretId := models.GetConfigVal("tencent_secret_id")          //
	secretKey := models.GetConfigVal("tencent_secret_key")        //
	appSecretKey := models.GetConfigVal("tencent_app_secret_key") //
	captchaAppId := models.GetConfigVal("tencent_app_id")         //
	//sceneId := uint64(92201)	//场景 ID，网站或应用的业务下有多个场景使用此服务，通过此 ID 区分统计数据
	service := "captcha"
	version := "2019-07-22"
	action := "DescribeCaptchaResult"
	ip := c.ClientIP()
	credential := common.NewCredential(
		secretId,
		secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "captcha.tencentcloudapi.com"
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.Language = "en-US"
	//创建common client
	client := common.NewCommonClient(credential, regions.Guangzhou, cpf)

	captchaType := 9
	// 创建common request，依次传入产品名、产品版本、接口名称
	request := tchttp.NewCommonRequest(service, version, action)
	body := map[string]interface{}{
		"CaptchaType":  captchaType,
		"Ticket":       ticket,
		"UserIp":       ip,
		"Randstr":      randStr,
		"CaptchaAppId": com.StrTo(captchaAppId).MustInt(),
		"AppSecretKey": appSecretKey,
	}
	err := request.SetActionParameters(body)
	if err != nil {
		return false, err.Error()
	}

	//创建common response
	response := tchttp.NewCommonResponse()
	//发送请求
	err = client.Send(request, response)
	if err != nil {
		msg := fmt.Sprintf("fail to invoke api: %v \n", err)

		AddLogs("tencent_auth_error", msg) //写日志
		return false, "__T_CAPTCHA_FAIL"
	}

	// 获取响应结果
	result := ResultInfo{}
	//fmt.Println("TX IMG：", string(response.GetBody()))
	_ = json.Unmarshal(response.GetBody(), &result)
	fmt.Printf("%+v \n", result)
	if result.Response.CaptchaCode == 1 {
		return true, ""
	} else {

		AddLogs("tencent_auth_error", util.ItoS(result.Response.CaptchaCode)+" "+result.Response.CaptchaMsg) //写日志
		return false, "__T_CAPTCHA_FAIL"
	}
}

// google 人机验证
func CheckGoogleRecaptcha(c *gin.Context, response string) (res bool, msg string) {
	var apiUrl = "https://www.google.com/recaptcha/api/siteverify"
	ip := c.ClientIP()
	ipInfo := GetIpInfo(ip)
	isCN := ipInfo.CountryCode == "CN" || ipInfo.CountryCode == "内网" || ipInfo.CountryCode == "保留" || ipInfo.CountryCode == ""
	if isCN {
		apiUrl = "https://www.recaptcha.net/recaptcha/api/siteverify"
	}

	var body GoogleAuthResponse
	err, bodyStr := HttpPostFormHeader(apiUrl, map[string]string{
		"secret":   models.GetConfigVal("google_recaptcha_secret"),
		"response": response,
	}, map[string]interface{}{})
	json.Unmarshal([]byte(bodyStr), &body)

	fmt.Println("bodyStr", bodyStr)
	AddLogs("google_auth_body", bodyStr) //写日志

	if err != nil {
		return false, "__T_CAPTCHA_FAIL-- ..."
	}
	// Check recaptcha verification success.
	if !body.Success {
		return false, "__T_CAPTCHA_FAIL"
	}

	// Check response score.
	return true, "ok"
}

type GoogleAuthResponse struct {
	Success     bool   `json:"success"`
	ChallengeTs string `json:"challenge_ts"`
	Hostname    string `json:"hostname"`
}

// Post 请求 httpPostForm(带请求头)
func HttpPostFormHeader(postUrl string, param map[string]string, header map[string]interface{}) (err error, result string) {
	data := make(url.Values)
	for k, v := range param {
		data[k] = []string{v}
	}
	req, err := http.NewRequest("POST", postUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return err, ""
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if len(header) > 0 {
		for k, v := range header {
			req.Header.Set(k, v.(string))
		}
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return err, ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, ""
	}
	return nil, string(body)
}
