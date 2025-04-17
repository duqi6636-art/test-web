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
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
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
// 获取实例实时监控信息
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

	start := util.GetIso8601Time(int64(startInt)) + "+08:00"
	end := util.GetIso8601Time(int64(endInt)) + "+08:00"
	//fmt.Println(start)
	//fmt.Println(end)
	//start :=  "2025-03-08T15:51:23+08:00"
	//end := "2025-03-08T20:51:23+08:00"
	hostInfo := models.GetPoolFlowDayByIp(uid, host)
	if hostInfo.Id == 0 || hostInfo.InstanceId == "" || hostInfo.Region == "" {
		JsonReturn(c, -1, "Config Info Error", gin.H{})
		return
	}
	cateStr := map[string]string{
		"cpu":     "CPUUsage",      // CPU
		"traffic": "WanOuttraffic", //外网出带宽
		"tcp":     "TcpCurrEstab",  //TCP 连接数
		"mem":     "MemUsage",      //内存使用率
	}

	second := endInt - startInt
	if second > 86400*30 {
		JsonReturn(c, -1, "__T_TIME_CYCLE", gin.H{})
		return
	}

	period := uint64(300) //时间粒度
	if second <= 3600*2 { //1小时 使用 时间粒度为 60s
		period = 60
	} else if second <= 24*3600 { //小于等于 24小时  时间粒度为 300s
		period = 300
	} else if second <= 24*3600*7 { //小于等于 7天  时间粒度为 3600s
		period = 3600
	} else { //大于 7天  时间粒度为 86400s
		period = 86400
	}

	resInfo := ResultMonitorDataModel{}
	for k, v := range cateStr {
		res, msg, monitorInfo := GetTencentMonitorHandle(hostInfo.InstanceId, hostInfo.Region, v, start, end, period)
		if res == false {
			fmt.Println(msg)
			JsonReturn(c, -1, "Data Empty", gin.H{})
			return
		}
		dataPoint := monitorInfo.Response.DataPoints[0]
		dataInfo := ResultMonitorDataDetailModel{}
		dataInfo.Avg = dataPoint.AvgValues
		//dataInfo.Max = dataPoint.MaxValues
		//dataInfo.Min = dataPoint.MinValues
		dataArr := []string{}
		for _, dv := range dataPoint.Timestamps {
			stamp := dv + locSecond
			tStr := util.GetTimeStr(stamp, "Y-m-d H:i:s")
			dataArr = append(dataArr, tStr)
		}
		dataInfo.XData = dataArr
		if k == "cpu" {
			resInfo.Cpu = dataInfo
		}
		if k == "mem" {
			resInfo.Mem = dataInfo
		}
		if k == "traffic" {
			resInfo.Traffic = dataInfo
		}
		if k == "tcp" {
			resInfo.Tcp = dataInfo
		}
	}

	JsonReturn(c, 0, "success", resInfo)
	return

}

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
// @Router /center/monitor/stats_download [post]
// 获取实例实时监控信息下载
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

	start := util.GetIso8601Time(int64(startInt)) + "+08:00"
	end := util.GetIso8601Time(int64(endInt)) + "+08:00"
	hostInfo := models.GetPoolFlowDayByIp(uid, host)
	if hostInfo.Id == 0 || hostInfo.InstanceId == "" || hostInfo.Region == "" {
		JsonReturn(c, -1, "Config Info Error", gin.H{})
		return
	}
	cateStr := map[string]string{
		"cpu":     "CPUUsage",      // CPU
		"traffic": "WanOuttraffic", //外网出带宽
		"tcp":     "TcpCurrEstab",  //TCP 连接数
		"mem":     "MemUsage",      //内存使用率
	}

	second := endInt - startInt
	if second > 86400*30 {
		JsonReturn(c, -1, "__T_TIME_CYCLE", gin.H{})
		return
	}

	period := uint64(300) //时间粒度
	if second <= 3600*2 { //1小时 使用 时间粒度为 60s
		period = 60
	} else if second <= 24*3600 { //小于等于 24小时  时间粒度为 300s
		period = 300
	} else if second <= 24*3600*7 { //小于等于 7天  时间粒度为 3600s
		period = 3600
	} else { //大于 7天  时间粒度为 86400s
		period = 86400
	}

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
	resInfo := ResultMonitorDataModel{}
	for k, v := range cateStr {
		res, msg, monitorInfo := GetTencentMonitorHandle(hostInfo.InstanceId, hostInfo.Region, v, start, end, period)
		if res == false {
			fmt.Println(msg)
			JsonReturn(c, -1, "Data Empty", gin.H{})
			return
		}
		dataPoint := monitorInfo.Response.DataPoints[0]
		dataInfo := ResultMonitorDataDetailModel{}
		dataInfo.Avg = dataPoint.AvgValues
		dataArr := []string{}
		for _, dv := range dataPoint.Timestamps {
			stamp := dv + locSecond
			tStr := util.GetTimeStr(stamp, "Y-m-d H:i:s")
			dataArr = append(dataArr, tStr)
		}
		dataInfo.XData = dataArr
		if k == "cpu" {
			resInfo.Cpu = dataInfo
		}
		if k == "mem" {
			resInfo.Mem = dataInfo
		}
		if k == "traffic" {
			resInfo.Traffic = dataInfo
		}
		if k == "tcp" {
			resInfo.Tcp = dataInfo
		}
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
// 获取实例实时监控信息
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
	secretId := models.GetConfigVal("Tencent_SecretId")   //
	secretKey := models.GetConfigVal("Tencent_SecretKey") //

	hostInfo := models.GetPoolFlowDayByIp(uid, host)
	if hostInfo.Id == 0 || hostInfo.InstanceId == "" || hostInfo.Region == "" {
		JsonReturn(c, -1, "Config Info Error", gin.H{})
		return
	}
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
		JsonReturn(c, -1, "An API SDK error has returned", nil)
		return
	}
	if err != nil {
		mm := fmt.Sprintf("An API error has returned: %s", err)
		AddLogs("tx_cvm_restart_api"+hostInfo.Ip, mm) //写日志
		JsonReturn(c, -1, "An API error has returned", nil)
		return
	}

	AddLogs("tx_cvm_restart "+hostInfo.Ip, response.ToJsonString()) //写日志
	JsonReturn(c, 0, "success", nil)
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
// @Router /center/monitor/status [post]
// 获取实例实时监控信息
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
	secretId := models.GetConfigVal("Tencent_SecretId")   //
	secretKey := models.GetConfigVal("Tencent_SecretKey") //

	hostInfo := models.GetPoolFlowDayByIp(uid, host)
	if hostInfo.Id == 0 || hostInfo.InstanceId == "" || hostInfo.Region == "" {
		JsonReturn(c, -1, "Config Info Error", gin.H{})
		return
	}
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
		JsonReturn(c, -1, "An API SDK error has returned", nil)
		return
	}
	if err != nil {
		mm := fmt.Sprintf("An API error has returned: %s", err)
		AddLogs("tx_cvm_state_api"+hostInfo.Ip, mm) //写日志
		JsonReturn(c, -1, "An API error has returned", nil)
		return
	}

	// 输出json格式的字符串回包
	resul := ResponseTencentCvmStatusModel{}
	_ = json.Unmarshal([]byte(response.ToJsonString()), &resul)
	AddLogs("tx_cvm_state "+hostInfo.Ip, response.ToJsonString()) //写日志

	statusInfo := resul.Response.InstanceStatusSet[0]
	resInfo := map[string]interface{}{
		"host":  hostInfo.Ip,
		"state": statusInfo.InstanceState,
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
