package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

// GetAutoRenewConfig 获取自动续费配置信息
func GetAutoRenewConfig(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	resData := map[string]interface{}{}                     //返回数据信息
	confList := models.GetConfBalanceRenewList("renew", "") //配置信息
	resList := map[string][]models.ResConfBalanceRenewModel{}
	resStaticList := map[string][]models.ResConfStaticBalanceRenewModel{}

	userConfList := models.GetUserAutoRenewDetailList(uid, "") //查询用户配置信息
	userConfInfo := map[string]models.UserAutoRenewDetailModel{}
	openAlreadyArr := map[int]string{}
	openIpArr := map[string]int{}
	for _, val := range userConfList {
		cate := strings.ToLower(val.Cate)
		if cate == "static" || cate == "unlimited" {
			openIpArr[cate] = val.ExId
		} else {
			userConfInfo[cate] = val
		}
		if val.Status == 1 {
			openAlreadyArr[val.ConfId] = cate
		}
	}
	// 已配置信息
	hasConfigMap := map[string]models.UserAutoRenewAlreadyModel{}
	for _, conf := range confList {
		cate := strings.ToLower(conf.Cate)
		info := models.ResConfBalanceRenewModel{}
		idStr := util.ItoS(conf.Id)

		detailInfo, ok := userConfInfo[cate]

		name := conf.Name
		setValue := int64(0)
		isRenew := 0            //续费开关  1开启
		balance, sy_day := 1, 1 //设置的值
		if ok {
			setValue = detailInfo.Value
			balance = int(detailInfo.Balance)
			sy_day = detailInfo.SyDay
			if conf.Id == detailInfo.ConfId {
				isRenew = detailInfo.Status
			}
		}
		if isRenew != 1 {
			isRenew = 0
		}
		totalPrice := 0.0
		setValue = setValue / conf.UnitValue
		if conf.IsCustom == 1 {
			if setValue > 0 {
				//name = fmt.Sprintf("%d %s", setValue, conf.Unit)
				totalPrice = conf.Price * float64(setValue)
			}
		} else {
			if conf.Value > 0 {
				totalPrice = conf.Price * float64(conf.Value)
				setValue = int64(conf.Value)
			}
		}
		totalPrices := util.StoF(util.FtoS2(totalPrice, 3)) //数据保留三位小数
		info.Id = idStr

		// 处理名字中的加号，拆分成 Name 和 NameExtra
		nameExtra := ""
		if strings.Contains(name, "+") {
			parts := strings.SplitN(name, "+", 2)
			if len(parts) == 2 {
				name = strings.TrimSpace(parts[0])
				nameExtra = strings.TrimSpace(parts[1])
			}
		}
		info.Name = name
		info.NameExtra = nameExtra
		info.Price = conf.Price
		info.Unit = conf.Unit
		info.Value = int(setValue)
		info.TotalPrice = totalPrices
		info.IsCustom = conf.IsCustom
		info.IsRenew = isRenew
		info.Balance = balance
		info.Day = sy_day
		if cate == "static" {
			infoS := models.ResConfStaticBalanceRenewModel{}
			infoS.Id = idStr
			infoS.Name = conf.Name
			infoS.Unit = conf.Unit
			infoS.Value = conf.Value
			infoS.TotalPrice = 0
			infoS.SingleIspPrice = conf.SingleIspPrice
			infoS.DoubleIspPrice = conf.DoubleIspPrice
			infoS.LocalIspPrice = conf.LocalIspPrice
			resStaticList[conf.Cate] = append(resStaticList[conf.Cate], infoS)
		} else {
			resList[conf.Cate] = append(resList[conf.Cate], info)
		}
		_, okh := openAlreadyArr[conf.Id]
		if okh {
			if cate == "static" || cate == "unlimited" {
				names, okip := openIpArr[cate]
				if !okip {
					name = ""
				} else {
					name = util.ItoS(names)
				}
			}
			hasInfo := models.UserAutoRenewAlreadyModel{
				Cate: cate,
				Name: name,
			}
			hasConfigMap[cate] = hasInfo
		}
	}

	hasConfig := []models.UserAutoRenewAlreadyModel{}
	for _, v := range hasConfigMap {
		name := v.Name
		if v.Cate == "static" {
			err, ipInfo := models.GetIpStaticIpById(util.StoI(name))
			if err == nil && ipInfo.Ip != "" {
				name = ipInfo.Ip + " ..."
			}
		}
		if v.Cate == "unlimited" {
			hostIpInfo := models.GetPoolFlowDayById(uid, util.StoI(name))
			if hostIpInfo.Ip != "" {
				name = hostIpInfo.Ip + " ..."
			}
		}
		v.Name = name
		if name != "" {
			hasConfig = append(hasConfig, v)
		}
	}

	listByte, _ := json.Marshal(resList)
	json.Unmarshal(listByte, &resData) //组合列表数据

	hasInfo := models.GetUserAutoRenewInfo(uid)
	resData["static"] = resStaticList             // 静态列表配置
	resData["open"] = hasInfo.Open                // 开启状态
	resData["method"] = hasInfo.Method            // 静态扣款方式顺序
	resData["has_config"] = hasConfig             // 已配置的信息
	resData["email"] = hasInfo.Email              // 邮件地址
	resData["email_switch"] = hasInfo.EmailSwitch // 邮件开关
	if hasInfo.Email == "" {
		resData["email"] = user.Email
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resData)
	return
}

// SetAutoRenewSwitch 设置自动续费开关
func SetAutoRenewSwitch(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	openStr := strings.TrimSpace(c.DefaultPostForm("auto_renew_switch", ""))
	emailSwitchStr := strings.TrimSpace(c.DefaultPostForm("email_switch", ""))
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))

	// 验证邮件格式
	if email != "" {
		if !util.CheckEmail(email) {
			JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", nil)
			return
		}
	}

	open := util.StoI(openStr)
	emailSwitch := util.StoI(emailSwitchStr)

	hasInfo := models.GetUserAutoRenewInfo(uid)
	var err error

	if hasInfo.Id == 0 {
		addInfo := models.UserAutoRenewModel{}
		addInfo.Uid = uid
		addInfo.Username = user.Username
		addInfo.Open = open
		if email != "" {
			addInfo.Email = email
		} else {
			addInfo.Email = user.Email
		}
		addInfo.EmailSwitch = emailSwitch
		addInfo.Ip = c.ClientIP()
		addInfo.CreateTime = util.GetNowInt()
		err = models.AddUserAutoRenew(addInfo)
	} else {
		upInfo := map[string]interface{}{}

		if openStr != "" {
			upInfo["open"] = open
		}

		if email != "" {
			upInfo["email"] = email
		} else if hasInfo.Email == "" {
			upInfo["email"] = user.Email
		}

		if emailSwitchStr != "" {
			upInfo["email_switch"] = emailSwitch
		}

		upInfo["ip"] = c.ClientIP()
		upInfo["update_time"] = util.GetNowInt()
		err = models.EditUserAutoRenew(uid, upInfo)
	}

	if err != nil {
		JsonReturn(c, e.ERROR, "__T_FAIL", nil)
		return
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return
}

// SetAutoRenewConfig 添加/编辑自动续费配置
func SetAutoRenewConfig(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	statusStr := strings.TrimSpace(c.DefaultPostForm("status", ""))     //状态  1开启
	valueIdStr := strings.TrimSpace(c.DefaultPostForm("value_id", ""))  //自动续费ID
	balanceStr := strings.TrimSpace(c.DefaultPostForm("balance", ""))   //剩余余额值
	valueStr := strings.TrimSpace(c.DefaultPostForm("value", ""))       //自动续费值
	dayStr := strings.TrimSpace(c.DefaultPostForm("day", ""))           //剩余天数  剩余几天有效期时自动续费
	method := strings.TrimSpace(c.DefaultPostForm("method", "balance")) //扣费类型
	cate := strings.TrimSpace(c.DefaultPostForm("cate", "isp"))         //扣费类型
	if method == "" {
		method = "balance"
	}
	balance := util.StoI(balanceStr)
	value := util.StoI(valueStr)
	day := util.StoI(dayStr)
	status := util.StoI(statusStr)
	valueId := util.StoI(valueIdStr)

	// 如果用户关闭自动续费（status=0）或者没有选择套餐，允许保存
	if status == 0 || valueId == 0 {
		// 查找用户现有的自动续费配置列表（所有类型）
		hasInfoList := models.GetUserAutoRenewDetailList(uid, cate)

		if len(hasInfoList) > 0 {
			// 如果存在配置，将所有配置更新为关闭状态
			for _, hasInfo := range hasInfoList {
				upInfo := map[string]interface{}{}
				upInfo["status"] = status
				upInfo["balance"] = balance
				upInfo["expire_day"] = 30
				upInfo["sy_day"] = day
				upInfo["method"] = method
				upInfo["update_time"] = util.GetNowInt()
				upInfo["status"] = 0 // 强制设置为关闭状态
				upInfo["update_time"] = util.GetNowInt()

				err := models.EditUserAutoRenewDetail(hasInfo.Id, upInfo)
				if err != nil {
					JsonReturn(c, e.ERROR, "__T_FAIL", nil)
					return
				}
			}
		}
		// 如果没有现有配置且用户选择关闭，直接返回成功（无需创建记录）
		JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
		return
	}

	configInfo := models.GetConfBalanceRenewById(valueId)
	if configInfo.Id == 0 {
		JsonReturn(c, e.ERROR, "__T_CONFIG_NOT_EXIST", nil)
		return
	}
	if configInfo.Value == 0 && (value < configInfo.Min || value > configInfo.Max) {
		JsonReturn(c, e.ERROR, "__T_RENEW_NUMBER_ERROR", nil) //续费设置信息错误
		return
	}
	if configInfo.IsCustom == 0 {
		value = configInfo.Value
	}

	hasInfo := models.GetUserAutoRenewDetail(uid, configInfo.Cate, 0)

	var err error
	if hasInfo.Id == 0 {
		addInfo := models.UserAutoRenewDetailModel{}
		addInfo.Uid = uid
		addInfo.Username = user.Username
		addInfo.Email = user.Email
		addInfo.ConfId = configInfo.Id
		addInfo.Cate = configInfo.Cate
		addInfo.Balance = int64(balance)
		addInfo.Value = int64(value) * configInfo.UnitValue
		addInfo.ExpireDay = 30 //续费的有效期天数
		addInfo.SyDay = day
		addInfo.Method = method
		addInfo.Status = status
		addInfo.Ip = c.ClientIP()
		addInfo.CreateTime = util.GetNowInt()
		err = models.AddUserAutoRenewDetail(addInfo)
	} else {
		upInfo := map[string]interface{}{}
		upInfo["status"] = status
		upInfo["balance"] = balance
		upInfo["conf_id"] = configInfo.Id
		upInfo["value"] = int64(value) * configInfo.UnitValue
		upInfo["expire_day"] = 30
		upInfo["sy_day"] = day
		upInfo["method"] = method
		upInfo["update_time"] = util.GetNowInt()

		err = models.EditUserAutoRenewDetail(hasInfo.Id, upInfo)
	}

	if err != nil {
		JsonReturn(c, e.ERROR, "__T_FAIL", nil)
		return
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return
}

// GetAutoRenewOrder 获取用户自动续费扣款顺序
func GetAutoRenewOrder(c *gin.Context) {
	resCode, msg, user := DealUser(c)
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	uid := user.Id
	orderList := models.GetUserAutoRenewOrderList(uid)

	// 如果用户没有配置，初始化默认配置
	if len(orderList) == 0 {
		err := models.InitUserAutoRenewOrder(uid, user.Username, c.ClientIP())
		if err != nil {
			JsonReturn(c, e.ERROR, "__T_INIT_ORDER_ERROR", nil)
			return
		}
		orderList = models.GetUserAutoRenewOrderList(uid)
	}

	// 构建返回数据
	result := make(map[string]interface{})
	result["list"] = orderList
	result["total"] = len(orderList)

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
}

// SetAutoRenewOrder 设置用户自动续费扣款顺序配置
func SetAutoRenewOrder(c *gin.Context) {
	resCode, msg, user := DealUser(c)
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	ordersStr := strings.TrimSpace(c.DefaultPostForm("orders", ""))
	if ordersStr == "" {
		JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil)
		return
	}

	uid := user.Id
	nowTime := util.GetNowInt()
	ip := c.ClientIP()

	// 解析扣款顺序配置
	// 支持两种格式：
	// 1. 直接传递续费类型顺序，如 "isp,static,flow,rotating,unlimited"
	// 2. 传递类型和优先级，如 "isp:1,static:2,flow:3"
	orderItems := strings.Split(ordersStr, ",")
	var orders []models.UserAutoRenewOrderModel

	// 验证续费类型
	validTypes := []string{"isp", "static", "flow", "rotating", "unlimited"}

	for i, item := range orderItems {
		item = strings.TrimSpace(item)

		var renewType string
		var priority int

		// 检查是否包含冒号
		if strings.Contains(item, ":") {
			parts := strings.Split(item, ":")
			if len(parts) != 2 {
				JsonReturn(c, e.ERROR, "__T_ORDER_FORMAT_ERROR", nil)
				return
			}

			renewType = strings.TrimSpace(parts[0])
			priorityStr := strings.TrimSpace(parts[1])

			var err error
			priority, err = strconv.Atoi(priorityStr)
			if err != nil {
				JsonReturn(c, e.ERROR, "__T_PRIORITY_FORMAT_ERROR", nil)
				return
			}
		} else {
			renewType = item
			priority = i + 1 // 优先级从1开始，按传递顺序递增
		}

		// 验证续费类型是否有效
		isValid := false
		for _, validType := range validTypes {
			if renewType == validType {
				isValid = true
				break
			}
		}

		if !isValid {
			JsonReturn(c, e.ERROR, "__T_INVALID_RENEW_TYPE", nil)
			return
		}

		order := models.UserAutoRenewOrderModel{
			Uid:        uid,
			Username:   user.Username,
			RenewType:  renewType,
			Priority:   priority,
			Status:     1,
			Ip:         ip,
			UpdateTime: nowTime,
			CreateTime: nowTime,
		}

		orders = append(orders, order)
	}

	// 批量更新配置
	err := models.BatchUpdateUserAutoRenewOrder(uid, orders)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_UPDATE_ORDER_ERROR", nil)
		return
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
}
