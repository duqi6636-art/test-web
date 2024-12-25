package controller

import (
	"api-360proxy/pkg/ipdat"
	"api-360proxy/web/pkg/util"
	"github.com/gin-gonic/gin"
	"strings"
)

func Test(c *gin.Context) {
	vars := make(map[string]string)
	vars["email"] = "szh94220@gmail.com"
	vars["code"] = "111111"
	dealSendEmail(1, "szh94220@gmail.com", vars, "1111")
	JsonReturn(c, 0, "ok", "test")
}

func DealUuid(c *gin.Context) {
	listsInfo := map[int]string{}
	for i := 0; i < 20; i++ {
		//UUid := GetUuid()
		str := util.GetNowTimeStr()
		cdkStr := util.Md5(str)
		//cdkArr := strings.Split(cdkStr, "-")
		cdkInfo := cdkStr[0:6] + cdkStr[20:]
		listsInfo[i] = cdkStr + "__" + str + "____" + cdkInfo
	}
	JsonReturn(c, 0, "ok", listsInfo)
}

// @BasePath /api/v1
// @Summary 获取IP信息
// @Description 获取IP信息
// @Tags 检测相关
// @Accept x-www-form-urlencoded
// @Param ip formData string false "获取IP信息"
// @Produce json
// @Success 0 {array} ipdat.IpInfoStruct{}
// @Router /ip [post]
// @Router /ip [get]
func DealIp(c *gin.Context) {
	ip := strings.TrimSpace(c.DefaultPostForm("ip", ""))
	if ip == "" {
		ip = strings.TrimSpace(c.DefaultQuery("ip", ""))
	}
	if ip == "" {
		ip = c.ClientIP()
	}
	ipInfo, _ := ipdat.IPDat.GetIpInfo(ip)
	JsonReturn(c, 0, "ok", ipInfo)
}
