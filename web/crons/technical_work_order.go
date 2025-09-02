package crons

import (
	"api-360proxy/web/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

// TechnicalWorkOrderSendDingTalk 发送钉钉提醒
func TechnicalWorkOrderSendDingTalk() {
	titleMap1 := map[string]string{"1": "产品购买&续费", "2": "了解产品", "3": "其他"}
	titleMap2 := map[string]string{"1": "产品使用问题", "2": "付款问题", "3": "其他"}
	// 处理工单发送到钉钉
	var now = time.Now()
	list := models.GetTechnicalWorkOrderList("send_status = 0 AND status = 0")
	for _, val := range list {
		var title = ""
		if v, ok := titleMap1[val.Title]; ok {
			title = v
		}
		msg := fmt.Sprintf("售前工单预警：\n 用户名：%v \n 问题分类：%v \n 详情描述：%v \n 反馈时间：%v \n", val.Username, title, val.Content, time.Unix(val.CreateTime, 0).Format("2006-01-02 15:04"))
		if val.Type == 2 {
			if v, ok := titleMap2[val.Title]; ok {
				title = v
			}
			msg = fmt.Sprintf("售后工单预警：\n 用户名：%v \n 工单号：%v \n 问题分类：%v \n 详情描述：%v \n 反馈时间：%v \n", val.Username, val.Number, title, val.Content, time.Unix(val.CreateTime, 0).Format("2006-01-02 15:04"))
		}

		err := sendDingTalk(msg)
		if err != nil {
			log.Println("Cron Technical Work Order Send Ding Talk", err, models.GetConfigVal("Technical_Work_Order_DingTalk_WebHook"))
			continue
		}
		err = models.UpdateTechnicalWorkOrder("id = ?", map[string]interface{}{"send_status": 1, "send_time": now.Unix()}, val.Id)
		if err != nil {
			log.Println("Update Cron Technical Work Order Send Ding Talk Err:", err)
			continue
		}
	}
	// 处理工单已经发送过一次但是未处理的情况
	/*pre18h := now.Add(time.Hour * -18)
	list = models.GetTechnicalWorkOrderList("send_status = 1 AND type = 2 AND status = 0 AND send_time <= ?", pre18h.Unix())
	for _, val := range list {
		title := "售后工单预警"
		msg := fmt.Sprintf("售后工单预警：\n 用户名：%v \n 工单号：%v \n 问题分类：%v \n 详情描述：%v \n 反馈时间：%v \n", val.Username, val.Number, val.Title, val.Content, time.Unix(val.CreateTime, 0).Format("2006-01-02 15:04"))
		err := sendDingTalk(msg, title)
		if err != nil {
			log.Println("Cron Technical Work Order Send Ding Talk", err)
			continue
		}
		err = models.UpdateTechnicalWorkOrder("id = ?", map[string]interface{}{"send_time": now.Unix()}, val.Id)
		if err != nil {
			log.Println("Update Cron Technical Work Order Send Ding Talk Err:", err)
			continue
		}
	}*/
	// 处理用户超过48小时没有主动关闭工单
	pre48h := now.Add(time.Hour * -48)
	list = models.GetTechnicalWorkOrderList("status = 1 AND type = 2 AND last_time <= ?", pre48h.Unix())
	for _, val := range list {
		_ = models.UpdateTechnicalWorkOrder("id = ?", map[string]interface{}{"status": 3, "deal_status": 1, "finish_time": now.Unix()}, val.Id)
	}

}

func sendDingTalk(m string) error {
	// 参数配置
	//webhook := models.GetConfigVal("Technical_Work_Order_DingTalk_WebHook")
	webhook := "https://oapi.dingtalk.com/robot/send?access_token=c726d8d031f7b950e0f1725e78cdc8793a39079af49170829636f87b8bf5f8b4"
	isAtAll := false
	phoneArr := []string{}
	// 构造消息体（文本类型）
	textArr := map[string]interface{}{
		"content": m,
	}
	atArr := map[string]interface{}{
		"atMobiles": phoneArr, //被@人的手机号（在content里添加@人的手机号）
		"isAtAll":   isAtAll,  //是否@所有人
	}
	msg := map[string]interface{}{
		"msgtype": "text",
		"text":    textArr,
		"at":      atArr,
	}
	jsonData, _ := json.Marshal(msg)
	req, err := http.NewRequest("POST", webhook, bytes.NewBuffer(jsonData))
	req.Header.Add("content-type", "application/json;charset=utf-8")
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["errcode"] != float64(0) {
		return errors.New("send ding talk fail" + fmt.Sprintf("%v", result))
	}
	return nil
}
