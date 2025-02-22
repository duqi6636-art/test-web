package models

import "api-360proxy/web/pkg/util"

type CmUserWhitelistIp struct {
	Id          int    `json:"id"`
	Uid         int    `json:"uid"`          // 用户ID
	AccountId   int    `json:"account_id"`   // 子账号ID
	Username    string `json:"username"`     // 用户名
	WhitelistIp string `json:"whitelist_ip"` // IP
	Country     string `json:"country"`      // 用户地区-国家
	City        string `json:"city"`         // 用户地区-城市
	State       string `json:"state"`        // 用户地区-州省
	Asn         string `json:"asn"`          // 运营商ASN
	Minutes     int    `json:"minutes"`      // 粘性IP轮转时长
	Hostname    string `json:"hostname"`     // hostname:port
	Status      int    `json:"status"`       // 状态 1 正常 -1删除
	FlowType    int    `json:"flow_type"`    // 状态 1 普通流量  2不限量流量
	Remark      string `json:"remark"`       // 备注
	Ip          string `json:"ip"`           // 用户IP
	CreateTime  int    `json:"create_time"`  // 更新时间
}

type ResUserWhitelistIp struct {
	Id          int    `json:"id"`
	WhitelistIp string `json:"whitelist_ip"` // IP
	Country     string `json:"country"`      // 用户地区-国家
	Name        string `json:"name"`         // 用户地区-国家名称
	Img         string `json:"img"`          // 用户地区-国家图片
	City        string `json:"city"`         // 用户地区-城市
	State       string `json:"state"`        // 用户地区-州省
	Asn         string `json:"asn"`          // 运营商ASN
	Hostname    string `json:"hostname"`     // hostname:port
	HostValue   string `json:"host_value"`   // hostname:port
	Cate        int    `json:"cate"`         // 类型 1 sticky ip  2 random IP
	FlowType    int    `json:"flow_type"`    // 状态 1 普通流量  2不限量流量
	Minute      int    `json:"minute"`       // 粘性IP轮转时长
	Minutes     string `json:"minutes"`      // 粘性IP轮转时长 字符串 拼好的
	Remark      string `json:"remark"`       // 备注
	Status      string `json:"status"`       // 备注
	CreateTime  string `json:"create_time"`  // 更新时间
}

// 添加数据
func AddUserWhitelistIp(info CmUserWhitelistIp) (err error) {
	err = db.Table("cm_user_whitelist_ip").Create(&info).Error
	return
}

// 获取分页列表 By WhitelistIp
func GetWhitelistIpsPageByUid(uid, accountId int, search string, offset, limit int) (info []CmUserWhitelistIp) {
	dbs := db.Table("cm_user_whitelist_ip").Where("uid =?", uid).Where("account_id =?", accountId).Where("status =?", 1)
	if search != "" {
		dbs = dbs.Where("whitelist_ip =? or country =? or city =?", search, search, search)
	}
	dbs = dbs.Offset(offset).Limit(limit).Order("id desc").Find(&info)
	return
}

// 获取列表 By WhitelistIp
func GetWhitelistIpsByUid(uid, accountId, flowType int, search string, status int, startTime, endTime int) (info []CmUserWhitelistIp) {
	dbs := db.Table("cm_user_whitelist_ip").
		Where("uid =?", uid).
		Where("account_id =?", accountId).
		Where("flow_type =?", flowType).
		Where("status =?", 1)
	if search != "" {
		dbs = dbs.Where("whitelist_ip like ? or country =? or city =?", "%"+search+"%", search, search)
	}
	if status > 0 {
		dbs = dbs.Where("status =?", status)
	} else {
		dbs = dbs.Where("status >?", 0)
	}
	if startTime > 0 {
		dbs = dbs.Where("create_time >= ?", startTime)
	}
	if endTime > 86400 {
		dbs = dbs.Where("create_time <= ?", endTime)
	}
	dbs = dbs.Order("id desc").Find(&info)
	return
}

// 获取信息 By WhitelistIp
func GetUserWhitelistIpById(id int) (info CmUserWhitelistIp, err error) {
	err = db.Table("cm_user_whitelist_ip").Where("id =?", id).Where("status =?", 1).First(&info).Error
	return
}

// 获取信息 By WhitelistIp
func GetUserWhitelistIpByIp(ip string, flowType int) (info CmUserWhitelistIp, err error) {
	err = db.Table("cm_user_whitelist_ip").Where("whitelist_ip =?", ip).Where("flow_type =?", flowType).Where("status =?", 1).First(&info).Error
	return
}

// 更新信息
func EditUserWhitelistIp(id int, params interface{}) (err error) {
	err = db.Table("cm_user_whitelist_ip").Where("id =?", id).Update(params).Error
	return
}

func DeleteUserWhitelist(id int) (err error) {
	err = db.Table("cm_user_whitelist_ip").Where("id=?", id).Delete(&CmUserWhitelistIp{}).Error
	return err
}

// API -提取流量API白名单
type MdUserWhitelistApi struct {
	Id          int    `json:"id"`
	Uid         int    `json:"uid"`          // 用户ID
	Username    string `json:"username"`     // 用户名
	WhitelistIp string `json:"whitelist_ip"` // IP
	Country     string `json:"country"`      // 国家
	Status      int    `json:"status"`       // 状态 1 正常 -1删除
	FlowType    int    `json:"flow_type"`    // 状态 1 普通流量  2不限量流量
	Remark      string `json:"remark"`       // 备注
	Ip          string `json:"ip"`           // 用户IP
	Cate        int    `json:"cate"`         // 类型 1 个人中心API  2 url
	CreateTime  int    `json:"create_time"`  // 更新时间
}

type ResUserWhitelistApi struct {
	Id          string `json:"id"`
	WhitelistIp string `json:"whitelist_ip"` // IP
	FlowType    int    `json:"flow_type"`    // 状态 1 普通流量  2不限量流量
	Remark      string `json:"remark"`       // 备注
	Status      string `json:"status"`       // 状态 1 正常 2禁用
	CreateTime  string `json:"create_time"`  // 更新时间
}
type ResFlowApiWhiteApi struct {
	Ip     string `json:"ip"`     // IP
	Remark string `json:"remark"` // 备注
}

var userWhiteApiTable = "cm_user_whitelist_api"

// 添加数据
func AddFlowApiWhite(info MdUserWhitelistApi) (err error) {
	err = db.Table(userWhiteApiTable).Create(&info).Error
	return
}

// 获取列表 By Uid
func GetFlowApiWhiteByUid(uid, flow_type int, search string, status int, startTime, endTime int) (info []MdUserWhitelistApi) {
	dbs := db.Table(userWhiteApiTable).Where("uid =? and flow_type = ?", uid, flow_type)
	if search != "" {
		dbs = dbs.Where("whitelist_ip =?", search)
	}
	if status > 0 {
		dbs = dbs.Where("status =?", status)
	} else {
		dbs = dbs.Where("status >?", 0)
	}
	if startTime > 0 {
		dbs = dbs.Where("create_time >= ?", startTime)
	}
	if endTime > 86400 {
		dbs = dbs.Where("create_time <= ?", endTime)
	}
	dbs = dbs.Order("id desc").Find(&info)
	return
}

// 获取信息 By Id
func GetFlowApiWhiteById(id int) (info MdUserWhitelistApi, err error) {
	err = db.Table(userWhiteApiTable).Where("id =?", id).First(&info).Error
	return
}

// 获取信息 By WhitelistIp
func GetFlowApiWhiteByIp(ip string, cate, status int) (info MdUserWhitelistApi, err error) {
	dbs := db.Table(userWhiteApiTable).Where("whitelist_ip =?", ip).Where("flow_type =?", cate)
	if status > 0 {
		dbs = dbs.Where("status =?", status)
	} else {
		dbs = dbs.Where("status >?", 0)
	}
	err = dbs.First(&info).Error
	return
}

// 获取信息 By Uid WhitelistIp
func GetFlowApiWhiteByUidIp(uid int, ip string, cate int) (info MdUserWhitelistApi, err error) {
	err = db.Table(userWhiteApiTable).Where("uid =?", uid).Where("whitelist_ip =?", ip).Where("flow_type =?", cate).Where("status >?", 0).First(&info).Error
	return
}

// 更新信息
func EdiFlowApiWhiteById(id int, params interface{}) (err error) {
	err = db.Table(userWhiteApiTable).Where("id =?", id).Update(params).Error
	return
}

type ApiProxyClientInfo struct {
	Id            int    `json:"id"`
	ProxyName     string `json:"proxy_name"`
	DeductionMode string `json:"deduction_mode"`
	ServerPort    int    `json:"server_port"`
	FixedPort     int    `json:"fixed_port"`
	Host          string `json:"host"`
	State         int    `json:"state"`
	Protocol      string `json:"protocol"`
	Numbering     string `json:"numbering"`
	SessionTime   int    `json:"session_time"`
	Area          string `json:"area"`
	Tag           string `json:"tag"`
	StartPort     int    `json:"start_port"`
	PortNumb      int    `json:"port_numb"`
	OpenedPort    string `json:"opened_port"`
	BlockChinaIp  int    `json:"block_china_ip"`
	CreateTime    int    `json:"create_time"`
	UpdateTime    int    `json:"update_time"`
}

func GetApiProxyClientBy() (proxyClient []ApiProxyClientInfo) {
	db.Table("cm_api_proxy_client").Where("area <> ?", "").Find(&proxyClient)
	return
}

func GetApiProxyAll(cate string) (proxyClient ApiProxyClientInfo, err error) {
	fixedPort := util.StoI(cate) - 1
	err = db.Table("cm_api_proxy_client").
		Where("area = ?", "").
		Where("fixed_port = ?", fixedPort).
		Where("tag like ?", "%asa%").
		First(&proxyClient).Error
	return
}

func GetApiProxyClientByArea(area string, cate string) (proxyClient ApiProxyClientInfo, err error) {
	fixedPort := util.StoI(cate) - 1
	err = db.Table("cm_api_proxy_client").
		Where("fixed_port = ?", fixedPort).
		Where("area = ?", area).
		First(&proxyClient).Error
	return
}
