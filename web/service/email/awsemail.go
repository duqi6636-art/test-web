package email

import (
	semail "api-360proxy/service/email"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

//var Sn = models.GetConfigV("AWS_SN")
//var Secret = models.GetConfigV("AWS_Secret")
//var AccSn = models.GetConfigV("AWS_AccSn")
//const (
//	//Sn     = "78a4ff2528ed72f28689a6b8dfa7c074"
//	//Secret = "nXTeBVJX8gat24SW"
//	//AccSn  = "5da2a9e270f2bf7dbb061d3907690e5e"
//	Sn        = "c25c71e1a6fb88823b624bb1fd67dd93"
//	Secret    = "ld6nSddkZ1luQyOo"
//	AccSn    = "32a4086dd8cecab07255bf9944d1f940"
//)

type AwsAuthModel struct {
	Expire int64  `json:"expire"`
	Token  string `json:"token"`
}

type AwsAuthReq struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Ret  AwsAuthModel `json:"ret"`
}
type AwsSendReq struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Ret       interface{} `json:"ret"`
	Timestamp int         `json:"timestamp"`
}

func AwsSendEmail(email string, email_type int, vars map[string]string, ip string) bool {
	tplSn := ""
	//if email_type == 1 { //邮件验证码
	//	tplSn = models.GetConfigVal("AWS_TplSnCode")
	//} else if email_type == 2 { // 账号密码
	//	tplSn = models.GetConfigVal("AWS_TplSnUser")
	//} else if email_type == 3 { // 注册成功
	//	tplSn = models.GetConfigVal("AWS_TplSnReg")
	//} else if email_type == 4 { //支付验证
	//	tplSn = models.GetConfigVal("AWS_TplUserPay")
	//} else if email_type == 5 { //支付成功
	//	tplSn = models.GetConfigVal("AWS_TplPayOk")
	//} else if email_type == 6 { //提现成功
	//	tplSn = models.GetConfigVal("AWS_TplUserTx")
	//} else if email_type == 7 { //流量不足
	//	tplSn = models.GetConfigVal("AWS_TplSnLimitFlow")
	//} else if email_type == 9 { //新用户谷歌登录
	//	tplSn = models.GetConfigVal("AWS_NewGoogleLogin")
	//} else {
	//	tplSn = ""
	//}
	default_mail := models.GetConfigVal("default_email")
	confEmail := models.GetConfEmail(default_mail, email_type)
	tplSn = confEmail.TplId
	from := confEmail.FromTo
	fromConf := models.GetConfigVal("AWS_From")
	if from == "" {
		from = fromConf
	}

	if tplSn == "" || from == "" {
		return false
	}
	fmt.Println(tplSn)
	//获取token
	keyInfo := "AWS_TOKEN"
	err, AwsToken := models.GetConfigs(keyInfo)
	token := ""

	if err != nil || AwsToken.Value == "" {
		var res = false
		res, token = GetAwsToken()
		if !res || token == "" {
			return false
		}
		AwsToken.Value = token
		models.UpConfigs(keyInfo, token)
	} else {
		token = AwsToken.Value
	}
	vs, _ := json.Marshal(vars)
	AccSn := models.GetConfigV("AWS_AccSn")
	sendRes, msg, result := semail.AwsSend(email, string(vs), token, tplSn, AccSn, from)
	fmt.Println(sendRes)
	fmt.Println(msg)
	models.AddLogEmail(email, string(vs), tplSn, result, ip)
	if sendRes == -2 {
		models.UpConfigs(keyInfo, "")
		return AwsSendEmail(email, email_type, vars, ip)
	}
	if sendRes != 0 {
		return false
	}
	return true
}
func AwsSendEmailMarket(email, code string, vars map[string]string, ip string) bool {
	tplSn := ""
	default_mail := models.GetConfigVal("default_email")
	confEmail := models.GetConfEmailBy8(default_mail, code)
	tplSn = confEmail.TplId
	from := confEmail.FromTo
	fromConf := models.GetConfigVal("AWS_From")
	if from == "" {
		from = fromConf
	}

	if tplSn == "" || from == "" {
		return false
	}
	//获取token
	keyInfo := "AWS_TOKEN"
	err, AwsToken := models.GetConfigs(keyInfo)
	token := ""

	if err != nil || AwsToken.Value == "" {
		var res = false
		res, token = GetAwsToken()
		if !res || token == "" {
			return false
		}
		AwsToken.Value = token
		models.UpConfigs(keyInfo, token)
	} else {
		token = AwsToken.Value
	}

	if len(confEmail.PackageName) > 0 && len(confEmail.PackageUrl) > 0 {

		vars["packageName"] = confEmail.PackageName
		vars["packageUrl"] = confEmail.PackageUrl
	}

	vs, _ := json.Marshal(vars)
	AccSn := models.GetConfigVal("AWS_AccSn")
	//from := models.GetConfigVal("AWS_From")
	sendRes, msg, result := semail.AwsSend(email, string(vs), token, tplSn, AccSn, from)
	fmt.Println(sendRes)
	fmt.Println(msg)
	models.AddLogEmail(email, string(vs), tplSn, result, ip)
	// 写入发送记录日志
	if sendRes == -2 {
		models.UpConfigs(keyInfo, "")
		return AwsSendEmailMarket(email, code, vars, ip)
	}
	if sendRes != 0 {
		return false
	}
	return true
}

/*
*
获取token
*/
func GetAwsToken() (bool, string) {
	getAwsAuth := AwsAuthReq{}
	time_str := strconv.FormatInt(time.Now().Unix(), 10)
	Sn := models.GetConfigV("AWS_SN")
	Secret := models.GetConfigV("AWS_Secret")
	getAwsAuthApiResult := HttpPostForm("http://api.martianinc.co:5705/auth/token", map[string]string{
		"sn":        Sn,
		"timestamp": time_str,
		"sign":      util.Md5(Sn + time_str + Secret),
		"force":     "true",
	})
	if getAwsAuthApiResult == "" {
		return false, "-1"
	}
	err := json.Unmarshal([]byte(getAwsAuthApiResult), &getAwsAuth)
	if err != nil {
		return false, err.Error()
	}
	if getAwsAuth.Code == 2 {
		GetAwsToken()
	}
	if getAwsAuth.Code != 0 {
		return false, getAwsAuth.Msg
	}
	if getAwsAuth.Ret.Token == "" {
		return false, "-2"
	}
	return true, getAwsAuth.Ret.Token
}

// Post 请求 httpPostForm
func HttpPostForm(postUrl string, param map[string]string) string {
	data := make(url.Values)
	for k, v := range param {
		data[k] = []string{v}
	}
	resp, err := http.PostForm(postUrl, data)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(body)
}
