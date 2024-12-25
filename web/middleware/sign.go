package middleware

import (
	"api-360proxy/web/models"
	"github.com/gin-gonic/gin"
)

// 验签名
func SignMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//var maps = make(map[string]interface{})
		//maps["_device_os"] = c.GetHeader("_device_os")
		//maps["_language"] = c.GetHeader("_language")
		//maps["_os_version"] = c.GetHeader("_os_version")
		//maps["_oem"] = c.GetHeader("_oem")
		//maps["_prj_name"] = c.GetHeader("_prj_name")
		//maps["_salt"] = c.GetHeader("_salt")
		//maps["_session"] = c.GetHeader("_session")
		//maps["_uid"] = c.GetHeader("_uid")
		//maps["_username"] = c.GetHeader("_username")
		//maps["_sn"] = c.GetHeader("_sn")
		//maps["_timestamp"] = c.GetHeader("_timestamp")
		//maps["_version"] = c.GetHeader("_version")
		//maps["_version_show"] = c.GetHeader("_version_show")
		//maps["_brand"] = c.GetHeader("_brand")
		//maps["_channel"] = c.GetHeader("_channel")
		//sign := c.GetHeader("_sign")
		//var (
		//	keyList []string
		//	arr     []string
		//)
		//for k := range maps {
		//	keyList = append(keyList, k)
		//}
		//sort.Strings(keyList)
		//// 去除空值
		//for _, k := range keyList {
		//	if maps[k] != nil && maps[k] != "" {
		//		arr = append(arr, fmt.Sprintf("%s=%s", k, maps[k]))
		//	}
		//}
		//// 校验签名
		//signData := strings.Join(arr, "&") + models.GetConfigVal("md_en_api_secret_token")
		//// 刷新缓存接口不校验签名
		//if c.Request.RequestURI == "/api/system/fresh_config_cache" {
		//	c.Next()
		//} else {
		//	if sign != utils.Md5(signData) && setting.AppConfig.VerifySignature {
		//		c.Abort()
		//		controller.JsonReturn(c, -1, "Sign error", nil)
		//		return
		//	}
		//}
		lan := GetLanguage(c)
		param := models.SignParam{}
		param.DeviceOs = c.GetHeader("_device_os")
		//param.Language = c.GetHeader("_language")
		param.Language = lan
		param.OsVersion = c.GetHeader("_os_version")
		param.Oem = c.GetHeader("_oem")
		param.Salt = c.GetHeader("_salt")
		param.Session = c.GetHeader("_session")
		param.Sn = c.GetHeader("_sn")
		param.Timestamp = c.GetHeader("_timestamp")
		param.Version = c.GetHeader("_version")
		param.VersionShow = c.GetHeader("_version_show")
		param.Brand = c.GetHeader("_brand")
		param.Channel = c.GetHeader("_channel")
		c.Set("HeaderData", param)
		c.Next()
	}
}

