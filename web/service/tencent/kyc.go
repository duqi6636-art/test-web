package tencent

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type KycAccessTokenResponse struct {
	Code            string `json:"code"`
	Msg             string `json:"msg"`
	TransactionTime string `json:"transactionTime"`
	AccessToken     string `json:"access_token"`
	ExpireTime      string `json:"expire_time"`
	ExpireIn        int    `json:"expire_in"`
}

//const kycAppId = "TIDAOuE0"
//const kycSecret = "Jd9sfvQYpGDJKpl9nug64TI5gkVWx2qlCvfAJBOV9A5Bm6oKfH2SG9U2OLZhDRqa"

// 获取 KYC 访问令牌
func getKycAccessToken() (string, error) {

	kycAppId := models.GetConfigVal("tencent_kyc_app_id")
	kycSecret := models.GetConfigVal("tencent_kyc_secret")
	// 构造请求的URL
	baseURL := "https://kyc1.qcloud.com/api/oauth2/access_token"
	params := url.Values{}
	params.Add("appId", kycAppId)
	params.Add("secret", kycSecret)
	params.Add("grant_type", "client_credential")
	params.Add("version", "1.0.0")

	// 拼接URL
	kycAccesstokenUrl := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// 发送GET请求
	resp, err := http.Get(kycAccesstokenUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取返回的响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析JSON响应
	var response KycAccessTokenResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	// 如果返回成功，返回access_token
	if response.Code == "0" {
		return response.AccessToken, nil
	}

	// 如果请求失败，返回错误信息
	return "", fmt.Errorf("Error: %s", response.Msg)
}

// Ticket 结构体，表示每个票据的详细信息
type Ticket struct {
	Value      string `json:"value"`
	ExpireIn   int    `json:"expire_in"`   // 票据的有效期（单位：秒）
	ExpireTime string `json:"expire_time"` // 票据的过期时间（时间戳）
}

// ApiTicketResponse 结构体映射返回的JSON数据
type ApiTicketResponse struct {
	Code            string   `json:"code"`
	Msg             string   `json:"msg"`
	TransactionTime string   `json:"transactionTime"`
	Tickets         []Ticket `json:"tickets"` // 一个包含票据的数组
}

// 获取 API 签名票据
func getApiTicket(accessToken, ticketType string, userId string) ([]Ticket, error) {

	kycAppId := models.GetConfigVal("tencent_kyc_app_id")
	//kycSecret := models.GetConfigVal("tencent_kyc_secret")
	// 构造请求的URL
	baseURL := "https://kyc1.qcloud.com/api/oauth2/api_ticket"
	params := url.Values{}
	params.Add("appId", kycAppId)
	params.Add("access_token", accessToken)
	params.Add("type", ticketType)
	params.Add("version", "1.0.0")
	if userId != "" {
		params.Add("user_id", userId)
	}

	// 拼接URL
	kycSignTicketUrl := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// 发送GET请求
	resp, err := http.Get(kycSignTicketUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取返回的响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析JSON响应
	var response ApiTicketResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	// 如果返回成功，返回票据列表
	if response.Code == "0" {
		return response.Tickets, nil
	}
	// 如果请求失败，返回错误信息
	return nil, fmt.Errorf("Error: %s", response.Msg)
}

// sign 函数生成签名，values 为传入的参数列表（不包括 ticket），ticket 为额外的参数
func getKycSign(values []string, ticket string) (string, error) {
	// 检查传入的 values 是否为 nil
	if values == nil {
		return "", fmt.Errorf("values is nil")
	}

	// 移除 values 中的 nil 值
	var filteredValues []string
	for _, v := range values {
		if v != "" {
			filteredValues = append(filteredValues, v)
		}
	}

	// 添加 ticket 到 values 列表
	filteredValues = append(filteredValues, ticket)

	// 排序 values 列表
	sort.Strings(filteredValues)

	// 拼接所有的字符串
	var sb strings.Builder
	for _, v := range filteredValues {
		sb.WriteString(v)
	}

	// 计算 SHA-1 哈希值
	hash := sha1.New()
	hash.Write([]byte(sb.String()))
	hashed := hash.Sum(nil)

	// 返回哈希值的十六进制字符串，并转换为大写
	return strings.ToUpper(hex.EncodeToString(hashed)), nil
}

// 定义返回的结构体
type FaceVerifyResponse struct {
	Code            string `json:"code"`
	Msg             string `json:"msg"`
	BizSeqNo        string `json:"bizSeqNo"`
	TransactionTime string `json:"transactionTime"`
	Result          struct {
		BizSeqNo        string `json:"bizSeqNo"`
		TransactionTime string `json:"transactionTime"`
		OrderNo         string `json:"orderNo"`
		FaceId          string `json:"faceId"`
		OptimalDomain   string `json:"optimalDomain"`
		Success         bool   `json:"success"`
	} `json:"result"`
}

// GetH5FaceId 获取 Face ID 并生成二维码 URL，供用户进行身份验证。
func GetH5FaceId(orderNo, name, idNo, userId string) (qrcodeUrl string, err error, faceId string) {

	kycAppId := models.GetConfigVal("tencent_kyc_app_id")
	kycSecret := models.GetConfigVal("tencent_kyc_secret")
	if kycAppId == "" || kycSecret == "" {
		return "", fmt.Errorf("Config Info error"), ""
	}
	// 第一步：获取 KYC 访问令牌
	accessToken, err := getKycAccessToken()
	if err != nil {
		return "", fmt.Errorf("获取 KYC 访问令牌失败: %v", err), ""
	}
	fmt.Println("accessToken:", accessToken)

	// 第二步：获取 API 签名票据
	ticketList, err := getApiTicket(accessToken, "SIGN", "")
	if err != nil || len(ticketList) == 0 {
		return "", fmt.Errorf("获取 API 签名票据失败: %v", err), ""
	}
	ticket := ticketList[0].Value

	// 第三步：生成请求的随机 nonce 值
	nonce := util.GenerateRandomString(32)

	// 第四步：准备请求参数并进行签名
	param := []string{nonce, userId, "1.0.0", kycAppId}
	sign, err := getKycSign(param, ticket)
	if err != nil {
		return "", fmt.Errorf("生成 KYC 签名失败: %v", err), ""
	}

	// 第五步：发送 FaceVerifyRequest 请求到 KYC API
	faceVerifyReq := FaceVerifyRequest{
		AppId:         kycAppId,
		OrderNo:       orderNo,
		Name:          name,
		IdNo:          idNo,
		UserId:        userId,
		Version:       "1.0.0",
		Nonce:         nonce,
		Sign:          sign,
		LiveInterType: "1", // 表示真人验证类型
	}

	// 发送请求并处理响应
	err, response := sendRequest(faceVerifyReq)
	if err != nil {
		return "", fmt.Errorf("发送身份验证请求失败: %v", err), ""
	}

	// 第六步：检查响应是否成功
	if response.Code == "0" && response.Result.Success {
		// 成功：生成回调 URL
		returnUrl := url.QueryEscape(strings.TrimRight(models.GetConfigV("API_DOMAIN_URL"), "/") + "/notify/tencent_kyc")

		// 准备返回参数
		returnParams := []string{
			kycAppId, "1.0.0", nonce, orderNo, response.Result.FaceId, userId,
		}

		// 第七步：获取用于返回请求的 nonce 票据
		nonceTicketList, err := getApiTicket(accessToken, "NONCE", userId)
		if err != nil || len(nonceTicketList) == 0 {
			return "", fmt.Errorf("获取 nonce 票据失败: %v", err), ""
		}
		nonceTicket := nonceTicketList[0].Value

		// 第八步：为回调 URL 请求生成签名
		sign, err = getKycSign(returnParams, nonceTicket)
		if err != nil {
			return "", fmt.Errorf("生成回调签名失败: %v", err), ""
		}

		// 第九步：生成二维码 URL，用户扫描进行验证
		qrcodeUrl = fmt.Sprintf(
			"https://"+response.Result.OptimalDomain+"/api/web/login?appId=%s&version=%s&nonce=%s&orderNo=%s&faceId=%s&url=%s&from=browser&userId=%s&sign=%s&redirectType=1&resultType=",
			kycAppId, "1.0.0", nonce, orderNo, response.Result.FaceId, returnUrl, userId, sign,
		)

		// 设置返回的 faceId 以供后续使用
		faceId = response.Result.FaceId

		// 返回成功结果
		return qrcodeUrl, nil, faceId
	} else {
		// 失败：返回响应中的错误信息
		return "", fmt.Errorf("KYC 验证失败: %s", response.Msg), ""
	}
}

// 定义请求参数结构体
type FaceVerifyRequest struct {
	AppId           string `json:"appId"`
	OrderNo         string `json:"orderNo"`
	Name            string `json:"name"`
	IdNo            string `json:"idNo"`
	UserId          string `json:"userId"`
	SourcePhotoStr  string `json:"sourcePhotoStr,omitempty"`
	SourcePhotoType string `json:"sourcePhotoType,omitempty"`
	Version         string `json:"version"`
	Sign            string `json:"sign"`
	Nonce           string `json:"nonce"`
	LiveInterType   string `json:"liveInterType"`
}

func sendRequest(requestData FaceVerifyRequest) (err error, response FaceVerifyResponse) {
	// 请求 URL
	faceIdUrl := "https://kyc1.qcloud.com/api/server/getAdvFaceId?orderNo=" + requestData.OrderNo // 更改为实际的 API 端点

	// 将请求体转为 JSON 格式
	data, err := json.Marshal(requestData)
	if err != nil {
		return
	}

	// 发送 POST 请求
	resp, err := http.Post(faceIdUrl, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// 打印响应
	fmt.Println("Response Body:", string(respBody))

	// 解析 JSON 响应
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return
	}

	if response.Code != "0" {
		err = fmt.Errorf("Error: %s", response.Msg)
		return
	}
	return nil, response
}

// 回调签名验证
func VerifyFace(params []string, newSign string) bool {
	accessToken, _ := getKycAccessToken()
	ticketList, err := getApiTicket(accessToken, "SIGN", "")
	if err != nil {
		return false
	}
	ticket := ticketList[0].Value
	sign, _ := getKycSign(params, ticket)
	if sign != newSign {
		return false
	}
	return true
}
