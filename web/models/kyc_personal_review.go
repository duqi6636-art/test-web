package models

import "api-360proxy/web/pkg/util"

// KycManualReview 人工审核记录表
type KycManualReview struct {
	Id               int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Uid              int    `json:"uid"`                // 用户ID
	ApplicantID      string `json:"applicant_id"`       // 申请人ID
	Identity         string `json:"identity"`           // 身份证
	WaterBill        string `json:"water_bill"`         // 水费账单URL
	ElectricityBill  string `json:"electricity_bill"`   // 电费账单URL
	CreditCardBill   string `json:"credit_card_bill"`   //信用卡账单
	Email            string `json:"email"`              //邮箱
	Address          string `json:"address"`            //地址信息
	OtherDocuments   string `json:"other_documents"`    // 其他证明文件JSON
	ReviewStatus     int    `json:"review_status"`      // 审核状态：0=待审核,1=审核中，2=审核通过，3=审核失败，4=未提交审核 -1=提交失败
	ReviewReason     string `json:"review_reason"`      // 审核原因
	Reviewer         string `json:"reviewer"`           // 审核人
	ThirdPartyReqId  string `json:"third_party_req_id"` // 第三方请求ID
	ThirdPartyStatus int    `json:"third_party_status"` // 第三方状态：0=未提交，1=已提交，2=审核通过，3=审核拒绝
	ThirdPartyResult string `json:"third_party_result"` // 第三方审核结果
	SubmitTime       int    `json:"submit_time"`        // 提交时间
	ReviewTime       int    `json:"review_time"`        // 审核时间
	UpdateTime       int    `json:"update_time"`        // 更新时间
	CreateTime       int    `json:"create_time"`        // 创建时间
}

var (
	kycManualReviewTable = "cm_kyc_personal_review"
)

// AddKycManualReview 添加人工审核记录
func AddKycManualReview(review KycManualReview) (int, error) {
	review.CreateTime = util.GetNowInt()
	review.UpdateTime = util.GetNowInt()
	err := db.Table(kycManualReviewTable).Create(&review).Error
	return review.Id, err
}

// GetKycManualReviewByUid 根据用户ID获取审核记录
func GetKycManualReviewByUid(uid int) (review KycManualReview) {
	db.Table(kycManualReviewTable).Where("uid = ?", uid).Order("id DESC").First(&review)
	return
}

// UpdateKycReviewThirdPartyInfo 更新KYC审核记录的第三方信息
func UpdateKycReviewThirdPartyInfo(reviewId int, updateData map[string]interface{}) error {
	return db.Table(kycManualReviewTable).Where("id = ?", reviewId).Updates(updateData).Error
}

// GetKycReviewByThirdPartyId 根据第三方请求ID获取个人KYC记录
func GetKycReviewByThirdPartyId(thirdPartyReqId int) KycManualReview {
	var review KycManualReview
	db.Table(kycManualReviewTable).Where("third_party_req_id = ?", thirdPartyReqId).First(&review)
	return review
}

// UpdateKycReviewInfo 更新个人KYC审核信息
func UpdateKycReviewInfo(reviewId int, updateData map[string]interface{}) error {
	return db.Table(kycManualReviewTable).Where("id = ?", reviewId).Updates(updateData).Error
}

// GetUserKycReviewStatus 获取用户KYC审核状态
func GetUserKycReviewStatus(uid int) map[string]interface{} {
	review := GetKycManualReviewByUid(uid)

	if review.Id == 0 {
		return map[string]interface{}{
			"status": 4,
			"msg":    "你还未提交个人认证申请",
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
		"kyc_id":       review.Id,
		"applicant_id": review.ApplicantID,
		"status":       review.ReviewStatus,
		"status_text":  statusMap[review.ReviewStatus],
		"reason":       review.ReviewReason,
		"reviewer":     review.Reviewer,
		"submit_time":  review.SubmitTime,
		"review_time":  review.ReviewTime,
	}
}

// CheckUserCanSubmitKyc 检查用户是否可以提交KYC审核
func CheckUserCanSubmitKyc(uid int) (bool, string) {
	review := GetKycManualReviewByUid(uid)

	// 如果没有审核记录，可以提交
	if review.Id == 0 {
		return true, ""
	}

	// 如果审核被拒绝或提交失败，可以重新提交
	if review.ReviewStatus == 3 || review.ReviewStatus == -1 || review.ReviewStatus == 4 {
		return true, ""
	}

	// 如果审核通过，不能重新提交
	if review.ReviewStatus == 2 {
		return false, "您的实名认证已通过，无需重复提交"
	}

	// 如果正在审核中，不能重新提交
	if review.ReviewStatus == 1 || review.ReviewStatus == 0 {
		return false, "您的实名认证正在审核中，请耐心等待"
	}

	return false, "未知状态，请联系客服"
}
