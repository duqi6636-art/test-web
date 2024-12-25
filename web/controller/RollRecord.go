package controller

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"github.com/gin-gonic/gin"
	"strings"
)

// 返回信息值
type ResRollOrderList struct {
	Minute   int     `json:"minute"`
	PakName  string  `json:"pak_name"`
	Email    string  `json:"email"`
	Give     int     `json:"give"`
	Discount float64 `json:"discount"`
}

// 返回信息值
type ResRollTxList struct {
	Money float64 `json:"money"`
	Email string  `json:"email"`
	//Date  string  `json:"date"`
	Minute int `json:"minute"`
}

// 购买订单滚动列表
// @BasePath /api/v1
// @Summary 获取滚动订单信息
// @Description 获取滚动订单信息
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param limit formData string false "每页显示条数"
// @Produce json
// @Success 0 {object} []ResRollOrderList{}
// @Router /web/roll/order [post]
func GetRollOrderList(c *gin.Context) {
	limitStr := c.DefaultPostForm("limit", "20")
	limit := util.StoI(limitStr)
	if limit == 0 {
		limit = 20
	}
	_, orders := models.GetOrderListByLimit(limit)
	nowTime := util.GetNowInt()
	result := []ResRollOrderList{}
	for _, v := range orders {
		res := ResRollOrderList{}
		timeNum := (nowTime - v.CreateTime) / 60
		minuteStr := timeNum % 30
		if minuteStr == 0 {
			minuteStr = 1
		}
		pName := v.PakName
		email := ""
		if v.Email != "" {
			arrStr := strings.Split(v.Email, "@")
			email = v.Email[0:3] + "******" + "@" + arrStr[1]
		} else {
			email = v.Username[0:3] + "******.com"
		}
		res.PakName = pName
		res.Email = email
		res.Give = v.Give
		res.Discount = v.Discount
		res.Minute = minuteStr
		result = append(result, res)
	}

	JsonReturn(c, 0, "__T_SUCCESS", result)
	return
}

// 返佣提现滚动列表
// @BasePath /api/v1
// @Summary 获取滚动提现信息
// @Description 获取滚动提现信息
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param limit formData string false "每页显示条数"
// @Produce json
// @Success 0 {object} []ResRollTxList{}
// @Router /web/roll/withdrawal [post]
func GetRollTxList(c *gin.Context) {
	limitStr := c.DefaultPostForm("limit", "20")
	limit := util.StoI(limitStr)
	if limit == 0 {
		limit = 20
	}
	moneyLists := models.GetWithdrawalPageBy(0, 0, limit)
	result := []ResRollTxList{}
	nowTime := util.GetNowInt()
	for _, v := range moneyLists {
		res := ResRollTxList{}
		timeNum := (nowTime - v.CreateTime) / 60
		minuteStr := timeNum % 30
		if minuteStr == 0 {
			minuteStr = 1
		}
		email := ""
		emailStr := ""
		if v.Email != "" {
			arrStr := strings.Split(v.Email, "@")
			email = v.Email[0:3] + "******" + "@" + arrStr[1]
			emailStr = strings.ToLower(arrStr[1])
		} else {
			email = v.Username[0:3] + "******.com"
		}
		res.Email = email
		res.Money = v.Money
		//res.Date = util.GetTimeStr(v.CreateTime, "Y-m-d")
		res.Minute = minuteStr
		if emailStr != "qq.com" {
			result = append(result, res)
		}
	}

	JsonReturn(c, 0, "__T_SUCCESS", result)
	return
}
