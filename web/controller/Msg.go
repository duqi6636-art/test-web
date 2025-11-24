package controller

import (
	"api-360proxy/pkg/ipdat"
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

// @BasePath /api/v1
// @Summary 支付反馈
// @Description 支付反馈
// @Tags 公告/反馈
// @Accept x-www-form-urlencoded
// @Param cate formData string false "分类"
// @Param lang formData string false "语言"
// @Param email formData string true "邮箱"
// @Param content formData string true "内容"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /web/msg/feedback [post]
func FeedbackWeb(c *gin.Context) {
	params := GetParams(c)
	ip := c.ClientIP()
	cate := c.DefaultPostForm("cate", "")
	lang := c.DefaultPostForm("lang", "")
	email := c.DefaultPostForm("email", "")
	content := c.DefaultPostForm("content", "")
	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", nil)
		return
	}
	if !util.CheckEmail(email) {
		JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", gin.H{})
		return
	}
	if cate == "" && content == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}

	var (
		uid = 0
	)
	if params.Session != "" {
		_, uid = GetUIDbySession(params.Session)
	}

	nowTime := util.GetNowInt()
	ipInfo, _ := ipdat.IPDat.GetIpInfo(ip)
	fb := models.FeedbackWeb{
		Cate:       cate,
		Content:    content,
		Platform:   "web",
		Email:      email,
		Uid:        uid,
		CreateTime: nowTime,
		Lang:       lang,
		Ip:         ip,
		Country:    ipInfo.Country,
	}
	err := models.AddFeedBackWeb(fb)
	if err != nil {
		JsonReturn(c, -1, "__T_FAIL", nil)
		return
	}
	JsonReturn(c, 0, "__T_SUCCESS", nil)
	return

}

// @BasePath /api/v1
// @Summary 申请协助反馈
// @Description 申请协助反馈
// @Tags 公告/反馈
// @Accept x-www-form-urlencoded
// @Param cate formData string true "类型"
// @Param lang formData string false "语言"
// @Param image formData string false "图片"
// @Param email formData string true "联系方式"
// @Param content formData string true "问题描述"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /web/msg/feedback_assistance [post]
func FeedbackAssistance(c *gin.Context) {

	params := GetParams(c)
	ip := c.ClientIP()
	feedbackType, err := strconv.Atoi(c.DefaultPostForm("type", "")) // 反馈类型 0-普通 1-协助

	if err != nil {

		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}
	questionType := c.DefaultPostForm("questionType", "") // 问题类别
	description := c.DefaultPostForm("description", "")   // 问题描述
	occupation := c.DefaultPostForm("occupation", "")     // 职业
	lang := c.DefaultPostForm("lang", "")                 // 语言
	contact := c.DefaultPostForm("contact", "")           // 联系方式
	pictures := c.DefaultPostForm("pictures", "")

	if questionType == "" {
		JsonReturn(c, -1, "__T_QUESTIONTYPE_ERROR", nil)
		return
	}
	if description == "" {
		JsonReturn(c, -1, "__T_DESCRIPTION_ERROR", nil)
		return
	}

	if feedbackType == 1 {

		if occupation == "" {
			JsonReturn(c, -1, "__T_OCCUPATION_ERROR", nil)
			return
		}
		if pictures == "" {
			JsonReturn(c, -1, "__T_PICTURES_ERROR", nil)
			return
		}
	}
	if contact == "" {
		JsonReturn(c, e.ERROR, "__T_CONTACT_ERROR", nil)
		return
	}

	var (
		uid       = 0
		userEmail = ""
		nickname  = ""
	)
	if params.Session != "" {
		_, uid = GetUIDbySession(params.Session)
		_, user := models.GetUserById(uid)
		userEmail = user.Email
		nickname = user.Nickname
	}

	cate := ""
	switch questionType {
	case "usage":
		cate = "產品使用問題"
	case "account":
		cate = "帳戶問題"
	case "payment":
		cate = "付款問題"
	case "register":
		cate = "註冊問題"
	case "assist":
		cate = "協助調查"
	case "other":
		cate = "其他"
	default:
		cate = questionType
	}

	nowTime := util.GetNowInt()
	ipInfo, _ := ipdat.IPDat.GetIpInfo(ip)
	fb := models.FeedbackWeb{
		Cate:       cate,
		Type:       feedbackType,
		Content:    description,
		Platform:   "web",
		Email:      contact,
		Img:        pictures,
		Uid:        uid,
		CreateTime: nowTime,
		Occupation: occupation,
		Lang:       lang,
		Ip:         ip,
		Country:    ipInfo.Country,
		Nickname:   nickname,
		UserEmail:  userEmail,
	}
	err = models.AddFeedBackWeb(fb)
	if err != nil {
		JsonReturn(c, -1, "__T_FAIL", nil)
		return
	}
	JsonReturn(c, 0, "__T_SUCCESS", nil)
	return
}

// @Summary 代理管理器反馈
// @Description 代理管理器
// @Tags 公告/反馈
// @Accept x-www-form-urlencoded
// @Param cate formData string false "分类"
// @Param lang formData string false "语言"
// @Param email formData string true "邮箱"
// @Param content formData string true "内容"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /mp/feedback [post]
func FeedbackAgentManager(c *gin.Context) {
	params := GetParams(c)
	ip := c.ClientIP()
	cate := c.DefaultPostForm("cate", "")
	lang := c.DefaultPostForm("lang", "")
	email := c.DefaultPostForm("email", "")
	content := c.DefaultPostForm("content", "")
	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", nil)
		return
	}
	if !util.CheckEmail(email) {
		JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", gin.H{})
		return
	}
	if cate == "" || content == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
		return
	}

	var (
		uid = 0
	)
	if params.Session != "" {
		_, uid = GetUIDbySession(params.Session)
	}

	nowTime := util.GetNowInt()
	ipInfo, _ := ipdat.IPDat.GetIpInfo(ip)
	fb := models.FeedbackWeb{
		Cate:       cate,
		Content:    content,
		Platform:   "agent",
		Email:      email,
		Uid:        uid,
		CreateTime: nowTime,
		Lang:       lang,
		Ip:         ip,
		Country:    ipInfo.Country,
	}
	err := models.AddFeedBackWeb(fb)
	if err != nil {
		JsonReturn(c, -1, "__T_FAIL", nil)
		return
	}
	JsonReturn(c, 0, "__T_SUCCESS", nil)
	return

}

// @BasePath /api/v1
// @Summary 不限量套餐反馈
// @Description 不限量套餐反馈
// @Tags 公告/反馈
// @Accept x-www-form-urlencoded
// @Param cate formData string false "分类"
// @Param lang formData string false "语言"
// @Param email formData string true "邮箱"
// @Param content formData string true "内容"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /web/msg/feedback_unlimited [post]
func UnlimitedFeedback(c *gin.Context) {
	params := GetParams(c)
	ip := c.ClientIP()
	lang := c.DefaultPostForm("lang", "")
	email := c.DefaultPostForm("email", "")
	content := c.DefaultPostForm("content", "")
	config := c.DefaultPostForm("config", "")
	bandwidth := c.DefaultPostForm("bandwidth", "")
	cate := c.DefaultPostForm("cate", "")
	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", nil)
		return
	}
	//if !util.CheckEmail(email) {
	//	JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", gin.H{})
	//	return
	//}

	if cate == "" {
		cate = "不限量套餐定制"
	}
	if cate == "不限量套餐定制" {
		if config == "" || bandwidth == "" {
			JsonReturn(c, -1, "__T_UNLIMITED_CONFIG_EMPTY", nil)
			return
		}
	} else {
		if config == "" {
			JsonReturn(c, -1, "__T_UNLIMITED_CONFIG_EMPTY", nil)
			return
		}
	}
	uid := 0
	if params.Session != "" {
		_, uid = GetUIDbySession(params.Session)
	}

	nowTime := util.GetNowInt()
	ipInfo, _ := ipdat.IPDat.GetIpInfo(ip)
	fb := models.FeedbackWeb{
		Cate:       cate,
		Content:    content,
		Platform:   "web",
		Email:      email,
		Config:     config,
		Bandwidth:  bandwidth,
		Uid:        uid,
		CreateTime: nowTime,
		Lang:       lang,
		Ip:         ip,
		Country:    ipInfo.Country,
	}
	err := models.AddFeedBackWeb(fb)
	if err != nil {
		JsonReturn(c, -1, "__T_FAIL", nil)
		return
	}
	// 预警：接口被请求时即触发（不限量定制：支持规则模板），包含【产品】、用户、并发、带宽
	userStr := email
	if uid > 0 {
		if err, u := models.GetUserById(uid); err == nil && u.Id > 0 && strings.TrimSpace(u.Username) != "" {
			userStr = u.Username
		}
	}

	runtime := map[string]any{
		"username":    userStr,
		"concurrency": config,
		"bandwidth":   bandwidth,
	}

	fallback := fmt.Sprintf("预警：【cherry】用户【%s】并发【%s】带宽【%s】", userStr, config, bandwidth)
	models.SendProductAlertWithRule("feedback_flow", runtime, fallback)

	JsonReturn(c, 0, "__T_SUCCESS", nil)
	return
}
