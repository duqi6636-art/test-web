package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

// 点击统计
// @BasePath /api/v1
// @Summary 点击统计
// @Description 点击统计
// @Tags 上报接口
// @Accept x-www-form-urlencoded
// @Param code formData string false "标识"
// @Param session formData string false "session"
// @Param param formData string false "参数"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/stats/active [post]
func StatsActiveClick(c *gin.Context) {
	code := strings.TrimSpace(c.DefaultPostForm("code", ""))
	session := strings.TrimSpace(c.DefaultPostForm("session", ""))
	param := strings.TrimSpace(c.DefaultPostForm("param", ""))
	lang := strings.ToLower(c.DefaultPostForm("lang", "en"))

	uid := 0
	if session != "" {
		var res = false
		res, uid = GetUIDbySession(session)
		fmt.Println(res)
	}
	ip := c.ClientIP()
	nowTime := util.GetNowInt()

	var res error
	info := models.StatsClick{}
	info.Uid = uid
	info.Code = code
	info.Param = param
	info.Language = lang
	info.Platform = "web"
	info.Ip = ip
	info.CreateTime = nowTime
	info.Today = util.GetTodayTime()
	res = models.CreateStatsClick(info)

	JsonReturn(c, 0, "__T_SUCCESS", res)
	return
}

// @BasePath /api/v1
// @Summary 用户案例
// @Description 用户案例
// @Tags 个人中心
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param case formData string false "案例"
// @Param other formData string false "其他"
// @Produce json
// @Success 0 {array} error
// @Router /web/user/up_case [post]
func UserCase(c *gin.Context) {
	cases := strings.TrimSpace(c.DefaultPostForm("case", ""))
	other := strings.TrimSpace(c.DefaultPostForm("other", ""))

	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	caseArr := strings.Split(cases, ",")
	if len(caseArr) == 0 && other == "" {
		JsonReturn(c, e.ERROR, "__T_CASE_TEXT_INFO", nil)
		return
	}
	if len(caseArr) > 3 {
		JsonReturn(c, e.ERROR, "__T_CASE_LIMIT", nil)
		return
	}

	ip := c.ClientIP()
	nowTime := util.GetNowInt()
	var res error
	info := models.UserCaseModel{}
	info.Uid = user.Id
	info.Username = user.Username
	info.Email = user.Email
	info.Case = cases
	info.Other = other
	info.Ip = ip
	info.CreateTime = nowTime
	res = models.AddUserCase(info)

	JsonReturn(c, 0, "__T_SUCCESS", res)
	return
}

func PackStatus(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	fmt.Println(user.Id)
}
