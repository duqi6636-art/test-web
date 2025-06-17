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

// 释放实例
func ReleaseInstances(host, instanceId string) (bool, string) {
	secretKeyId := models.GetConfigVal("Zenlayer_SecretKeyId")             //
	secretKeyPassword := models.GetConfigVal("Zenlayer_SecretKeyPassword") //

	client, _ := zec.NewClientWithSecretKey(secretKeyId, secretKeyPassword)

	instanceIds := []string{}
	instanceIds = []string{instanceId}
	// Prepare the request
	request := zec.NewReleaseInstancesRequest()
	request.InstanceIds = instanceIds

	// Make the API call
	response, err := client.ReleaseInstances(request)
	if err != nil {
		AddLogs("zenlayer_release_sdk"+host, "Error creating ZEC instances") //写日志
		return false, "Error Api Call instances"
	}
	// Handle the response
	if response.Response.RequestId == "" {
		AddLogs("zenlayer_release_api"+host, "API request failed") //写日志
		return false, "API request failed"
	}
	bytes, _ := json.Marshal(response.Response)
	AddLogs("zenlayer_release "+host, string(bytes)) //写日志

	return true, response.Response.RequestId
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

	request.ZoneId = models.GetConfigVal("Zenlayer_ZoneId") // 实例所在节点ID
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

// 获取镜像ID 列表
func DescribeImages(c *gin.Context) {
	zoneId := c.DefaultPostForm("zoneId", "")
	secretKeyId := models.GetConfigVal("Zenlayer_SecretKeyId")             //
	secretKeyPassword := models.GetConfigVal("Zenlayer_SecretKeyPassword") //

	client, _ := zec.NewClientWithSecretKey(secretKeyId, secretKeyPassword)
	if zoneId == "" {
		zoneId = models.GetConfigVal("Zenlayer_ZoneId") // 实例所在节点ID
	}
	// Prepare the request
	request := zec.NewDescribeImagesRequest()

	request.ZoneId = zoneId
	request.Category = "Ubuntu" //镜像所属分类。 CentOS Windows Ubuntu Debian 等
	//request.ImageType = "CUSTOM_IMAGE" //镜像类型 PUBLIC_IMAGE-公共镜像。 CUSTOM_IMAGE-自定义镜像。
	response, err := client.DescribeImages(request)

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
