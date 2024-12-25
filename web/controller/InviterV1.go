package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"math"
	"strings"
)

// 获取邀请信息
// @BasePath /api/v1
// @Summary 获取邀请信息v1
// @Description 获取邀请信息v1
// @Tags 邀请返佣
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} map[string]interface{} "invite：用户邀请总人数，invite_url：邀请链接，today：今日佣金，total_money：累计总奖励，used_money：已用金额，total_withdrawal_money：累计提现金额，total_exchange_money：累计兑换金额，redeemed_isp：已兑换的ISP，redeemed_flow：已兑换的流量，ratio：佣金比例，isp_ratio：ISP佣金比例，flow_ratio：流量佣金比例，isp：已兑换的ISP，flow：已兑换的流量，max_withdrawal_money：最高提现金额"
// @Router /web/invite_v1/info [post]
func GetUserInviteInfoV1(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	//查询用户邀请记录
	err, inviteList := models.GetInviterListByMap(map[string]interface{}{"inviter_id": user.Id})
	inviteNum := 0 // 用户邀请总人数
	if err == nil && len(inviteList) > 0 {
		inviteNum = len(inviteList)
	}

	//查询用户佣金记录
	err, moneyList := models.GetMoneyList(user.Id)
	totalMoney := user.TotalMoney // 累计总奖励
	todayMoney := 0.00            // 今日佣金
	today := util.GetTodayTime()
	if err == nil && len(moneyList) > 0 {
		for _, v := range moneyList {
			if v.Today == today {
				todayMoney = todayMoney + (v.Money * v.Ratio)
			}
		}
	}

	// 查询用户邀请信息
	_, pUser := models.GetUserInviterByMap(map[string]interface{}{"uid": user.Id})
	inviterCode := pUser.InviterCode                                                              // 用户邀请码
	links := strings.TrimSpace(models.GetConfigV("inviter_share_url")) + "?invite=" + inviterCode // 邀请链接

	// 查询兑换记录
	err, exchangelist := models.GetExchangeListByUid(user.Id)
	totalExchangeMoney := 0.00
	var isp, flow int64
	if err == nil && len(exchangelist) > 0 {
		for _, v := range exchangelist {
			totalExchangeMoney = totalExchangeMoney + v.Money
			if v.Code == 12 {
				isp = isp + v.Ip
			} else if v.Code == 13 {
				flow = flow + (v.Ip / 1024 / 1024 / 1024)
			}
		}
	}

	// 查询提现记录
	withList := models.GetWithdrawalListBy(user.Id)
	totalWithdrawalMoney := 0.0
	for _, v := range withList {
		if v.Status == 2 {
			totalWithdrawalMoney = totalWithdrawalMoney + v.TrueMoney
		}
	}

	// 获取比例
	ispRatio := models.GetExchangeRatio("isp")
	flowRatio := models.GetExchangeRatio("flow")
	if ispRatio == 0.00 || flowRatio == 0.00 {
		ispRatio = 0.04  // 默认isp兑换比例
		flowRatio = 2.00 // 默认流量兑换比例
	}

	// 最低提现金额：$50
	minMoneyStr := models.GetConfigVal("inviter_tx_money_min")
	minMoney := 0.00
	if minMoneyStr == "" {
		minMoney = 50
	} else {
		minMoney = util.StoF(minMoneyStr)
	}
	// 获取保证金比例
	marginRatio := util.StoF(models.GetConfigVal("inviter_tx_margin_ratio")) // 保证金比例
	if marginRatio == 0.00 {
		marginRatio = 0.10
	}
	maxWithdrawalMoney := 0.00
	if user.UsedMoney-(user.UsedMoney*marginRatio) > minMoney {
		maxWithdrawalMoney = user.UsedMoney - (user.UsedMoney * marginRatio)
	}

	resData := map[string]interface{}{}
	resData["invite_num"] = inviteNum                             // 累计邀请人数
	resData["invite"] = inviterCode                               // 用户邀请码
	resData["invite_url"] = links                                 // 邀请链接
	resData["today"] = fmt.Sprintf("%.2f", todayMoney)            // 今日佣金
	resData["total_money"] = totalMoney                           // 累计总奖励
	resData["used_money"] = user.UsedMoney                        // 可用佣金
	resData["total_withdrawal_money"] = totalWithdrawalMoney      // 累计提现金额
	resData["total_exchange_money"] = totalExchangeMoney          // 累计兑换金额
	resData["redeemed_isp"] = isp                                 // 已兑换isp数量
	resData["redeemed_flow"] = flow                               // 已兑换flow数量
	resData["ratio"] = fmt.Sprintf("%.2f", pUser.Ratio*100) + "%" // 当前佣金比例
	resData["isp_ratio"] = fmt.Sprintf("%.2f", ispRatio)          // isp兑换比例
	resData["flow_ratio"] = fmt.Sprintf("%.2f", flowRatio)        // 流量兑换比例
	resData["isp"] = math.Floor(user.UsedMoney / ispRatio)        // 当前可兑换isp的最大值
	resData["flow"] = math.Floor(user.UsedMoney / flowRatio)      // 当前可兑换流量的最大值
	resData["max_withdrawal_money"] = maxWithdrawalMoney          // 当可提现最大金额
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resData)
	return
}

// 佣金兑换isp或流量
// @BasePath /api/v1
// @Summary 佣金兑换isp或流量v1
// @Description 佣金兑换isp或流量v1
// @Tags 邀请返佣
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param cate formData string true "兑换的类型   isp  /  flow"
// @Param number formData string true "兑换的数量"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/invite_v1/ex_balance [post]
func ExBalanceV1(c *gin.Context) {
	cate := c.DefaultPostForm("cate", "")  // 兑换的类型   isp  /  flow
	num := c.DefaultPostForm("number", "") // 兑换的数量
	number := util.StoI(num)
	if cate == "" || num == "" || number <= 0 {
		JsonReturn(c, e.ERROR, "__T_FAIL", nil)
		return
	}

	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	// 获取比例
	ratio := models.GetExchangeRatio(cate)
	if ratio == 0.00 {
		JsonReturn(c, e.ERROR, "__T_FAIL", nil)
		return
	}
	money := 0.00
	code := 0
	// 计算需扣除的金额
	if cate == "isp" {
		money = float64(number) * ratio
		code = 12
	} else if cate == "flow" {
		money = float64(number) * ratio
		code = 13
	}

	if ratio == 0.00 || money == 0.00 || user.UsedMoney <= 0 || user.UsedMoney < money {
		JsonReturn(c, e.ERROR, "__T_COMMISSION_LOW", nil)
		return
	}

	nowTime := util.GetNowInt()

	//写佣金记录
	moneyLog := models.UserMoneyLog{}
	moneyLog.Uid = user.Id
	moneyLog.Code = code
	moneyLog.Money = money
	moneyLog.Ip = int64(number)
	moneyLog.Ratio = ratio
	moneyLog.Mark = -1
	moneyLog.Status = 1
	moneyLog.CreateTime = nowTime
	moneyLog.Today = util.GetTodayTime()
	res := models.CreateUserMoneyLog(moneyLog)
	if res == nil {
		updateParams := make(map[string]interface{})
		updateParams["used_money"] = user.UsedMoney - money
		if cate == "isp" {
			// 更新用户isp余额
			updateParams["balance"] = user.Balance + number
		} else if cate == "flow" {
			/// 获取用户信息
			userFlowInfo := models.GetUserFlowInfo(user.Id)
			flows := int64(number) * 1021 * 1024 * 1024
			//流量余额变动日志
			go models.AddUserFlowsChangeLog(models.UserFlowsChangeLog{
				Uid:         user.Id,
				OldFlows:    userFlowInfo.Flows,
				NewFlows:    userFlowInfo.Flows + flows,
				ChangeFlows: flows,
				ChangeType:  "activity_ex_flow",
				ChangeTime:  nowTime,
				Mark:        1,
			})

			if userFlowInfo.ID == 0 {
				//创建用户余额IP
				socksIp := models.UserFlow{}
				socksIp.Uid = user.Id
				socksIp.Flows = flows
				socksIp.AllFlow = flows
				socksIp.Status = 1
				socksIp.ExpireTime = 30*86400 + nowTime
				socksIp.CreateTime = nowTime
				models.CreateUserFlow(socksIp)
			} else {
				upParam := make(map[string]interface{})
				upParam["all_flow"] = flows + userFlowInfo.AllFlow
				upParam["flows"] = flows + userFlowInfo.Flows
				if userFlowInfo.ExpireTime < nowTime {
					upParam["expire_time"] = nowTime + 30*86400
				}
				models.EditUserFlow(userFlowInfo.ID, upParam)
			}
			// 更新用户流量余额
			editError := models.UpdateUserById(user.Id, updateParams)
			fmt.Println("update_user_used_money", editError)
		}
		// 更新用户信息
		editError := models.UpdateUserById(user.Id, updateParams)
		fmt.Println("update_user_used_money", editError)
		JsonReturn(c, e.SUCCESS, "__T_APPLE_SUCCESS", nil)
		return
	}
	JsonReturn(c, e.ERROR, "__T_APPLE_FAIL", nil)
	return
}

// 用户提现
// @BasePath /api/v1
// @Summary 申请提现v1
// @Description 申请提现v1
// @Tags 邀请返佣
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param money formData string true "提现金额"
// @Param wallet formData string true "钱包地址"
// @Param email formData string true "联系邮箱"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/invite_v1/withdrawal [post]
func WithdrawalV1(c *gin.Context) {
	moneyStr := strings.TrimSpace(c.DefaultPostForm("money", "")) //提现金额
	wallet := strings.TrimSpace(c.DefaultPostForm("wallet", ""))  //钱包地址
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))    //联系邮箱
	resCode, msg, user := DealUser(c)                             //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	money := util.StoF(moneyStr)
	if money == 0 {
		JsonReturn(c, e.ERROR, "__T_MONEY_ERROR", nil)
		return
	}
	if wallet == "" {
		JsonReturn(c, e.ERROR, "__T_WALLET_ERROR", nil)
		return
	}
	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", nil)
		return
	}
	if !util.CheckEmail(email) {
		JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", nil)
		return
	}

	// 最低提现金额：$50
	minMoneyStr := models.GetConfigVal("inviter_tx_money_min")
	minMoney := 0.00
	if minMoneyStr == "" {
		minMoney = 50
		minMoneyStr = "50"
	} else {
		minMoney = util.StoF(minMoneyStr)
	}
	if money < minMoney {
		JsonReturn(c, e.ERROR, "__T_MONEY_NUMBER_MIN-- $"+minMoneyStr, nil)
		return
	}

	// 账户会保留可提取总额的10%作为保证金。
	// 获取保证金比例
	marginRatio := util.StoF(models.GetConfigVal("inviter_tx_margin_ratio")) // 保证金比例
	if marginRatio == 0.00 {
		marginRatio = 0.10
	}

	margin := money * marginRatio
	if margin+money > user.UsedMoney {
		JsonReturn(c, e.ERROR, "__T_APPLY_FAIL", nil)
		return
	}

	//写提现申请记录
	txLog := models.UserWithdrawalModel{}
	txLog.Uid = user.Id
	txLog.Username = user.Username
	txLog.Email = email
	txLog.Money = money
	txLog.TrueMoney = money
	txLog.Wallet = wallet
	txLog.Ip = c.ClientIP()
	txLog.CreateTime = util.GetNowInt()
	txLog.Status = 1
	res := models.AddWithdrawal(txLog)
	if res == nil {
		// 更新用户可用余额
		updateParams := make(map[string]interface{})
		updateParams["used_money"] = user.UsedMoney - money
		editError := models.UpdateUserById(user.Id, updateParams)
		fmt.Println("update_user_used_money", editError)
		// 更新用户IP余额
		JsonReturn(c, e.SUCCESS, "__T_APPLE_SUCCESS", nil)
		return
	}
	JsonReturn(c, e.ERROR, "__T_APPLE_FAIL", nil)
	return
}

// 查询用户返佣记录
// @BasePath /api/v1
// @Summary  邀请记录   /  兑换记录   /   提现记录
// @Description  邀请记录   /  兑换记录   /   提现记录
// @Tags 邀请返佣
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param cate formData string true "类型 invite：邀请返佣记录，exchange：佣金兑换，withdrawal：佣金提现"
// @Param limit formData string true "每页显示数量"
// @Param page formData string true "当前页"
// @Produce json
// @Success 0 {object} interface{} "total：总记录数，total_page：总页数，lists：记录列表 []models.ResUserOrderLogV1{}"
// @Router /web/invite_v1/get_user_money_log [post]
func GetUserMoneyLog(c *gin.Context) {
	cate := c.DefaultPostForm("cate", "invite") // 类型 invite：邀请返佣记录，exchange：佣金兑换，withdrawal：佣金提现
	limitStr := c.DefaultPostForm("limit", "10")
	pageStr := c.DefaultPostForm("page", "1")

	if limitStr == "" {
		limitStr = "10"
	}
	if pageStr == "" {
		pageStr = "1"
	}
	orderField := "m.id desc"
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	limit := util.StoI(limitStr)
	page := util.StoI(pageStr)
	offset := (page - 1) * limit

	// 查询邀请返佣记录
	if cate == "invite" {
		// 邀请支付记录
		list := models.GetMoneyListByInvitePageV1(uid, offset, limit, orderField)

		lists := []models.ResUserOrderLogV1{}
		for _, v := range list {
			info := models.ResUserOrderLogV1{}
			info.Email = v.Email
			info.RegTime = util.GetTimeStr(v.RegTime, "Y-m-d")
			info.Money = v.Money
			info.Commission = v.Money * v.Ratio
			lists = append(lists, info)
		}
		totalList := models.GetMoneyListByInvite(uid, "id")
		totalPage := int(math.Ceil(float64(len(totalList)) / float64(limit)))
		result := map[string]interface{}{
			"total":      len(totalList),
			"total_page": totalPage,
			"lists":      lists,
		}
		JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
		return
	} else if cate == "exchange" {
		// 佣金兑换记录
		list := models.GetUserExchangeList(uid, offset, limit)

		lists := []models.ResUserOrderLogV1{}
		for _, v := range list {
			info := models.ResUserOrderLogV1{}
			if v.Code == 12 {
				info.Type = "IP"
				info.Value = fmt.Sprintf("%dIPs", v.Value)
			} else if v.Code == 13 {
				info.Type = "Traffic"
				info.Value = fmt.Sprintf("%dGB", v.Value)
			}
			info.Money = v.Money
			info.CreateTime = util.GetTimeStr(v.CreateTime, "Y-m-d H:i")
			lists = append(lists, info)
		}
		total := models.GetUserExchangeCount(uid)
		totalPage := int(math.Ceil(float64(total) / float64(limit)))
		result := map[string]interface{}{
			"total":      total,
			"total_page": totalPage,
			"lists":      lists,
		}
		JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
		return
	} else if cate == "withdrawal" {
		// 佣金提现记录
		list := models.GetWithdrawalPageBy(uid, offset, limit)

		lists := []models.ResUserOrderLogV1{}
		for _, v := range list {
			info := models.ResUserOrderLogV1{}
			info.Money = v.Money
			info.CreateTime = util.GetTimeStr(v.CreateTime, "Y-m-d H:i")
			info.Status = v.Status
			lists = append(lists, info)
		}
		total := models.GetWithdrawalCount(uid)
		totalPage := int(math.Ceil(float64(total) / float64(limit)))
		result := map[string]interface{}{
			"total":      total,
			"total_page": totalPage,
			"lists":      lists,
		}
		JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
		return
	}
	// 若参数不对或用户没有数据，返回空数组
	result := map[string]interface{}{
		"total":      0,
		"total_page": 0,
		"lists":      []string{},
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
	return
}
