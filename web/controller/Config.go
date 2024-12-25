package controller

import (
	"api-360proxy/web/models"
	"github.com/gin-gonic/gin"
)

// @BasePath /api/v1
// @Summary 初始化插件配置
// @Description 初始化插件配置
// @Tags 插件配置
// @Accept x-www-form-urlencoded
// @Produce json
// @Success 0 {array}  map[string]interface{}
// @Router /init_plugin [post]
func InitPlugin(c *gin.Context) {
	data := map[string]interface{}{}
	jumpList := map[string]map[string]string{}
	err, list := models.FindAllJumpHelp("plugin")
	if err != nil || len(list) == 0 {
		JsonReturn(c, 0, "__T_SUCCESS", data)
		return
	}

	for _, v := range list {
		if jumpList[v.Cate] == nil {
			jumpList[v.Cate] = make(map[string]string)
		}
		jumpList[v.Cate][v.Code] = v.Value
	}
	data["jump_list"] = jumpList

	JsonReturn(c, 0, "__T_SUCCESS", data)
	return

}

// @BasePath /api/v1
// @Summary 初始化oem下载地址
// @Description 初始化oem下载地址
// @Tags 插件配置
// @Accept x-www-form-urlencoded
// @Produce json
//
//	@Success 0 {object}  map[string]string " "windows_down：windowsSocks5代理客户端下载地址,mac_down：macSocks5代理客户端下载地址,proxyManage_down：代理管理器下载地址,google_down：google插件下载地址,firefox_down：火狐插件下载地址"
//
// @Router /init_oem [post]
func InitOem(c *gin.Context) {
	data := map[string]interface{}{}

	//windowsSocks5代理客户端下载地址
	windowsDown, ok := models.OemVersion["normal_1"]
	if !ok {
		data["windows_down"] = ""
	} else {
		data["windows_down"] = windowsDown.DownUrl
	}
	//macSocks5代理客户端下载地址
	macDown, ok := models.OemVersion["mac_2"]
	if !ok {
		data["mac_down"] = ""
	} else {
		data["mac_down"] = macDown.DownUrl
	}
	//代理管理器下载地址
	proxyManageDown, ok := models.OemVersion["proxy_manage_6"]
	if !ok {
		data["proxyManage_down"] = ""
	} else {
		data["proxyManage_down"] = proxyManageDown.DownUrl
	}
	//google 插件 下载地址
	googleDown, ok := models.OemVersion["google_5"]
	if !ok {
		data["google_down"] = ""
	} else {
		data["google_down"] = googleDown.DownUrl
	}
	//火狐 插件 下载地址
	firefoxDown, ok := models.OemVersion["firefox_5"]
	if !ok {
		data["firefox_down"] = ""
	} else {
		data["firefox_down"] = firefoxDown.DownUrl
	}
	JsonReturn(c, 0, "__T_SUCCESS", data)
	return

}
