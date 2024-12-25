package controller

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/ipdat"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"strings"
)

type RequestInfo struct {
	Code int         `json:"code"` //洲
	Msg  string      `json:"msg"`  //城市
	Data Ip2AreaInfo `json:"data"` //省份
}
type Ip2AreaInfo struct {
	Continent   string `json:"continent"`    //洲
	City        string `json:"city"`         //城市
	Province    string `json:"province"`     //省份
	Country     string `json:"country"`      //国家
	CountryCode string `json:"country_code"` //国家英文简写
	ZipCode     string `json:"zip_code"`     //邮编
	Asn         string `json:"asn"`          //asn
	Isp         string `json:"isp"`          //运营商
	TimeZone    string `json:"time_zone"`    //时区
	Lat         string `json:"lat"`          //纬度
	Lon         string `json:"lon"`          //经度
	Ip          string `json:"ip"`           //IP
}

// 获取ip 地区信息
func GetIpInfo(ip string) Ip2AreaInfo {
	url := models.GetConfigVal("ip2area_url")
	if url == "" {
		url = "http://129.226.147.198:9000"
	}
	url = strings.Trim(url, "/") + "/get_area"
	param := map[string]string{
		"ip": ip,
	}
	infoStr := util.HttpPostForm(url, param)
	resInfo := RequestInfo{}
	err := json.Unmarshal([]byte(infoStr), &resInfo)
	ipInfo := resInfo.Data
	if err != nil || resInfo.Code != 0 {
		ipInfoOld, _ := ipdat.IPDat.GetIpInfo(ip)
		ipInfo.Country = ipInfoOld.Country
		ipInfo.CountryCode = ipInfoOld.CountryCode
		ipInfo.Ip = ipInfoOld.Ip
	}
	return ipInfo
}

func (ip Ip2AreaInfo) String() string {
	return ip.Country + "|" + ip.Province + "|" + ip.City + "|" + ip.Isp + ip.ZipCode
}
