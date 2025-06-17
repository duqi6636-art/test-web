package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
	"time"
)

// @BasePath /api/v1
// @Summary 不限量配置使用统计
// @Schemes
// @Description 不限量配置使用统计
// @Tags 个人中心 - 不限量
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param host formData string true "不限量机器"
// @Param lang formData string false "语言"
// @Param start_time formData string false "开始时间"
// @Param end_time formData string false "结束时间"
// @Produce json
// @Success 0 {} {}
// @Router /center/monitor/stats [post]
func TencentCvmMonitor(c *gin.Context) {
	host := c.DefaultPostForm("host", "")
	if host == "" {
		JsonReturn(c, -1, "__T_CHOOSE_HOST", gin.H{})
		return
	}
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	startTime := c.DefaultPostForm("start_time", "")
	endTime := c.DefaultPostForm("end_time", "")
	timeZone := c.DefaultPostForm("timezone", "")
	if timeZone == "" {
		timeZone = "Asia/Shanghai"
	}

	startInt := 0
	endInt := 0
	nowTime := util.GetNowInt()
	locSecond := 0                //两个时区相差的秒数
	txTimezone := "Asia/Shanghai" //云服务器所在时区
	if startTime == "" {
		startInt = util.GetTodayTime()
	} else {
		startInt = int(util.GetTimeByTimezone(timeZone, startTime, txTimezone)) //获取时区转变后的时间戳
		timeA := util.StoI(util.GetTimeStamp(startTime, "Y-m-d H:i:s"))
		locSecond = timeA - startInt
	}
	if endTime == "" {
		endInt = nowTime
	} else {
		endInt = int(util.GetTimeByTimezone(timeZone, endTime, txTimezone)) //获取时区转变后的时间戳
	}
	if startInt > endInt { //置换开始时间和结束时间
		a := startInt
		startInt = endInt
		endInt = a
	}

	hostInfo := models.GetPoolFlowDayByIp(uid, host)
	if hostInfo.Id == 0 {
		JsonReturn(c, -1, "Config Info Error", gin.H{})
		return
	}

	second := endInt - startInt
	if second > 86400*30 {
		JsonReturn(c, -1, "__T_TIME_CYCLE", gin.H{})
		return
	}

	resInfo := ResultMonitorDataModel{}
	monitorList := models.GetUserUnlimitedCvmListBy(uid, host, startInt, endInt)
	cpuInfo := ResultMonitorDataDetailModel{}
	tcpInfo := ResultMonitorDataDetailModel{}
	memInfo := ResultMonitorDataDetailModel{}
	trafficInfo := ResultMonitorDataDetailModel{}
	for _, val := range monitorList {
		stamp := val.Period + locSecond
		tStr := util.GetTimeStr(stamp, "Y-m-d H:i:s")
		cpuInfo.Avg = append(cpuInfo.Avg, val.CpuAvg)
		cpuInfo.XData = append(cpuInfo.XData, tStr)

		tcpInfo.Avg = append(tcpInfo.Avg, val.TcpAvg)
		tcpInfo.XData = append(tcpInfo.XData, tStr)

		memInfo.Avg = append(memInfo.Avg, val.MemAvg)
		memInfo.XData = append(memInfo.XData, tStr)

		trafficInfo.Avg = append(trafficInfo.Avg, val.BandwidthAvg)
		trafficInfo.XData = append(trafficInfo.XData, tStr)
	}

	resInfo.Cpu = cpuInfo
	resInfo.Tcp = tcpInfo
	resInfo.Mem = memInfo
	resInfo.Traffic = trafficInfo

	JsonReturn(c, 0, "success", resInfo)
	return

}

// @BasePath /api/v1
// @Summary 获取实例实时监控信息下载
// @Schemes
// @Description 获取实例实时监控信息下载
// @Tags 个人中心 - 不限量
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param host formData string true "不限量机器"
// @Param lang formData string false "语言"
// @Param start_time formData string false "开始时间"
// @Param end_time formData string false "结束时间"
// @Produce json
// @Success 0 {} {}
// @Router /center/monitor/stats_download [post]
func TencentCvmMonitorDownload(c *gin.Context) {
	host := c.DefaultPostForm("host", "")
	if host == "" {
		JsonReturn(c, -1, "__T_CHOOSE_HOST", gin.H{})
		return
	}
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id
	startTime := c.DefaultPostForm("start_time", "")
	endTime := c.DefaultPostForm("end_time", "")
	timeZone := c.DefaultPostForm("timezone", "")
	if timeZone == "" {
		timeZone = "Asia/Shanghai"
	}

	startInt := 0
	endInt := 0
	nowTime := util.GetNowInt()
	locSecond := 0                //两个时区相差的秒数
	txTimezone := "Asia/Shanghai" //云服务器所在时区
	if startTime == "" {
		startInt = util.GetTodayTime()
	} else {
		startInt = int(util.GetTimeByTimezone(timeZone, startTime, txTimezone)) //获取时区转变后的时间戳
		timeA := util.StoI(util.GetTimeStamp(startTime, "Y-m-d H:i:s"))
		locSecond = timeA - startInt
	}
	if endTime == "" {
		endInt = nowTime
	} else {
		endInt = int(util.GetTimeByTimezone(timeZone, endTime, txTimezone)) //获取时区转变后的时间戳
	}
	if startInt > endInt { //置换开始时间和结束时间
		a := startInt
		startInt = endInt
		endInt = a
	}

	hostInfo := models.GetPoolFlowDayByIp(uid, host)
	if hostInfo.Id == 0 {
		JsonReturn(c, -1, "Config Info Error", gin.H{})
		return
	}

	second := endInt - startInt
	if second > 86400*30 {
		JsonReturn(c, -1, "__T_TIME_CYCLE", gin.H{})
		return
	}

	resInfo := ResultMonitorDataModel{}
	monitorList := models.GetUserUnlimitedCvmListBy(uid, host, startInt, endInt)
	cpuInfo := ResultMonitorDataDetailModel{}
	tcpInfo := ResultMonitorDataDetailModel{}
	memInfo := ResultMonitorDataDetailModel{}
	trafficInfo := ResultMonitorDataDetailModel{}
	for _, val := range monitorList {
		stamp := val.Period + locSecond
		tStr := util.GetTimeStr(stamp, "Y-m-d H:i:s")
		cpuInfo.Avg = append(cpuInfo.Avg, val.CpuAvg)
		cpuInfo.XData = append(cpuInfo.XData, tStr)

		tcpInfo.Avg = append(tcpInfo.Avg, val.TcpAvg)
		tcpInfo.XData = append(tcpInfo.XData, tStr)

		memInfo.Avg = append(memInfo.Avg, val.MemAvg)
		memInfo.XData = append(memInfo.XData, tStr)

		trafficInfo.Avg = append(trafficInfo.Avg, val.BandwidthAvg)
		trafficInfo.XData = append(trafficInfo.XData, tStr)
	}

	resInfo.Cpu = cpuInfo
	resInfo.Tcp = tcpInfo
	resInfo.Mem = memInfo
	resInfo.Traffic = trafficInfo

	title := []string{"Time", "Cpu", "TCP", "Mem", "Bandwidth"} //导出数据表头
	csvData := [][]string{}
	csvData = append(csvData, title)

	type MonitorKvModel struct {
		Cpu      float64 `json:"cpu"`
		Tcp      float64 `json:"tcp"`
		Mem      float64 `json:"mem"`
		Traffic  float64 `json:"traffic"`
		Datetime string  `json:"datetime"`
	}

	var output []MonitorKvModel
	for i := 0; i < len(resInfo.Cpu.XData); i++ {
		// 检查索引是否越界
		cpu := 0.0
		if i < len(resInfo.Cpu.Avg) {
			cpu = resInfo.Cpu.Avg[i]
		}
		// 检查索引是否越界
		tcp := 0.0
		if i < len(resInfo.Tcp.Avg) {
			tcp = resInfo.Tcp.Avg[i]
		}
		// 检查索引是否越界
		mem := 0.0
		if i < len(resInfo.Mem.Avg) {
			mem = resInfo.Mem.Avg[i]
		}
		// 检查索引是否越界
		traffic := 0.0
		if i < len(resInfo.Traffic.Avg) {
			traffic = resInfo.Traffic.Avg[i]
		}

		item := MonitorKvModel{
			Cpu:      cpu,
			Tcp:      tcp,
			Mem:      mem,
			Traffic:  traffic,
			Datetime: resInfo.Cpu.XData[i],
		}
		output = append(output, item)
	}

	// 组合导出数据
	for _, val := range output {
		info := []string{}
		info = append(info, val.Datetime)
		info = append(info, util.FtoS2(val.Cpu, 2))
		info = append(info, util.FtoS2(val.Tcp, 2))
		info = append(info, util.FtoS2(val.Mem, 2))
		info = append(info, util.FtoS2(val.Traffic, 3))
		csvData = append(csvData, info)
	}

	err := DownloadCsv(c, host, csvData)
	fmt.Println(err)
	return
}

// @BasePath /api/v1
// @Summary 不限量配置使用重启
// @Schemes
// @Description 不限量配置使用重启
// @Tags 个人中心 - 不限量
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param host formData string true "不限量机器"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {} {}
// @Router /center/monitor/restart [post]
func TencentCvmRestart(c *gin.Context) {
	host := c.DefaultPostForm("host", "")
	if host == "" {
		JsonReturn(c, -1, "__T_CHOOSE_HOST", gin.H{})
		return
	}
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	hostInfo := models.GetPoolFlowDayByIp(uid, host)
	if hostInfo.Id == 0 || hostInfo.InstanceId == "" || hostInfo.Region == "" {
		JsonReturn(c, -1, "Config Info Error", gin.H{})
		return
	}
	state := ""
	var res bool
	if hostInfo.Supplier == "zenlayer" {
		res, state = RebootInstances(hostInfo)
	} else {
		res, state = TencentRestartInstances(hostInfo)
	}
	if res == false {
		JsonReturn(c, -1, state, nil)
		return
	}
	JsonReturn(c, 0, "success", nil)
	return

}

// @BasePath /api/v1
// @Summary 不限量配置获取状态
// @Schemes
// @Description 不限量配置获取状态
// @Tags 个人中心 - 不限量
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登录凭证信息"
// @Param host formData string true "不限量机器"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {} {}
// @Router /center/monitor/status [post]
func TencentCvmDescribeStatus(c *gin.Context) {
	host := c.DefaultPostForm("host", "")
	if host == "" {
		JsonReturn(c, -1, "__T_CHOOSE_HOST", gin.H{})
		return
	}
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := user.Id

	hostInfo := models.GetPoolFlowDayByIp(uid, host)
	if hostInfo.Id == 0 || hostInfo.InstanceId == "" || hostInfo.Region == "" {
		JsonReturn(c, -1, "Config Info Error", gin.H{})
		return
	}
	state := ""
	var res bool
	if hostInfo.Supplier == "zenlayer" {
		res, state = DescribeInstancesStatus(hostInfo)
	} else {
		res, state = TencentDescribeInstancesStatus(hostInfo)
	}
	if res == false {
		JsonReturn(c, -1, state, nil)
		return
	}

	resInfo := map[string]interface{}{
		"host":  hostInfo.Ip,
		"state": state,
	}
	JsonReturn(c, 0, "success", resInfo)
	return

}

func GetTencentMonitorHandle(instanceId, region, metricName, start, end string, period uint64) (res bool, msg string, info ResponseTencentMonitorModel) {
	secretId := models.GetConfigVal("Tencent_SecretId")   //
	secretKey := models.GetConfigVal("Tencent_SecretKey") //
	//获取相关配置
	if period == 0 {
		period = 300
	}

	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	credential := common.NewCredential(
		secretId,
		secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "monitor.tencentcloudapi.com"
	cpf.HttpProfile.ReqMethod = "POST"
	//创建common client

	client, _ := monitor.NewClient(credential, region, cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := monitor.NewGetMonitorDataRequest()

	request.Namespace = common.StringPtr("QCE/CVM")
	request.MetricName = common.StringPtr(metricName)
	request.Period = common.Uint64Ptr(period)   // 时间粒度  默认展示  5分钟数据
	request.StartTime = common.StringPtr(start) //起始时间 Timestamp ISO8601 如2018-09-22T19:51:23+08:00
	request.EndTime = common.StringPtr(end)     //结束时间 Timestamp ISO8601 如2019-03-24T10:51:23+08:00
	request.Instances = []*monitor.Instance{
		&monitor.Instance{
			Dimensions: []*monitor.Dimension{
				&monitor.Dimension{
					Name:  common.StringPtr("InstanceId"),
					Value: common.StringPtr(instanceId), //实例ID
				},
			},
		},
	}
	request.SpecifyStatistics = common.Int64Ptr(7) //返回多种统计方式数据。avg, max, min (1,2,4)可以自由组合
	fmt.Println("request", request)

	// 返回的resp是一个GetMonitorDataResponse的实例，与请求对象对应
	response, err := client.GetMonitorData(request)
	fmt.Println("response", response)

	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		mm := fmt.Sprintf("An API TencentCloudSDKError: %s", err)
		return false, mm, info
	}
	if err != nil {
		//panic(err)
		mm := fmt.Sprintf("An API error has returned: %s", err)
		return false, mm, info
	}
	// 输出json格式的字符串回包
	resul := ResponseTencentMonitorModel{}
	_ = json.Unmarshal([]byte(response.ToJsonString()), &resul)
	AddLogs("tx_Monitor "+metricName, response.ToJsonString()) //写日志
	return true, "success", resul
}

type ResponseTencentMonitorModel struct {
	Response struct {
		Period     int    `json:"Period"`
		MetricName string `json:"MetricName"`
		DataPoints []struct {
			Dimensions []struct {
				Name  string `json:"Name"`
				Value string `json:"Value"`
			} `json:"Dimensions"`
			Timestamps []int `json:"Timestamps"`
			//Values     []interface{} `json:"Values"`
			//MaxValues  []float64     `json:"MaxValues"`
			//MinValues  []float64     `json:"MinValues"`
			AvgValues []float64 `json:"AvgValues"`
		} `json:"DataPoints"`
		StartTime time.Time `json:"StartTime"`
		EndTime   time.Time `json:"EndTime"`
		Msg       string    `json:"Msg"`
		RequestId string    `json:"RequestId"`
	} `json:"Response"`
}

type ResultMonitorDataModel struct {
	Cpu     ResultMonitorDataDetailModel `json:"cpu"`
	Tcp     ResultMonitorDataDetailModel `json:"tcp"`
	Mem     ResultMonitorDataDetailModel `json:"mem"`
	Traffic ResultMonitorDataDetailModel `json:"traffic"`
}

type ResultMonitorDataDetailModel struct {
	Avg   []float64 `json:"avg"`
	Max   []float64 `json:"max"`
	Min   []float64 `json:"min"`
	XData []string  `json:"x_data"`
}

type ResponseTencentCvmStatusModel struct {
	Response struct {
		TotalCount        int `json:"TotalCount"`
		InstanceStatusSet []struct {
			InstanceId    string `json:"InstanceId"`
			InstanceState string `json:"InstanceState"`
		} `json:"InstanceStatusSet"`
		RequestId string `json:"RequestId"`
	} `json:"Response"`
}

// 退还实例 释放实例
func ReturnTerminateCvm(c *gin.Context) {
	host := c.DefaultPostForm("host", "")
	uidStr := c.DefaultPostForm("uid", "")
	if host == "" {
		JsonReturn(c, -1, "__T_CHOOSE_HOST", gin.H{})
		return
	}
	uid := util.StoI(uidStr)
	//获取相关配置
	region := models.GetConfigVal("TencentCVM_Region") //可用区
	if region == "" {
		region = "na-ashburn"
	}
	hostInfo := models.GetPoolFlowDayByIp(uid, host)
	if hostInfo.Id == 0 || hostInfo.InstanceId == "" {
		JsonReturn(c, -1, "Config Info Error", gin.H{})
		return
	}

	msg := ""
	var res bool
	if hostInfo.Supplier == "zenlayer" {
		res, msg = ReleaseInstances(hostInfo.Ip, hostInfo.InstanceId)
	} else {
		if hostInfo.Region == "" {
			JsonReturn(c, -1, "Config Region Error", gin.H{})
			return
		}
		res, msg = TencentTerminateInstances(hostInfo.Region, hostInfo.InstanceId)
	}

	if res == false {
		JsonReturn(c, -1, msg, nil)
		return
	}
	JsonReturn(c, 0, "success", nil)
	return
}

type InstanceIdsModel struct {
	Response struct {
		InstanceIDSet []string `json:"InstanceIdSet"`
		RequestID     string   `json:"RequestId"`
	} `json:"Response"`
}

type RequestCommonModel struct {
	Response struct {
		RequestID string `json:"RequestId"`
	} `json:"Response"`
}
