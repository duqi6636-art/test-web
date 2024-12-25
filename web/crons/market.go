package crons

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	emailSender "api-360proxy/web/service/email"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"strings"
	"time"
)

// 邮件营销
func MarketDoSending() {

	openConfig := models.GetConfigVal("email_market_open") // 营销邮件开启
	if openConfig == "1" {
		lists := models.GetEmailMarketListBy() // 从邮件发送队列里面获取未发送的邮件

		if len(lists) < 1 {
			fmt.Println("no data")
		} else {
			default_mail := models.GetConfigVal("default_email")
			for _, v := range lists {
				email := v.Email

				if email != "" {
					isSend := 3
					if v.IsEmail == 1 {
						result := false
						//发送邮件
						vars := make(map[string]string)
						vars["email"] = email
						if default_mail == "aws_mail" {
							result = emailSender.AwsSendEmailMarket(email, v.Code, vars, "market_email")
							fmt.Println("send result aws:", result)
						}
						if default_mail == "tencent_mail" {
							result = emailSender.TencentSendEmailMarket(email, v.Code, vars, "market_email")
							fmt.Println("send result tencent:", result)
						}

						if result == true {
							isSend = 1
						} else {
							isSend = 2
						}
					}
					models.EditEmailMarket(v.Id, isSend)
					time.Sleep(1)
				}

				// 是否发送优惠券
				if v.IsCoupon == 1 {

					regConfig := models.GetConfigVal("reg_coupon_24h")     // 新用户优惠券配置
					repayConfig := models.GetConfigVal("repay_coupon_12h") // 老用户优惠券
					//发送优惠券
					if v.Cate == "reg" {
						if regConfig != "" {
							CouponInfoAuto(v.Uid, v.Username, regConfig)
						}
					} else {
						if repayConfig != "" {
							CouponInfoAuto(v.Uid, v.Username, repayConfig)
						}
					}
				}
			}
			fmt.Println("---------------")
		}

		fmt.Println("---end---")
		return
	}
}

// 检测用户券信息和发放券
func CouponInfoAuto(uid int, username, couponInfo string) {
	if couponInfo != "" {
		couArr := strings.Split(couponInfo, ",")           // 可领取的优惠券数组
		_, couponUserList := models.GetCouponList(uid, "") //获取用户已领取的券
		var hasCidArr []int
		for _, userV := range couponUserList {
			hasCidArr = append(hasCidArr, userV.Cid)
		}

		for _, val := range couArr {
			cid := util.StoI(val)

			// 用户没有领取过 才会继续往下执行
			if !util.InArrayInt(cid, hasCidArr) {

				nowTime := util.GetNowInt()
				/// 获取优惠券信息
				err, couponInfo := models.GetCouponById(cid)
				if err != nil {
					continue
				}
				expire := nowTime + couponInfo.ExpiryDay*86400
				codeStr := GetUuid()
				codeArr := strings.Split(codeStr, "-")
				lens := len(codeArr)
				code := codeArr[0] + "-" + codeArr[lens-2] + "-" + codeArr[lens-1]
				info := models.CouponList{}
				info.Name = couponInfo.Name
				info.Title = couponInfo.Title
				info.Cate = couponInfo.Cate
				info.Type = couponInfo.Type
				info.Value = couponInfo.Value
				info.UserType = couponInfo.UserType
				info.Status = couponInfo.Status
				info.UseType = couponInfo.UseType
				info.Meals = couponInfo.Meals
				info.ExpiryDay = couponInfo.ExpiryDay
				info.UseCycle = couponInfo.UseCycle
				info.UseNumber = couponInfo.UseNumber
				info.Platform = couponInfo.Platform
				info.GroupId = couponInfo.GroupId
				info.Cron = couponInfo.Cron
				info.Cid = couponInfo.Id
				info.Code = code
				info.Expire = expire
				info.BindUid = uid
				info.BindUsername = username
				info.CreateTime = nowTime
				err = models.AddCoupon(info)
				fmt.Println("coupon:", err)
			}
		}
	}
}

// 递归生成唯一uuid
func GetUuid() string {
	username := uuid.NewV4()
	return username.String()
}
