package controller
/*
import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
	"strings"
)

// 获取子账户列表
func GetUserAccountList(c *gin.Context) {
	username := strings.TrimSpace(c.DefaultPostForm("username", "")) // 用户名
	status := com.StrTo(c.DefaultPostForm("status", "10")).MustInt()

	resCode,msg,userInfo := DealUser(c)		//处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id
	_, accountLists := models.GetUserAccountList(uid, username, status)

	data := []models.ResUserAccount{}
	for _, v := range accountLists {
		flow := int(v.LimitFlow / 1024 / 1024 / 1024)
		useFlow := int(v.UseFlow / 1024 / 1024)
		exceed := 0
		if v.UseFlow >= v.LimitFlow {
			exceed = 1
		}
		info := models.ResUserAccount{}
		info.Id = v.Id
		info.Account = v.Account
		info.Password = v.Password
		info.LimitFlow = util.ItoS(flow) + " GB"
		info.UseFlow = util.ItoS(useFlow) + " MB"
		info.Master = v.Master
		info.Status = v.Status
		info.Exceed = exceed		//1 超出  0未超出
		info.Remark = v.Remark
		info.CreateTime = Time2DateEn(v.CreateTime)
		data = append(data, info)
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", data)
	return
}

// 添加 流量账号子账户
func AddUserFlowAccount(c *gin.Context) {
	accountId := com.StrTo(c.DefaultPostForm("account_id", "0")).MustInt() // 帐密ID
	username := strings.TrimSpace(c.DefaultPostForm("username", ""))       // 用户名
	password := strings.TrimSpace(c.DefaultPostForm("password", ""))       // 密码
	remark := strings.TrimSpace(c.DefaultPostForm("remark", ""))           // 备注
	flowStr := strings.TrimSpace(c.DefaultPostForm("flow", ""))            // 流量限制
	status := com.StrTo(c.DefaultPostForm("status", "1")).MustInt()

	resCode,msg,userInfo := DealUser(c)		//处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id

	if username == "" {
		JsonReturn(c, e.ERROR, "__T_USERNAME_ERROR", nil)
		return
	}
	if password == "" {
		JsonReturn(c, e.ERROR, "__T_PASSWORD_ERROR", nil)
		return
	}
	if flowStr == "" {
		JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil)
		return
	}
	if !util.CheckUserAccount(username) {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_USERNAME_ERROR", nil)
		return
	}
	if !util.CheckUserPassword(password) {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_PASSWORD_ERROR", nil)
		return
	}
	if accountId > 0 {
		_, hasAccount := models.GetUserAccountNeqId(accountId, username)
		if hasAccount.Id != 0 {
			JsonReturn(c, e.ERROR, "__T_ACCOUNT_EXIST", nil)
			return
		}
	} else {
		_, hasAccount := models.GetUserAccount(0, username)
		if hasAccount.Id != 0 {
			JsonReturn(c, e.ERROR, "__T_ACCOUNT_EXIST", nil)
			return
		}
	}

	data := models.UserAccount{}
	data.Status = 1
	data.Remark = remark
	// 查询该用户下面是否存在相同账户 若存在则修改账户信息
	accountInfo := models.UserAccount{}
	if accountId > 0 {
		accountInfo, _ = models.GetUserAccountById(accountId)
		upMap := map[string]interface{}{}
		if accountInfo.Master == 0 {
			upMap["account"] = username
		}
		upMap["limit_flow"] = int64(util.StoI(flowStr)) * 1024 * 1024 * 1024
		upMap["password"] = password
		upMap["status"] = status
		upMap["remark"] = remark
		models.UpdateUserAccountById(accountInfo.Id, upMap)
	} else {
		flow := int64(util.StoI(flowStr)) * 1024 * 1024 * 1024
		data.Uid = uid
		data.Account = username
		data.Password = password
		data.Master = 0
		data.Status = status
		data.LimitFlow = flow
		data.CreateTime = util.GetNowInt()
		err := models.AddProxyAccount(data)
		fmt.Println(err)
	}

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return
}

// 删除 流量子账号
func DelUserAccount(c *gin.Context) {
	session := c.DefaultPostForm("session", "")                            // session
	accountId := com.StrTo(c.DefaultPostForm("account_id", "0")).MustInt() // 帐密ID
	if session == "" {
		JsonReturn(c, -1, "__T_SESSION_ERROR", gin.H{})
		return
	}
	errs, uid := GetUIDbySession(session)
	if errs == false || uid == 0 {
		JsonReturn(c, -1, "__T_SESSION_EXPIRE", gin.H{})
		return
	}
	if accountId == 0 {
		JsonReturn(c, -1, "__T_PARAM_ERROR", gin.H{})
		return
	}

	accountInfo, _ := models.GetUserAccountById(accountId)
	upMap := map[string]interface{}{}
	if accountInfo.Master != 0 {
		JsonReturn(c, -1, "__T_MASTER_ACCOUNT_NO", gin.H{})
		return
	}
	upMap["status"] = -1
	upMap["update_time"] = util.GetNowInt()
	b := models.UpdateUserAccountById(accountInfo.Id, upMap)
	if !b {
		JsonReturn(c, e.ERROR, "__T_FAIL", nil)
		return
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return
}

// 账号启用  / 禁用
func AccountEnableOrDisable(c *gin.Context) {
	session := c.DefaultPostForm("session", "")                            // session
	accountId := com.StrTo(c.DefaultPostForm("account_id", "0")).MustInt() // 帐密ID
	if session == "" {
		JsonReturn(c, -1, "__T_SESSION_ERROR", gin.H{})
		return
	}
	errs, uid := GetUIDbySession(session)
	if errs == false || uid == 0 {
		JsonReturn(c, -1, "__T_SESSION_EXPIRE", gin.H{})
		return
	}
	if accountId == 0 {
		JsonReturn(c, -1, "__T_PARAM_ERROR", gin.H{})
		return
	}

	accountInfo, _ := models.GetUserAccountById(accountId)
	status := accountInfo.Status
	if status == 1 {
		status = 0
	} else if status == 0 {
		status = 1
	}
	upMap := map[string]interface{}{}
	if accountInfo.Master != 0 {
		JsonReturn(c, -1, "__T_MASTER_ACCOUNT_NO", gin.H{})
		return
	}
	upMap["status"] = status
	upMap["update_time"] = util.GetNowInt()
	b := models.UpdateUserAccountById(accountInfo.Id, upMap)
	if !b {
		JsonReturn(c, e.ERROR, "__T_FAIL", nil)
		return
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
	return
}


 */