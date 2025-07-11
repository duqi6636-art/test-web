package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
)

// @Summary 获取 PublicApi 信息
// @Router /center/docs/info [post]
func GetDocApiUserInfo(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	//uid := user.ID
	// IP信息
	tokenInfo := GetDcoApiToken(c, user)
	resultTokenInfo := models.ResultUserDocApiToken{}
	resultTokenInfo.Username = tokenInfo.Username
	resultTokenInfo.Email = tokenInfo.Email
	resultTokenInfo.Token = tokenInfo.Token
	resultTokenInfo.Key = tokenInfo.Key

	apiUrl := models.GetConfigVal("DOC_API_URL")
	resData := map[string]interface{}{
		"token_info": resultTokenInfo,
		"api_url":    apiUrl,
	}
	JsonReturn(c, 0, "__T_SUCCESS", resData)
	return
}

// @Summary 刷新 PublicApi 令牌信息
// @Router /center/docs/reset_key [post]
func RefreshDocApiKey(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	uid := user.Id

	_, tokenInfo := models.GetUserDocApiToken(uid)

	apiKey := util.RandStr("r", 16)
	upInfo := map[string]interface{}{
		"key":         apiKey,
		"update_time": util.GetNowInt(),
		"ip":          c.ClientIP(),
	}
	err := models.EditUserDocApiTokenById(tokenInfo.Id, upInfo)
	if err != nil {
		apiKey = tokenInfo.Key
	}
	resData := map[string]interface{}{
		"key":   apiKey,
		"token": tokenInfo.Token,
	}
	JsonReturn(c, 0, "__T_SUCCESS", resData)
	return
}

// 获取 用户 DocApi Token 信息
func GetDcoApiToken(c *gin.Context, userInfo models.Users) models.UserDocApiTokenModel {
	// 查询该用户下面是否存在相同账户 若存在则修改账户信息
	accountInfo := models.UserDocApiTokenModel{}
	_, accountInfo = models.GetUserDocApiToken(userInfo.Id)
	nowTime := util.GetNowInt()
	if accountInfo.Id == 0 { //不存在就默认添加一个
		apiKey := util.RandStr("r", 12)
		data := models.UserDocApiTokenModel{}
		data.Status = 1
		data.Uid = userInfo.Id
		data.Username = userInfo.Username
		data.Email = userInfo.Email
		data.Token = userInfo.Sn
		data.Key = apiKey
		data.Ip = c.ClientIP()
		data.CreateTime = nowTime
		err := models.AddUserDocApiToken(data)
		fmt.Println("err:", err)
		accountInfo = data
	}

	return accountInfo
}
