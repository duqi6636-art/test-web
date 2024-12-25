package middleware

import (
	"api-360proxy/pkg/ipdat"
	"api-360proxy/web/controller"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"github.com/gin-gonic/gin"
	"strings"
)

func AreaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIp := c.ClientIP()
		ipInfo ,_:= ipdat.IPDat.GetIpInfo(clientIp)
		var whiteIpRes = false
		whiteIpInfo := models.GetWhiteIpInfoMap(map[string]interface{}{"ip": clientIp})
		if whiteIpInfo.Id != 0 {
			whiteIpRes = true
		}

		if whiteIpRes == false && ipInfo.Country == "中国"{
			if ipInfo.City == ""{
				c.Abort()
				controller.JsonReturn(c, 1000, "Due to policy, this service is not avaiable in mainland China.", nil)
			}
			if ipInfo.City == "香港"  || ipInfo.City == "台湾" || ipInfo.City == "澳门"{

			}else{
				nowTime := util.GetNowInt()
				models.CreateLimitIp(models.LimitIp{
					Ip				:clientIp,
					CreateTime		:nowTime,
					CreateTimeShow	:util.GetTimeStr(nowTime, "Y-m-d H:i:s"),
				})
				c.Abort()
				controller.JsonReturn(c, 1000, "Due to policy, this service is not avaiable in mainland China.", nil)
			}
		}

		c.Next()
		return
	}
}

// 获取语言
func GetLanguage(c *gin.Context) string {
	lan := ""
	var name string = c.FullPath()
	if strings.Contains(name, "/kr/") == true {
		lan = "/kr"
	} else if strings.Contains(name, "/vn/") == true {
		lan = "/vn"
	} else if strings.Contains(name, "/ru/") == true {
		lan = "/ru"
	} else if strings.Contains(name, "/id/") == true {
		lan = "/id"
	} else if strings.Contains(name, "/in/") == true {
		lan = "/in"
	} else if strings.Contains(name, "/bd/") == true {
		lan = "/bd"
	} else if strings.Contains(name, "/my/") == true {
		lan = "/my"
	} else if strings.Contains(name, "/ar/") == true {
		lan = "/ar"
	} else if strings.Contains(name, "/hk/") == true {
		lan = "/hk"
	} else if strings.Contains(name, "/pt/") == true {
		lan = "/pt"
	} else if strings.Contains(name, "/tr/") == true {
		lan = "/tr"
	}
	return lan
}

