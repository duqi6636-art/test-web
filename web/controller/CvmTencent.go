package controller

import (
	"api-360proxy/web/models"
	"encoding/json"
	"fmt"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

// 获取实例状态
func TencentDescribeInstancesStatus(hostInfo models.PoolFlowDayDetailModel) (bool, string) {
	secretId := models.GetConfigVal("Tencent_SecretId")   //
	secretKey := models.GetConfigVal("Tencent_SecretKey") //
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	credential := common.NewCredential(
		secretId,
		secretKey)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := cvm.NewClient(credential, hostInfo.Region, cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := cvm.NewDescribeInstancesStatusRequest()

	// 返回的resp是一个DescribeInstancesStatusResponse的实例，与请求对象对应
	response, err := client.DescribeInstancesStatus(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		mm := fmt.Sprintf("An API SDK error has returned: %s", err)
		AddLogs("tx_cvm_state_sdk"+hostInfo.Ip, mm) //写日志
		return false, "An API SDK error has returned"

	}
	if err != nil {
		mm := fmt.Sprintf("An API error has returned: %s", err)
		AddLogs("tx_cvm_state_api"+hostInfo.Ip, mm) //写日志
		return false, "An API error has returned"
	}

	// 输出json格式的字符串回包
	resul := ResponseTencentCvmStatusModel{}
	_ = json.Unmarshal([]byte(response.ToJsonString()), &resul)
	AddLogs("tx_cvm_state "+hostInfo.Ip, response.ToJsonString()) //写日志
	statusInfo := resul.Response.InstanceStatusSet[0]

	return true, statusInfo.InstanceState
}

// 重启实例
func TencentRestartInstances(hostInfo models.PoolFlowDayDetailModel) (bool, string) {
	secretId := models.GetConfigVal("Tencent_SecretId")   //
	secretKey := models.GetConfigVal("Tencent_SecretKey") //
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	credential := common.NewCredential(
		secretId,
		secretKey)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := cvm.NewClient(credential, hostInfo.Region, cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := cvm.NewRebootInstancesRequest()

	request.InstanceIds = common.StringPtrs([]string{hostInfo.InstanceId})
	request.StopType = common.StringPtr("SOFT")

	// 返回的resp是一个RebootInstancesResponse的实例，与请求对象对应
	response, err := client.RebootInstances(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		mm := fmt.Sprintf("An API SDK error has returned: %s", err)
		AddLogs("tx_cvm_restart_sdk"+hostInfo.Ip, mm) //写日志
		return false, "An API SDK error has returned"
	}
	if err != nil {
		mm := fmt.Sprintf("An API error has returned: %s", err)
		AddLogs("tx_cvm_restart_api"+hostInfo.Ip, mm) //写日志
		return false, "An API error has returned"
	}

	AddLogs("tx_cvm_restart "+hostInfo.Ip, response.ToJsonString()) //写日志
	return true, "success"
}

// 退还/释放实例
func TencentTerminateInstances(region, instanceId string) (bool, string) {

	secretId := models.GetConfigVal("Tencent_SecretId")   //
	secretKey := models.GetConfigVal("Tencent_SecretKey") //
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	credential := common.NewCredential(
		secretId,
		secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"
	cpf.HttpProfile.ReqMethod = "POST"
	//创建common client

	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := cvm.NewClient(credential, region, cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := cvm.NewTerminateInstancesRequest()

	request.InstanceIds = common.StringPtrs([]string{instanceId})
	//request.ReleasePrepaidDataDisks = common.BoolPtr(false)

	// 返回的resp是一个TerminateInstancesResponse的实例，与请求对象对应
	response, err := client.TerminateInstances(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		mm := fmt.Sprintf("An API TencentCloudSDKError: %s", err)
		AddLogs("TxCVMReturnTencentCloudSDKError-", mm) //写日志
		return false, mm
	}
	if err != nil {
		//panic(err)
		mm := fmt.Sprintf("An API error has returned: %s", err)
		AddLogs("TxCVMReturnError-", mm) //写日志
		return false, mm
	}

	AddLogs("TxCVMReturnInstance-", response.ToJsonString()) //写日志

	result := RequestCommonModel{}
	_ = json.Unmarshal([]byte(response.ToJsonString()), &result)

	return true, result.Response.RequestID
}
