package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
)

// EnterpriseKycSubmitRequest 企业认证提交请求
type EnterpriseKycSubmitRequest struct {
	WaterBill        string `json:"water_bill" form:"water_bill"`                                  // 水费账单
	ElectricityBill  string `json:"electricity_bill" form:"electricity_bill"`                      // 电费账单
	ProofDocuments   string `json:"proof_documents" form:"proof_documents"`                        // 租赁合同
	CompanyDocuments string `json:"company_documents" form:"company_documents" binding:"required"` // 公司信息文件URLs(企业证件)
	BankDocuments    string `json:"bank_documents" form:"bank_documents" binding:"required"`       // 银行证明文件URLs（银行证明）
}

// GetEnterpriseKycStatus 查询企业认证状态
func GetEnterpriseKycStatus(c *gin.Context) {
	// 用户认证检查
	resCode, msg, user := DealUser(c)
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	// 获取企业认证状态
	status := models.GetEnterpriseKycReviewStatus(user.Id)
	JsonReturn(c, e.SUCCESS, "success", status)
}

// SubmitEnterpriseKyc 提交企业认证
func SubmitEnterpriseKyc(c *gin.Context) {
	// 用户认证检查
	resCode, msg, user := DealUser(c)
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	// 解析请求参数
	var req EnterpriseKycSubmitRequest
	if err := c.ShouldBind(&req); err != nil {
		JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil)
		return
	}

	// 检查企业是否可以提交认证
	canSubmit, reason := models.CheckEnterpriseCanSubmitKyc(uid)
	if !canSubmit {
		JsonReturn(c, e.ERROR, reason, nil)
		return
	}

	if req.WaterBill == "" && req.ElectricityBill == "" && req.ProofDocuments == "" {
		JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil)
		return
	}

	// 生成申请人ID
	applicantID := fmt.Sprintf("ENT_KYC_%d_%d", uid, util.GetNowInt())

	// 获取用户历史购买记录类型
	orderTypes := getUserOrderTypes(uid)

	// 创建企业认证记录
	kyc := models.EnterpriseKyc{
		Uid:                  uid,
		ApplicantID:          applicantID,
		CorporateCertificate: req.CompanyDocuments,
		LeaseContract:        req.ProofDocuments,
		BankMonthBill:        req.BankDocuments,
		WaterBill:            req.WaterBill,
		ElectricityBill:      req.ElectricityBill,
		ReviewStatus:         0, // 待审核
		OrderType:            orderTypes,
	}

	// 保存认证记录
	kycId, err := models.AddEnterpriseKyc(kyc)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_SUBMIT_FAILED", nil)
		return
	}

	// 异步调用第三方审核接口
	go func() {
		err := submitEnterpriseToThirdParty(kycId, kyc, orderTypes)
		if err != nil {
			// 更新状态为提交失败
			updateData := map[string]interface{}{
				"third_party_req_id": int(0),
				"third_party_status": 0, // 未提交
				"third_party_result": "submit failed",
				"review_status":      -1, //提交失败
				"update_time":        util.GetNowInt(),
				"submit_time":        util.GetNowInt(),
			}
			log.Println("updateData:", updateData)
			log.Println("submitEnterpriseToThirdParty err:", err)
			err = models.UpdateEnterpriseKycThirdPartyInfo(kycId, updateData)
			log.Println("UpdateEnterpriseKycThirdPartyInfo err:", err)
		}
	}()

	// 返回结果
	response := map[string]interface{}{
		"kyc_id":       kycId,
		"applicant_id": applicantID,
		"status":       "submitted",
		"message":      "您的企业认证申请已提交，我们将在24小时内完成审核",
	}

	JsonReturn(c, e.SUCCESS, "success", response)
}

func submitEnterpriseToThirdParty(kycId int, kyc models.EnterpriseKyc, orderTypes string) error {
	// 获取用户信息
	err, userInfo := models.GetUserById(kyc.Uid)
	if userInfo.Id == 0 {
		return fmt.Errorf("用户不存在")
	}

	// 准备提交数据
	// 水、电、租赁合同任意一项
	submitData := map[string]interface{}{
		"oa_platform_name":      "360cherry",
		"uid":                   kyc.Uid,
		"account":               userInfo.Username,
		"account_type":          1, // 1: 官网账号 2：代理商
		"reg_time":              userInfo.CreateTime,
		"apply_time":            util.GetNowInt(),
		"verify_type":           2, // 2：企业认证
		"source":                3, // 3：平台自审
		"water_bill":            kyc.WaterBill,
		"electricity_bill":      kyc.ElectricityBill,
		"corporate_certificate": kyc.CorporateCertificate,
		"bank_month_bill":       kyc.BankMonthBill,
		"lease_contract":        kyc.LeaseContract,
		"order_type":            orderTypes,
	}

	// 生成签名
	departmentId := models.GetConfigVal("third_party_department_id")
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signKey := models.GetConfigVal("third_party_sign_key")

	sign := generateThirdPartySign(departmentId, timestamp, signKey)

	// 调用第三方API
	apiUrl := models.GetConfigVal("third_party_kyc_api_url")
	if apiUrl == "" {
		return fmt.Errorf("第三方企业认证接口URL未配置")
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
			"third_party_status": 1, // 已提交
			"review_status":      1, //审核中
			"third_party_result": "submitted",
			"update_time":        util.GetNowInt(),
			"submit_time":        util.GetNowInt(),
		}
		log.Println("updateData:", updateData)
		err := models.UpdateEnterpriseKycThirdPartyInfo(kycId, updateData)
		if err != nil {
			log.Printf("Failed to update enterprise KYC third party info: kycId=%d, error=%v", kycId, err)
			return err
		} else {
			log.Printf("Successfully updated enterprise KYC third party info for kyc: %d with ID: %d", kycId, int(response.Data.Id))
		}

	} else {
		return fmt.Errorf("ID Failed")
	}

	return nil
}
