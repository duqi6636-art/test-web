package controller

import (
	"api-360proxy/statistics/pkg/util"
	"api-360proxy/web/models"
	"encoding/json"
	"fmt"
	"strings"
)

var StaticZToken = "KkSEJVVWZtAthPonYtYZcppgacbbuZRC"

// ResponseStaticZtListModel 请求资源中台列表信息返回值
type ResponseStaticZtListModel struct {
	Code int `json:"code"`
	Data struct {
		Limit int                 `json:"limit"`
		List  []StaticZtListModel `json:"list"`
		Page  int                 `json:"page"`
		Total int                 `json:"total"`
	} `json:"data"`
	Msg string `json:"msg"`
}

// StaticZtListModel 请求资源中台列表信息返回值
type StaticZtListModel struct {
	Id             int     `json:"id"`
	IpType         int     `json:"ip_type"`
	IpAttribute    string  `json:"ip_attribute"`
	Cate           int     `json:"cate"`
	Region         string  `json:"region"`
	RegionCode     string  `json:"region_code"`
	Ip             string  `json:"ip"`
	Status         int     `json:"status"`
	Oem            string  `json:"oem"`
	Isp            string  `json:"isp"`
	Operator       string  `json:"operator"`
	Asn            string  `json:"asn"`
	Price          float64 `json:"price"`
	CreateTime     int     `json:"create_time"`
	UsedNum        int     `json:"used_num"`
	CalmStatus     int     `json:"calm_status"`
	CalmExpireTime int     `json:"calm_expire_time"`
}

// StaticZtList 资源列表
func StaticZtList(regionSn string) (bool, string, []StaticZtListModel) {
	lists := []StaticZtListModel{}

	// IP属性
	attribute := "isp,double_isp,native_ct_isp"
	openIPType := models.GetConfigVal("zt_static_ip_open_type")
	if openIPType == "" {
		openIPType = "1"
	}

	// 查询静态可用IP列表
	apiUrl := models.GetConfigVal("zt_static_ip_list") //资源中台静态IP池列表url
	if apiUrl == "" {
		apiUrl = "http://xc-static-mid-api.worldrift.com/api/get_static_ip_list"
	}
	oemName := models.GetConfigVal("zt_static_oem_name")
	if oemName == "" {
		oemName = "cherry"
	}
	token := models.GetConfigVal("zt_static_ip_token") //资源中台账户Token
	if token == "" {
		token = StaticZToken
	}
	headerInfo := map[string]interface{}{
		"token": token,
	}
	data_info := map[string]string{
		"oem":     oemName,    //被授权可使用的项目
		"cate":    openIPType, //IP分类，0=独享、1=共享
		"ip_type": "2",        //IP类型，1=数据中心，2=静态住宅
		//"region":    "",        //地区代码 如 fr-paris
		"attribute": attribute, //IP属性，isp=单ISP、double_isp=双ISP、native_isp=原生、native_hq_isp=优质原生、native_ct_isp=原生电信  多个用,逗号隔开
		"page":      "1",       //页码：不传默认1
		"limit":     "100",     //每页数：不传默认10
	}
	if regionSn != "" {
		data_info["region"] = regionSn
	}
	AddLogs("StaticZtList requestStr", fmt.Sprintf("%+v", data_info))
	err, requestStr := util.HttpPostFormHeader(apiUrl, data_info, headerInfo)
	if err != nil {
		AddLogs("StaticZtList request error", err.Error())
	}

	responseInfo := ResponseStaticZtListModel{}
	err1 := json.Unmarshal([]byte(requestStr), &responseInfo)
	if err1 != nil {
		AddLogs("StaticZtList", err1.Error())
		return false, "__T_IP_RESOURCE_ERROR", lists
	}
	if responseInfo.Code != 200 {
		AddLogs("StaticZtList", responseInfo.Msg) //写日志
		return false, "__T_IP_RESOURCE_ERROR--1", lists
	}
	lists = responseInfo.Data.List
	return true, "", lists
}

// ResponseStaticOpenZtModel 请求资源中台开通信息返回值
type ResponseStaticOpenZtModel struct {
	Code int      `json:"code"`
	Data []string `json:"data"`
	Msg  string   `json:"msg"`
}

type ResponseStaticZtModel struct {
	Code int               `json:"code"`
	Data map[string]string `json:"data"`
	Msg  string            `json:"msg"`
}

type ResponseZtRegionStockModel struct {
	Code int `json:"code"`
	Data struct {
		RegionStock map[string]int `json:"region_stock"`
	} `json:"data"`
	Msg string `json:"msg"`
}

// 请求资源中台释放信息返回值
type ResponseStaticReleaseZtModel struct {
	Code int    `json:"code"`
	Data string `json:"data"`
	Msg  string `json:"msg"`
}

// StaticZtOpen 资源开通
func StaticZtOpen(uid, durationTime int, ipStr, orderId, regionSn string) (bool, string) {
	// IP属性
	attribute := "isp,double_isp,native_ct_isp"
	openIPType := models.GetConfigVal("zt_static_ip_open_type")
	if openIPType == "" {
		openIPType = "1"
	}

	apiUrl := models.GetConfigVal("zt_static_ip_open") //资源中台开通
	if apiUrl == "" {
		apiUrl = "http://xc-static-mid-api.worldrift.com/api/get_ip"
	}
	oemName := models.GetConfigVal("zt_static_oem_name")
	if oemName == "" {
		oemName = "cherry"
	}
	token := models.GetConfigVal("zt_static_ip_token") //资源中台账户Token
	if token == "" {
		token = StaticZToken
	}
	headerInfo := map[string]interface{}{
		"token": token,
	}
	expire_duration_str := models.GetConfigVal("zt_static_ip_expire_duration") //单位 秒
	expire_duration := util.StoI(expire_duration_str)
	if expire_duration == 0 {
		expire_duration = 7 * 86400
	}
	duration := durationTime + expire_duration //开通时长 秒  冷静期配置
	data_info := map[string]string{
		"oem":          oemName,             //被授权可使用的项目
		"uid":          util.ItoS(uid),      //项目开通的用户ID
		"order_id":     orderId,             //项目开通的订单号
		"num":          "1",                 //开通的数量
		"duration":     util.ItoS(duration), //开通的时长（秒）
		"region":       regionSn,            //开通的地区代码 如 fr-paris
		"own_type":     openIPType,          //开通的IP分类，0=独享、1=共享、2=定制
		"type":         "2",                 //IP类型，1=数据中心，2=静态住宅
		"zj_open_type": "2",                 //选择开通方式，1=随机、2=自选
		"ips":          ipStr,               //自选IP，多个用,号隔开
		"attribute":    attribute,           //IP属性，isp=单ISP、double_isp=双ISP、native_isp=原生、native_hq_isp=优质原生、native_ct_isp=原生电信  多个用,逗号隔开
	}

	//获取提取信息-开通扣费  ，请求资源中台
	fmt.Println(data_info)
	err1, requestStr := util.HttpPostFormHeader(apiUrl, data_info, headerInfo)
	fmt.Println(err1)
	fmt.Println(requestStr)
	AddLogs("IpOpen", requestStr) //写日志
	responseInfo := ResponseStaticOpenZtModel{}
	err1 = json.Unmarshal([]byte(requestStr), &responseInfo)
	if err1 != nil {
		return false, "__T_IP_OPEN_ERROR"
	}
	if responseInfo.Code != 200 {
		AddLogs("IpOpenErr01", responseInfo.Msg) //写日志
		return false, "__T_IP_OPEN_ERROR -- 1"
	}

	return true, ""
}

// ResponseZtStaticStatusModel 资源中台 IP状态
type ResponseZtStaticStatusModel struct {
	Code int                   `json:"code"`
	Data []StaticZtStatusModel `json:"data"`
	Msg  string                `json:"msg"`
}

type StaticZtStatusModel struct {
	Ip     string `json:"ip"`
	Status int    `json:"status"`
}

// StaticZtStatus 获取IP状态
func StaticZtStatus(ipStatusArr []string) (bool, string, []StaticZtStatusModel) {
	lists := []StaticZtStatusModel{}

	// 查询静态可用IP列表
	apiUrl := models.GetConfigVal("zt_static_ip_status") //资源中台静态IP状态
	if apiUrl == "" {
		apiUrl = "http://xc-static-mid-api.worldrift.com/api/get_ip_status"
	}
	oemName := models.GetConfigVal("zt_static_oem_name")
	if oemName == "" {
		oemName = "cherry"
	}
	token := models.GetConfigVal("zt_static_ip_token") //资源中台账户Token
	if token == "" {
		token = StaticZToken
	}
	headerInfo := map[string]interface{}{
		"token": token,
	}
	ipStr := strings.Join(ipStatusArr, ",")
	data_info := map[string]string{
		"oem": oemName, //被授权可使用的项目
		"ips": ipStr,   //IP列表
	}
	err, requestStr := util.HttpPostFormHeader(apiUrl, data_info, headerInfo)
	fmt.Println(err)
	fmt.Println(requestStr)

	responseInfo := ResponseZtStaticStatusModel{}
	err1 := json.Unmarshal([]byte(requestStr), &responseInfo)
	if err1 != nil {
		return false, "__T_IP_RESOURCE_ERROR", lists
	}
	if responseInfo.Code != 200 {
		AddLogs("IpStatus", requestStr) //写日志
		return false, "__T_IP_RESOURCE_ERROR--1", lists
	}
	lists = responseInfo.Data
	return true, "", lists
}

// StaticZtOpenRenew 资源续费
func StaticZtOpenRenew(uid, durationTime int, ipStr, orderId string) (bool, string) {
	apiUrl := models.GetConfigVal("zt_static_ip_renew") //资源中台续费
	if apiUrl == "" {
		apiUrl = "http://xc-static-mid-api.worldrift.com/api/renew_ip"
	}
	oemName := models.GetConfigVal("zt_static_oem_name")
	if oemName == "" {
		oemName = "cherry"
	}
	token := models.GetConfigVal("zt_static_ip_token") //资源中台账户Token
	if token == "" {
		token = StaticZToken
	}
	headerInfo := map[string]interface{}{
		"token": token,
	}
	expire_duration_str := models.GetConfigVal("zt_static_ip_expire_duration") //单位 秒
	expire_duration := util.StoI(expire_duration_str)
	if expire_duration == 0 {
		expire_duration = 7 * 86400
	}
	duration := durationTime + expire_duration //开通时长 秒  冷静期配置

	data_info := map[string]string{
		"oem":      oemName,             //被授权可使用的项目
		"uid":      util.ItoS(uid),      //项目开通的用户ID
		"order_id": orderId,             //项目开通的订单号
		"duration": util.ItoS(duration), //开通的时长（秒）
		"ip":       ipStr,               //自选IP，多个用,号隔开
	}

	//获取提取信息-开通扣费  ，请求资源中台
	fmt.Println(data_info)
	err1, requestStr := util.HttpPostFormHeader(apiUrl, data_info, headerInfo)
	fmt.Println(err1)
	fmt.Println(requestStr)
	AddLogs("IpRenew", requestStr) //写日志
	responseInfo := ResponseStaticZtModel{}
	err1 = json.Unmarshal([]byte(requestStr), &responseInfo)
	if err1 != nil {
		return false, "__T_IP_RENEW_ERROR"
	}
	if responseInfo.Code != 200 {
		AddLogs("IpRenewErr01", responseInfo.Msg) //写日志
		return false, "__T_IP_RENEW_ERROR -- 1"
	}
	return true, ""
}

// StaticZtRelease 资源释放
func StaticZtRelease(uid int, ipStr string) (bool, string) {
	apiUrl := models.GetConfigVal("zt_static_ip_release") //资源中台开通
	if apiUrl == "" {
		apiUrl = "http://xc-static-mid-api.worldrift.com/api/release"
	}
	oemName := models.GetConfigVal("zt_static_oem_name")
	if oemName == "" {
		oemName = "cherry"
	}
	token := models.GetConfigVal("zt_static_ip_token") //资源中台账户Token
	if token == "" {
		token = StaticZToken
	}
	headerInfo := map[string]interface{}{
		"token": token,
	}

	data_info := map[string]string{
		"oem": oemName,        //被授权可使用的项目
		"uid": util.ItoS(uid), //项目开通的用户ID
		"ips": ipStr,          //自选IP，多个用,号隔开
	}

	//获取提取信息-开通扣费  ，请求资源中台
	fmt.Println(data_info)
	err1, requestStr := util.HttpPostFormHeader(apiUrl, data_info, headerInfo)
	fmt.Println(err1)
	fmt.Println(requestStr)
	AddLogs("IpRelease", requestStr) //写日志
	responseInfo := ResponseStaticReleaseZtModel{}
	err1 = json.Unmarshal([]byte(requestStr), &responseInfo)
	if err1 != nil {
		return false, "__T_IP_RELEASE_ERROR"
	}
	if responseInfo.Code != 200 {
		AddLogs("IpReleaseErr01", responseInfo.Msg) //写日志
		return false, "__T_IP_RELEASE_ERROR -- 1"
	}

	return true, ""
}

func StaticZtRegionStockList() (bool, string, map[string]int) {
	lists := map[string]int{}
	attribute := "isp,double_isp,native_ct_isp"
	openIPType := models.GetConfigVal("zt_static_ip_open_type")
	if openIPType == "" {
		openIPType = "1"
	}
	apiUrl := models.GetConfigVal("zt_static_region_stock_list")
	if apiUrl == "" {
		apiUrl = "http://xc-static-mid-api.worldrift.com/api/get_region_stock_list"
	}
	oemName := models.GetConfigVal("zt_static_oem_name")
	if oemName == "" {
		oemName = "cherry"
	}
	token := models.GetConfigVal("zt_static_ip_token")
	if token == "" {
		token = StaticZToken
	}
	headerInfo := map[string]interface{}{
		"token": token,
	}
	data_info := map[string]string{
		"oem":       oemName, //被授权可使用的项目
		"ip_type":   "2",     //IP类型，1=数据中心，2=静态住宅
		"cate":      openIPType,
		"attribute": attribute, //IP属性，isp=单ISP、double_isp=双ISP、native_isp=原生、native_hq_isp=优质原生、native_ct_isp=原生电信  多个用,逗号隔开
	}

	err, requestStr := util.HttpPostFormHeader(apiUrl, data_info, headerInfo)
	if err != nil {
		AddLogs("RegionStock HttpPostFormHeader", err.Error())
		return false, "__T_IP_RELEASE_ERROR", lists
	}
	responseInfo := ResponseZtRegionStockModel{}
	err1 := json.Unmarshal([]byte(requestStr), &responseInfo)
	if err1 != nil {
		AddLogs("RegionStock Unmarshal", err1.Error())
		return false, "__T_IP_RESOURCE_ERROR", lists
	}
	if responseInfo.Code != 200 {
		AddLogs("RegionStock code", responseInfo.Msg)
		return false, "__T_IP_RESOURCE_ERROR--1", lists
	}
	lists = responseInfo.Data.RegionStock
	return true, "", lists
}
