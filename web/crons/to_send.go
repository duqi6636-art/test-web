package crons

import (
	"api-360proxy/web/models"
	emailSender "api-360proxy/web/service/email"
	"fmt"
	"time"
)

func CheckDoSending() {
	lists, err := models.UserFlowLists()

	if err != nil {
		fmt.Println("no data")
	} else {
		for _, v := range lists {
			email := v.Email

			if email != "" {
				default_mail := models.GetConfigVal("default_email")
				result := false
				//发送邮件
				vars := make(map[string]string)
				vars["email"] = email
				if default_mail == "aws_mail" {
					result = emailSender.AwsSendEmail(email, 7, vars, "limit_flow")
					fmt.Println(result)
				}
				if default_mail == "tencent_mail" {
					result = emailSender.TencentSendEmail(email, 7, vars, "limit_flow")
					fmt.Println(result)
				}
				if result {
					upinfo := map[string]interface{}{}
					upinfo["send_has"] = 1
					models.EditUserFlow(v.ID, upinfo)
					time.Sleep(1)
				} else {
					upinfo := map[string]interface{}{}
					upinfo["send_has"] = 2
					models.EditUserFlow(v.ID, upinfo)
					time.Sleep(1)
				}
			}
		}

	}
	fmt.Println("---------------")
	return
}

func CheckLongIspDoSending() {
	lists, err := models.UserLongIspSendFlowLists()

	if err != nil {
		fmt.Println("no data")
	} else {
		for _, v := range lists {
			email := v.Email

			if email != "" {
				default_mail := models.GetConfigVal("default_email")
				result := false
				//发送邮件
				vars := make(map[string]string)
				vars["email"] = email
				if default_mail == "aws_mail" {
					result = emailSender.AwsSendEmail(email, 7, vars, "long_limit_flow")
					fmt.Println(result)
				}
				if default_mail == "tencent_mail" {
					result = emailSender.TencentSendEmail(email, 7, vars, "long_limit_flow")
					fmt.Println(result)
				}
				if result {
					upinfo := map[string]interface{}{}
					upinfo["send_has"] = 1
					models.EditUserDynamicIspByUid(v.Uid, upinfo)
					time.Sleep(1)
				} else {
					upinfo := map[string]interface{}{}
					upinfo["send_has"] = 2
					models.EditUserDynamicIspByUid(v.Uid, upinfo)
					time.Sleep(1)
				}
			}
		}

	}
	fmt.Println("---------------")
	return
}
