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

	"github.com/gin-gonic/gin"
)

// AddDomainWhiteApply 添加域名白名单申请
func AddDomainWhiteApply(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c)
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	uid := userInfo.Id
	username := userInfo.Username
	domainArr := c.DefaultPostForm("domain_arr", "")
	remark := c.DefaultPostForm("remark", "")

	domainList := strings.Split(domainArr, ",")
	if len(domainList) == 0 || domainList[0] == "" {
		JsonReturn(c, e.ERROR, "__T_DOMAIN_TIP", nil)
		return
	}

	successDomains := []string{}
	failedDomains := []string{}
	invalidDomains := []string{}
	// 保存到数据库
	for _, domain := range domainList {
		domain = strings.TrimSpace(domain)
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

			addInfo := models.MdUserApplyDomain{
				Uid:        uid,
				Username:   username,
				Domain:     domain,
				Status:     0, // 审核中
				Remark:     remark,
				CreateTime: util.GetNowInt(),
				UpdateTime: util.GetNowInt(),
			}

			if err := models.AddUserDomainWhite(addInfo); err != nil {
				failedDomains = append(failedDomains, domain)
				continue
			}
			successDomains = append(successDomains, domain)
		}
	}

	// 批量提交到第三方
	if len(successDomains) > 0 {
		go func() {
			err := submitDomainsToThirdPartyBatch(uid, username, remark, successDomains)
			if err != nil {
				updateData := map[string]interface{}{
					"third_party_req_id": int(0),
					"third_party_status": 0, // 未提交
					"third_party_result": "submit failed",
					"status":             -1,
					"update_time":        util.GetNowInt(),
					"submit_time":        util.GetNowInt(),
				}
				err := models.UpdateDomainApplyByUserAndDomains(uid, successDomains, updateData)
				log.Printf("Failed to submit domains to third party: %v", err)
			}
		}()
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

func submitDomainsToThirdPartyBatch(uid int, username, remark string, domains []string) error {
	// 准备提交数据
	submitData := map[string]interface{}{
		"oa_platform_name": "360cherry",
		"uid":              uid,
		"account":          username,
		"account_type":     1,
		"apply_time":       time.Now().Unix(),
		"apply_remark":     remark,
		"reg_time":         time.Now().Unix(),
		"domains":          strings.Join(domains, ","),
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

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}
	log.Println("response:", response)

	// 在响应处理部分
	if data, ok := response["data"].(map[string]interface{}); ok {
		if idValue, exists := data["id"]; exists {
			if id, ok := idValue.(float64); ok {
				// 根据域名批量更新
				updateData := map[string]interface{}{
					"third_party_req_id": int(id),
					"third_party_status": 1,
					"third_party_result": "submitted",
					"update_time":        util.GetNowInt(),
					"status":             1,
					"submit_time":        util.GetNowInt(),
				}
				log.Println("updateData:", updateData)
				err := models.UpdateDomainApplyByUserAndDomains(uid, domains, updateData)
				if err != nil {
					log.Printf("Failed to update third party info: uid=%d, domains=%v, error=%v", uid, domains, err)
					return err
				} else {
					log.Printf("Successfully updated third party info for domains: %v with ID: %d", domains, int(id))
				}
			} else {
				log.Printf("ID type assertion failed, got: %T", idValue)
				return fmt.Errorf("third_party_req_id assertion failed")
			}
		}
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
