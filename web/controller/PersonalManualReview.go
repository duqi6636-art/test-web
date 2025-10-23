package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"api-360proxy/web/service/email"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/utils"
)

// KycFileUploadResponse 文件上传响应
type KycFileUploadResponse struct {
	FileUrl  string `json:"file_url"`
	FileSize int64  `json:"file_size"`
}

// ThirdPartyKycResponse 第三方KYC提交响应结构
type ThirdPartyKycResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Id int `json:"id"` // 申请记录ID
	} `json:"data"`
}

// KycSubmitRequest KYC提交请求
type KycSubmitRequest struct {
	Email           string `json:"email" form:"email" binding:"required"`    // 邮箱
	AddressInfo     string `json:"address" form:"address"`                   //地址信息
	Country         string `json:"country" form:"country"`                   //国家
	Identity        string `json:"identity" form:"identity"`                 //身份证
	WaterBill       string `json:"water_bill" form:"water_bill"`             // 水费
	ElectricityBill string `json:"electricity_bill" form:"electricity_bill"` //电费
	CreditCardBill  string `json:"credit_card_bill" form:"credit_card_bill"` // 信用卡
}

type KycAccessTokenResponse struct {
	Code            string `json:"code"`
	Msg             string `json:"msg"`
	TransactionTime string `json:"transactionTime"`
	AccessToken     string `json:"access_token"`
	ExpireTime      string `json:"expire_time"`
	ExpireIn        int    `json:"expire_in"`
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

type VerifyRequest struct {
	AppId   string `json:"appId"`
	Version string `json:"version"`
	Nonce   string `json:"nonce"`
	OrderNo string `json:"orderNo"`
	Sign    string `json:"sign"`
	GetFile string `json:"getFile"`
}

// TencentKycPhotoResponse 腾讯KYC查询照片响应结构体
type TencentKycPhotoResponse struct {
	Code            string `json:"code"`
	Msg             string `json:"msg"`
	BizSeqNo        string `json:"bizSeqNo"`
	TransactionTime string `json:"transactionTime"`
	Result          struct {
		OriCode      string `json:"oriCode"`
		OrderNo      string `json:"orderNo"`
		LiveRate     string `json:"liveRate"`
		Similarity   string `json:"similarity"`
		OccurredTime string `json:"occurredTime"`
		AppId        string `json:"appId"`
		Photo        string `json:"photo"` // Base64编码的照片
		BizSeqNo     string `json:"bizSeqNo"`
		TrtcFlag     string `json:"trtcFlag"` // TRTC标识
	} `json:"result"`
}

// UploadKycDocument 上传KYC证明材料
func UploadKycDocument(c *gin.Context) {
	// 用户认证检查
	resCode, msg, _ := DealUser(c)
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	// 获取上传文件
	f, err := c.FormFile("file")
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_UPLOAD_FILE_ERROR", nil)
		return
	}

	// 验证文件大小（最大10MB）
	if f.Size > 10*1024*1024 {
		JsonReturn(c, e.ERROR, "__T_FILE_SIZE_TOO_LARGE", nil)
		return
	}

	// 验证文件类型
	fileExt := strings.ToLower(path.Ext(f.Filename))
	allowedExts := []string{".png", ".jpg", ".jpeg"}
	if !utils.Contains(allowedExts, fileExt) {
		JsonReturn(c, e.ERROR, "__T_INVALID_FILE_TYPE", nil)
		return
	}

	// 打开文件
	fd, err := f.Open()
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_UPLOAD_FILE_ERROR", nil)
		return
	}
	defer fd.Close()

	// 上传到文件服务器
	resourceDomainLocal := models.GetConfigVal("resource_domain_local")
	uploadUrl := resourceDomainLocal + "/upload_file"
	rep, err := util.HttpPostMultiPart(uploadUrl, "kyc", fd, f.Filename)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_UPLOAD_FILE_ERROR", nil)
		return
	}

	// 解析上传结果
	var uploadResult Uploads
	if err := json.Unmarshal([]byte(rep), &uploadResult); err != nil {
		JsonReturn(c, e.ERROR, "__T_UPLOAD_FILE_ERROR", nil)
		return
	}

	if uploadResult.Code != 0 {
		JsonReturn(c, e.ERROR, "__T_UPLOAD_FILE_ERROR", nil)
		return
	}

	// 获取文件URL
	fileUrl, ok := uploadResult.Data["path_url"].(string)
	if !ok {
		JsonReturn(c, e.ERROR, "__T_UPLOAD_FILE_ERROR", nil)
		return
	}

	// 返回结果
	response := KycFileUploadResponse{
		FileUrl:  fileUrl,
		FileSize: f.Size,
	}
	JsonReturn(c, e.SUCCESS, "success", response)
}

// GetKycReviewStatus 查询KYC审核状态
func GetKycReviewStatus(c *gin.Context) {
	// 用户认证检查
	resCode, msg, user := DealUser(c)
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	// 获取审核状态
	status := models.GetUserKycReviewStatus(user.Id)
	JsonReturn(c, e.SUCCESS, "success", status)
}

// SubmitKycManualReview 提交KYC人工审核
func SubmitKycManualReview(c *gin.Context) {
	// 用户认证检查
	resCode, msg, user := DealUser(c)
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	uid := user.Id

	// 解析请求参数
	var req KycSubmitRequest
	if err := c.ShouldBind(&req); err != nil {
		JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil)
		return
	}

	if req.CreditCardBill == "" && req.ElectricityBill == "" && req.WaterBill == "" {
		JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil)
		return
	}

	// 检查用户是否可以提交审核
	canSubmit, reason := models.CheckUserCanSubmitKyc(uid)
	if !canSubmit {
		JsonReturn(c, e.ERROR, reason, nil)
		return
	}

	// 生成申请人ID
	applicantID := fmt.Sprintf("KYC_%d_%d", uid, util.GetNowInt())

	// 获取用户历史购买记录类型
	orderTypes := getUserOrderTypes(uid)

	// 创建审核记录
	review := models.KycManualReview{
		Uid:             uid,
		ApplicantID:     applicantID,
		WaterBill:       req.WaterBill,
		ElectricityBill: req.ElectricityBill,
		CreditCardBill:  req.CreditCardBill,
		Email:           req.Email,
		Identity:        req.Identity,
		Address:         req.AddressInfo,
		ReviewStatus:    0, // 待审核
		OrderType:       orderTypes,
	}

	// 保存审核记录
	reviewId, err := models.AddKycManualReview(review)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_SUBMIT_FAILED", nil)
		return
	}

	// 异步调用第三方审核接口
	go func() {
		err := submitToThirdParty(reviewId, review, orderTypes)
		if err != nil {
			AddLogs("submitToThirdParty", fmt.Sprintf("KYC ID: %d, 第三方提交失败: %s", reviewId, err.Error()))
			// 更新状态为提交失败
			updateData := map[string]interface{}{
				"third_party_req_id": int(0),
				"third_party_status": 0, // 未提交
				"third_party_result": "submit failed",
				"review_status":      -1,
				"update_time":        util.GetNowInt(),
				"submit_time":        util.GetNowInt(),
			}
			models.UpdateKycReviewThirdPartyInfo(reviewId, updateData)
			_ = models.UpdateUserKycByUid(review.Uid, map[string]interface{}{
				"operator": "1",
			})
		}
	}()

	// 返回结果
	response := map[string]interface{}{
		"id":           reviewId,
		"applicant_id": applicantID,
		"status":       "submitted",
		"message":      "您的实名认证申请已提交，我们将在24小时内完成审核",
	}

	JsonReturn(c, e.SUCCESS, "success", response)
}

func submitToThirdParty(reviewId int, kyc models.KycManualReview, orderTypes string) error {
	// 获取用户信息
	err, userInfo := models.GetUserById(kyc.Uid)
	if err != nil || userInfo.Id == 0 {
		return fmt.Errorf("用户不存在:%s", err.Error())
	}

	// 获取用户KYC信息以获取link_url
	userKycInfo := models.GetUserKycByUid(kyc.Uid)
	if userKycInfo.Uid == 0 {
		return fmt.Errorf("未找到用户KYC信息")
	}

	// 判断认证类型（国内/国外）
	isDomestic := isDomesticKyc(userKycInfo.LinkUrl)
	// 解析腾讯KYC链接参数（如果是国内认证）
	var photoURL string
	var certImages string
	var certVideo string
	var tencentNonce, tencentOrderNo string
	if isDomestic {
		// 解析链接参数
		tencentNonce, tencentOrderNo, err = parseTencentKycUrl(userKycInfo.LinkUrl)
		if err != nil {
			return fmt.Errorf("解析腾讯KYC链接失败:%s", err.Error())
		}

		// 获取照片路由地址
		kycAppId := models.GetConfigVal("tencent_kyc_app_id")
		// 第一步：获取 KYC 访问令牌
		accessToken, err := getKycAccessToken()
		if err != nil {
			return fmt.Errorf("获取KYC访问令牌失败: %s", err.Error())
		} else {
			// 第二步：获取 API 签名票据
			ticketList, err := getApiTicket(accessToken, "SIGN", "")
			if err != nil || len(ticketList) == 0 {
				return fmt.Errorf("获取API签名票据失败:%s", err.Error())
			} else {
				ticket := ticketList[0].Value
				// 使用解析出的参数获取照片
				param := []string{tencentNonce, tencentOrderNo, "1.0.0", kycAppId}
				sign, err := getKycSign(param, ticket)
				if err == nil {
					VerifyReq := VerifyRequest{
						AppId:   kycAppId,
						OrderNo: tencentOrderNo,
						Version: "1.0.0",
						Nonce:   tencentNonce,
						Sign:    sign,
						GetFile: "2",
					}
					err, response := sendRequest(VerifyReq)
					if err == nil {
						photoURL, err = SaveTencentKycPhoto(response.Result.Photo, tencentOrderNo)
						if err != nil {
							AddLogs("SaveTencentKycPhoto", fmt.Sprintf("保存照片失败: %s", err.Error()))
						}
					} else {
						AddLogs("SaveTencentKycPhoto", fmt.Sprintf("查询照片失败: %s", err.Error()))
					}
				}
			}
		}
	} else {
		// Onfido认证：从数据库获取已下载的证件图片和视频
		// 解析JSON格式的URL数组并转换为逗号分割的字符串
		if userKycInfo.CertImages != "" {
			var imageUrls []string
			if err := json.Unmarshal([]byte(userKycInfo.CertImages), &imageUrls); err == nil {
				certImages = strings.Join(imageUrls, ",")
			} else {
				// 如果解析失败，直接使用原始值
				certImages = userKycInfo.CertImages
				AddLogs("Unmarshal CertImages", fmt.Sprintf("解析证件图片JSON失败: %s", err.Error()))
			}
		}

		if userKycInfo.CertVideo != "" {
			var videoUrls []string
			if err := json.Unmarshal([]byte(userKycInfo.CertVideo), &videoUrls); err == nil {
				certVideo = strings.Join(videoUrls, ",")
			} else {
				// 如果解析失败，直接使用原始值
				certVideo = userKycInfo.CertVideo
				AddLogs("Unmarshal CertVideo", fmt.Sprintf("解析人脸视频JSON失败: %s", err.Error()))
			}
		}

		log.Printf("Onfido认证 - 证件图片: %s, 人脸视频: %s", certImages, certVideo)
	}

	// 准备提交数据
	submitData := map[string]interface{}{
		"oa_platform_name": "360cherry",
		"uid":              kyc.Uid,
		"account":          userInfo.Username,
		"account_type":     1, // 1：普通账号，2：代理商
		"reg_time":         userInfo.CreateTime,
		"apply_time":       util.GetNowInt(),
		"verify_type":      1, // 1：个人，2：企业
		"source":           3, // 3：平台自审
		"water_bill":       kyc.WaterBill,
		"electricity_bill": kyc.ElectricityBill,
		"credit_card_bill": kyc.CreditCardBill,
		"order_type":       orderTypes,
	}

	// 如果是国内认证，添加照片路由地址和腾讯KYC参数
	if isDomestic {
		// 添加照片路由地址
		if kyc.Identity != "" {
			submitData["face_photo"] = kyc.Identity
		} else if photoURL != "" {
			submitData["face_photo"] = photoURL
		}
		submitData["identity_certificate"] = kyc.Identity
	} else {
		// Onfido认证：添加证件图片和人脸视频
		if certImages != "" {
			submitData["identity_certificate"] = certImages // 证件图片
		}
		if certVideo != "" {
			submitData["face_photo"] = certVideo // 人脸视频
		}
	}

	// 生成签名
	departmentId := models.GetConfigVal("third_party_department_id")
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signKey := models.GetConfigVal("third_party_sign_key")

	sign := generateThirdPartySign(departmentId, timestamp, signKey)

	// 调用第三方API
	apiUrl := models.GetConfigVal("third_party_kyc_api_url")
	if apiUrl == "" {
		return fmt.Errorf("第三方KYC接口URL未配置")
	}

	jsonData, err := json.Marshal(submitData)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %v", err)
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("departmentId", departmentId)
	req.Header.Set("timestamp", timestamp)
	req.Header.Set("sign", sign)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 处理响应
	if resp.StatusCode != 200 {
		return fmt.Errorf("API调用失败: %d", resp.StatusCode)
	}

	var response ThirdPartyKycResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	log.Println("KYC third party response:", response)

	if response.Code != 0 {
		return fmt.Errorf("第三方API错误: code=%d, msg=%s", response.Code, response.Msg)
	}

	if response.Data.Id > 0 {
		updateData := map[string]interface{}{
			"third_party_req_id": int(response.Data.Id),
			"third_party_status": 1,
			"review_status":      1,
			"third_party_result": "submitted",
			"update_time":        util.GetNowInt(),
			"submit_time":        util.GetNowInt(),
		}
		log.Println("updateData:", updateData)
		err := models.UpdateKycReviewThirdPartyInfo(reviewId, updateData)
		if err != nil {
			return fmt.Errorf("UpdateKycReviewThirdPartyInfo reviewId=%d, error=%v", reviewId, err)
		}
		_ = models.UpdateUserKycByUid(kyc.Uid, map[string]interface{}{
			"operator": "1",
		})
	} else {
		return fmt.Errorf("response ID failed")
	}
	return nil
}

// 生成第三方签名
func generateThirdPartySign(departmentId string, timestamp string, signKey string) string {
	params := map[string]string{
		"departmentId": departmentId,
		"timestamp":    timestamp,
	}

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var arr []string
	for _, k := range keys {
		if params[k] != "" {
			arr = append(arr, fmt.Sprintf("%s=%s", k, params[k]))
		}
	}

	signData := strings.Join(arr, "&") + "&key=" + signKey

	h := md5.New()
	h.Write([]byte(signData))
	sign := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	return sign
}

// parseTencentKycUrl 解析腾讯KYC链接中的参数
func parseTencentKycUrl(linkUrl string) (nonce, orderNo string, err error) {
	if linkUrl == "" {
		return "", "", fmt.Errorf("链接为空")
	}

	// 解析URL
	parsedUrl, err := url.Parse(linkUrl)
	if err != nil {
		return "", "", fmt.Errorf("解析URL失败: %v", err)
	}

	// 获取查询参数
	queryParams := parsedUrl.Query()
	nonce = queryParams.Get("nonce")
	orderNo = queryParams.Get("orderNo")

	if nonce == "" || orderNo == "" {
		return "", "", fmt.Errorf("未找到nonce或orderNo参数")
	}

	return nonce, orderNo, nil
}

// isDomesticKyc 判断是否为国内认证
func isDomesticKyc(linkUrl string) bool {
	// 如果链接包含腾讯KYC域名，则为国内认证
	return strings.Contains(linkUrl, "kyc1.qcloud.com")
}

func sendRequest(requestData VerifyRequest) (err error, res TencentKycPhotoResponse) {
	// 请求 URL
	faceIdUrl := "https://kyc1.qcloud.com/api/v2/base/queryfacerecord?orderNo=" + requestData.OrderNo

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

	// 解析 JSON 响应
	var response TencentKycPhotoResponse
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

// SaveBase64ImageToRemote 将Base64图片上传到远程服务器
func SaveBase64ImageToRemote(base64Str, fileName string) (string, error) {
	if base64Str == "" {
		return "", fmt.Errorf("base64字符串为空")
	}

	// 处理data:image前缀
	var base64Data string
	if i := strings.Index(base64Str, ","); i != -1 {
		base64Data = base64Str[i+1:]
	} else {
		base64Data = base64Str
	}

	// 解码Base64
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %v", err)
	}

	// 创建临时文件用于上传
	tempFile, err := ioutil.TempFile("", "kyc_*.jpg")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name()) // 上传完成后删除临时文件
	defer tempFile.Close()

	// 写入临时文件
	if _, err := tempFile.Write(imageData); err != nil {
		return "", fmt.Errorf("写入临时文件失败: %v", err)
	}

	// 重置文件指针到开头
	tempFile.Seek(0, 0)

	// 获取远程服务器配置
	resourceDomainLocal := models.GetConfigVal("resource_domain_local")
	if resourceDomainLocal == "" {
		return "", fmt.Errorf("远程服务器配置为空")
	}

	// 构建上传URL
	uploadUrl := resourceDomainLocal + "/upload_img"

	// 上传到远程服务器
	rep, err := util.HttpPostMultiPart(uploadUrl, "file", tempFile, fileName)
	if err != nil {
		return "", fmt.Errorf("上传到远程服务器失败: %v", err)
	}

	// 解析上传结果
	var uploadResult struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal([]byte(rep), &uploadResult); err != nil {
		return "", fmt.Errorf("解析上传结果失败: %v", err)
	}

	if uploadResult.Code != 0 {
		return "", fmt.Errorf("上传失败: %s", uploadResult.Msg)
	}

	// 获取文件URL
	fileUrl, ok := uploadResult.Data["path_url"].(string)
	if !ok {
		return "", fmt.Errorf("获取文件URL失败")
	}

	return fileUrl, nil
}

// SaveTencentKycPhoto 保存腾讯KYC照片到远程服务器
func SaveTencentKycPhoto(base64Photo, orderNo string) (string, error) {
	if base64Photo == "" {
		return "", fmt.Errorf("照片数据为空")
	}

	// 生成唯一文件名
	fileName := fmt.Sprintf("face_%d_%s_%d.jpg", orderNo, time.Now().Unix())

	// 上传到远程服务器
	photoUrl, err := SaveBase64ImageToRemote(base64Photo, fileName)
	if err != nil {
		return "", fmt.Errorf("上传照片到远程服务器失败: %v", err)
	}

	return photoUrl, nil
}

// 获取 KYC 访问令牌
func getKycAccessToken() (string, error) {
	kycAppId := models.GetConfigVal("tencent_kyc_app_id")
	kycSecret := models.GetConfigVal("tencent_kyc_secret")
	var errc error
	if kycAppId == "" || kycSecret == "" {
		return "config Info", errc
	}
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

// UnifiedKycCallbackData KYC回调数据结构
type UnifiedKycCallbackData struct {
	Id             int    `json:"id"`               // 第三方请求ID
	Account        string `json:"account"`          // 客户账号
	OaPlatformName string `json:"oa_platform_name"` // 平台名称
	ApplyTime      int    `json:"apply_time"`       // 申请时间
	VerifyType     int    `json:"verify_type"`      // 认证类型：1=个人，2=企业
	AuditStatus    int    `json:"audit_status"`     // 审核状态：1=通过，2=驳回
	AuditUserName  string `json:"audit_user_name"`  // 审核人姓名
	AuditTime      int    `json:"audit_time"`       // 审核时间
	AuditRemark    string `json:"audit_remark"`     // 审核备注
}

// EnterpriseKycNotify 认证结果回调
func EnterpriseKycNotify(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			// 获取堆栈信息
			stack := make([]byte, 4096)
			length := runtime.Stack(stack, false)
			stackTrace := string(stack[:length])
			// 记录详细的panic信息
			panicMsg := fmt.Sprintf("[PANIC] EnterpriseKycNotify: %v\nStack trace:\n%s", r, stackTrace)
			AddLogs("EnterpriseKycNotify", panicMsg)
			// 确保响应已经发送
			if !c.Writer.Written() {
				JsonReturn(c, e.ERROR, "Internal server error", nil)
			}
		}
	}()
	// 验证签名
	signature := c.GetHeader("sign")
	departmentId := c.GetHeader("departmentId")
	timestamp := c.GetHeader("timestamp")

	if departmentId == "" || timestamp == "" || signature == "" {
		AddLogs("EnterpriseKycNotify", "Missing headers")
		JsonReturn(c, e.ERROR, "Missing headers", nil)
		return
	}

	if c.Request.Body == nil {
		AddLogs("EnterpriseKycNotify", "Empty request body")
		JsonReturn(c, e.ERROR, "Empty request body", nil)
		return
	}

	reqBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		AddLogs("EnterpriseKycNotify", "read body fail"+err.Error())
		JsonReturn(c, e.ERROR, "read body fail", nil)
		return
	}

	// 验证签名
	signKey := models.GetConfigVal("third_party_sign_key")
	if signKey == "" {
		AddLogs("EnterpriseKycNotify", "third_party_sign_key is null")
		JsonReturn(c, e.ERROR, "third_party_sign_key failed", nil)
		return
	}
	expectedSign := generateThirdPartySign(departmentId, timestamp, signKey)
	if signature != expectedSign {
		AddLogs("EnterpriseKycNotify", fmt.Sprintf("签名验证失败: expected: %s, received: %s", expectedSign, signature))
		JsonReturn(c, e.ERROR, "Invalid signature", nil)
		return
	}

	// 解析回调数据
	var callbackData UnifiedKycCallbackData
	if err := json.Unmarshal(reqBody, &callbackData); err != nil {
		AddLogs("EnterpriseKycNotify", fmt.Sprintf("JSON解析失败: %v, body: %s", err, string(reqBody)))
		JsonReturn(c, e.ERROR, "json unmarshal error", nil)
		return
	}

	// 根据认证类型分别处理
	hasError := false
	errorMsg := ""
	switch callbackData.VerifyType {
	case 1: // 个人认证
		if err := handlePersonalKycCallback(callbackData); err != nil {
			hasError = true
			errorMsg = err.Error()
			AddLogs("EnterpriseKycNotify", errorMsg)
			JsonReturn(c, e.ERROR, err.Error(), nil)
			return
		}
	case 2: // 企业认证
		if err := handleEnterpriseKycCallback(callbackData); err != nil {
			hasError = true
			errorMsg = err.Error()
			AddLogs("EnterpriseKycNotify", errorMsg)
			JsonReturn(c, e.ERROR, err.Error(), nil)
			return
		}
	default:
		hasError = true
		errorMsg = "未知的认证类型"
		AddLogs("EnterpriseKycNotify", fmt.Sprintf("未知的认证类型: %d", callbackData.VerifyType))
		JsonReturn(c, e.ERROR, "unknow callback type", nil)
		return
	}

	// 记录处理结果
	if hasError {
		AddLogs("EnterpriseKycNotify", fmt.Sprintf("处理失败: %s", errorMsg))
	} else {
		AddLogs("EnterpriseKycNotify", "处理成功")
	}

	JsonReturn(c, e.SUCCESS, "success", nil)
}

// handlePersonalKycCallback 处理个人认证回调
func handlePersonalKycCallback(callbackData UnifiedKycCallbackData) error {
	// 根据第三方请求ID查找个人认证记录
	review := models.GetKycReviewByThirdPartyId(callbackData.Id)
	if review.Id == 0 {
		return fmt.Errorf("KYC not found by ThirdPartyId: %d", callbackData.Id)
	}

	// 转换审核状态
	var reviewStatus int
	switch callbackData.AuditStatus {
	case 1: // 通过
		reviewStatus = 2
	case 2: // 驳回
		reviewStatus = 3
	}

	// 更新个人认证状态
	updateData := map[string]interface{}{
		"review_status":      reviewStatus,
		"review_reason":      callbackData.AuditRemark,
		"reviewer":           callbackData.AuditUserName,
		"review_time":        callbackData.AuditTime,
		"third_party_status": reviewStatus,
		"third_party_result": callbackData.AuditRemark,
		"update_time":        util.GetNowInt(),
	}

	err := models.UpdateKycReviewInfo(review.Id, updateData)
	if err != nil {
		return fmt.Errorf("UpdateKycReviewInfo failed: %w", err)
	}

	// 如果审核通过，更新用户KYC状态
	//if reviewStatus == 2 {
	//	models.UpdateUserKycStatus(review.Uid, 1) // 更新用户KYC状态为已认证
	//}
	_ = models.UpdateUserKycByUid(review.Uid, map[string]interface{}{
		"operator": "1",
	})
	// 异步发送邮件和站内信通知
	go sendKycReviewNotifications(review.Uid, reviewStatus, callbackData.AuditTime, callbackData.AuditRemark, "personal")

	return nil
}

// handleEnterpriseKycCallback 处理企业认证回调
func handleEnterpriseKycCallback(callbackData UnifiedKycCallbackData) error {
	// 根据第三方请求ID查找企业认证记录
	kyc := models.GetEnterpriseKycByThirdPartyId(callbackData.Id)
	if kyc.Id == 0 {
		return fmt.Errorf("GetEnterpriseKycByThirdPartyId failed")
	}

	// 转换审核状态
	var reviewStatus int
	switch callbackData.AuditStatus {
	case 1: // 通过
		reviewStatus = 2
	case 2: // 驳回
		reviewStatus = 3
	}

	if reviewStatus == 0 {
		return fmt.Errorf("KYC reviewStatus: %d", reviewStatus)
	}

	// 更新企业认证状态
	updateData := map[string]interface{}{
		"review_status":      reviewStatus,
		"review_reason":      callbackData.AuditRemark,
		"reviewer":           callbackData.AuditUserName,
		"review_time":        callbackData.AuditTime,
		"third_party_status": reviewStatus,
		"third_party_result": callbackData.AuditRemark,
		"update_time":        util.GetNowInt(),
	}

	err := models.UpdateEnterpriseKycInfo(kyc.Id, updateData)
	if err != nil {
		return fmt.Errorf("UpdateEnterpriseKycInfo failed: %w", err)
	}

	// 如果审核通过，更新用户企业认证状态
	//if reviewStatus == 2 {
	//	models.UpdateUserEnterpriseKycStatus(kyc.Uid, 1) // 更新用户企业认证状态为已认证
	//}
	// 异步发送邮件和站内信通知
	go sendKycReviewNotifications(kyc.Uid, reviewStatus, callbackData.AuditTime, callbackData.AuditRemark, "enterprise")
	return nil
}

// sendKycReviewEmail 统一的KYC审核结果邮件发送函数
func sendKycReviewEmail(uid int, reviewStatus int) {
	// 获取用户信息
	err, user := models.GetUserById(uid)
	if err != nil || user.Id == 0 {
		fmt.Printf("Failed to get user info for uid %d: %v\n", uid, err)
		return
	}

	// 检查用户邮箱
	if user.Email == "" {
		fmt.Printf("User email is empty for uid: %d\n", uid)
		return
	}

	// 获取邮件服务配置
	useEmail := models.GetConfigVal("default_email")
	if useEmail == "" {
		fmt.Printf("Email service not configured\n")
		return
	}

	// 准备基础邮件变量
	params := make(map[string]string)
	// 根据KYC类型设置认证类型文本和邮件类型
	emailType := 11
	var authStatusText string
	var descriptionText string

	switch reviewStatus {
	case 2: // 审核通过
		descriptionText = "Hello, your real-name authentication application has been approved. You can now [immediately] check your status and experience more services."
		authStatusText = ""
	case 3: // 审核拒绝
		descriptionText = "Hello, your real-name authentication submission failed. Please review your uploaded information and re-authenticate."
		authStatusText = "Re-authentication"
	default:
		params["auth_status"] = "Invalid review status for personal KYC"
	}

	params["auth_description"] = descriptionText
	params["auth_status"] = authStatusText

	params["email"] = "support@cherryproxy.com"
	params["whatsapp"] = "+85267497336"
	params["telegram"] = "@Olivia_257856"
	params["team_name"] = "Cherry Proxy Team"
	log.Println(params)

	// 根据邮件服务类型发送邮件
	var result bool
	switch useEmail {
	case "aws_mail":
		result = email.AwsSendEmail(user.Email, emailType, params, "")
	case "tencent_mail":
		result = email.TencentSendEmail(user.Email, emailType, params, "")
	default:
		fmt.Printf("Unsupported email service: %s\n", useEmail)
		return
	}

	if result {
		fmt.Printf("%s KYC review email sent successfully to %s\n", authStatusText, user.Email)
	} else {
		fmt.Printf("Failed to send %s KYC review email to %s\n", authStatusText, user.Email)
	}
}

// sendPersonalKycReviewMsg 发送个人认证审核结果站内信
func sendPersonalKycReviewMsg(uid int, reviewStatus int, reviewReason string) {
	nowTime := util.GetNowInt()
	// 获取用户信息
	err, user := models.GetUserById(uid)
	if err != nil || user.Id == 0 {
		fmt.Printf("Failed to get user info for uid %d: %v\n", uid, err)
		return
	}

	// 根据审核状态构造站内信内容
	var msgCate, code, title, brief, content, titleZh, briefZh, contentZh string
	sort := 10

	switch reviewStatus {
	case 2: // 审核通过
		msgCate = "personal_kyc_pass"
		code = "personal"
		title = "Personal Authentication Approved"
		brief = "Your personal authentication has been approved."
		content = "<p>Dear CherryProxy user:</p><p>Hello, your real-name authentication application has been approved. You can now experience many more services.</p><p>If you have any questions, please feel free to contact us through our official email!</p><p>Email: support@cherryproxy.com</p><p>WhatsApp：+85267497336</p><p>Cherry Proxy Team</p>"
		titleZh = "個人認證已通過"
		briefZh = "您的個人認證已通過審核。"
		contentZh = "<p>尊敬的CherryProxy用戶：</p><p>您好，您提交的實名認證審核已通過，您現在可以體驗更多服務。</p><p>如果您有任何問題，請隨時通過我們的官方郵箱聯繫我們！</p><p>郵箱：support@cherryproxy.com</p><p>WhatsApp：+85267497336</p><p>Cherry Proxy團隊</p>"
	case 3: // 审核拒绝
		msgCate = "personal_kyc_reject"
		code = "personal"
		title = "Personal Authentication Rejected"
		brief = "Your personal authentication has been rejected."
		content = "<p>Dear CherryProxy user:</p><p>Hello, your real-name authentication submission failed. Please review your uploaded information and re-authenticate.</p><p>If you have any questions, please feel free to contact us through our official email address!</p><p>Email: support@cherryproxy.com</p><p>WhatsApp: +85267497336</p><p>Telegram: @Olivia_257856</p><p>Cherry Proxy Team</p>"
		titleZh = "個人認證未通過"
		briefZh = "您的個人認證未通過審核。"
		contentZh = "<p>尊敬的CherryProxy用戶：</p><p>您好，您提交的實名認證審核未通過，請檢查上傳的信息並重新認證。</p><p>如果您有任何問題，請隨時通過我們的官方郵箱聯繫我們！</p><p>郵箱：support@cherryproxy.com</p><p>Whatsapp：+85267497336</p><p>Cherry Proxy團隊</p>"
	default:
		fmt.Printf("Invalid review status for personal KYC message: %d\n", reviewStatus)
		return
	}

	// 构造站内信
	msgInfo := models.CmNoticeMsg{
		Title:      title,
		Brief:      brief,
		Content:    content,
		TitleZh:    titleZh,
		BriefZh:    briefZh,
		ContentZh:  contentZh,
		ShowType:   1, // 显示类型：1-普通通知
		CreateTime: nowTime,
		Cate:       msgCate,
		Uid:        uid,
		ReadTime:   0, // 0-未读
		PushTime:   nowTime,
		Sort:       sort,
		Admin:      code,
	}

	// 添加站内信
	msgList := []models.CmNoticeMsg{msgInfo}
	if err := models.BatchAddNoticeMsgLog(msgList); err != nil {
		fmt.Printf("Failed to add personal KYC message for uid %d: %v\n", uid, err)
	} else {
		fmt.Printf("Personal KYC review message added successfully for uid: %d\n", uid)
	}
}

// sendEnterpriseKycReviewMsg 发送企业认证审核结果站内信
func sendEnterpriseKycReviewMsg(uid int, reviewStatus int, reviewReason string) {
	nowTime := util.GetNowInt()
	// 获取用户信息
	err, user := models.GetUserById(uid)
	if err != nil || user.Id == 0 {
		fmt.Printf("Failed to get user info for uid %d: %v\n", uid, err)
		return
	}

	// 根据审核状态构造站内信内容
	var msgCate, code, title, brief, content, titleZh, briefZh, contentZh string
	sort := 10

	switch reviewStatus {
	case 2: // 审核通过
		msgCate = "enterprise_kyc_pass"
		code = "enterprise"
		title = "Enterprise Authentication Approved"
		brief = "Your enterprise authentication has been approved."
		content = "<p>Dear CherryProxy user:</p><p>Hello, your real-name authentication application has been approved. You can now experience many more services.</p><p>If you have any questions, please feel free to contact us through our official email!</p><p>Email: support@cherryproxy.com</p><p>WhatsApp：+85267497336</p><p>Cherry Proxy Team</p>"
		titleZh = "企業認證已通過"
		briefZh = "您的企業認證已通過審核。"
		contentZh = "<p>尊敬的CherryProxy用戶：</p><p>您好，您提交的實名認證審核已通過，您現在可以體驗更多服務。</p><p>如果您有任何問題，請隨時通過我們的官方郵箱聯繫我們！</p><p>郵箱：support@cherryproxy.com</p><p>WhatsApp：+85267497336</p><p>Cherry Proxy團隊</p>"
	case 3: // 审核拒绝
		msgCate = "enterprise_kyc_reject"
		code = "enterprise"
		title = "Enterprise Authentication Rejected"
		brief = "Your enterprise authentication has been rejected."
		content = "<p>Dear CherryProxy user:</p><p>Hello, your real-name authentication submission failed. Please review your uploaded information and re-authenticate.</p><p>If you have any questions, please feel free to contact us through our official email address!</p><p>Email: support@cherryproxy.com</p><p>WhatsApp: +85267497336</p><p>Telegram: @Olivia_257856</p><p>Cherry Proxy Team</p>"
		titleZh = "企業認證未通過"
		briefZh = "您的企業認證未通過審核。"
		contentZh = "<p>尊敬的CherryProxy用戶：</p><p>您好，您提交的實名認證審核未通過，請檢查上傳的信息並重新認證。</p><p>如果您有任何問題，請隨時通過我們的官方郵箱聯繫我們！</p><p>郵箱：support@cherryproxy.com</p><p>Whatsapp：+85267497336</p><p>Cherry Proxy團隊</p>"
	default:
		fmt.Printf("Invalid review status for enterprise KYC message: %d\n", reviewStatus)
		return
	}

	// 构造站内信
	msgInfo := models.CmNoticeMsg{
		Title:      title,
		Brief:      brief,
		Content:    content,
		TitleZh:    titleZh,
		BriefZh:    briefZh,
		ContentZh:  contentZh,
		ShowType:   1, // 显示类型：1-普通通知
		CreateTime: nowTime,
		Cate:       msgCate,
		Uid:        uid,
		ReadTime:   0, // 0-未读
		PushTime:   nowTime,
		Sort:       sort,
		Admin:      code,
	}

	// 添加站内信
	msgList := []models.CmNoticeMsg{msgInfo}
	if err := models.BatchAddNoticeMsgLog(msgList); err != nil {
		fmt.Printf("Failed to add enterprise KYC message for uid %d: %v\n", uid, err)
	} else {
		fmt.Printf("Enterprise KYC review message added successfully for uid: %d\n", uid)
	}
}

// sendKycReviewNotifications 统一处理KYC审核结果的邮件和站内信通知
func sendKycReviewNotifications(uid int, reviewStatus int, auditTime int, reviewReason string, kycType string) {
	// 发送邮件通知
	sendKycReviewEmail(uid, reviewStatus)

	// 根据KYC类型发送不同的站内信
	if kycType == "personal" {
		sendPersonalKycReviewMsg(uid, reviewStatus, reviewReason)
	} else if kycType == "enterprise" {
		sendEnterpriseKycReviewMsg(uid, reviewStatus, reviewReason)
	} else {
		fmt.Printf("Invalid KYC type for notifications: %s\n", kycType)
	}
}
