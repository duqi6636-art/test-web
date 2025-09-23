package controller

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"sort"
	"strings"
	"time"
)

func GetNoticeMsg(c *gin.Context) {
	sessionId := c.DefaultPostForm("session", "")
	lang := strings.ToLower(c.DefaultPostForm("lang", "")) //语言

	// 语言处理
	if lang == "" {
		lang = "en"
	}
	uid := 0
	if sessionId != "" {
		_, uid = GetUIDbySession(sessionId)
	}

	msgRes := []models.ResNoticeV1{}
	msgNoReadNum := 0 //未读消息数量

	// 获取消息列表，添加错误处理
	noticeMsg, err := models.GetNoticeMsgList(uid, 100)
	if err != nil {
		JsonReturn(c, -1, "获取消息列表失败", nil)
		return
	}

	// 处理消息数据
	for _, v := range noticeMsg {
		isRead := 0
		if v.ReadTime > 0 {
			isRead = 1
		} else {
			msgNoReadNum++
		}

		info := models.ResNoticeV1{
			Id:         v.Id,
			Title:      v.Title,
			Brief:      v.Brief,
			Content:    v.Content,
			CreateTime: util.Time2DateEn(v.CreateTime),
			IsRead:     isRead,
			Sort:       v.Sort,
			Timestamp:  v.CreateTime,
		}
		if lang == "en" {
			info.Title = v.Title
			info.Brief = v.Brief
			info.Content = v.Content
		} else {
			info.Title = v.TitleZh
			info.Brief = v.BriefZh
			info.Content = v.ContentZh
		}

		msgRes = append(msgRes, info)
	}
	//排序
	sort.SliceStable(msgRes, func(i, j int) bool {
		if msgRes[i].Sort == msgRes[j].Sort {
			// 如果 Sort 相同，按 Timestamp 降序
			return msgRes[i].Timestamp > msgRes[j].Timestamp
		}
		// 否则按 Sort 降序
		return msgRes[i].Sort > msgRes[j].Sort
	})

	response := map[string]interface{}{
		"list":          msgRes,
		"un_read_count": msgNoReadNum,
		"total":         len(noticeMsg),
	}

	JsonReturn(c, 0, "success", response)
	return
}

// MoreReadNotice 批量操作公告 读取/删除
func MoreReadNotice(c *gin.Context) {
	typeStr := strings.TrimSpace(c.DefaultPostForm("type", "read"))
	typeStr = strings.ToLower(typeStr)
	sessionId := c.DefaultPostForm("session", "")
	lang := strings.ToLower(c.DefaultPostForm("lang", ""))   //语言
	idStr := strings.TrimSpace(c.DefaultPostForm("ids", "")) // 获取用户传入的公告ID列表，逗号分隔

	if lang == "" {
		lang = "en"
	}
	uid := 0
	if sessionId != "" {
		_, uid = GetUIDbySession(sessionId)
	}

	// 用户未登录，无法进行操作
	if uid <= 0 {
		JsonReturn(c, -1, "请先登录", nil)
		return
	}

	// 判断是否有传入ID列表
	if idStr != "" {
		// 处理批量操作指定ID的公告
		idArr := strings.Split(idStr, ",")
		idIntArr := []int{}
		for _, idStrItem := range idArr {
			id := util.StoI(idStrItem)
			if id > 0 {
				idIntArr = append(idIntArr, id)
			}
		}

		// 如果没有有效的ID，返回错误
		if len(idIntArr) == 0 {
			JsonReturn(c, -1, "无效的公告ID", nil)
			return
		}

		// 只处理用户指定的公告ID
		for _, id := range idIntArr {
			isDel := 1
			if typeStr == "delete" {
				isDel = -1
			}

			isRead := 0
			if typeStr == "read" {
				isRead = int(time.Now().Unix())
			}
			// 添加或更新用户公告记录
			info := map[string]interface{}{
				"read_time": isRead,
				"status":    isDel,
			}
			if !models.EditNoticeMsg(id, info) {
				fmt.Println("添加用户公告记录失败")
			}
		}
	}
	JsonReturn(c, 0, "success", gin.H{})
	return
}
