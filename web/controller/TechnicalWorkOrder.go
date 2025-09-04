package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
	"net/http"
	"net/url"
	"time"
)

// SubmitTechnicalWorkOrder 提交技术工单
func SubmitTechnicalWorkOrder(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	produceType := c.DefaultPostForm("type", "")
	title := c.DefaultPostForm("title", "")
	content := c.DefaultPostForm("content", "")
	files := c.DefaultPostForm("files", "")
	email := c.DefaultPostForm("email", "")
	fileInfo := c.DefaultPostForm("file_info", "")
	contactType := c.DefaultPostForm("contact_type", "")
	contactValue := c.DefaultPostForm("contact_value", "")

	// 必填字段验证
	if produceType == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}
	t := com.StrTo(produceType).MustInt()
	if t < 1 || t > 2 {
		JsonReturn(c, -1, "__T_INVALID_TYPE", nil)
		return
	}

	if title == "" {
		JsonReturn(c, -1, "__T_TITLE_REQUIRED", nil)
		return
	}
	if content == "" {
		JsonReturn(c, -1, "__T_CONTENT_REQUIRED", nil)
		return
	}
	if len(content) > 800 {
		JsonReturn(c, -1, "__T_CONTENT_TOO_LONG", nil)
		return
	}

	if email == "" {
		JsonReturn(c, -1, "__T_EMAIL_REQUIRED", nil)
		return
	}

	if !util.CheckEmail(email) {
		JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", nil)
		return
	}

	if contactType != "" && !util.InArrayString(contactType, []string{"whatsapp", "email", "others"}) {
		JsonReturn(c, -1, "__T_INVALID_CONTACT_TYPE", nil)
		return
	}

	var now = time.Now()
	pre24 := now.AddDate(0, 0, -1)
	list := models.GetTechnicalWorkOrderList("uid = ? AND create_time >= ? AND type = ? AND status >= 0", uid, pre24.Unix(), t)
	if len(list) >= 5 {
		JsonReturn(c, e.ERROR, "__24_HOURS_MAX_5", nil)
		return
	}
	ip := c.ClientIP()
	ipInfo := GetIpInfo(ip)
	data := models.TechnicalWorkOrder{
		Uid:          uid,
		Username:     user.Username,
		Type:         t,
		Status:       0,
		Files:        files,
		Email:        email,
		Title:        title,
		Content:      content,
		LastTime:     now.Unix(),
		CreateTime:   now.Unix(),
		UserIp:       ip,
		UserRegion:   ipInfo.CountryCode,
		DealStatus:   0,
		FileInfo:     fileInfo,
		ContactType:  contactType,
		ContactValue: contactValue,
	}
	err := models.CreateTechnicalWorkOrder(&data)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_CREATE_ERR", nil)
		return
	}
	// 更新工单编号
	workOrderNumber := fmt.Sprintf("360CHERRY%v%v", now.Format("20060102"), data.Uid)
	_ = models.UpdateTechnicalWorkOrder("id = ?", map[string]interface{}{"number": workOrderNumber}, data.Id)

	// 发送钉钉通知
	go func() {
		message := fmt.Sprintf("📋 新工单提交通知\n\n"+
			"工单编号: %s\n"+
			"用户: %s\n"+
			"邮箱: %s\n"+
			"标题: %s\n"+
			"内容: %s\n"+
			"联系方式: %s - %s\n"+
			"提交时间: %s\n"+
			"用户IP: %s\n"+
			"用户地区: %s",
			workOrderNumber,
			user.Username,
			email,
			title,
			content,
			contactType,
			contactValue,
			now.Format("2006-01-02 15:04:05"),
			ip,
			ipInfo.CountryCode)

		err := sendDingTalkNotification(message)
		if err != nil {
			fmt.Printf("钉钉通知发送失败: %v\n", err)
		}
	}()
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
}

// RevocationTechnicalWorkOrder 撤回工单
func RevocationTechnicalWorkOrder(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	id := c.DefaultPostForm("id", "") // 工单id
	two := models.GetTechnicalWorkOrder("id = ? AND uid= ? AND status = 0 AND type = 2", id, uid)
	if two.Id <= 0 {
		JsonReturn(c, e.ERROR, "__T_WORD_ORDER_DOESNT_EXIST", nil)
		return
	}
	err := models.UpdateTechnicalWorkOrder("id = ?", map[string]interface{}{"status": -1, "deal_status": 1}, id)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_REVOCATION_WORK_ORDER_ERR", nil)
		return
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
}

// DelTechnicalWorkOrder 删除工单
func DelTechnicalWorkOrder(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	// 工单id
	id := c.DefaultPostForm("id", "")
	two := models.GetTechnicalWorkOrder("id = ? AND uid= ? AND type = 2 ", id, uid)
	if two.Id <= 0 {
		JsonReturn(c, e.ERROR, "__T_WORD_ORDER_DOESNT_EXIST", nil)
		return
	}
	err := models.UpdateTechnicalWorkOrder("id = ?", map[string]interface{}{"status": -2, "deal_status": 1}, id)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_REVOCATION_WORK_ORDER_ERR", nil)
		return
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
}

// FinishTechnicalWorkOrder 结束工单
func FinishTechnicalWorkOrder(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	id := c.DefaultPostForm("id", "") // 工单id
	two := models.GetTechnicalWorkOrder("id = ? AND uid= ? AND type = 2 AND status = 1", id, uid)
	if two.Id <= 0 {
		JsonReturn(c, e.ERROR, "__T_WORD_ORDER_DOESNT_EXIST", nil)
		return
	}
	err := models.UpdateTechnicalWorkOrder("id = ?", map[string]interface{}{"status": 2, "deal_status": 1, "finish_time": time.Now().Unix()}, id)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_REVOCATION_WORK_ORDER_ERR", nil)
		return
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
}

// ListTechnicalWorkOrder 工单列表
func ListTechnicalWorkOrder(c *gin.Context) {
	resCode, msg, user := DealUser(c)
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	status := c.DefaultPostForm("status", "")
	number := c.DefaultPostForm("number", "")
	lang := c.DefaultPostForm("lang", "en")
	startTime := com.StrTo(c.DefaultPostForm("start_time", "0")).MustInt()
	endTime := com.StrTo(c.DefaultPostForm("end_time", "0")).MustInt()

	query := "uid = ? AND type = 2 AND status > -2"
	args := []interface{}{user.Id}
	if status != "" {
		statusInt := com.StrTo(status).MustInt()
		if statusInt == 2 {
			query += " AND (status = 2 or status = 3)"
		} else {
			query += " AND status = ?"
			args = append(args, statusInt)
		}
	}

	if number != "" {
		query += " AND (number like ? or content like ?)"
		args = append(args, "%"+number+"%", "%"+number+"%")
	}

	if startTime > 0 && endTime > 0 {
		query += " AND create_time >= ? AND create_time <= ?"
		args = append(args, startTime, endTime)
	}

	list := models.GetTechnicalWorkOrderList(query, args...)
	var rsp = make([]models.ListTechnicalWorkOrderRsp, 0)
	for _, two := range list {
		fileInfo := make([]interface{}, 0)
		err := json.Unmarshal([]byte(two.FileInfo), &fileInfo)
		if err != nil {
			fileInfo = []interface{}{}
		}
		rsp = append(rsp, models.ListTechnicalWorkOrderRsp{
			Id:           two.Id,
			Number:       two.Number,
			Uid:          two.Uid,
			Username:     two.Username,
			Type:         two.Type,
			Status:       two.Status,
			DealStatus:   two.DealStatus,
			Files:        two.Files,
			Email:        two.Email,
			Title:        two.Title,
			Content:      two.Content,
			CreateTime:   util.GetTimeHIByLang(int(two.CreateTime), lang),
			FileInfo:     fileInfo,
			FinishTime:   util.GetTimeHIByLang(int(two.FinishTime), lang),
			ReadStatus:   two.ReadStatus,
			ContactType:  two.ContactType,
			ContactValue: two.ContactValue,
		})
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", rsp)
}

// AddTechnicalWorkOrderDetails 添加工单详情
func AddTechnicalWorkOrderDetails(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	id := com.StrTo(c.DefaultPostForm("id", "")).MustInt() // 工单id
	files := c.DefaultPostForm("files", "")
	content := c.DefaultPostForm("content", "")
	two := models.GetTechnicalWorkOrder("id = ? AND (status = 0 OR status = 1)", id)
	if two.Id <= 0 {
		JsonReturn(c, e.ERROR, "__T_WORD_ORDER_DOESNT_EXIST", nil)
		return
	}
	var now = time.Now()
	data := models.TechnicalWorkOrderDetails{
		Uid:        uid,
		Type:       1,
		Files:      files,
		TwoId:      id,
		Content:    content,
		CreateTime: now.Unix(),
	}
	err := models.CreateTechnicalWorkOrderDetails(&data)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_ADD_DETAILS_WORK_ORDER_ERR", err.Error())
		return
	}
	_ = models.UpdateTechnicalWorkOrder("id = ?", map[string]interface{}{"last_time": now.Unix(), "deal_status": 0}, id)
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
}

// TechnicalWorkOrderDetailsList 工单详情列表
func TechnicalWorkOrderDetailsList(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	id := com.StrTo(c.DefaultPostForm("id", "")).MustInt() // 工单id
	lang := c.DefaultPostForm("lang", "en")
	// 获取工单内容
	two := models.GetTechnicalWorkOrder("id = ?", id)
	if two.Id <= 0 || two.Status < -1 {
		JsonReturn(c, e.ERROR, "__T_WORD_ORDER_DOESNT_EXIST", nil)
		return
	}
	if two.ReadStatus == 1 {
		query := "uid = ? AND id = ?"
		args := []interface{}{uid, two.Id}
		err := models.UpdateTechnicalWorkOrder(query, map[string]interface{}{"read_status": 0}, args...)
		if err != nil {
			JsonReturn(c, e.ERROR, "__T_FAIL", nil)
			return
		}
	}
	fileInfo := make([]interface{}, 0)
	err := json.Unmarshal([]byte(two.FileInfo), &fileInfo)
	if err != nil {
		fileInfo = []interface{}{}
	}
	twoRsp := models.ListTechnicalWorkOrderRsp{
		Id:         two.Id,
		Number:     two.Number,
		Uid:        two.Uid,
		Username:   two.Username,
		Type:       two.Type,
		Status:     two.Status,
		DealStatus: two.DealStatus,
		Files:      two.Files,
		Email:      two.Email,
		Title:      two.Title,
		Content:    two.Content,
		CreateTime: util.GetTimeHIByLang(int(two.CreateTime), lang),
		FileInfo:   fileInfo,
		FinishTime: util.GetTimeHIByLang(int(two.FinishTime), lang),
	}
	list := models.GetTechnicalWorkOrderDetailsList("uid = ? AND two_id = ?", uid, id)
	rspList := make([]models.TechnicalWorkOrderDetailsListRsp, 0)
	for _, details := range list {
		rspList = append(rspList, models.TechnicalWorkOrderDetailsListRsp{
			Id:         details.Id,
			TwoId:      details.TwoId,
			Uid:        details.Uid,
			Type:       details.Type,
			Files:      details.Files,
			Content:    details.Content,
			CreateTime: util.GetTimeHIByLang(int(details.CreateTime), lang),
		})
	}
	rsp := map[string]interface{}{"list": rspList, "data": twoRsp}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", rsp)
}

// GetTechnicalWorkOrderReadStatus 获取工单是否有未读消息
func GetTechnicalWorkOrderReadStatus(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	var readStatus = 0
	two := models.GetTechnicalWorkOrder("uid = ? AND type = 2 AND status >= 0 AND read_status = 1", uid)
	if two.Id > 0 {
		readStatus = 1
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", map[string]interface{}{"read_status": readStatus})
}

// SetTechnicalWorkOrderReadStatus 修改工单的未读状态
func SetTechnicalWorkOrderReadStatus(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	id := com.StrTo(c.DefaultPostForm("id", "0")).MustInt() // 工单id
	query := "uid = ? AND read_status = 1"
	args := []interface{}{uid}
	if id > 0 {
		query += " AND id = ? "
		args = append(args, id)
	}
	err := models.UpdateTechnicalWorkOrder(query, map[string]interface{}{"read_status": 0}, args...)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_FAIL", nil)
		return
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
}

// sendDingTalkNotification 发送钉钉通知（支持加签验证）
func sendDingTalkNotification(message string) error {
	// 从配置中获取钉钉webhook地址和密钥
	webhook := models.GetConfigVal("Technical_Work_Order_DingTalk_WebHook")
	secret := models.GetConfigVal("Technical_Work_Order_DingTalk_Secret")

	if webhook == "" {
		// 如果配置为空，使用默认地址
		webhook = "https://oapi.dingtalk.com/robot/send?access_token=c726d8d031f7b950e0f1725e78cdc8793a39079af49170829636f87b8bf5f8b4"
	}

	if secret == "" {
		secret = "SECe8c45ad94ba174e5df380e504b90293abcf09e03532d48f9ea8a2544018f67d9"
	}

	// 添加关键词前缀，确保通过安全验证
	safeMessage := "【工单通知】" + message

	// 如果配置了密钥，则使用加签验证
	var finalWebhook string
	if secret != "" {
		// 生成时间戳（毫秒）
		timestamp := time.Now().UnixNano() / 1e6

		// 生成签名字符串
		stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)

		// 使用HMAC-SHA256生成签名
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(stringToSign))
		signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

		// 构造带签名的webhook URL
		finalWebhook = fmt.Sprintf("%s&timestamp=%d&sign=%s",
			webhook, timestamp, url.QueryEscape(signature))
	} else {
		// 不使用加签，直接使用原webhook
		finalWebhook = webhook
	}

	// 构造钉钉消息体
	msgData := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": safeMessage,
		},
		"at": map[string]interface{}{
			"atMobiles": []string{}, // 可以配置需要@的手机号
			"isAtAll":   false,      // 是否@所有人
		},
	}

	// 序列化为JSON
	jsonData, err := json.Marshal(msgData)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", finalWebhook, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json;charset=utf-8")

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查钉钉API返回的错误码
	if errCode, ok := result["errcode"]; ok {
		if code, ok := errCode.(float64); ok && code != 0 {
			return fmt.Errorf("钉钉API返回错误: %v, 错误信息: %v", result["errcode"], result["errmsg"])
		}
	}

	return nil
}
