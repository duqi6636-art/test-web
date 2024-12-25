package controller

import (
	"api-360proxy/pkg/ipdat"
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

// 获取积分信息
// @BasePath /api/v1
// @Summary 获取积分信息
// @Description 获取积分信息
// @Tags 个人中心 - 积分
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} map[string]interface{} "country：国家，all_score：获取的总积分，use_score：已使用积分，score：剩余积分，expire：过期时间，feedback：反馈弹窗的状态 0无反馈 1待处理 2成功  3失败未填写反馈 4 失败已填写反馈"
// @Router /web/score/info [post]
func GetUserScore(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id
	userScore := models.GetUserScoreInfo(uid)
	score, allScore, useScore := 0, 0, 0
	expire := ""
	if userScore.Id > 0 {
		score = userScore.Score
		allScore = userScore.AllScore
		useScore = allScore - score
		if useScore < 0 {
			useScore = 0
		}
		if userScore.ExpireTime > 0 {
			expire = util.GetTimeStr(userScore.ExpireTime, "d-m-Y")
		}
	}
	hasInfo := models.GetScoreFeedback(uid)
	canFeedback := 0
	if hasInfo.ID > 0 {
		if hasInfo.Status == 0 {
			canFeedback = 1
		} else if hasInfo.Status == 1 {
			canFeedback = 2
		} else {
			if hasInfo.Content == "" {
				canFeedback = 3
			} else {
				canFeedback = 4
			}
		}
	}
	ip := c.ClientIP()
	ipInfo, _ := ipdat.IPDat.GetIpInfo(ip)
	resInfo := map[string]interface{}{}
	resInfo["country"] = ipInfo.CountryCode
	resInfo["all_score"] = allScore
	resInfo["use_score"] = useScore
	resInfo["score"] = score
	resInfo["expire"] = expire
	resInfo["feedback"] = canFeedback //反馈弹窗的状态 0无反馈 1待处理 2成功  3失败未填写反馈 4 失败已填写反馈
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resInfo)
	return
}

// 积分兑换流量
// @BasePath /api/v1
// @Summary 积分兑换流量
// @Description 积分兑换流量
// @Tags 个人中心 - 积分
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param flow formData string true "兑换流量数量"
// @Produce json
// @Success 0 {object} map[string]interface{} "all_score：总积分，use_score：已使用积分，score：兑换后剩余积分"
// @Router /web/score/exchange [post]
func ExScoreFlow(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id
	userScore := models.GetUserScoreInfo(uid)
	score := 0
	if userScore.Id > 0 {
		score = userScore.Score
	}
	if score <= 0 {
		JsonReturn(c, e.ERROR, "__T_SCORE_NO_ENOUGH", nil)
		return
	}
	flowStr := strings.TrimSpace(c.DefaultPostForm("flow", "")) // 流量限制
	if flowStr == "" {
		JsonReturn(c, e.ERROR, "__T_FLOW_NUMBER", nil)
		return
	}
	flowInt := util.StoI(flowStr)
	scoreFlow := models.GetConfScoreByFlow(flowInt)
	if scoreFlow.Id == 0 {
		JsonReturn(c, e.ERROR, "__T_FLOW_NUMBER", nil)
		return
	}
	if userScore.Score < scoreFlow.Score {
		JsonReturn(c, e.ERROR, "__T_SCORE_NO_ENOUGH", nil)
		return
	}

	nowTime := util.GetNowInt()
	// 异步处理生成 cdk   ----start
	info := models.PushScoreFlow{}
	info.Uid = uid
	info.Name = scoreFlow.Name
	info.Score = scoreFlow.Score
	info.Flow = scoreFlow.Flow
	info.Ip = c.ClientIP()
	info.CreateTime = nowTime
	listStr, _ := json.Marshal(info)
	resP := models.RedisLPUSH("list_score", string(listStr))
	fmt.Println(resP)
	// 异步处理生成 cdk   ----end
	scoreNew := 0
	scoreNew = userScore.Score - scoreFlow.Score

	resInfo := map[string]interface{}{}
	resInfo["all_score"] = userScore.AllScore
	resInfo["use_score"] = userScore.AllScore - scoreNew
	resInfo["score"] = scoreNew
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resInfo)
	return
}

// 积分兑换不限量流量
// @BasePath /api/v1
// @Summary 积分兑换不限量流量
// @Description 积分兑换不限量流量
// @Tags 个人中心 - 积分
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param flow formData string true "兑换流量天数"
// @Produce json
// @Success 0 {object} map[string]interface{} "all_score：总积分，use_score：已使用积分，score：兑换后剩余积分"
// @Router /web/score/exchange_flow_day [post]
func ExScoreFlowDay(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id
	userScore := models.GetUserScoreInfo(uid)
	score := 0
	if userScore.Id > 0 {
		score = userScore.Score
	}
	if score <= 0 {
		JsonReturn(c, e.ERROR, "__T_SCORE_NO_ENOUGH", nil)
		return
	}
	dayStr := strings.TrimSpace(c.DefaultPostForm("day", "")) // 要兑换的天数
	if dayStr == "" {
		JsonReturn(c, e.ERROR, "__T_FLOW_NUMBER", nil)
		return
	}
	dayInt := util.StoI(dayStr)
	scoreFlow := models.GetConfScoreByDay(dayInt)
	if scoreFlow.Id == 0 {
		JsonReturn(c, e.ERROR, "__T_FLOW_NUMBER", nil)
		return
	}
	if userScore.Score < scoreFlow.Score {
		JsonReturn(c, e.ERROR, "__T_SCORE_NO_ENOUGH", nil)
		return
	}

	nowTime := util.GetNowInt()
	// 异步处理生成 cdk   ----start
	info := models.PushScoreFlowDay{}
	info.Uid = uid
	info.Name = scoreFlow.Name
	info.Score = scoreFlow.Score
	info.Day = scoreFlow.Day
	info.Ip = c.ClientIP()
	info.CreateTime = nowTime
	listStr, _ := json.Marshal(info)
	resP := models.RedisLPUSH("list_score_flow_day", string(listStr))
	fmt.Println(resP)
	// 异步处理生成 cdk   ----end
	scoreNew := 0
	scoreNew = userScore.Score - scoreFlow.Score

	resInfo := map[string]interface{}{}
	resInfo["all_score"] = userScore.AllScore
	resInfo["use_score"] = userScore.AllScore - scoreNew
	resInfo["score"] = scoreNew
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resInfo)
	return
}


// 获取积分记录
// @BasePath /api/v1
// @Summary 获取积分记录
// @Description 获取积分记录
// @Tags 个人中心 - 积分
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Produce json
// @Success 0 {object} map[string]interface{} "score：积分记录（值为[]ResLogUserScore{}模型），use：使用记录（值为[]ResLogUserScore{}模型）"
// @Router /web/score/record [post]
func ScoreRecord(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id

	lists := models.GetLogUserScore(uid)
	scoreList := []ResLogUserScore{}
	useList := []ResLogUserScore{}
	for _, v := range lists {
		info := ResLogUserScore{}
		info.Id = v.Id
		info.Uid = v.Uid
		info.Name = v.Name
		info.Money = util.FtoS2(v.Money, 2)
		info.Score = v.Score
		info.CreateTime = util.GetTimeStr(v.CreateTime, "d/m/Y")
		if v.Mark > 0 {
			scoreList = append(scoreList, info)
		} else {
			useList = append(useList, info)
		}
	}
	resInfo := map[string]interface{}{}
	resInfo["score"] = scoreList
	resInfo["use"] = useList
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", resInfo)
	return

}

// 免费获取积分上报信息
// @BasePath /api/v1
// @Summary 免费获取积分上报信息
// @Description 免费获取积分上报信息
// @Tags 个人中心 - 积分
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param lang formData string true "语言"
// @Param email formData string true "邮箱"
// @Param phone formData string true "手机号"
// @Param company formData string true "公司名称"
// @Param address formData string false "公司地址"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /web/score/free [post]
func UpFreeWeb(c *gin.Context) {
	ip := c.ClientIP()

	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id

	lang := c.DefaultPostForm("lang", "")
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))
	phone := strings.TrimSpace(c.DefaultPostForm("phone", ""))
	company := strings.TrimSpace(c.DefaultPostForm("company", ""))
	address := strings.TrimSpace(c.DefaultPostForm("address", ""))
	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", nil)
		return
	}
	if !util.CheckEmail(email) {
		JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", gin.H{})
		return
	}
	if phone == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR_PHONE", nil)
		return
	}
	if company == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR_COMPANY", nil)
		return
	}
	hasInfo := models.GetScoreFeedback(uid)
	if hasInfo.ID > 0 {
		JsonReturn(c, -1, "__T_HAS_SCORE_FEEDBACK", nil)
		return
	}
	nowTime := util.GetNowInt()
	fb := models.ScoreFeedback{
		Uid:        uid,
		Email:      email,
		Phone:      phone,
		Company:    company,
		Address:    address,
		Status:     0,
		Lang:       lang,
		Ip:         ip,
		CreateTime: nowTime,
	}
	err := models.AddScoreFeedback(fb)
	if err != nil {
		JsonReturn(c, -1, "__T_FAIL", nil)
		return
	}
	JsonReturn(c, 0, "__T_SUCCESS", nil)
	return

}

// 反馈上报信息
// @BasePath /api/v1
// @Summary 反馈上报信息
// @Description 反馈上报信息
// @Tags 个人中心 - 积分
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登录信息"
// @Param content formData string true "反馈内容"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /web/score/feedback [post]
func UpFeedbackWeb(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id
	content := strings.TrimSpace(c.DefaultPostForm("content", ""))

	if content == "" {
		JsonReturn(c, -1, "__T_EMPTY_CONTENT", nil)
		return
	}
	hasInfo := models.GetScoreFeedback(uid)
	if hasInfo.ID == 0 {
		JsonReturn(c, -1, "__T_HAS_NO_FEEDBACK", nil)
		return
	}
	isUp := 0
	if hasInfo.Status == 2 && hasInfo.Content == "" {
		isUp = 1
	}
	if isUp == 0 {
		JsonReturn(c, -1, "__T_HAS_SCORE_FEEDBACK", nil)
		return
	}

	nowTime := util.GetNowInt()
	resInfo := map[string]interface{}{}
	resInfo["content"] = content
	resInfo["update_time"] = nowTime

	err := models.EditScoreFeedback(hasInfo.ID, resInfo)
	if err != nil {
		JsonReturn(c, -1, "__T_FAIL", nil)
		return
	}
	JsonReturn(c, 0, "__T_SUCCESS", nil)
	return
}

type ResLogUserScore struct {
	Id         int    `json:"id"`
	Uid        int    `json:"uid"`         // 用户id
	Name       string `json:"name"`        // 名称
	Score      int    `json:"score"`       // 积分值
	Money      string `json:"money"`       // 支付金额
	CreateTime string `json:"create_time"` // 创建时间
}
