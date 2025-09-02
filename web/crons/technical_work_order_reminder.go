package crons

import (
	"api-360proxy/web/models"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

// CheckOverdueTechnicalWorkOrders 检查超过18小时未处理的工单并发送钉钉通知
func CheckOverdueTechnicalWorkOrders() {
	// 计算18小时前的时间戳
	now := time.Now()
	overdue := now.Add(-1 * time.Hour).Unix()

	// 查询超过18小时未处理的工单（status=0表示未处理，send_status=0表示未发送过通知）
	workOrders := models.GetTechnicalWorkOrderList(
		"create_time <= ? AND status = 0 AND send_status = 0",
		overdue,
	)

	if len(workOrders) == 0 {
		fmt.Println("没有超时未处理的工单")
		return
	}

	log.Printf("发现 %d 个超过18小时未处理的工单\n", len(workOrders))

	// 为每个超时工单发送通知
	for _, workOrder := range workOrders {
		// 构造通知消息
		message := fmt.Sprintf(
			"⚠️ 工单超时提醒\n\n"+
				"工单编号: %s\n"+
				"用户: %s\n"+
				"邮箱: %s\n"+
				"标题: %s\n"+
				"内容: %s\n"+
				"提交时间: %s\n"+
				"超时时长: %.1f小时\n"+
				"⚠️ 请及时处理！",
			workOrder.Number,
			workOrder.Username,
			workOrder.Email,
			workOrder.Title,
			workOrder.Content,
			time.Unix(workOrder.CreateTime, 0).Format("2006-01-02 15:04:05"),
			float64(now.Unix()-workOrder.CreateTime)/3600, // 转换为小时
		)

		// 发送钉钉通知
		err := sendWorkOrderReminderNotification(message)
		if err != nil {
			fmt.Printf("工单 %s 钉钉通知发送失败: %v\n", workOrder.Number, err)
			continue
		}

		// 更新工单的发送状态，避免重复发送
		updateData := map[string]interface{}{
			"send_status": 1,
			"send_time":   now.Unix(),
		}
		err = models.UpdateTechnicalWorkOrder("id = ?", updateData, workOrder.Id)
		if err != nil {
			fmt.Printf("更新工单 %s 发送状态失败: %v\n", workOrder.Number, err)
		} else {
			fmt.Printf("工单 %s 超时通知发送成功\n", workOrder.Number)
		}
	}
}

// sendWorkOrderReminderNotification 发送工单提醒钉钉通知
func sendWorkOrderReminderNotification(message string) error {
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
	safeMessage := "【工单提醒】" + message

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
