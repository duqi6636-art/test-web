package controller

//import (
//	"api-360proxy/web/e"
//	"api-360proxy/web/models"
//	"api-360proxy/web/pkg/util"
//	"encoding/json"
//	"fmt"
//	"github.com/gin-gonic/gin"
//	"math"
//	"strings"
//	"time"
//)
//
////兑换券
//func ExchangeCdk(c *gin.Context) {
//	resCode, msg, user := DealUser(c) //处理用户信息
//	if resCode != e.SUCCESS {
//		JsonReturn(c, resCode, msg, nil)
//		return
//	}
//	uid := user.Id
//	username := user.Username
//	code := strings.TrimSpace(c.DefaultPostForm("code", ""))
//	if code == "" {
//		JsonReturn(c, -1, "__T_EX_EMPTY", nil)
//		return
//	}
//
//	nowTime := util.GetNowInt()
//	balance := 0 //余额
//	res := false
//	var err error
//	exInfo := models.ExchangeList{}
//	err, exInfo = models.GetExchangeInfo(code)
//	if err != nil && exInfo.Id == 0 {
//		JsonReturn(c, -1, "__T_EX_CODE_ERROR", nil)
//		return
//	}
//	if exInfo.Status == 2 || exInfo.BindUid > 0 {
//		JsonReturn(c, -1, "__T_EX_USED", nil)
//		return
//	}
//
//	if exInfo.Expire > 0 && exInfo.Expire < nowTime {
//		JsonReturn(c, -1, "__T_EX_EXPIRED", nil)
//		return
//	}
//	balance = exInfo.Value
//	if exInfo.UseCycle >= 1 {
//		_, couponUse := models.GetExchangeByUsePlatform(uid, exInfo.UseCycle, exInfo.Platform)
//		if len(couponUse) >= exInfo.UseNumber {
//			JsonReturn(c, -1, "__T_COUPON_USED_LIMIT", nil)
//			return
//		}
//	}
//	// 同一分组下只可兑换一次
//	if exInfo.GroupId > 0 {
//		cdkInfo, _ := models.GetExchangeByGroupId(uid, exInfo.GroupId)
//		if cdkInfo.Id != 0 && exInfo.Id != cdkInfo.Id {
//			JsonReturn(c, -1, "__T_COUPON_USED_LIMIT", nil)
//			return
//		}
//	}
//
//	_, userInfo := models.GetUserById(uid)
//	// 用户类型和券类型不匹配
//	is_pay := "no_pay"
//	if userInfo.IsPay == "true" {
//		is_pay = "payed"
//	}
//	if exInfo.UserType != "all" && is_pay != exInfo.UserType {
//		JsonReturn(c, -1, "__T_COUPON_COMPARE", nil)
//		return
//	}
//
//
//
//
//
//
//
//	if exInfo.UseType == 2 {
//		info := exInfo
//		info.Id = 0
//		info.Expire = exInfo.Expire
//		info.BindUid = userInfo.Id
//		info.BindUsername = userInfo.Username
//		info.CreateTime = nowTime
//		err = models.AddExchange(info)
//		if err == nil {
//			res = true
//		}
//	} else {
//		exParam := map[string]interface{}{}
//		exParam["status"] = 2
//		exParam["use_time"] = nowTime
//		exParam["bind_uid"] = uid
//		exParam["bind_username"] = username
//		res = models.EditExchangeByCode(code, exParam)
//	}
//
//	if res {
//		// 处理用户余额问题
//		updateParams := make(map[string]interface{})
//		updateParams["balance"] = userInfo.Balance + balance
//		editError := models.EditUserByMap(map[string]interface{}{"id": userInfo.Id}, updateParams)
//		if editError != nil {
//			JsonReturn(c, -1, "__T_EX_ERROR", nil)
//			return
//		}
//		data := map[string]interface{}{
//			"user_balance": userInfo.Balance + balance,
//		}
//		JsonReturn(c, 0, "__T_EX_SUCCESS", data)
//		return
//	}
//	if models.GetConfigVal("dns_domain") != "" {
//		b := CreateUserDomain(userInfo)
//		if !b {
//			fmt.Println("域名生成失败")
//		}
//	}
//	JsonReturn(c, -1, "__T_EX_ERROR--!", nil)
//	return
//}
//
//// 生成兑换券
//func Generate(c *gin.Context) {
//	number := strings.TrimSpace(c.DefaultPostForm("number", "")) // 兑换ip数量
//	resCode, msg, user := DealUser(c)                            //处理用户信息
//	if resCode != e.SUCCESS {
//		JsonReturn(c, resCode, msg, nil)
//		return
//	}
//	uid := user.Id
//	numbers := util.StoI(number)
//	if numbers == 0 {
//		JsonReturn(c, e.ERROR, "__T_AMOUNT_ERROR", nil)
//		return
//	}
//	// 查询用户代理商余额记录
//	agentInfo := models.GetAgentBalanceByUid(uid)
//	if agentInfo.Id == 0 {
//		JsonReturn(c, e.ERROR, "__PC_NO_IP_INFO", nil)
//		return
//	}
//	minNumberStr := models.GetConfigV("generate_min_money") //生成 最小数量
//	minNumber := 0
//	if minNumberStr == "" {
//		minNumber = 50
//		minNumberStr = "50"
//	} else {
//		minNumber = util.StoI(minNumberStr)
//	}
//	if numbers < minNumber {
//		JsonReturn(c, e.ERROR, "__T_AMOUNT_MIN-- "+" "+minNumberStr, nil)
//		return
//	}
//	balStr := models.GetConfigV("user_balance_min") //用户余额最少剩余多少才可以生成
//	bal := 0
//	if balStr == "" {
//		bal = 200
//	} else {
//		bal = util.StoI(balStr)
//	}
//
//	if agentInfo.Balance < bal || agentInfo.Balance < numbers {
//		JsonReturn(c, e.ERROR, "__PC_NO_IP_INFO", nil)
//		return
//	}
//
//	nowTime := util.GetNowInt()
//	//写入兑换券记录
//	cdkStr := GetUuid()
//	cdkArr := strings.Split(cdkStr, "-")
//	lens := len(cdkArr)
//	code := strings.ToUpper(GenValidateCode(6) + cdkArr[lens-1])
//	info := models.ExchangeList{}
//	info.Cid = 0
//	info.Cate = 2
//	info.Uid = uid
//	info.Code = code
//	info.Name = "ISP"
//	info.BindUid = 0
//	info.BindUsername = ""
//	info.Status = 1
//	info.UseTime = 0
//	info.Title = "Exchange" + util.ItoS(numbers) + " IPs"
//	info.Value = numbers
//	info.UserType = "all"
//	info.UseType = 1
//	info.Expire = 0
//	info.ExpiryDay = 0
//	info.UseCycle = 0
//	info.UseNumber = 0
//	info.Platform = 0
//	info.GroupId = 0
//	info.CreateTime = nowTime
//	err := models.AddExchange(info)
//
//	data := map[string]interface{}{}
//	data["code"] = code
//	if err == nil {
//		// 写入日志
//		models.AddAgentExchange(user, agentInfo, numbers)
//		// 更新用户可用余额
//		agentBalance := make(map[string]interface{})
//		agentBalance["balance"] = agentInfo.Balance - numbers
//		editError := models.EditAgentBalanceBy(map[string]interface{}{"id": agentInfo.Id}, agentBalance)
//		fmt.Println("update_user_agent_balance", editError)
//		JsonReturn(c, e.SUCCESS, "__T_EX_GENERATE_OK", data)
//		return
//	}
//	JsonReturn(c, e.ERROR, "error", nil)
//	return
//}
//
//// 直接提现到本账户ip余额
//func DirectConversion(c *gin.Context) {
//	number := strings.TrimSpace(c.DefaultPostForm("number", "")) // 兑换ip数量
//	resCode, msg, user := DealUser(c)                            //处理用户信息
//	if resCode != e.SUCCESS {
//		JsonReturn(c, resCode, msg, nil)
//		return
//	}
//	numbers := util.StoI(number)
//	if numbers == 0 {
//		JsonReturn(c, e.ERROR, "__T_AMOUNT_ERROR", nil)
//		return
//	}
//	// 查询用户代理商余额记录
//	agentInfo := models.GetAgentBalanceByUid(user.Id)
//	if agentInfo.Id == 0 && agentInfo.Balance > 0 {
//		JsonReturn(c, e.ERROR, "__PC_NO_IP_INFO", nil)
//		return
//	}
//
//	if agentInfo.Balance < numbers {
//		JsonReturn(c, e.ERROR, "__PC_NO_IP_INFO", nil)
//		return
//	}
//
//	b := models.AgentAndUser(user, agentInfo, numbers)
//
//	if b {
//		if models.GetConfigVal("dns_domain") != "" {
//			b := CreateUserDomain(user)
//			if !b {
//				fmt.Println("域名生成失败")
//			}
//		}
//		JsonReturn(c, e.SUCCESS, "__T_EX_SUCCESS", nil)
//		return
//	}
//	JsonReturn(c, e.ERROR, "error", nil)
//	return
//}
//
//// 兑换券 列表
//func ExchangeList(c *gin.Context) {
//	limitStr := c.DefaultPostForm("limit", "10")
//	pageStr := c.DefaultPostForm("page", "1")
//
//	if limitStr == "" {
//		limitStr = "10"
//	}
//	if pageStr == "" {
//		pageStr = "1"
//	}
//
//	resCode, msg, user := DealUser(c) //处理用户信息
//	if resCode != e.SUCCESS {
//		JsonReturn(c, resCode, msg, nil)
//		return
//	}
//	uid := user.Id
//
//	limit := util.StoI(limitStr)
//	page := util.StoI(pageStr)
//	offset := (page - 1) * limit
//
//	exLists, _ := models.GetExchangeByCate(uid, offset, limit)
//	lists := []models.ResExchange{}
//	for _, v := range exLists {
//		info := models.ResExchange{}
//		info.Code = v.Code
//		info.Value = v.Value
//		info.BindUsername = v.BindUsername
//		info.BindUid = v.BindUid
//		info.Status = v.Status
//		info.Create_time = util.GetTimeStr(v.CreateTime, "Y/m/d H:i:s")
//		lists = append(lists, info)
//	}
//	totalList, _ := models.GetExchangeListByCate(uid)
//	totalPage := int(math.Ceil(float64(len(totalList)) / float64(limit)))
//	result := map[string]interface{}{
//		"total":      len(totalList),
//		"total_page": totalPage,
//		"lists":      lists,
//	}
//	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
//	return
//}
//
//// 禁用 代理商cdk券
//func ForbidEx(c *gin.Context) {
//	code := strings.TrimSpace(c.DefaultPostForm("code", ""))
//	resCode, msg, user := DealUser(c) //处理用户信息
//	if resCode != e.SUCCESS {
//		JsonReturn(c, resCode, msg, nil)
//		return
//	}
//	if code == "" {
//		JsonReturn(c, -1, "__T_EX_EMPTY", nil)
//		return
//	}
//	uid := user.Id
//	time.Sleep(1)
//
//	err, exInfo := models.GetExchangeInfo(code)
//	if err != nil && exInfo.Id == 0 {
//		JsonReturn(c, -1, "__T_EX_CODE_ERROR", nil)
//		return
//	}
//	if exInfo.Status == 3 {
//		JsonReturn(c, -1, "__T_EX_FORBID", nil)
//		return
//	}
//	if exInfo.Status == 2 || exInfo.BindUid > 0 {
//		JsonReturn(c, -1, "__T_EX_USED", nil)
//		return
//	}
//
//	// 查询用户代理商余额记录
//	agentBalance := 0
//	agentInfo := models.GetAgentBalanceByUid(uid)
//	if agentInfo.Id == 0 {
//		JsonReturn(c, e.ERROR, "__PC_NO_IP_INFO", nil)
//		return
//	}
//
//	total := agentInfo.Total
//	if agentInfo.Balance < 0 {
//		agentInfo.Balance = 0
//	}
//	agentBalance = agentInfo.Balance
//	newBalance := exInfo.Value + agentBalance
//	if newBalance > total {
//		JsonReturn(c, e.ERROR, "error", nil)
//		return
//	}
//
//	nowTime := util.GetNowInt()
//	data := map[string]interface{}{}
//	data["status"] = 3
//	data["use_time"] = nowTime
//	res := models.EditExchangeByCode(code, data)
//	if res == true {
//		where := map[string]interface{}{"id": agentInfo.Id}
//		info := map[string]interface{}{"balance": exInfo.Value + agentBalance}
//		err := models.EditAgentBalanceBy(where, info)
//		fmt.Println("update_user_agent_balance", err)
//		JsonReturn(c, e.SUCCESS, "__T_SUCCESS", data)
//		return
//	}
//	JsonReturn(c, e.ERROR, "error", nil)
//	return
//}


