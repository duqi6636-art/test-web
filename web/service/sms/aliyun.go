package sms

import (
	"errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
)

func SendAliSMS(code ,phone, accessKeyId,accessKeySecret,aliTmpCode,aliSign string) (response *dysmsapi.SendSmsResponse, err error) {
	if accessKeyId=="" || accessKeySecret=="" || aliTmpCode=="" || aliSign=="" {
		err = errors.New("阿里短信配置信息不全")
		return
	}
	client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", accessKeyId, accessKeySecret)
	if err != nil {
		return
	}
	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = phone
	request.SignName = aliSign
	request.TemplateCode = aliTmpCode
	request.TemplateParam = code
	response, err = client.SendSms(request)
	/*if err!=nil{
		return
	}
	//结果判断
	if response.Code != "OK" {
		err = errors.New(response.Message)
		return
	}*/
	return
}

