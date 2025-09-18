package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	emailSender "api-360proxy/web/service/email"
	"github.com/gin-gonic/gin"
)

// DomainRemarkPair 域名和备注的配对结构
type DomainRemarkPair struct {
	Domain string `json:"domain"`
	Remark string `json:"remark"`
}

type ThirdPartyResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Id int `json:"id"` // 申请记录ID
	} `json:"data"`
}

// AddDomainWhiteApply 添加域名白名单申请
func AddDomainWhiteApply(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c)
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	uid := userInfo.Id
	username := userInfo.Username

	domainsJson := c.DefaultPostForm("domains", "")
	if domainsJson == "" {
		JsonReturn(c, e.ERROR, "域名数据不能为空", nil)
		return
	}

	// 解析域名和备注的JSON数组
	var domainRemarkPairs []DomainRemarkPair
	if err := json.Unmarshal([]byte(domainsJson), &domainRemarkPairs); err != nil {
		JsonReturn(c, e.ERROR, "域名数据解析失败", nil)
		return
	}

	if len(domainRemarkPairs) == 0 {
		JsonReturn(c, e.ERROR, "__T_DOMAIN_TIP", nil)
		return
	}

	successDomains := []string{}
	failedDomains := []string{}
	invalidDomains := []string{}
	// 保存到数据库
	for _, pair := range domainRemarkPairs {
		domain := strings.TrimSpace(pair.Domain)
		remark := strings.TrimSpace(pair.Remark)

		if domain != "" {
			// 检查是否已存在
			if models.CheckUserDomainExists(uid, domain) {
				log.Printf("Domain %s already exists for user %d, skipping", domain, uid)
				continue
			}
			// 域名合法性校验
			isValid, errMsg := util.ValidateDomainAdvanced(domain)
			if !isValid {
				log.Printf("Invalid domain %s for user %d: %s", domain, uid, errMsg)
				invalidDomains = append(invalidDomains, fmt.Sprintf("%s (%s)", domain, errMsg))
				continue
			}

			// 检查域名是否在黑名单中
			isInBlacklist := models.CheckDomainInBlacklist(domain)

			var addInfo models.MdUserApplyDomain
			if isInBlacklist {
				// 在黑名单中，需要第三方审核
				addInfo = models.MdUserApplyDomain{
					Uid:        uid,
					Username:   username,
					Domain:     domain,
					Status:     0, // 审核中
					Remark:     remark,
					CreateTime: util.GetNowInt(),
					UpdateTime: util.GetNowInt(),
				}

				domainID, err := models.AddUserDomainWhite(addInfo)
				if err != nil {
					failedDomains = append(failedDomains, domain)
					continue
				}

				// 异步提交到第三方审核
				go func(id int, info models.MdUserApplyDomain) {
					err := submitDomainsToThirdPartyBatch(id, info)
					if err != nil {
						updateData := map[string]interface{}{
							"third_party_req_id": int(0),
							"third_party_status": 0, // 未提交
							"third_party_result": "submit failed",
							"status":             -1,
							"update_time":        util.GetNowInt(),
							"submit_time":        util.GetNowInt(),
						}
						err := models.UpdateDomainApplyID(id, updateData)
						log.Printf("Failed to submit domain %s to third party: %v", info.Domain, err)
					}
				}(domainID, addInfo)

				log.Printf("Domain %s is in blacklist, submitted for third-party review", domain)
			} else {
				// 不在黑名单中，直接审核通过
				addInfo = models.MdUserApplyDomain{
					Uid:        uid,
					Username:   username,
					Domain:     domain,
					Status:     2, // 直接审核通过
					Remark:     remark,
					CreateTime: util.GetNowInt(),
					UpdateTime: util.GetNowInt(),
					ReviewTime: util.GetNowInt(), // 设置审核时间
				}

				_, err := models.AddUserDomainWhite(addInfo)
				if err != nil {
					failedDomains = append(failedDomains, domain)
					continue
				}

				log.Printf("Domain %s is not in blacklist, automatically approved", domain)
			}

			successDomains = append(successDomains, domain)
		}
	}

	// 构建返回结果
	result := map[string]interface{}{
		"success_domains": successDomains,
		"failed_domains":  failedDomains,
		"invalid_domains": invalidDomains,
		"message":         fmt.Sprintf("已提交%d个域名进行审核", len(successDomains)),
	}

	// 如果有无效域名，在消息中提示
	if len(invalidDomains) > 0 {
		result["message"] = fmt.Sprintf("已提交%d个域名进行审核，%d个域名格式无效", len(successDomains), len(invalidDomains))
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
}

func submitDomainsToThirdPartyBatch(id int, domainData models.MdUserApplyDomain) error {
	// 获取用户信息
	err, userInfo := models.GetUserById(domainData.Uid)
	if err != nil || userInfo.Id == 0 {
		return fmt.Errorf("用户不存在")
	}

	// 准备提交数据
	submitData := map[string]interface{}{
		"oa_platform_name": "360cherry",
		"uid":              domainData.Uid,
		"account":          domainData.Username,
		"account_type":     1,
		"apply_time":       time.Now().Unix(),
		"apply_remark":     domainData.Remark,
		"reg_time":         userInfo.CreateTime,
		"domains":          domainData.Domain,
	}

	// 生成签名
	departmentId := models.GetConfigVal("third_party_department_id")
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signKey := models.GetConfigVal("third_party_sign_key")
	sign := generateThirdPartySign(departmentId, timestamp, signKey)

	// 调用第三方API
	apiUrl := models.GetConfigVal("third_party_domain_review_api_url")
	if apiUrl == "" {
		return fmt.Errorf("third_party_domain_review_api_url not found")
	}

	jsonData, _ := json.Marshal(submitData)
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

	var response ThirdPartyResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}
	log.Println("response:", response)

	if response.Code != e.SUCCESS {
		return fmt.Errorf("API 调用失败 code: %d", response.Code)
	}

	if response.Data.Id > 0 {
		// 根据域名批量更新
		updateData := map[string]interface{}{
			"third_party_req_id": int(response.Data.Id),
			"third_party_status": 1,
			"third_party_result": "submitted",
			"update_time":        util.GetNowInt(),
			"status":             1,
			"submit_time":        util.GetNowInt(),
		}
		log.Println("updateData:", updateData)
		err := models.UpdateDomainApplyID(id, updateData)
		if err != nil {
			log.Printf("Failed to update third party info: uid=%d, domains=%v, error=%v", domainData.Uid, domainData.Domain, err)
			return err
		} else {
			log.Printf("Successfully updated third party info for domains: %v with ID: %d", domainData.Domain, response.Data.Id)
		}
	} else {
		log.Printf("ID type assertion failed")
		return fmt.Errorf("third_party_req_id assertion failed")
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

// DomainWhiteList 域名白名单列表
func DomainWhiteList(c *gin.Context) {
	//language := c.DefaultPostForm("lang", "en")
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id

	domaoinList := models.GetUserDomainWhiteByUid(uid)
	resList := []models.ResUserApplyDomain{}
	for _, domaoin := range domaoinList {
		resInfo := models.ResUserApplyDomain{
			Id:         domaoin.Id,
			Domain:     domaoin.Domain,
			Status:     domaoin.Status,
			SubmitTime: domaoin.SubmitTime,
			ReviewTime: domaoin.ReviewTime,
			Remark:     domaoin.Remark,
		}
		//if domaoin.SubmitTime > 0 {
		//	resInfo.SubmitTime = util.GetTimeHISByLang(domaoin.SubmitTime, language)
		//}
		resList = append(resList, resInfo)
	}

	JsonReturn(c, e.SUCCESS, "__SUCCESS", resList)
	return
}

func DomainWhiteReviewNotify(c *gin.Context) {
	// 验证签名
	departmentId := c.GetHeader("departmentId")
	timestamp := c.GetHeader("timestamp")
	sign := c.GetHeader("sign")

	if departmentId == "" || timestamp == "" || sign == "" {
		log.Println("Missing headers")
		JsonReturn(c, e.SUCCESS, "Missing headers", nil)
		return
	}

	signKey := models.GetConfigVal("third_party_sign_key")
	expectedSign := generateThirdPartySign(departmentId, timestamp, signKey)
	if sign != expectedSign {
		log.Println("Invalid signature, expected:", expectedSign, "got:", sign)
		JsonReturn(c, e.ERROR, "Invalid signature", nil)
		return
	}

	// 解析回调数据
	reqBody, _ := io.ReadAll(c.Request.Body)
	log.Printf("Received callback data: %s", string(reqBody))

	var callbackData models.DomainReviewCallbackData
	if err := json.Unmarshal(reqBody, &callbackData); err != nil {
		log.Printf("Parse error: %v", err)
		JsonReturn(c, e.ERROR, "Parse error", nil)
		return
	}

	// 处理审核结果
	err := processDomainsByStatus(callbackData)
	if err != nil {
		JsonReturn(c, e.ERROR, "process processDomainsByStatus failed", nil)
		return
	}
	// 发送邮件通知
	go sendDomainReviewNotificationEmail(callbackData)

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
}

func processDomainsByStatus(data models.DomainReviewCallbackData) error {
	// 处理通过的域名（批量）
	applyId := int(data.Id)
	if data.PassDomains != "" {
		domains := parseAndCleanDomains(data.PassDomains)
		if len(domains) > 0 {
			updateData := map[string]interface{}{
				"status":             2,
				"review_time":        util.GetNowInt(),
				"update_time":        util.GetNowInt(),
				"audit_user_name":    data.AuditUserName,
				"third_party_result": "",
			}

			err := models.UpdateDomainApplyByDomains(applyId, domains, updateData)
			if err != nil {
				log.Printf("Failed to batch update approved domains: %v", err)
				return err
			} else {
				log.Printf("Successfully batch updated %d approved domains", len(domains))
			}
		}
	}

	// 处理未通过的域名（批量）
	if data.NoPassDomains != "" {
		domains := parseAndCleanDomains(data.NoPassDomains)
		if len(domains) > 0 {
			updateData := map[string]interface{}{
				"status":             4,
				"review_time":        util.GetNowInt(),
				"update_time":        util.GetNowInt(),
				"audit_user_name":    data.AuditUserName,
				"third_party_result": data.AuditRemark,
			}

			err := models.UpdateDomainApplyByDomains(applyId, domains, updateData)
			if err != nil {
				log.Printf("Failed to batch update rejected domains: %v", err)
				return err
			} else {
				log.Printf("Successfully batch updated %d rejected domains", len(domains))
			}
		}
	}
	return nil
}

func parseAndCleanDomains(domainStr string) []string {
	domains := strings.Split(domainStr, ",")
	var cleanDomains []string
	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		if domain != "" {
			cleanDomains = append(cleanDomains, domain)
		}
	}
	return cleanDomains
}

// sendDomainReviewNotificationEmail 发送域名审核结果邮件通知
func sendDomainReviewNotificationEmail(callbackData models.DomainReviewCallbackData) {
	log.Printf("开始发送域名审核邮件通知: %+v", callbackData)

	userInfo, err := getUserInfoFromDomains(callbackData)
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
		return
	}

	if userInfo.Email == "" {
		log.Printf("用户邮箱为空，无法发送邮件通知")
		return
	}

	// 构建邮件内容
	emailParams := buildDomainReviewEmailParams(callbackData, userInfo)

	// 发送邮件
	success := sendDomainReviewEmail(userInfo.Email, callbackData.AuditStatus, emailParams)

	if success {
		log.Printf("域名审核邮件通知发送成功: %s", userInfo.Email)
	} else {
		log.Printf("域名审核邮件通知发送失败: %s", userInfo.Email)
	}
}

// getUserInfoFromDomains 从域名信息中获取用户信息
func getUserInfoFromDomains(callbackData models.DomainReviewCallbackData) (models.Users, error) {
	var userInfo models.Users

	// 从通过的域名或未通过的域名中获取用户信息
	domains := parseAndCleanDomains(callbackData.PassDomains)
	if len(domains) == 0 {
		domains = parseAndCleanDomains(callbackData.NoPassDomains)
	}

	if len(domains) == 0 {
		return userInfo, fmt.Errorf("没有找到域名信息")
	}

	// 通过域名查找申请记录，获取用户ID
	applyRecord, err := models.GetDomainApplyByDomainAndThirdPartyId(domains[0], int(callbackData.Id))
	if err != nil {
		return userInfo, fmt.Errorf("查找域名申请记录失败: %v", err)
	}

	// 根据用户ID获取用户信息
	err, userInfo = models.GetUserById(applyRecord.Uid)
	if err != nil {
		return userInfo, fmt.Errorf("获取用户信息失败: %v", err)
	}

	return userInfo, nil
}

// buildDomainReviewEmailParams 构建邮件参数
func buildDomainReviewEmailParams(callbackData models.DomainReviewCallbackData, userInfo models.Users) map[string]string {
	params := make(map[string]string)

	// 基础信息
	params["submitDate"] = util.GetTimeStr(callbackData.AuditTime, "2006-01-02 15:04:05")

	// 域名信息（保留原有字段以兼容其他地方的使用）
	if callbackData.PassDomains != "" {
		params["approvedDomains"] = callbackData.PassDomains
	} else {
		params["approvedDomains"] = ""
	}

	if callbackData.NoPassDomains != "" {
		params["rejectedDomains"] = callbackData.NoPassDomains
	} else {
		params["rejectedDomains"] = ""
	}

	// 联系信息
	params["supportEmail"] = "support@cherryproxy.com"
	params["whatsappContact"] = "+85267497336"
	params["teamName"] = "Cherry Proxy Team"

	return params
}

// sendDomainReviewEmail 发送域名审核邮件
func sendDomainReviewEmail(email string, auditStatus int, params map[string]string) bool {
	// 获取默认邮件服务配置
	defaultMail := models.GetConfigVal("default_email")

	// 根据审核状态选择邮件类型
	emailType := 16

	// 发送邮件
	var success bool
	switch defaultMail {
	case "aws_mail":
		success = emailSender.AwsSendEmail(email, emailType, params, "")
	case "tencent_mail":
		success = emailSender.TencentSendEmail(email, emailType, params, "")
	default:
		log.Printf("不支持的邮件服务类型: %s", defaultMail)
		return false
	}
	return success
}
