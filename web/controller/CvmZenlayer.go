package controller

import (
	"api-360proxy/web/models"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"log"
	"strings"
)

// 重启实例
func RebootInstances(hostInfo models.PoolFlowDayDetailModel) (bool, string) {
	secretKeyId := models.GetConfigVal("Zenlayer_SecretKeyId")             //
	secretKeyPassword := models.GetConfigVal("Zenlayer_SecretKeyPassword") //
	instanceIds := []string{}
	instanceIds = []string{hostInfo.InstanceId}
	fmt.Println("instanceIds:", instanceIds)

	client, _ := zec.NewClientWithSecretKey(secretKeyId, secretKeyPassword)

	// Prepare the request
	request := zec.NewRebootInstancesRequest()
	request.InstanceIds = instanceIds

	// Make the API call
	response, err := client.RebootInstances(request)
	if err != nil {
		AddLogs("zenlayer_restart_sdk"+hostInfo.Ip, "Error creating ZEC instances") //写日志
		return false, "Error Api Call instances"
	}
	// Handle the response
	if response.Response.RequestId == "" {
		AddLogs("zenlayer_restart_api"+hostInfo.Ip, "Error creating ZEC instances") //写日志
		return false, "API request failed"

	}

	bytes, _ := json.Marshal(response.Response)
	AddLogs("zenlayer_restart "+hostInfo.Ip, string(bytes)) //写日志
	return true, response.Response.RequestId
}

// 获取实例状态
func DescribeInstancesStatus(hostInfo models.PoolFlowDayDetailModel) (bool, string) {
	secretKeyId := models.GetConfigVal("Zenlayer_SecretKeyId")             //
	secretKeyPassword := models.GetConfigVal("Zenlayer_SecretKeyPassword") //

	client, _ := zec.NewClientWithSecretKey(secretKeyId, secretKeyPassword)
	instanceIds := []string{}
	instanceIds = []string{hostInfo.InstanceId}
	// Prepare the request
	request := zec.NewDescribeInstancesStatusRequest()
	request.InstanceIds = instanceIds
	response, err := client.DescribeInstancesStatus(request)

	if err != nil {
		AddLogs("zenlayer_state_sdk"+hostInfo.Ip, "Error creating ZEC instances") //写日志
		return false, "Error Api Call instances"
	}
	// Handle the response
	if response.Response.RequestId == "" {
		AddLogs("zenlayer_state_api"+hostInfo.Ip, "API request failed") //写日志
		return false, "API request failed"
	}
	bytes, _ := json.Marshal(response.Response)
	AddLogs("zenlayer_state "+hostInfo.Ip, string(bytes)) //写日志
	stateInfo := response.Response.DataSet[0]
	return true, stateInfo.InstanceStatus

}

// 虚拟机实例列表
func DescribeInstances(c *gin.Context) {
	instanceIds := c.DefaultPostForm("instance_id", "")
	//publicIPs := c.DefaultPostForm("public_ip", "")
	secretKeyId := models.GetConfigVal("Zenlayer_SecretKeyId")             //
	secretKeyPassword := models.GetConfigVal("Zenlayer_SecretKeyPassword") //

	client, _ := zec.NewClientWithSecretKey(secretKeyId, secretKeyPassword)
	instanceIdArr := strings.Split(instanceIds, ",")
	//publicIpArr := strings.Split(publicIPs, ",")

	// Prepare the request
	request := zec.NewDescribeInstancesRequest()
	if instanceIds != "" && len(instanceIdArr) > 0 {
		request.InstanceIds = instanceIdArr
	}

	request.ZoneId = "na-west-1a" // 实例所在节点ID
	response, err := client.DescribeInstances(request)

	if err != nil {
		log.Fatalf("Error creating ZEC instances: %v", err)
	}
	// Handle the response
	if response.Response.RequestId == "" {
		log.Fatalf("API request failed: %#v", response.Response)
	}

	fmt.Printf("Successfully: %#v\n", response.Response)
	fmt.Printf("Successfully. Request ID: %#v\n", response.Response.RequestId)

	JsonReturn(c, 0, "__T_SUCCESS", response.Response)
	return

}

// 获取实例状态
func DescribeInstancesStatus1(c *gin.Context) {
	instanceIds := c.DefaultPostForm("instance_id", "")
	secretKeyId := models.GetConfigVal("Zenlayer_SecretKeyId")             //
	secretKeyPassword := models.GetConfigVal("Zenlayer_SecretKeyPassword") //

	client, _ := zec.NewClientWithSecretKey(secretKeyId, secretKeyPassword)
	instanceIdArr := strings.Split(instanceIds, ",")
	// Prepare the request
	request := zec.NewDescribeInstancesStatusRequest()
	request.InstanceIds = instanceIdArr
	response, err := client.DescribeInstancesStatus(request)

	if err != nil {
		log.Fatalf("Error creating ZEC instances: %v", err)
	}
	// Handle the response
	if response.Response.RequestId == "" {
		log.Fatalf("API request failed: %#v", response.Response)
	}

	fmt.Printf("Successfully: %#v\n", response.Response)
	fmt.Printf("Successfully. Request ID: %#v\n", response.Response.RequestId)

	JsonReturn(c, 0, "__T_SUCCESS", response.Response)
	return

}

type DescribeInstancesStatusModels struct {
	RequestId string `json:"requestId"`
	DataSet   []struct {
		InstanceId     string `json:"instanceId"`
		InstanceStatus string `json:"instanceStatus"`
	} `json:"dataSet"`
	TotalCount int `json:"totalCount"`
}
