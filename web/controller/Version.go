package controller

import (
	"api-360proxy/web/models"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

// 版本更新
func VersionCheck(c *gin.Context)  {
	platform := c.DefaultPostForm("platform","")
	versionStr := c.DefaultPostForm("version","")

	versionStr = strings.ReplaceAll(versionStr, ".", "")

	// 转换为整数
	version,_ :=  strconv.Atoi(versionStr)    // 版本

	var updateMap = map[string]interface{}{}

	versionInfo := models.GetVersionInfo(strings.ToLower(platform))

	if versionInfo.Id == 0{
		JsonReturn(c,2,"no update 1",gin.H{})
		return
	}
	//fmt.Println(version,"====",versionInfo.Version)
	if versionInfo.Version <= version{
		JsonReturn(c,2,"no update 2",gin.H{})
		return
	}

	updateMap["web_url"] = versionInfo.Url
	updateMap["download"] = versionInfo.DownUrl
	updateMap["content"] = versionInfo.Desc
	updateMap["version"] = versionInfo.ShowVersion
	JsonReturn(c,0,"update",updateMap)
	return
}

