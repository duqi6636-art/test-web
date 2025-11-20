package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/setting"
	"api-360proxy/web/pkg/util"
	emailSender "api-360proxy/web/service/email"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var AesKey = "CherryProxyYxorpYrrehc" //AES加密密钥

// 根据session获取用户ID 多设备登录
func GetUIDbySession(session string) (bool, int) {
	if session == "" {
		return false, 0
	}
	ses, err := models.GetSessionBySn(session)
	if err != nil {
		return false, 0
	}
	return true, ses.Uid
}

// 获取处理用户信息
func DealUser(c *gin.Context) (code int, msg string, user models.Users) {
	session := c.DefaultPostForm("session", "")
	//fmt.Println("session ==== ", session)
	if session == "" {
		return e.SESSION_EXPIRED, "__T_SESSION_ERROR1", models.Users{}
	}
	var err error
	res, uid := GetUIDbySession(session)
	if !res {
		return e.SESSION_EXPIRED, "__T_SESSION_ERROR2", user
	}
	err, user = models.GetUserById(uid)

	if err != nil {
		return e.ERROR, "__T_USER_INFO_ERROR", user
	}

	if user.Status == 2 {
		return e.ERROR, "__T_ACCOUNT_DISABLED", user
	}
	return e.SUCCESS, "ok", user
}

func GetParams(c *gin.Context) models.SignParam {
	language := c.DefaultPostForm("lang", "en")
	session := c.DefaultPostForm("session", "")
	deviceNum := c.DefaultPostForm("device_num", "")
	TimeZone := c.DefaultPostForm("time_zone", "")
	platform := c.DefaultPostForm("platform", "web")
	if language == "" {
		language = setting.AppConfig.DefaultLanguage
	}
	params := models.SignParam{}
	params.DeviceOs = "web"
	params.Oem = "web"
	params.Brand = "web"
	params.Language = language
	params.OsVersion = "1.0"
	params.VersionShow = "1.0.1"
	params.Version = "1"
	params.Session = session
	params.DeviceNum = deviceNum
	params.TimeZone = TimeZone
	params.Platform = platform
	return params
}

func FreshCache(c *gin.Context) {
	models.FreshConfigCache()
	models.InitLang()
	_, config := models.FindConfigs()
	res := map[string]interface{}{
		"config": config,
		"lang":   models.LangMap,
	}
	fmt.Println("res ==== ", res)
	JsonReturn(c, 0, "success", nil)
	return
}

// json 返回值
func JsonReturn(c *gin.Context, code int, msg string, data interface{}) {
	//c.Header("Content-Type", "text/html; charset=utf-8")
	lang := GetParams(c).Language
	msgLan := msg
	if lang == "" {
		lang = setting.AppConfig.DefaultLanguage
	}
	langCode := msg
	langMsg := ""
	if util.StrContain(msg, "--") {
		msgArr := strings.Split(msg, "--")
		langCode = msgArr[0]
		if len(msgArr) > 1 {
			langMsg = msgArr[1]
		}
	}
	msgLan = models.GetLang(langCode, lang)
	if msgLan == "" {
		msgLan = models.GetLang(langCode, "en")
	}

	if data == nil {
		data = gin.H{}
	}
	var res = gin.H{
		"code":     code,
		"msg":      msgLan + langMsg,
		"data":     data,
		"msg_code": msg,
	}
	c.JSON(http.StatusOK, res)
}

// 返回语言文案
func TextReturn(c *gin.Context, msg string) string {
	lang := GetParams(c).Language
	msgLan := msg
	if lang == "" {
		lang = setting.AppConfig.DefaultLanguage
	}
	langCode := msg
	langMsg := ""
	if util.StrContain(msg, "--") {
		msgArr := strings.Split(msg, "--")
		langCode = msgArr[0]
		if len(msgArr) > 1 {
			langMsg = msgArr[1]
		}
	}
	msgLan = models.GetLang(langCode, lang)
	return msgLan + langMsg
}

// 随机生成数字
func GenValidateCode(width int) string {
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.Seed(time.Now().UnixNano())

	var sb strings.Builder
	for i := 0; i < width; i++ {
		fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
	}
	return sb.String()
}

// 获取数字ID
func GetUsername() string {
	username := GenValidateCode(8)
	//查询是否存在
	err, userInfo := models.GetUserInfo(map[string]interface{}{
		"username": username,
	})
	if err == nil && userInfo.Id > 0 {
		return GetUsername()
	}

	return username
}

// 获取默认密码
func GenUserPwd() string {
	userPwd := ""
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	str1 := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str1)
	result1 := []byte{}
	for i := 0; i < 3; i++ {
		result1 = append(result1, bytes[r.Intn(len(bytes))])
	}
	str2 := "0123456789"
	bytes2 := []byte(str2)
	result2 := []byte{}
	for i := 0; i < 3; i++ {
		result2 = append(result2, bytes2[r.Intn(len(bytes2))])
	}
	userPwd = string(result1) + string(result2)
	return userPwd
}

// 递归生成唯一用户uuid
func GetUserUuid() string {
	username := uuid.NewV4()
	if models.ExistUserByUuid(username.String()) {
		return GetUserUuid()
	}
	return username.String()
}

// 递归生成唯一uuid
func GetUuid() string {
	username := uuid.NewV4()
	return username.String()
}

/*
*
导出csv文件
参数：
c：gin包
fileName：文件名
data：导出的数据（包括表头信息）
*/
func DownloadCsv(c *gin.Context, fileName string, data [][]string) error {
	//内容先写入buffer缓存
	buf := new(bytes.Buffer)
	//写入UTF-8 BOM,此处如果不写入就会导致写入的汉字乱码
	buf.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(buf)
	err := w.WriteAll(data)
	if err != nil {
		return err
	}
	w.Flush()

	//设置http表头为下载
	c.Writer.Header().Add("Content-type", "application/octet-stream")
	c.Writer.Header().Add("Accept-Ranges", "bytes")
	c.Header("Content-Disposition", "attachment; filename="+fileName+".csv")
	_, err = io.Copy(c.Writer, buf)
	return err
}

// 实现加密 使用 CBC
func AesEcrypt(origData []byte, key []byte) ([]byte, error) {
	//创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//获取块的大小
	blockSize := block.BlockSize()
	//对数据进行填充，让数据长度满足需求
	origData = PKCS7Padding(origData, blockSize)
	//采用AES加密方法中CBC加密模式
	blocMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	//执行加密
	blocMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// 实现解密  使用 CBC
func AesDeCrypt(cypted []byte, key []byte) ([]byte, error) {
	//创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//获取块大小
	blockSize := block.BlockSize()
	//创建加密客户端实例
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(cypted))
	//这个函数也可以用来解密
	blockMode.CryptBlocks(origData, cypted)
	//去除填充字符串
	origData, err = PKCS7UnPadding(origData)
	if err != nil {
		return nil, err
	}
	return origData, err
}

// PKCS7 填充模式
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	//Repeat()函数的功能是把切片[]byte{byte(padding)}复制padding个，然后合并成新的字节切片返回
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// 填充的反向操作，删除填充字符串
func PKCS7UnPadding(origData []byte) ([]byte, error) {
	//获取数据长度
	length := len(origData)
	if length == 0 {
		return nil, errors.New("string error！")
	}
	//获取填充字符串长度
	unpadding := int(origData[length-1])
	if unpadding > length {
		return nil, errors.New("invalid padding size")
	}
	//截取切片，删除填充字节，并且返回明文
	return origData[:(length - unpadding)], nil
}

// json 返回值
func JsonReturnShow(c *gin.Context, code int, msg string, data interface{}) {
	lang := GetParams(c).Language
	msgLan := msg
	if lang == "" {
		lang = setting.AppConfig.DefaultLanguage
	}
	langCode := msg
	langMsg := ""
	if util.StrContain(msg, "--") {
		msgArr := strings.Split(msg, "--")
		langCode = msgArr[0]
		if len(msgArr) > 1 {
			langMsg = msgArr[1]
		}
	}
	msgLan = models.GetLang(langCode, lang)
	if msgLan == "" {
		msgLan = models.GetLang(langCode, "en")
	}
	if data == nil {
		data = gin.H{}
	}
	var res = gin.H{
		"code":         code,
		"msg":          msgLan + langMsg,
		"data":         data,
		"request_ip":   c.ClientIP(),
		"request_time": util.GetNowInt(),
	}
	c.JSON(http.StatusOK, res)
}

func Connect(c *gin.Context) {
	JsonReturn(c, 200, "success", "")
	return
}

func dealSendEmail(email_type int, email string, params map[string]string, ip string) {
	useEmail := models.GetConfigVal("default_email")

	if useEmail == "aws_mail" { //亚马逊
		result := emailSender.AwsSendEmail(email, email_type, params, ip)
		fmt.Println(result)
	}

	if useEmail == "tencent_mail" { //腾讯
		for k, v := range params {
			params[k] = strings.Replace(v, "https://www.cherryproxy.com/", "", -1)
			params[k] = strings.Replace(v, "https://center.cherryproxy.com/", "", -1)
		}
		result := emailSender.TencentSendEmail(email, email_type, params, ip)
		fmt.Println(result)
	}
}

// 日志记录
func AddLogs(code, data string) {
	models.AddLog(models.LogModel{
		Code:       code,
		Text:       data,
		CreateTime: util.GetTimeStr(util.GetNowInt(), "Y-m-d H:i:s"),
	})
}

type dingMsgV struct {
	MsgType string                 `json:"msgtype"`
	Text    map[string]string      `json:"text"`
	At      map[string]interface{} `json:"at"`
}

// 统一的产品侧规则驱动预警，支持模板与回退

func SendProductAlertWithRule(ruleKey string, runtime map[string]any, fallbackTpl string) {
	rule, ruleErr := models.GetAlertRule(ruleKey)
	if ruleErr == nil && rule.ID > 0 && strings.TrimSpace(rule.WebhookURL) != "" {
		msg, renderErr := RenderMessage(strings.TrimSpace(rule.Context), runtime)
		if renderErr != nil || strings.TrimSpace(msg) == "" {
			AddLogs("SendProductAlertWithRule", fmt.Sprintf("render fail: %v", renderErr))
			msg = fallbackTpl
		}
		if sendErr := SendDingTalkURL(strings.TrimSpace(rule.WebhookURL), msg); sendErr != nil {
			AddLogs("SendProductAlertWithRule", sendErr.Error())
		}
	}
}

// 直接使用完整 webhook URL 发送；若提供 secret 则自动加签

func SendDingTalkURL(webhookURL, content string) error {
	url := webhookURL
	isAtAll := false
	phoneArr := []string{}

	atArr := map[string]interface{}{
		"atMobiles": phoneArr,
		"isAtAll":   isAtAll,
	}
	body := dingMsgV{
		MsgType: "text",
		Text:    map[string]string{"content": content},
		At:      atArr,
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// 将 rule.Context 作为模板

func RenderMessage(contextTpl string, runtimeVars map[string]interface{}) (string, error) {
	data := map[string]interface{}{}
	if runtimeVars != nil {
		for k, v := range runtimeVars {
			data[k] = v
		}
	}
	tplText := strings.TrimSpace(contextTpl)
	tpl, err := template.New("alert").Option("missingkey=zero").Parse(tplText)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
