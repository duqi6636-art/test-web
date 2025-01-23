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
// @Summary 滑块验证
// @Description 滑块验证
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
	ticket := c.DefaultPostForm("ticket", "")
	randstr := c.DefaultPostForm("randstr", "")
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
		"Randstr":      randstr,
		"CaptchaAppId": com.StrTo(captchaAppId).MustInt(),
		"AppSecretKey": appSecretKey,
	}
	err := request.SetActionParameters(body)
	if err != nil {
		JsonReturn(c, -1, err.Error(), gin.H{})
		return
	}

	//创建common response
	response := tchttp.NewCommonResponse()
	//发送请求
	err = client.Send(request, response)
	if err != nil {
		mm := fmt.Sprintf("fail to invoke api: %v \n", err)
		AddLogs("tencent_sign_error", mm) //写日志
		JsonReturn(c, -1, "__T_CAPTCHA_FAIL", gin.H{})
		return
	}

	// 获取响应结果
	result := ResultInfo{}
	//fmt.Println("TX IMG：", string(response.GetBody()))
	_ = json.Unmarshal(response.GetBody(), &result)

	fmt.Printf("%+v \n", result)
	if result.Response.CaptchaCode == 1 {
		err, user := models.GetUserByEmail(email)
		// 注册/绑定  判断用户是否存在
		if code_type == "reg" || code_type == "bind" {
			if err == nil && user.Id > 0 {
				JsonReturn(c, -1, "__T_ACCOUNT_EXIST", gin.H{})
				return
			}
		}

		// 登录 找回密码  判断用户是否存在
		if code_type == "login" || code_type == "find" {
			if err != nil {
				JsonReturn(c, -1, "__T_USER_NOT_EXIST", gin.H{})
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
	} else {

		AddLogs("tencent_sign_error", util.ItoS(result.Response.CaptchaCode)+" "+result.Response.CaptchaMsg) //写日志
		JsonReturn(c, -1, "__T_CAPTCHA_FAIL",nil)
		return
	}
}


// 处理验证
func CaptchaHandle(c *gin.Context,ticket, randStr string) (bool, string) {
	if ticket == "" || randStr == "" {
		return false, "__T_CAPTCHA_FAIL"
	}
	secretId := models.GetConfigVal("tencent_secret_id") //
	secretKey := models.GetConfigVal("tencent_secret_key") //
	appSecretKey := models.GetConfigVal("tencent_app_secret_key") //
	captchaAppId := models.GetConfigVal("tencent_app_id") //
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
	}else{

		AddLogs("tencent_auth_error", util.ItoS(result.Response.CaptchaCode)+" "+result.Response.CaptchaMsg) //写日志
		return false, "__T_CAPTCHA_FAIL"
	}
}



//func GetAuthSignCopy(c *gin.Context)  {
//	ticket := c.DefaultPostForm("ticket","")
//	randstr := c.DefaultPostForm("randstr","")
//
//	secretId := "IKIDLacR4vwtiXGsV3TuIcqvqY594qT9fvAF"
//	secretKey := "0wfDew43tZ1MWaVJzVqKl5ISIJBsSuRz"
//	credential := common.NewCredential(
//		secretId,
//		secretKey,
//	)
//	appSecretKey := "UsiU7uHCQK4lGGfLbxUp2rilj"
//	captchaAppId := uint64(189983884)
//	//sceneId := uint64(92201)	//场景 ID，网站或应用的业务下有多个场景使用此服务，通过此 ID 区分统计数据
//
//	cpf := profile.NewClientProfile()
//	cpf.HttpProfile.ReqMethod = "POST"
//	cpf.HttpProfile.ReqTimeout = 5
//	cpf.SignMethod = "HmacSHA1"
//	cpf.Language = "en-US"
//	ip := c.ClientIP()
//	client, _ := cap.NewClient(credential, "", cpf)
//
//	captchaType := uint64(9)
//	request := cap.NewDescribeCaptchaResultRequest()
//
//	request = &cap.DescribeCaptchaResultRequest{
//		BaseRequest: &tchttp.BaseRequest{},
//		CaptchaType: &captchaType,
//		Ticket:&ticket,
//		UserIp:&ip,
//		Randstr:&randstr,
//		CaptchaAppId:&captchaAppId,
//		AppSecretKey:&appSecretKey,
//		//SceneId:&sceneId,
//	}
//
//	fmt.Printf("%+v \n",request)
//	response, err := client.DescribeCaptchaResult(request)
//	// Handle the exception
//	if _, ok := err.(*errors.TencentCloudSDKError); ok {
//		fmt.Printf("An API error has returned: %s", err)
//		fmt.Println("\n")
//		fmt.Printf("%s", response.ToJsonString())
//		return
//	}
//	// unexpected errors
//	if err != nil {
//		//panic(err)
//		fmt.Println(err)
//	}
//	// Print the returned json string
//	fmt.Printf("%s", response.ToJsonString())
//}
