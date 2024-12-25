package controller

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"sort"
	"strings"
)

// @BasePath /api/v1
// @Summary 获取弹窗信息
// @Description 获取弹窗信息
// @Tags 公告/反馈
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登陆信息"
// @Param lang formData string false "语言"
// @Param place formData string false "//位置   index首页   right右下角 footer 底部"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /web/msg/popup [post]
func GetPopup(c *gin.Context) {
	sessionId := c.DefaultPostForm("session", "")
	lang := strings.ToLower(c.DefaultPostForm("lang", "en"))
	platform := "web"
	place := c.DefaultPostForm("place", "index") //位置   index首页   right右下角 footer 底部
	if place == "" {
		place = "index"
	}
	if lang == "" {
		lang = "en"
	}

	is_pay := 0
	if sessionId != "" {
		_, uid := GetUIDbySession(sessionId)
		_, user := models.GetUserById(uid)
		if user.IsPay == "true" {
			is_pay = 1
		}
	}
	user_type := ""
	if is_pay == 1 {
		user_type = "payed"
	} else {
		user_type = "no_pay"
	}
	info, err := models.GetAdvertInfo(lang, "popup", user_type, platform, place) //查询弹窗

	if err == nil {
		JsonReturn(c, 0, "success", info)
		return
	}
	JsonReturn(c, -1, "no data", nil)
	return
}

// @BasePath /api/v1
// @Summary 获取banner
// @Description 获取banner
// @Tags 公告/反馈
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登陆信息"
// @Param lang formData string false "语言"
// @Param place formData string false "//位置   index首页   right右下角 footer 底部"
// @Produce json
// @Success 0 {object} models.Advert{} "成功"
// @Router /web/msg/banner [post]
func GetBanner(c *gin.Context) {
	sessionId := c.DefaultPostForm("session", "")
	lang := strings.ToLower(c.DefaultPostForm("lang", "en"))
	platform := "web"
	place := c.DefaultPostForm("place", "index") //位置   index首页   right右下角 footer 底部
	if place == "" {
		place = "index"
	}
	if lang == "" {
		lang = "en"
	}
	is_pay := 0
	if sessionId != "" {
		_, uid := GetUIDbySession(sessionId)
		_, user := models.GetUserById(uid)
		if user.IsPay == "true" {
			is_pay = 1
		}
	}
	user_type := ""
	if is_pay == 1 {
		user_type = "payed"
	} else {
		user_type = "no_pay"
	}
	info, err := models.GetAdvertInfo(lang, "banner", user_type, platform, place) //查询弹窗

	if err == nil {
		JsonReturn(c, 0, "success", info)
		return
	}
	JsonReturn(c, -1, "no data", nil)
	return
}

// @BasePath /api/v1
// @Summary 获取公告信息
// @Description 获取公告信息
// @Tags 公告/反馈
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登陆信息"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {object} map[string]interface{} "list 公告列表，red_dots：红点数量，pop：弹窗信息"
// @Router /web/msg/notice [post]
func GetNotice(c *gin.Context) {
	sessionId := c.DefaultPostForm("session", "")
	lang := strings.ToLower(c.DefaultPostForm("lang", "en"))
	if lang == "" {
		lang = "en"
	}
	uid := 0
	if sessionId != "" {
		_, uid = GetUIDbySession(sessionId)
	}
	noticeList, err := models.GetNoticeList()
	if err != nil || len(noticeList) == 0 {

	} else {
		//noticeCodes := []string{}
		//for _, v := range noticeList {
		//	noticeCodes = append(noticeCodes, v.Code)
		//}
		//enNoticeList, _ := models.GetOtherNoticeList("en", noticeCodes)
		//noticeList = append(noticeList, enNoticeList...)
		// 使用 sort.Slice 方法进行排序
		sort.Slice(noticeList, func(i, j int) bool {
			return noticeList[i].ReleaseTime > noticeList[j].ReleaseTime
		})
	}
	resList,noRead,_  := DealMsgCate(noticeList, uid, lang)

	red_dots := 0
	if noRead > 0 {
		red_dots = 1
	}
	if len(noticeList) > 0 {
		pop := models.ResNotice{}
		if len(resList) > 0 {
			if resList[0].Type == 1 && resList[0].IsRead == 0 {
				// 查询用户弹窗关闭记录
				info := models.GetUserTodayPopoInfo(uid, util.GetTodayTime())
				if info.Id == 0 {
					pop = resList[0]
				}
			}
		}
		JsonReturn(c, 0, "success", map[string]interface{}{
			"list":     resList,
			"red_dots": red_dots,
			"pop":      pop,
		})
		return
	}
	JsonReturn(c, 0, "no data", map[string]interface{}{
		"list":     []string{},
		"red_dots": red_dots,
		"pop":      gin.H{},
	})
	return
}


// 处理消息分类
func DealMsgCate(noticeList []models.CmNotice, uid int, lang string) ([]models.ResNotice, int, int) {
	isPay := "no_pay"
	isAgent := 0
	if uid > 0 {
		_, user := models.GetUserById(uid)
		if user.IsPay == "true" {
			isPay = "payed"
		}
		agentInfo := models.GetAgentByUid(uid)
		if agentInfo.Id != 0 {
			isAgent = 1
		}
	}
	nowTime := util.GetTodayTime() - 86400*6
	hasNew := 0 // 是否有新的消息
	noticeLists := []models.CmNotice{}

	for _, v := range noticeList {
		if v.Cate == "all" {
			noticeLists = append(noticeLists, v)
		} else {
			if uid > 0 {
				if v.Cate == isPay {
					noticeLists = append(noticeLists, v)
				} else if v.Cate == "agent" {
					if isAgent == 1 {
						noticeLists = append(noticeLists, v)
					}
				} else {
					uidStr := "," + util.ItoS(uid) + ","
					confUser := "," + strings.Trim(v.Users, ",") + ","
					if v.Users != "" && util.StrContain(confUser, uidStr) {
						noticeLists = append(noticeLists, v)
					}
				}
			}
		}
	}
	userRead := map[int]int{}
	userDel := map[int]int{}
	if uid > 0 {
		userList, _ := models.GetUserNoticeList(uid)
		for _, v := range userList {
			if v.IsDel == 1 {
				userDel[v.Nid] = v.CreateTime
			} else {
				userRead[v.Nid] = v.CreateTime
			}
		}
	}
	//通用  News
	noticeRes := []models.ResNotice{}
	noticeNoReadNum := 0 //未读消息数量
	for _, v := range noticeLists {
		isDel := 0
		if uid > 0 && len(userDel) > 0 {
			_, ok := userDel[v.Id]
			if ok {
				isDel = 1
			} else {
				isDel = 0
			}
		}
		if isDel == 0 {
			isRead := 0
			if uid > 0 && len(userRead) > 0 {
				_, ok := userRead[v.Id]
				if ok {
					isRead = 1
				} else {
					isRead = 0
					noticeNoReadNum = noticeNoReadNum + 1
				}
			}
			title := v.Title
			brief := v.Brief
			content := v.Content
			if lang == "zh-tw" {
				if v.TitleZh != "" {
					title = v.TitleZh
				}
				if v.BriefZh != "" {
					brief = v.BriefZh
				}
				if v.ContentZh != "" {
					content = v.ContentZh
				}
			}
			info := models.ResNotice{}
			info.Id = v.Id
			info.Type = v.Type
			info.Title = title
			info.Brief = brief
			info.Content = content
			info.CreateTime = util.Time2DateEn(v.CreateTime)
			info.ReleaseTime = util.Time2DateEn(v.ReleaseTime)
			info.IsRead = isRead
			noticeRes = append(noticeRes, info)
			if v.CreateTime > nowTime {
				hasNew = 1
			}
		}
	}

	return noticeRes, noticeNoReadNum, hasNew
}


//func GetNotice_bak(c *gin.Context) {
//	sessionId := c.DefaultPostForm("session", "")
//	lang := strings.ToLower(c.DefaultPostForm("lang", "en"))
//	if lang == "" {
//		lang = "en"
//	}
//	uid := 0
//	if sessionId != "" {
//		_, uid = GetUIDbySession(sessionId)
//	}
//	noticeList, err := models.GetNoticeList(lang)
//	if err != nil || len(noticeList) == 0 {
//		noticeList, err = models.GetNoticeList("en")
//	} else {
//		noticeCodes := []string{}
//		for _, v := range noticeList {
//			noticeCodes = append(noticeCodes, v.Code)
//		}
//		enNoticeList, _ := models.GetOtherNoticeList("en", noticeCodes)
//		noticeList = append(noticeList, enNoticeList...)
//		// 使用 sort.Slice 方法进行排序
//		sort.Slice(noticeList, func(i, j int) bool {
//			return noticeList[i].ReleaseTime > noticeList[j].ReleaseTime
//		})
//	}
//
//	red_dots := 0
//	if len(noticeList) > 0 {
//		resList := []models.ResNotice{}
//		userRead := map[string]int{}
//		userDel := map[string]int{}
//		if uid > 0 {
//			userList, _ := models.GetUserNoticeList(uid)
//			for _, v := range userList {
//				if v.IsDel == 1 {
//					userDel[v.Code] = v.CreateTime
//				} else {
//					userRead[v.Code] = v.CreateTime
//				}
//			}
//		}
//		for _, v := range noticeList {
//			isDel := 0
//			if uid > 0 && len(userDel) > 0 {
//				_, ok := userDel[v.Code]
//				if ok {
//					isDel = 1
//				}
//			}
//			if isDel == 0 {
//				isRead := 0
//				if uid > 0 && len(userRead) > 0 {
//					_, ok := userRead[v.Code]
//					if ok {
//						isRead = 1
//					}
//				}
//				info := models.ResNotice{}
//				info.Id = v.Id
//				info.Title = v.Title
//				info.Content = v.Content
//				info.Type = v.Type
//				info.CreateTime = util.Time2DateEn(v.CreateTime)
//				info.ReleaseTime = util.Time2DateEn(v.ReleaseTime)
//				info.IsRead = isRead
//				info.Brief = v.Brief
//				resList = append(resList, info)
//				if red_dots == 0 && isRead == 0 {
//					red_dots = 1
//				}
//			}
//		}
//
//		pop := models.ResNotice{}
//		if len(resList) > 0 {
//			if resList[0].Type == 1 && resList[0].IsRead == 0 {
//				// 查询用户弹窗关闭记录
//				info := models.GetUserTodayPopoInfo(uid, util.GetTodayTime())
//				if info.Id == 0 {
//					pop = resList[0]
//				}
//			}
//		}
//
//		JsonReturn(c, 0, "success", map[string]interface{}{
//			"list":     resList,
//			"red_dots": red_dots,
//			"pop":      pop,
//		})
//		return
//	}
//	JsonReturn(c, 0, "no data", map[string]interface{}{
//		"list":     []string{},
//		"red_dots": red_dots,
//		"pop":      gin.H{},
//	})
//	return
//}

// @BasePath /api/v1
// @Summary 读取公告
// @Description 读取公告
// @Tags 公告/反馈
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登陆信息"
// @Param lang formData string false "语言"
// @Param ids formData string true "公告id列表，逗号分隔"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /web/msg/read_notice [post]
func ReadNotice(c *gin.Context) {
	idStr := c.DefaultPostForm("ids", "")
	sessionId := c.DefaultPostForm("session", "")
	lang := strings.ToLower(c.DefaultPostForm("lang", "en"))
	if lang == "" {
		lang = "en"
	}
	uid := 0
	if sessionId != "" {
		_, uid = GetUIDbySession(sessionId)
	}

	idArr := strings.Split(idStr, ",")
	// 查询公告信息
	list, err := models.GetNoticeListByIds(idArr)
	if len(list) == 0 || err != nil {
		JsonReturn(c, 0, "success", gin.H{})
		return
	}
	for _, v := range list {
		if uid > 0 {
			info := models.CmNoticeUser{}
			info.Nid = v.Id
			info.Uid = uid
			info.Ip = c.ClientIP()
			info.Lang = lang
			info.CreateTime = util.GetNowInt()
			info.Code = v.Code
			err := models.AddUserNotice(info)
			fmt.Println(err)
		}
	}

	JsonReturn(c, 0, "success", gin.H{})
	return
}

// 删除公告
// @BasePath /api/v1
// @Summary 删除公告
// @Description 删除公告
// @Tags 公告/反馈
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登陆信息"
// @Param lang formData string false "语言"
// @Param ids formData string true "公告id列表，逗号分隔"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /web/msg/del_notice [post]
func DelNotice(c *gin.Context) {
	idStr := c.DefaultPostForm("ids", "")
	sessionId := c.DefaultPostForm("session", "")
	lang := strings.ToLower(c.DefaultPostForm("lang", "en"))
	if lang == "" {
		lang = "en"
	}
	uid := 0
	if sessionId != "" {
		_, uid = GetUIDbySession(sessionId)
	}
	idArr := strings.Split(idStr, ",")
	fmt.Println(idArr)
	// 查询公告信息
	list, err := models.GetNoticeListByIds(idArr)
	if len(list) == 0 || err != nil {
		JsonReturn(c, 0, "success", gin.H{})
		return
	}
	for _, v := range list {
		if uid > 0 {
			info := models.CmNoticeUser{}
			info.Nid = v.Id
			info.Uid = uid
			info.Ip = c.ClientIP()
			info.Lang = lang
			info.CreateTime = util.GetNowInt()
			info.IsDel = 1
			info.Code = v.Code
			err := models.AddUserNotice(info)
			fmt.Println(err)
		}
	}

	JsonReturn(c, 0, "success", gin.H{})
	return
}

// 获取公告弹窗
// @BasePath /api/v1
// @Summary 获取公告弹窗
// @Description 获取公告弹窗
// @Tags 公告/反馈
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登陆信息"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {object} models.ResNotice{} "成功"
// @Router /web/msg/notice_pop [post]
func GetNoticePop(c *gin.Context) {
	sessionId := c.DefaultPostForm("session", "")
	lang := strings.ToLower(c.DefaultPostForm("lang", "en"))
	var langLastNotice models.CmNotice
	if lang != "en" {
		langLastNotice, _ = models.GetLastNotice(lang)
	}
	enLastNotice, _ := models.GetLastNotice(lang)
	uid := 0
	if sessionId != "" {
		_, uid = GetUIDbySession(sessionId)
	}
	lastNotice := models.CmNotice{}
	if langLastNotice.Code == enLastNotice.Code {
		lastNotice = langLastNotice
	} else {
		lastNotice = enLastNotice
	}

	userNoticeInfo, _ := models.GetUserNoticeInfo(uid, lastNotice.Id)
	if userNoticeInfo.Id > 0 {
		JsonReturn(c, 0, "success", nil)
		return
	} else {
		info := models.ResNotice{}
		info.Id = lastNotice.Id
		info.Title = lastNotice.Title
		info.Content = lastNotice.Content
		info.Type = lastNotice.Type
		info.CreateTime = util.Time2DateEn(lastNotice.CreateTime)
		info.IsRead = 0
		info.Brief = lastNotice.Brief
		JsonReturn(c, 0, "success", info)
		return
	}
}

// 关闭公告弹窗
// @BasePath /api/v1
// @Summary 关闭公告弹窗
// @Description 关闭公告弹窗
// @Tags 公告/反馈
// @Accept x-www-form-urlencoded
// @Param session formData string false "用户登陆信息"
// @Produce json
// @Success 0 {object} interface{} "成功"
// @Router /web/msg/close_notice_pop [post]
func CloseNoticePop(c *gin.Context) {
	sessionId := c.DefaultPostForm("session", "")
	uid := 0
	if sessionId != "" {
		_, uid = GetUIDbySession(sessionId)
	}

	models.CreateUserTodayPopoInfo(models.CmUserTodayPopo{
		Uid:   uid,
		Today: util.GetTodayTime(),
	})

	JsonReturn(c, 0, "success", gin.H{})
	return
}
