package routers

import (
	proxyManageController "api-360proxy/web/controller/proxy_manager"
	"github.com/gin-gonic/gin"
)

func proxyManageRouter(router *gin.Engine) {

	proxy := router.Group("/proxy_manage")
	proxy.POST("/set_config", proxyManageController.SetConfig) // 设置配置
	proxy.POST("/get_config", proxyManageController.GetConfig) // 获取配置

}
