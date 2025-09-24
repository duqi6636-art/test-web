package models

import (
	"api-360proxy/web/pkg/util"
	"log"
)

type UserKyc struct {
	Id                int    `json:"id"`                   // id
	Uid               int    `json:"uid"`                  // uid
	ApplicantID       string `json:"applicant_id"`         // 申请人id
	FirstName         string `json:"first_name" `          // first name
	LastName          string `json:"last_name"`            // last name
	LastWorkflowRunId string `json:"last_workflow_run_id"` // last_workflow_run_id
	LinkUrl           string `json:"link_url"`             // link_url
	CertIssuePlace    string `json:"cert_issue_place" `    // 证件签发地
	CertNumber        string `json:"cert_number" `         // 证件号
	CertCate          string `json:"cert_cate"`            // 证件类型
	CertImages        string `json:"cert_images"`          // 证件图片
	CertVideo         string `json:"cert_video"`           // 视频
	Status            string `json:"status"`               // 状态:0=进行中,1=正常,2=认证失败 review,  3=认证失败 declined,4=认证失败 证件号是否已被认证
	CreateTime        int64  `json:"create_time"`          // 添加时间
	UpdateTime        int64  `json:"update_time"`          // 更新时间
	ExpireTime        int64  `json:"expire_time"`          // 过期时间
	Country           string `json:"country"`              // 国家
}

var userKycTable = "cm_user_kyc"

func GetUserKycByUid(uid int) (info UserKyc) {
	db.Table(userKycTable).Where("uid = ?", uid).First(&info)
	return info
}

func AddUserKycData(m UserKyc) (err error) {
	err = db.Table(userKycTable).Create(&m).Error
	return err
}

func UpdateUserKycByUid(uid int, data interface{}) (err error) {
	err = db.Table(userKycTable).Where("uid = ?", uid).Update(data).Error
	return err
}

func GetUserKycBy(where map[string]interface{}) (info UserKyc) {
	db.Table(userKycTable).Where(where).First(&info)
	return info
}

func GetUserKycListBy(where interface{}) (list []UserKyc) {
	db.Table(userKycTable).Where(where).Find(&list)
	return list
}

func DeleteUserKycByUid(uid int) (err error) {
	err = db.Table(userKycTable).Where("uid = ?", uid).Delete(&UserKyc{}).Error
	return err
}

// 判断证件是否已认证
func CheckCertNumber(number string) bool {
	var count int
	db.Table(userKycTable).Where("cert_number = ?", number).Where("status = '1' or status = '2'").Count(&count)
	if count > 0 {
		return true
	}
	return false
}

type UserKycHistory struct {
	Id                int    `json:"id"`                   // id
	Uid               int    `json:"uid"`                  // uid
	ApplicantID       string `json:"applicant_id" `        // 申请人id
	FirstName         string `json:"first_name" `          // first name
	LastName          string `json:"last_name" `           // last name
	LastWorkflowRunId string `json:"last_workflow_run_id"` // last_workflow_run_id
	LinkUrl           string `json:"link_url"`             // link_url
	CertIssuePlace    string `json:"cert_issue_place"`     // 证件签发地
	CertNumber        string `json:"cert_number"`          // 证件号
	CertCate          string `json:"cert_cate"`            // 证件类型
	CertImages        string `json:"cert_images"`          // 证件图片
	Status            string `json:"status"`               // 状态:0=进行中,1=正常,2=认证失败 review,  3=认证失败 declined,4=认证失败 证件号是否已被认证
	CreateTime        int64  `json:"create_time"`          // 添加时间
	UpdateTime        int64  `json:"update_time"`          // 更新时间
	ExpireTime        int64  `json:"expire_time"`          // 过期时间
	SnapshotTime      int64  `json:"snapshot_time"`        // 快照时间
}

var userKycHistoryTable = "cm_user_kyc_history"

func GetKycHistoryInfoByUid(uid int) (info UserKyc) {
	db.Table(userKycHistoryTable).Where("uid = ?", uid).First(&info)
	return info
}

func AddKycHistoryData(m UserKycHistory) (err error) {
	err = db.Table(userKycHistoryTable).Create(&m).Error
	return err
}

func UpdateKycHistoryByUid(uid int, data interface{}) (err error) {
	err = db.Table(userKycHistoryTable).Where("uid = ?", uid).Update(data).Error
	return err
}

func GetKycHistoryBy(where interface{}) (info UserKyc) {
	db.Table(userKycHistoryTable).Where(where).First(&info)
	return info
}

func GetKycHistoryByCount(uid int) int {
	var count int
	todayZeroTime := util.GetTodayTime()
	db.Table(userKycHistoryTable).Where("create_time >= ?", todayZeroTime).Where("uid =?", uid).Count(&count)
	return count
}

// CheckUserKycStatus 检查用户实名认证状态
// 返回值：0-未实名，1-已实名，2-认证中，3-认证失败
func CheckUserKycStatus(uid int) int {
	// 检查个人认证状态
	userKyc := GetUserKycByUid(uid)
	userCertified := false

	if userKyc.Uid != 0 {
		switch userKyc.Status {
		case "1":
			// 检查是否过期
			nowTime := util.GetNowInt()
			if userKyc.ExpireTime <= 0 || int64(nowTime) <= userKyc.ExpireTime {
				userCertified = true
			}
		}
	}

	// 检查个人人工认证状态
	personalCertified := false
	personalKyc := GetKycManualReviewByUid(uid)
	if personalKyc.ReviewStatus == 2 {
		// 企业认证通过
		personalCertified = true
	}

	// 检查企业认证状态
	enterpriseCertified := false
	enterpriseKyc := GetEnterpriseKycByUid(uid)

	if enterpriseKyc.ReviewStatus == 2 {
		// 企业认证通过
		enterpriseCertified = true
	}

	// 判断最终状态
	if userCertified || personalCertified || enterpriseCertified {
		// 任一认证通过，视为已实名
		return 1
	}
	// 未实名认证
	return 0
}

// CheckUserNeedKyc 检查用户是否需要实名认证（根据配置的launch_date判断）
func CheckUserNeedKyc(createTime int) bool {
	// 获取总开关配置
	err, kycSwitchConfig := GetConfigs("static_extraction_switch")
	if err != nil {
		return false
	}
	// 将开关值转换为int
	kycSwitch := util.StoI(kycSwitchConfig.Value)
	// 如果开关为0，直接返回false，不需要实名认证
	if kycSwitch == 0 {
		return false
	}

	// 获取时间戳配置
	err, kycRequiredTimeConfig := GetConfigs("launch_date")
	if err != nil {
		return false
	}
	// 将配置值转换为int
	kycRequiredTime := util.StoI(kycRequiredTimeConfig.Value)
	log.Println("createTime", createTime)
	log.Println("kycRequiredTime", kycRequiredTime)
	// 只有当用户创建时间大于等于配置的时间戳，并且开关为1时，才需要去判断实名认证
	return createTime >= kycRequiredTime
}
