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
	userKyc := GetUserKycByUid(uid)
	if userKyc.Uid == 0 {
		return 0 // 未实名
	}

	switch userKyc.Status {
	case "1":
		// 检查是否过期
		nowTime := util.GetNowInt()
		if userKyc.ExpireTime > 0 && int64(nowTime) > userKyc.ExpireTime {
			return 0 // 已过期，视为未实名
		}
		return 1 // 已实名
	case "0":
		return 2 // 认证中
	default:
		return 3 // 认证失败
	}
}

// CheckUserNeedKyc 检查用户是否需要实名认证（根据配置的launch_date判断）
func CheckUserNeedKyc(createTime int) bool {
	err, kycRequiredTimeConfig := GetConfigs("launch_date")
	// 2025-09-01 00:00:00 UTC
	defaultTime := 1756665600
	if err != nil {
		// 如果配置不存在，使用默认时间：2025年9月1日
		return createTime >= defaultTime
	}

	// 将配置值转换为int
	kycRequiredTime := util.StoI(kycRequiredTimeConfig.Value)
	if kycRequiredTime == 0 {
		// 如果配置值无效，使用默认时间
		return createTime >= defaultTime
	}
	log.Println("createTime", createTime)
	log.Println("kycRequiredTime", kycRequiredTime)

	return createTime >= kycRequiredTime
}
