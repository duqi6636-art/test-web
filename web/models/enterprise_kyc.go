package models

import "api-360proxy/web/pkg/util"

// EnterpriseKyc 企业认证记录表
type EnterpriseKyc struct {
	Id                   int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Uid                  int    `json:"uid"`                   // 用户ID
	ApplicantID          string `json:"applicant_id"`          // 申请人ID
	CorporateCertificate string `json:"corporate_certificate"` // 企业证件URL
	LeaseContract        string `json:"lease_contract"`        // 租赁合同URL
	BankMonthBill        string `json:"bank_month_bill"`       // 银行月结单URL
	WaterBill            string `json:"water_bill"`            // 水费账单URL
	ElectricityBill      string `json:"electricity_bill"`      // 电费账单URL
	OtherDocuments       string `json:"other_documents"`       // 其他证明文件JSON
	ReviewStatus         int    `json:"review_status"`         // 审核状态：0=待审核,1=审核中，2=审核通过，3=审核拒绝，4=未提交审核，-1=提交失败
	ReviewReason         string `json:"review_reason"`         // 审核原因
	Reviewer             string `json:"reviewer"`              // 审核人
	ThirdPartyReqId      string `json:"third_party_req_id"`    // 第三方请求ID
	ThirdPartyStatus     int    `json:"third_party_status"`    // 第三方状态：0=未提交，1=已提交，2=审核通过，3=审核拒绝
	ThirdPartyResult     string `json:"third_party_result"`    // 第三方审核结果
	SubmitTime           int    `json:"submit_time"`           // 提交时间
	ReviewTime           int    `json:"review_time"`           // 审核时间
	UpdateTime           int    `json:"update_time"`           // 更新时间
	CreateTime           int    `json:"create_time"`           // 创建时间
}

var (
	enterpriseKycTable = "cm_enterprise_kyc"
)

// AddEnterpriseKyc 添加企业认证记录
func AddEnterpriseKyc(kyc EnterpriseKyc) (int, error) {
	kyc.CreateTime = util.GetNowInt()
	kyc.UpdateTime = util.GetNowInt()
	err := db.Table(enterpriseKycTable).Create(&kyc).Error
	return kyc.Id, err
}

// GetEnterpriseKycByUid 根据用户ID获取企业认证记录
func GetEnterpriseKycByUid(uid int) (kyc EnterpriseKyc) {
	db.Table(enterpriseKycTable).Where("uid = ?", uid).Order("id DESC").First(&kyc)
	return
}

// UpdateEnterpriseKycThirdPartyInfo 更新企业认证第三方信息
func UpdateEnterpriseKycThirdPartyInfo(kycId int, updateData map[string]interface{}) error {
	return db.Table(enterpriseKycTable).Where("id = ?", kycId).Updates(updateData).Error
}

// GetEnterpriseKycByThirdPartyId 根据第三方请求ID获取企业认证记录
func GetEnterpriseKycByThirdPartyId(thirdPartyReqId int) EnterpriseKyc {
	var kyc EnterpriseKyc
	db.Table(enterpriseKycTable).Where("third_party_req_id = ?", thirdPartyReqId).First(&kyc)
	return kyc
}

// UpdateEnterpriseKycInfo 更新企业认证信息
func UpdateEnterpriseKycInfo(kycId int, updateData map[string]interface{}) error {
	return db.Table(enterpriseKycTable).Where("id = ?", kycId).Updates(updateData).Error
}

// GetEnterpriseKycReviewStatus 获取企业认证审核状态
func GetEnterpriseKycReviewStatus(uid int) map[string]interface{} {
	var kyc EnterpriseKyc
	db.Table(enterpriseKycTable).Where("uid = ?", uid).Order("id desc").First(&kyc)

	if kyc.Id == 0 {
		return map[string]interface{}{
			"status":  4, // 未申请
			"message": "您还未提交企业认证申请",
		}
	}

	statusMap := map[int]string{
		0:  "待审核",
		1:  "审核中",
		2:  "审核通过",
		3:  "审核拒绝",
		-1: "提交失败",
	}

	return map[string]interface{}{
		"kyc_id":       kyc.Id,
		"applicant_id": kyc.ApplicantID,
		"status":       kyc.ReviewStatus,
		"status_text":  statusMap[kyc.ReviewStatus],
		"reason":       kyc.ReviewReason,
		"reviewer":     kyc.Reviewer,
		"submit_time":  kyc.SubmitTime,
		"review_time":  kyc.ReviewTime,
	}
}

// CheckEnterpriseCanSubmitKyc 检查企业是否可以提交认证
func CheckEnterpriseCanSubmitKyc(uid int) (bool, string) {
	kyc := GetEnterpriseKycByUid(uid)

	// 如果没有认证记录，可以提交
	if kyc.Id == 0 {
		return true, ""
	}

	// 如果审核被拒绝或提交失败，可以重新提交
	if kyc.ReviewStatus == 3 || kyc.ReviewStatus == -1 || kyc.ReviewStatus == 4 {
		return true, ""
	}

	// 如果审核通过，不能重新提交
	if kyc.ReviewStatus == 2 {
		return false, "您的实名认证已通过，无需重复提交"
	}

	// 如果正在审核中，不能重新提交
	if kyc.ReviewStatus == 1 || kyc.ReviewStatus == 0 {
		return false, "您的实名认证正在审核中，请耐心等待"
	}

	return false, "未知状态，请联系客服"
}
