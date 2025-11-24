package models

import (
	"api-360proxy/web/pkg/util"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type DingTalkRobot struct {
	Webhook string
	Secret  string // 可选：若使用安全加签
}

type RobotMessage struct {
	MsgType  string       `json:"msgtype"`
	Text     *TextMsg     `json:"text,omitempty"`
	Markdown *MarkdownMsg `json:"markdown,omitempty"`
	At       *AtConfig    `json:"at,omitempty"`
}

type TextMsg struct {
	Content string `json:"content"`
}

type MarkdownMsg struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type AtConfig struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	IsAtAll   bool     `json:"isAtAll,omitempty"`
}

// 日志记录

func AddLogs(code, data string) {
	AddLog(LogModel{
		Code:       code,
		Text:       data,
		CreateTime: util.GetTimeStr(util.GetNowInt(), "Y-m-d H:i:s"),
	})
}

// 发送消息
func (d *DingTalkRobot) sendMessage(msg RobotMessage) error {
	finalURL := d.Webhook

	// 若启用加签
	if d.Secret != "" {
		ts := time.Now().UnixMilli()
		sign := d.generateSign(ts)
		finalURL += fmt.Sprintf("&timestamp=%d&sign=%s", ts, url.QueryEscape(sign))
	}

	data, _ := json.Marshal(msg)

	client := http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("POST", finalURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP状态错误 %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}

	if json.Unmarshal(body, &result) != nil {
		return fmt.Errorf("解析钉钉响应失败: %s", string(body))
	}

	if result.Errcode != 0 {
		return fmt.Errorf("钉钉返回错误: %s (code: %d)", result.Errmsg, result.Errcode)
	}

	return nil
}

// 生成签名
func (d *DingTalkRobot) generateSign(timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, d.Secret)
	h := hmac.New(sha256.New, []byte(d.Secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// 构建 @ 的文本
func buildAtContent(content string, mobiles []string) string {
	if len(mobiles) == 0 {
		return content
	}
	return content + " @" + strings.Join(mobiles, " @")
}

// 发文本（可 @ 多人）

func (d *DingTalkRobot) SendText(content string, atMobiles []string) error {
	msg := RobotMessage{
		MsgType: "text",
		Text:    &TextMsg{Content: buildAtContent(content, atMobiles)},
		At:      &AtConfig{AtMobiles: atMobiles},
	}
	return d.sendMessage(msg)
}

// @ 所有人

func (d *DingTalkRobot) SendTextAtAll(content string) error {
	msg := RobotMessage{
		MsgType: "text",
		Text:    &TextMsg{Content: content},
		At:      &AtConfig{IsAtAll: true},
	}
	return d.sendMessage(msg)
}

// Markdown 消息

func (d *DingTalkRobot) SendMarkdown(title, text string, atMobiles []string, isAtAll bool) error {
	msg := RobotMessage{
		MsgType:  "markdown",
		Markdown: &MarkdownMsg{Title: title, Text: text},
		At:       &AtConfig{AtMobiles: atMobiles, IsAtAll: isAtAll},
	}
	return d.sendMessage(msg)
}

type DingSendOptions struct {
	AtMobiles   []string `json:"at_mobiles"`
	AtMode      string   `json:"at_mode"`
	MessageType string   `json:"message_type"`
	CallPhone   string   `json:"call_phone"`
	CallApiURL  string   `json:"call_api_url"`
	CallProduct string   `json:"call_product"`
	CallName    string   `json:"call_name"`
	Secret      string   `json:"secret"`
}

func TriggerPhoneCall(ctx context.Context, apiURL, phone, p, n string) error {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s?phone=%s&p=%s&n=%s", apiURL, phone, p, n), nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func SendProductAlertWithRule(ruleKey string, runtime map[string]any, fallbackTpl string) {
	rule, ruleErr := GetAlertRule(ruleKey)
	if ruleErr == nil && rule.ID > 0 && strings.TrimSpace(rule.WebhookURL) != "" {
		msg, renderErr := RenderMessage(strings.TrimSpace(rule.Context), runtime)
		if renderErr != nil || strings.TrimSpace(msg) == "" {
			AddLogs("SendProductAlertWithRule RenderMessage", fmt.Sprintf("render fail: %v", renderErr))
			msg = fallbackTpl
		}
		var opts DingSendOptions
		if strings.TrimSpace(rule.DParams) != "" {
			_ = json.Unmarshal([]byte(rule.DParams), &opts)
		}
		dingTalk := DingTalkRobot{
			Webhook: rule.WebhookURL,
			Secret:  strings.TrimSpace(opts.Secret),
		}
		atMode := strings.ToLower(strings.TrimSpace(opts.AtMode))
		effAtMobiles := []string{}
		if len(opts.AtMobiles) > 0 {
			effAtMobiles = opts.AtMobiles
		}
		effIsAtAll := false
		if atMode != "" {
			switch atMode {
			case "none":
				effAtMobiles = []string{}
				effIsAtAll = false
			case "all":
				effAtMobiles = []string{}
				effIsAtAll = true
			case "mobiles":
				effIsAtAll = false
			}
		}

		sendOne := func(content string) error {
			if effIsAtAll {
				return dingTalk.SendTextAtAll(content)
			}
			return dingTalk.SendText(content, effAtMobiles)
		}

		sendErr := sendOne(msg)
		if sendErr != nil {
			AddLogs("SendProductAlertWithRule "+ruleKey, sendErr.Error())
		}
		if strings.TrimSpace(opts.CallPhone) != "" && strings.TrimSpace(opts.CallApiURL) != "" {
			p := opts.CallProduct
			if strings.TrimSpace(p) == "" {
				p = rule.Name
			}
			n := opts.CallName
			if strings.TrimSpace(n) == "" {
				n = rule.Name
			}

			go func(callApi, phone, prod, name string) {
				defer func() {
					if r := recover(); r != nil {
						AddLogs("TriggerPhoneCall panic", fmt.Sprintf("panic: %v", r))
					}
				}()
				ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
				defer cancel()
				if err := TriggerPhoneCall(ctx, callApi, phone, prod, name); err != nil {
					AddLogs("TriggerPhoneCall err", err.Error())
				}
			}(opts.CallApiURL, opts.CallPhone, p, n)
		}
	}
}

// 将 rule.Context 作为模板

func RenderMessage(contextTpl string, runtimeVars map[string]interface{}) (string, error) {
	data := map[string]interface{}{}
	if runtimeVars != nil {
		for k, v := range runtimeVars {
			data[k] = v
		}
	}
	tplText := strings.TrimSpace(contextTpl)
	tpl, err := template.New("alert").Option("missingkey=zero").Parse(tplText)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
