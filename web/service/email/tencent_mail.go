package email

import (
	semail "api-360proxy/service/email"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"strings"
)

func TencentSendEmail(email string, email_type int, params map[string]string, ip string) bool {
	tplSn := ""
	default_mail := models.GetConfigVal("default_email")
	confEmail := models.GetConfEmail(default_mail, email_type)
	tplSn = confEmail.TplId
	tplSnSubject := confEmail.Subject

	secretId := models.GetConfigVal("Tencent_SecretId")
	secretKey := models.GetConfigVal("Tencent_SecretKey")
	tencentFromConf := models.GetConfigVal("Tencent_From")
	tencentFrom := confEmail.FromTo

	if tencentFrom == "" {
		tencentFrom = tencentFromConf
	}
	if tplSn == "" || tplSnSubject == "" {
		return false
	}
	if secretId == "" || secretKey == "" || tencentFrom == "" {
		return false
	}

	templateID := uint64(util.StoI(tplSn)) // 125615 // 腾讯云模板ID，可前往腾讯云控制台查看
	if templateID == 0 {
		return false
	}

	paramsByte, _ := json.Marshal(params)
	paramsStr := string(paramsByte)

	sendRes, result := semail.TencentSend(templateID, secretId, secretKey, email, paramsStr, tencentFrom, tplSnSubject)
	fmt.Println(sendRes)
	fmt.Println(result)
	models.AddLogEmail(email, paramsStr, tplSn, result, ip)

	if sendRes != 0 {
		return false
	}
	return true
}

func TencentSendEmailMarket(email string, code string, params map[string]string, ip string) bool {
	tplSn := ""
	default_mail := models.GetConfigVal("default_email")
	confEmail := models.GetConfEmailBy8(default_mail, code)
	tplSn = confEmail.TplId
	tplSnSubject := confEmail.Subject

	secretId := models.GetConfigVal("Tencent_SecretId")
	secretKey := models.GetConfigVal("Tencent_SecretKey")
	tencentFromConf := models.GetConfigVal("Tencent_From")
	tencentFrom := confEmail.FromTo

	if tencentFrom == "" {
		tencentFrom = tencentFromConf
	}
	if tplSn == "" || tplSnSubject == "" {
		fmt.Println("err ----1")
		return false
	}
	if secretId == "" || secretKey == "" || tencentFrom == "" {
		fmt.Println("err ----2")
		return false
	}

	templateID := uint64(util.StoI(tplSn)) // 125615 // 腾讯云模板ID，可前往腾讯云控制台查看
	if templateID == 0 {
		fmt.Println("err ----3")
		return false
	}

	if len(confEmail.PackageName) > 0 && len(confEmail.PackageUrl) > 0 {
		params["packageName"] = confEmail.PackageName
		params["packageUrl"] = strings.Replace(confEmail.PackageUrl, "https://www.cherryproxy.com/", "", -1)
	}
	paramsByte, _ := json.Marshal(params)
	paramsStr := string(paramsByte)

	sendRes, result := semail.TencentSend(templateID, secretId, secretKey, email, paramsStr, tencentFrom, tplSnSubject)
	fmt.Println(sendRes)
	fmt.Println(result)
	models.AddLogEmail(email, paramsStr, tplSn, result, ip)

	if sendRes != 0 {
		return false
	}
	return true
}
