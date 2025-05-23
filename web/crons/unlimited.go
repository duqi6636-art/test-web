package crons

import (
	"api-360proxy/web/models"
	"api-360proxy/web/service/email"
	"fmt"
	"strings"
	"time"
)

// 不限量预警任务发送邮件

func UnlimitedEarlyWarning() {
	// 获取所有打开开关的用户
	query := "status = ?"
	args := []interface{}{1}
	uewList := models.GetUnlimitedEarlyWarningList(query, args)
	var now = time.Now()
	after24 := now.AddDate(0, 0, -1)
	for _, uew := range uewList {
		var uewDetail = models.UnlimitedEarlyWarningDetail{Uid: uew.Uid}
		defaultEmail := models.GetConfigVal("default_email")
		uewDetailList := uewDetail.GetAll()
		vars := map[string]string{
			"email":  uew.Email,
			"detail": "",
		}
		var detList = make([]string, 0)
		for _, detail := range uewDetailList {
			var det = ""
			// 过滤未打开开关的
			if detail.Status != 1 {
				continue
			}
			// 过滤24小时内发送过的
			if detail.SendTime != 0 && detail.SendTime > after24.Unix() {
				continue
			}
			query = "uid = ? and ins_id = ? and period >= ?"
			afterDuration := now.Add(-time.Duration(detail.Duration) * time.Minute).Unix()
			args = []interface{}{detail.Uid, detail.InstanceId, afterDuration}
			cvmList := models.GetUserUnlimitedCvmList(query, args)
			// 没有读取到数据不发送
			if len(cvmList) <= 0 {
				continue
			}
			var isSendMap = map[string]bool{"cpu": true, "memory": true, "bandwidth": true, "concurrency": true}
			for _, cvm := range cvmList {
				if cvm.CpuAvg < float64(detail.Cpu) {
					isSendMap["cpu"] = false
				}
				if cvm.MemAvg < float64(detail.Memory) {
					isSendMap["memory"] = false
				}
				if cvm.BandwidthAvg < float64(detail.Bandwidth) {
					isSendMap["bandwidth"] = false
				}
				configTcp := cvm.Config * 1000
				userTcp := cvm.TcpAvg * 100 / float64(configTcp)
				if userTcp < float64(detail.Concurrency) {
					isSendMap["concurrency"] = false
				}
			}

			var infoList = make([]string, 0)
			for key, val := range isSendMap {
				if val {
					switch key {
					case "cpu":
						infoList = append(infoList, "CPU utilization")
					case "memory":
						infoList = append(infoList, "memory utilization")
					case "bandwidth":
						infoList = append(infoList, "bandwidth utilization")
					case "concurrency":
						infoList = append(infoList, "concurrency utilization")
					}
					// 更新一下发送的时间
					detail.SendTime = now.Unix()
					detail.Update()
				}
			}
			if len(infoList) > 0 {
				det += fmt.Sprintf("%v (%v)", detail.Ip, strings.Join(infoList, ","))
				detList = append(detList, det)
			}
		}
		if len(detList) > 0 {
			vars["detail"] = strings.Join(detList, "\n")
			// 发送邮件
			if defaultEmail == "" {
				result := email.AwsSendEmail(uew.Email, 10, vars, "unlimited_early_warning_email")
				fmt.Println("send unlimited early warning result aws:", result)
			} else {
				result := email.TencentSendEmail(uew.Email, 10, vars, "unlimited_early_warning_email")
				fmt.Println("send unlimited early warning result aws:", result)
			}
		}

	}
}
