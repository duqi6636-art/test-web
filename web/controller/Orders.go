package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"github.com/gin-gonic/gin"
	"strings"
)

// 用户订单列表
func UserOrderListBy(c *gin.Context) {
	pak_type := c.DefaultPostForm("pak_type", "")         //套餐类型 isp,normal,static,flow,flow_agent,agent,flow_day
	pay_status_str := c.DefaultPostForm("pay_status", "") //支付状态
	start_date := c.DefaultPostForm("start_time", "")     //开始时间
	end_date := c.DefaultPostForm("end_time", "")         //结束时间
	resCode, msg, user := DealUser(c)                     //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	start, end := 0, 0
	if start_date != "" && end_date != "" {
		start = util.StoI(util.GetTimeStamp(start_date, "Y-m-d"))
		end = util.StoI(util.GetTimeStamp(end_date, "Y-m-d"))
		end = end + 86400
	}
	pay_status := util.StoI(pay_status_str)
	pStatus := pay_status
	if pStatus == 2 {
		pStatus = 1
	}
	lists := models.GetOrderListBy(uid, start, end, pStatus, pak_type)

	_, confPayLists := models.GetPayPlatConfList()
	confPayArr := map[int]string{}
	for _, v := range confPayLists {
		confPayArr[v.Id] = v.ShowName
	}

	nowTime := util.GetNowInt()
	_, payPlatList := models.GetPayPlatConfList()
	resList := []models.ResSocksOrder{}
	for _, v := range lists {
		payPlat := "Other"
		for _, vp := range payPlatList {
			if v.PayPlat == vp.Id {
				payPlat = vp.ShowName
			}
		}
		pakType := ""
		char := "IPs"
		if v.PakType == "normal" {
			pakType = "Balance"
		} else if v.PakType == "static" {
			pakType = "Static Residential Proxies"
		} else if v.PakType == "flow" {
			pakType = "Residential Proxies"
			char = "GB"
		} else if v.PakType == "flow_agent" {
			pakType = "Residential Proxies (Business)"
			char = "GB"
		} else if v.PakType == "agent" {
			pakType = "Socks5 Proxies (Business)"
			char = "IPs"
		} else if v.PakType == "flow_day" {
			pakType = "Unlimited Residential Proxies(Bandwidth)"
			char = "Day"
		} else if v.PakType == "flow_day_port" {
			pakType = "Unlimited Residential Proxies(Port)"
			char = "Day"
		} else if v.PakType == "dynamic_isp" {
			pakType = "Dynamic ISP Proxies"
			char = "GB"
		} else if v.PakType == "balance" {
			pakType = "Balance Recharge"
			char = ""
		} else {
			pakType = "Socks5 Proxies"
		}
		pakName := v.PakName
		if v.PakType == "static" {
			text := strings.ToUpper(strings.Trim(v.PakRegion, ","))
			pakName = text + "-" + v.PakName
		}
		payStatus := ""
		if v.PayStatus == 3 {
			payStatus = "Payment successful"
		} else if v.PayStatus == 1 {
			if (nowTime - v.CreateTime) < 1800 {
				payStatus = "Payment pending"
			} else {
				payStatus = "Payment failed"
			}
		}

		patMethod, okp := confPayArr[v.PayPlat]
		if !okp {
			patMethod = "Other " + util.ItoS(v.PayPlat)
		}
		info := models.ResSocksOrder{}
		info.PakId = v.PakId
		info.OrderId = v.OrderId
		info.Name = v.Name
		info.Value = v.Value
		info.Total = v.Value + v.Give + v.Gift + v.Reward
		info.Give = v.Give + v.Gift + v.Reward
		info.Char = char
		if v.PakType == "flow" || v.PakType == "flow_agent" {
			info.Value = v.Value / 1024 / 1024 / 1024
			info.Total = (v.Value + v.Give + v.Gift + v.Reward) / 1024 / 1024 / 1024
			info.Give = (v.Give + v.Gift + v.Reward) / 1024 / 1024 / 1024
		}
		if v.PakType == "flow_day" {
			info.Value = v.Value / 86400
			info.Total = (v.Value + v.Give + v.Gift + v.Reward) / 86400
			info.Give = (v.Give + v.Gift + v.Reward) / 86400
		}
		info.TrueMoney = v.TrueMoney
		info.PayTime = util.GetTimeStr(v.PayTime, "d-m-Y H:i:s")
		info.PayPlat = payPlat
		info.PaySn = v.PaySn
		info.CreateTime = util.GetTimeStr(v.CreateTime, "d-m-Y H:i:s")
		info.Plat = patMethod
		info.PakName = pakName
		info.PakType = pakType
		info.PayStatus = payStatus
		info.Invoice = v.Invoice
		if pay_status > 0 {
			if pay_status == 1 {
				if (nowTime - v.CreateTime) >= 1800 {
					resList = append(resList, info)
				}
			} else if pay_status == 2 {
				if (nowTime - v.CreateTime) < 1800 {
					resList = append(resList, info)
				}
			} else {
				resList = append(resList, info)
			}
		} else {
			resList = append(resList, info)
		}
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resList)
	return
}
