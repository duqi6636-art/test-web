package models

import "api-360proxy/web/pkg/util"

// MdUserApplyDomain 添加申请域名白名单
type MdUserApplyDomain struct {
	Id               int    `json:"id"`
	Uid              int    `json:"uid"`                // 用户ID
	Username         string `json:"username"`           // 用户名
	Domain           string `json:"domain"`             // 域名
	Remark           string `json:"remark"`             // 备注
	Status           int    `json:"status"`             // 状态 0=待审核,1=审核中，2=审核通过，3=审核拒绝，4=审核失败，-1=提交失败
	AuditUserName    string `json:"audit_user_name"`    // 审核人姓名
	ThirdPartyReqId  int    `json:"third_party_req_id"` // 第三方请求ID
	ThirdPartyStatus int    `json:"third_party_status"` // 第三方审核状态
	ThirdPartyResult string `json:"third_party_result"` // 第三方审核结果
	SubmitTime       int    `json:"submit_time"`        // 提交时间
	ReviewTime       int    `json:"review_time"`        // 审核时间
	CreateTime       int    `json:"create_time"`        // 创建时间
	UpdateTime       int    `json:"update_time"`        // 更新时间
}

// DomainReviewCallbackData 第三方审核通知数据结构
type DomainReviewCallbackData struct {
	Id             int    `json:"id"`               // 记录ID
	Account        string `json:"account"`          // 账号
	OaPlatformName string `json:"oa_platform_name"` // 平台名称
	ApplyTime      int    `json:"apply_time"`       // 申请时间
	AuditStatus    int    `json:"audit_status"`     // 审核状态：0未审核,1全部通过,2全部驳回,3部分通过
	AuditUserName  string `json:"audit_user_name"`  // 审核人姓名
	AuditTime      int    `json:"audit_time"`       // 审核时间
	AuditRemark    string `json:"audit_remark"`     // 审核备注
	PassDomains    string `json:"pass_domains"`     // 通过审核的域名（逗号分隔）
	NoPassDomains  string `json:"no_pass_domains"`  // 未通过审核的域名（逗号分隔）
}

type ResUserApplyDomain struct {
	Id         int    `json:"id"`
	Domain     string `json:"domain"`      // 域名
	Status     int    `json:"status"`      // 状态
	Result     string `json:"result"`      // 审核结果
	SubmitTime int    `json:"submit_time"` // 提交时间
	ReviewTime int    `json:"review_time"` // 审核时间
	Remark     string `json:"remark"`      //备注信息
}

var userApplyDomainTable = "cm_manual_review_domain_white"

// CheckUserDomainExists 检查用户是否已申请过该域名
func CheckUserDomainExists(uid int, domain string) bool {
	var count int64
	db.Table(userApplyDomainTable).
		Where("uid = ? AND domain = ?", uid, domain).
		Count(&count)
	return count > 0
}

// AddUserDomainWhite 添加数据
func AddUserDomainWhite(info MdUserApplyDomain) (applyId int, err error) {
	info.CreateTime = util.GetNowInt()
	info.UpdateTime = util.GetNowInt()
	err = db.Table(userApplyDomainTable).Create(&info).Error
	applyId = info.Id
	return
}

// UpdateDomainApplyByUserAndDomains 批量更新指定用户的域名
func UpdateDomainApplyByUserAndDomains(uid int, domains []string, updateData map[string]interface{}) error {
	if len(domains) == 0 {
		return nil
	}
	return db.Table(userApplyDomainTable).
		Where("uid = ? AND domain IN (?)", uid, domains).
		Updates(updateData).Error
}

// UpdateDomainApplyID 更新指定ID信息
func UpdateDomainApplyID(uid int, updateData map[string]interface{}) error {
	return db.Table(userApplyDomainTable).
		Where("id = ?", uid).Updates(updateData).Error
}

// GetUserDomainWhiteByUid 获取列表 By Uid
func GetUserDomainWhiteByUid(uid int) (info []MdUserApplyDomain) {
	dbs := db.Table(userApplyDomainTable).Where("uid =?", uid)
	dbs = dbs.Order("id desc").Find(&info)
	return
}

// UpdateDomainApplyByDomains 批量根据域名更新申请状态
func UpdateDomainApplyByDomains(applyId int, domains []string, updateData map[string]interface{}) error {
	if len(domains) == 0 {
		return nil
	}
	return db.Table(userApplyDomainTable).Where("third_party_req_id = ? AND domain IN (?)", applyId, domains).Updates(updateData).Error
}

// GetDomainApplyByDomainAndThirdPartyId 根据域名和第三方ID查询域名申请记录
func GetDomainApplyByDomainAndThirdPartyId(domain string, thirdPartyId int) (MdUserApplyDomain, error) {
	var applyRecord MdUserApplyDomain
	err := db.Table(userApplyDomainTable).
		Where("domain = ? AND third_party_req_id = ?", domain, thirdPartyId).
		First(&applyRecord).Error
	return applyRecord, err
}

// CheckDomainInBlacklist 检查域名是否在黑名单中
func CheckDomainInBlacklist(domain string) bool {
	var count int64

	// 检查 cm_black_domain 表
	db.Table("cm_black_domain").Where("domain = ?", domain).Count(&count)
	if count > 0 {
		return true
	}

	// 检查 md_domain_black 表
	db.Table("cm_domain_black").Where("domain = ?", domain).Count(&count)
	if count > 0 {
		return true
	}

	return false
}
