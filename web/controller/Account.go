package controller

import (
	"api-360proxy/pkg/ipdat"
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/setting"
	"api-360proxy/web/pkg/util"
	emailSender "api-360proxy/web/service/email"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	googleidtokenverifier "github.com/movsb/google-idtoken-verifier"
	"github.com/mssola/user_agent"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// @BasePath /api/v1
// @Summary 邮箱注册
// @Description 邮箱注册
// @Tags 登陆注册相关
// @Accept x-www-form-urlencoded
// @Param email formData string true "邮箱"
// @Param code formData string false "验证码"
// @Param password formData string true "密码"
// @Param re_password formData string true "确认密码"
// @Param j_source formData string false "竞价平台"
// @Param j_code formData string false "竞价code"
// @Param j_domain formData string false "竞价域名"
// @Param origin formData string false "用户注册来源"
// @Param invite_type formData string false "邀请类型 1链接  2填写推广码"
// @Param invite formData string false "邀请码"
// @Produce json
// @Success 0 {object} models.ResUser{}
// @Router /web/user/email_reg [post]
func WebReg(c *gin.Context) {
	ip := c.ClientIP()
	params := GetParams(c)
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))
	code := strings.TrimSpace(c.DefaultPostForm("code", ""))
	re_password := strings.TrimSpace(c.DefaultPostForm("re_password", ""))
	password := strings.TrimSpace(c.DefaultPostForm("password", ""))
	jjSource := c.DefaultPostForm("j_source", "")                     // 竞价平台
	jjCode := c.DefaultPostForm("j_code", "")                         // 竞价code
	jjDomain := c.DefaultPostForm("j_domain", "")                     // 竞价域名
	origin := c.DefaultPostForm("origin", "")                         // 用户注册来源
	saltStr := strings.TrimSpace(c.DefaultPostForm("email_code", "")) // 加盐信息

	inviter_type := strings.TrimSpace(c.DefaultPostForm("invite_type", "")) // 邀请类型 1链接  2填写推广码
	inviter_code := strings.TrimSpace(c.DefaultPostForm("invite", ""))      // 邀请码

	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", map[string]string{"class_id": "email"})
		return
	}
	if saltStr == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_CODE_IS_MUST", map[string]string{"class_id": "email_code"})
		return
	}
	if !util.CheckEmail(email) {
		JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", map[string]string{"class_id": "email"})
		return
	}
	if password == "" {
		JsonReturn(c, e.ERROR, "__T_PLEASE_ENTER_PASSWORD", map[string]string{"class_id": "password"})
		return
	}
	if re_password == "" {
		JsonReturn(c, e.ERROR, "__T_PLEASE_ENTER_RE_PASSWORD", map[string]string{"class_id": "passwordRepeat"})
		return
	}
	checkRes, checkMsg := util.CheckPwdNew(password) //密码验证
	if !checkRes {
		JsonReturn(c, -1, checkMsg, map[string]string{"class_id": "password"})
		return
	}
	if re_password != password {
		JsonReturn(c, e.ERROR, "__T_TWO_PWD_NOT_MATCH", map[string]string{"class_id": "passwordRepeat"})
		return
	}
	// 滑块验证 -- start
	captchaSwitch := strings.TrimSpace(models.GetConfigVal("CaptchaRegisterSwitch")) // 滑块验证注册开关
	if captchaSwitch == "1" {
		ticket := c.DefaultPostForm("ticket", "")
		randStr := c.DefaultPostForm("randstr", "")
		if ticket == "" || randStr == "" {
			JsonReturn(c, e.ERROR, "__T_CAPTCHA_FAIL", nil)
			return
		}

		res, Msg := CaptchaHandle(c, ticket, randStr)
		if !res {
			JsonReturn(c, e.ERROR, Msg, nil)
			return
		}
	}
	// 滑块验证 -- end

	// ----------------- 注册限制频率 start -----------------
	resReg, salt, msgStr := DealLimitReg(c, email, ip, saltStr)
	fmt.Println("resReg", resReg, "msgStr", msgStr, "salt", salt)
	if resReg == -1 {
		JsonReturn(c, e.ERROR, msgStr, nil)
		return
	}
	if resReg == 1 { // 验证信息错误
		JsonReturn(c, e.ERROR, msgStr, map[string]string{"class_id": "code"})
		return
	}
	// ----------------注册限制频率 end --------------

	//获取邮箱验证开关
	codeId := 0
	config := strings.TrimSpace(models.GetConfigVal("EmailVerificationSwitch"))
	if config == "1" {
		if code == "" {
			JsonReturn(c, -1, "__T_VERIFY_EMPTY", map[string]string{"class_id": "code"})
			return
		}
		// 验证 验证码
		codeId, _ = models.CheckVerifyCode(code, email, "reg")
		if codeId == 0 {
			JsonReturn(c, -1, "__T_VERIFY_ERROR", map[string]string{"class_id": "code"})
			return
		}
	}

	//获取用户信息
	_, info := models.GetUserByEmail(email)
	if info.Id > 0 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_EXIST", map[string]string{"class_id": "email"})
		return
	}
	username := GetUsername()
	params.DeviceOs = "web"
	platform := "web"
	userAgent := c.GetHeader("user-agent")
	isMobile := strings.Contains(strings.ToLower(userAgent), "mobile")
	fmt.Println("isMobile", isMobile)
	if isMobile {
		params.DeviceOs = "wap"
		platform = "wap"
	}

	// 邀请信息
	inviter_id := 0
	inviter_username := ""
	if inviter_code != "" {
		if inviter_type == "" { //邀请类型   1 链接  2邀请码
			inviter_type = "1"
		}
		// 判断是否是活动邀请码
		if strings.Contains(inviter_code, "act_") {
			err, pUser := models.GetUserActivityInviterByMap(map[string]interface{}{"inviter_code": inviter_code})
			if err != nil || pUser.Id == 0 {
				JsonReturn(c, e.ERROR, "Invitation code error", nil)
				return
			}
			inviter_id = pUser.Uid
			inviter_username = pUser.InviterCode
		} else {
			err, pUser := models.GetUserInviterByMap(map[string]interface{}{"inviter_code": inviter_code})
			if err != nil || pUser.ID == 0 {
				JsonReturn(c, e.ERROR, "Invitation code error", nil)
				return
			}
			inviter_id = pUser.Uid
			inviter_username = pUser.InviterCode
		}
	}

	regResult, msg, user := DoRegister(ip, email, "", platform, password, params, username, origin, inviter_type, inviter_id, inviter_username, jjSource, jjCode, jjDomain, "", "", salt)

	if !regResult {
		JsonReturn(c, e.ERROR, msg, nil)
		return
	}
	if codeId > 0 {
		// 验证码验证完成后销毁
		cRes := models.UpdateCodeStatus(map[string]interface{}{"id": codeId})
		fmt.Println(cRes)
	}

	// 执行登录
	params.Salt = salt
	res, sessionRes := DoLogin(user, "", ip, params, false, c)
	if !res {
		JsonReturn(c, -1, sessionRes, nil)
		return
	}
	// 生成返回数据
	data := ResUserInfo(sessionRes, ip, user)
	JsonReturn(c, 0, "__T_REG_SUCCESS", data)
	return
}

// @BasePath /api/v1
// @Summary 邮箱登录
// @Description 邮箱登录
// @Tags 登陆注册相关
// @Accept x-www-form-urlencoded
// @Param email formData string true "邮箱"
// @Param code formData string false "验证码"
// @Param password formData string false "密码"
// @Produce json
// @Success 0 {object} models.ResUser{}
// @Router /web/user/login [post]
func Login(c *gin.Context) {
	params := GetParams(c)
	params.DeviceOs = "web"
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))
	code := strings.TrimSpace(c.DefaultPostForm("code", ""))
	password := c.DefaultPostForm("password", "")

	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", map[string]string{"class_id": "email"})
		return
	}
	if password == "" {
		JsonReturn(c, e.ERROR, "__T_PLEASE_ENTER_PASSWORD", map[string]string{"class_id": "password"})
		return
	}
	//获取邮箱验证开关
	codeId := 0
	config := strings.TrimSpace(models.GetConfigVal("EmailVerifyLoginSwitch"))
	if config == "1" {
		if code == "" {
			JsonReturn(c, -1, "__T_VERIFY_EMPTY", map[string]string{"class_id": "code"})
			return
		}
		// 验证 验证码
		codeId, _ = models.CheckVerifyCode(code, email, "login")
		if codeId == 0 {
			JsonReturn(c, -1, "__T_VERIFY_ERROR", map[string]string{"class_id": "code"})
			return
		}
	}
	//if !util.CheckPwd(password) {
	//	JsonReturn(c, e.ERROR, "__T_PASSWORD_FORMAT",nil)
	//	return
	//}
	err, info := models.GetUserByEmail(email)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_USER_NOT_EXIST", map[string]string{"class_id": "email"})
		return
	}
	if info.Status == 2 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_DISABLED", nil)
		return
	}
	if info.Status == 3 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_DISPUTE", nil)
		return
	}
	if !CheckPwd(info.Password, password, info.Username) {
		JsonReturn(c, e.ERROR, "__T_USER_PASS_WRONG", map[string]string{"class_id": "password"})
		return
	}
	// 执行登录
	ip := c.ClientIP()
	res, sessionRes := DoLogin(info, "", ip, params, false, c)
	if !res {
		JsonReturn(c, -1, sessionRes, map[string]string{"error_position": "0"})
		return
	}
	if codeId > 0 {
		// 验证码验证完成后销毁
		cRes := models.UpdateCodeStatus(map[string]interface{}{"id": codeId})
		fmt.Println(cRes)
	}
	// 生成返回数据
	data := ResUserInfo(sessionRes, ip, info)
	JsonReturn(c, 0, "__T_LOGIN_SUCCESS", data)
	return
}

// @BasePath /api/v1
// @Summary 谷歌登录
// @Description 谷歌登录
// @Tags 登陆注册相关
// @Accept x-www-form-urlencoded
// @Param token formData string true "谷歌ID令牌"
// @Param access_token formData string false "谷歌授权令牌"
// @Produce json
// @Success 0 {object} models.ResUser{}
// @Router /web/user/google_login [post]
func GoogleLogin(c *gin.Context) {

	// 获取谷歌ID令牌
	token := strings.TrimSpace(c.DefaultPostForm("token", ""))
	// 获取谷歌授权令牌
	accessToken := strings.TrimSpace(c.DefaultPostForm("access_token", ""))

	if token == "" && accessToken == "" {
		JsonReturn(c, e.ERROR, "Please Auth Login", nil)
		return
	}

	platform := c.DefaultPostForm("platform", "web")
	ip := c.ClientIP()
	params := GetParams(c)

	params.DeviceOs = "web"
	jjSource := c.DefaultPostForm("j_source", "") // 竞价平台
	jjCode := c.DefaultPostForm("j_code", "")     // 竞价code
	jjDomain := c.DefaultPostForm("j_domain", "") // 竞价域名
	origin := c.DefaultPostForm("origin", "")     // 用户注册来源

	var claims *googleidtokenverifier.ClaimSet
	var err error

	if token != "" {
		// 获取网页谷歌客户端ID
		clientID := models.GetConfigVal("google_login_client_id_web")
		// 去验证
		claims, err = googleidtokenverifier.Verify(token, clientID)
	}
	if accessToken != "" {
		// 构建请求
		req, errAccess := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
		if errAccess != nil {
			fmt.Println(errAccess)
		}
		// 设置 Authorization 头部
		req.Header.Set("Authorization", "Bearer "+accessToken)
		// 发送请求
		client := &http.Client{}
		resp, errAccess := client.Do(req)
		if errAccess != nil {
			fmt.Println(errAccess)
		}
		defer resp.Body.Close()

		// 解析响应
		err = json.NewDecoder(resp.Body).Decode(&claims)
		if err != nil {
			fmt.Println(err)
		}
	}
	if err != nil {
		JsonReturn(c, e.ERROR, "Login Error", nil)
		return
	}

	email := claims.Email
	nickname := claims.Name
	guid := claims.Sub

	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", map[string]string{"class_id": "email"})
		return
	}

	if !util.CheckEmail(email) {
		JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", map[string]string{"class_id": "email"})
		return
	}
	password := ""
	//获取用户信息
	err, userInfo := models.GetUserByGuid(guid)
	if userInfo.Id == 0 { // 未注册
		_, userEmailInfo := models.GetUserByEmail(email)
		if userEmailInfo.Id > 0 {
			if userEmailInfo.GoogleUid == "" {

				models.EditUserById(userEmailInfo.Id, map[string]interface{}{"google_uid": guid, "nickname": nickname})
			}
			userEmailInfo.GoogleUid = guid
			userInfo = userEmailInfo
		} else {
			// 注册用户
			regResult, msg, user := DoRegister(ip, email, nickname, platform, password, params, GetUsername(), origin, "", 0, "", jjSource, jjCode, jjDomain, "google", guid, "")

			if !regResult {
				JsonReturn(c, e.ERROR, msg, nil)
				return
			}
			// 执行登录
			res, sessionRes := DoLogin(user, "", ip, params, false, c)
			if !res {
				JsonReturn(c, -1, sessionRes, nil)
				return
			}
			// 生成返回数据
			data := ResUserInfo(sessionRes, ip, user)
			JsonReturn(c, 0, "__T_REG_SUCCESS", data)

			// 发送邮箱
			ipInfo, _ := ipdat.IPDat.GetIpInfo(ip)
			defaultMail := models.GetConfigVal("default_email")
			emailSender.SendGoogleLoginNewUserEmail(email, user.PlaintextPassword, ip, ipInfo.CountryCode, defaultMail)
			return
		}
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		JsonReturn(c, e.ERROR, "__T_USER_NOT_EXIST", map[string]string{"class_id": "email"})
		return
	}
	if userInfo.Status == 2 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_DISABLED", nil)
		return
	}
	if userInfo.Status == 3 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_DISPUTE", nil)
		return
	}
	// 执行登录

	res, sessionRes := DoLogin(userInfo, "", ip, params, false, c)
	if !res {
		JsonReturn(c, -1, sessionRes, map[string]string{"error_position": "0"})
		return
	}
	// 生成返回数据
	data := ResUserInfo(sessionRes, ip, userInfo)
	JsonReturn(c, 0, "__T_LOGIN_SUCCESS", data)
	return
}

var (
	githubOAuthConf = &oauth2.Config{
		ClientID:     "Ov23liJkGtErVxEm5E5x",
		ClientSecret: "92f92d228d6368c6a7c70383ed1fadef39942216",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
	oauthStateString = util.RandStr("r", 16)
)

type GithubUserResponse struct {
	Id                int    `json:"id"`
	AvatarUrl         string `json:"avatar_url"`
	Bio               string `json:"bio"`
	Blog              string `json:"blog"`
	Company           string `json:"company"`
	CreatedAt         string `json:"created_at"`
	Email             string `json:"email"`
	EventsUrl         string `json:"events_url"`
	Followers         int    `json:"followers"`
	FollowersUrl      string `json:"followers_url"`
	Following         int    `json:"following"`
	FollowingUrl      string `json:"following_url"`
	GistsUrl          string `json:"gists_url"`
	GravatarId        string `json:"gravatar_id"`
	Hireable          string `json:"hireable"`
	HtmlUrl           string `json:"html_url"`
	Location          string `json:"location"`
	Login             string `json:"login"`
	Name              string `json:"name"`
	NodeId            string `json:"node_id"`
	NotificationEmail string `json:"notification_email"`
	OrganizationsUrl  string `json:"organizations_url"`
	PublicGists       int    `json:"public_gists"`
	PublicRepos       int    `json:"public_repos"`
	ReceivedEventsUrl string `json:"received_events_url"`
	ReposUrl          string `json:"repos_url"`
	SiteAdmin         bool   `json:"site_admin"`
	StarredUrl        string `json:"starred_url"`
	SubscriptionsUrl  string `json:"subscriptions_url"`
	TwitterUsername   string `json:"twitter_username"`
	Type              string `json:"type"`
	UpdatedAt         string `json:"updated_at"`
	Url               string `json:"url"`
}

type GithubUserEmailResponse struct {
	Email      string `json:"email"`
	Verified   bool   `json:"verified"`
	Primary    bool   `json:"primary"`
	Visibility string `json:"visibility"`
}

func GithubLogin(c *gin.Context) {
	// 获取code
	code := strings.TrimSpace(c.DefaultPostForm("code", ""))
	if code == "" {
		JsonReturn(c, e.ERROR, "Please Auth Login", nil)
		return
	}
	accessToken, err := githubOAuthConf.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		JsonReturn(c, -1, "__T_LOGIN_ERROR", nil)
		return
	}
	platform := c.DefaultPostForm("platform", "web")
	ip := c.ClientIP()
	params := GetParams(c)

	params.DeviceOs = "web"
	jjSource := c.DefaultPostForm("j_source", "") // 竞价平台
	jjCode := c.DefaultPostForm("j_code", "")     // 竞价code
	jjDomain := c.DefaultPostForm("j_domain", "") // 竞价域名
	origin := c.DefaultPostForm("origin", "")     // 用户注册来源
	// 获取用户信息
	client := githubOAuthConf.Client(context.Background(), accessToken)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		fmt.Printf("client.Get() failed with '%s'\n", err)
		JsonReturn(c, -1, "__T_LOGIN_ERROR", nil)
		return
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll() failed with '%s'\n", err)
		JsonReturn(c, -1, "__T_LOGIN_ERROR", nil)
		return
	}

	githubUserInfo := GithubUserResponse{}
	if err := json.Unmarshal(contents, &githubUserInfo); err != nil {
		fmt.Printf("json.Unmarshal() failed with '%s'\n", err)
		JsonReturn(c, -1, "__T_LOGIN_ERROR", nil)
		return
	}
	clientEmail := githubOAuthConf.Client(context.Background(), accessToken)
	respEmail, err := clientEmail.Get("https://api.github.com/user/emails")
	if err != nil {
		fmt.Printf("clientEmail.Get() failed with '%s'\n", err)
		JsonReturn(c, -1, "__T_LOGIN_ERROR", nil)
		return
	}
	defer resp.Body.Close()

	contentsEmail, err := ioutil.ReadAll(respEmail.Body)
	if err != nil {
		fmt.Printf("contentsEmail.ReadAll() failed with '%s'\n", err)
		JsonReturn(c, -1, "__T_LOGIN_ERROR", nil)
		return
	}

	userEmail := []GithubUserEmailResponse{}
	if err := json.Unmarshal(contentsEmail, &userEmail); err != nil {
		fmt.Printf("json.Unmarshal() userEmail failed with '%s'\n", err)
		JsonReturn(c, -1, "__T_LOGIN_ERROR", nil)
		return
	}

	var email string
	for _, v := range userEmail {
		if v.Primary == true {
			email = v.Email
		}
	}
	if email == "" {
		email = userEmail[0].Email
	}

	nickname := githubUserInfo.Login
	githubId := githubUserInfo.Id

	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", map[string]string{"class_id": "email"})
		return
	}

	if !util.CheckEmail(email) {
		JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", map[string]string{"class_id": "email"})
		return
	}
	password := ""
	//获取用户信息
	err, userInfo := models.GetUserByGithubId(githubId)
	if userInfo.Id == 0 { // 未注册
		_, userEmailInfo := models.GetUserByEmail(email)
		if userEmailInfo.Id > 0 {
			if userEmailInfo.GithubId == 0 {
				models.EditUserById(userEmailInfo.Id, map[string]interface{}{"github_id": githubId, "nickname": nickname})
			}
			userEmailInfo.GithubId = githubId
			userInfo = userEmailInfo
		} else {
			// 注册用户
			regResult, msg, user := DoRegister(ip, email, nickname, platform, password, params, GetUsername(), origin, "", 0, "", jjSource, jjCode, jjDomain, "github", util.ItoS(githubId), "")

			if !regResult {
				JsonReturn(c, e.ERROR, msg, nil)
				return
			}
			// 执行登录
			res, sessionRes := DoLogin(user, "", ip, params, false, c)
			if !res {
				JsonReturn(c, -1, sessionRes, nil)
				return
			}
			// 生成返回数据
			data := ResUserInfo(sessionRes, ip, user)
			JsonReturn(c, 0, "__T_REG_SUCCESS", data)
			return
		}
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		JsonReturn(c, e.ERROR, "__T_USER_NOT_EXIST", map[string]string{"class_id": "email"})
		return
	}
	if userInfo.Status == 2 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_DISABLED", nil)
		return
	}
	if userInfo.Status == 3 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_DISPUTE", nil)
		return
	}
	// 执行登录
	res, sessionRes := DoLogin(userInfo, "", ip, params, false, c)
	if !res {
		JsonReturn(c, -1, sessionRes, map[string]string{"error_position": "0"})
		return
	}
	// 生成返回数据
	data := ResUserInfo(sessionRes, ip, userInfo)
	JsonReturn(c, 0, "__T_LOGIN_SUCCESS", data)
	return
}

var (
	mpGithubOAuthConf = &oauth2.Config{
		ClientID:     "Ov23liXNlUdgLFlOFHu7",
		ClientSecret: "4bec57391905d225aa3184e64d8612084136b478",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
)

func MPGithubLogin(c *gin.Context) {
	// 获取code
	code := strings.TrimSpace(c.DefaultPostForm("code", ""))
	if code == "" {
		JsonReturn(c, e.ERROR, "Please Auth Login", nil)
		return
	}
	accessToken, err := mpGithubOAuthConf.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		JsonReturn(c, -1, "__T_LOGIN_ERROR", nil)
		return
	}
	platform := c.DefaultPostForm("platform", "mp")
	ip := c.ClientIP()
	params := GetParams(c)

	params.DeviceOs = "mp"
	params.Platform = "mp"
	jjSource := c.DefaultPostForm("j_source", "") // 竞价平台
	jjCode := c.DefaultPostForm("j_code", "")     // 竞价code
	jjDomain := c.DefaultPostForm("j_domain", "") // 竞价域名
	origin := c.DefaultPostForm("origin", "")     // 用户注册来源
	// 获取用户信息
	client := mpGithubOAuthConf.Client(context.Background(), accessToken)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		fmt.Printf("client.Get() failed with '%s'\n", err)
		JsonReturn(c, -1, "__T_LOGIN_ERROR", nil)
		return
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll() failed with '%s'\n", err)
		JsonReturn(c, -1, "__T_LOGIN_ERROR", nil)
		return
	}

	githubUserInfo := GithubUserResponse{}
	if err := json.Unmarshal(contents, &githubUserInfo); err != nil {
		fmt.Printf("json.Unmarshal() failed with '%s'\n", err)
		JsonReturn(c, -1, "__T_LOGIN_ERROR", nil)
		return
	}
	clientEmail := mpGithubOAuthConf.Client(context.Background(), accessToken)
	respEmail, err := clientEmail.Get("https://api.github.com/user/emails")
	if err != nil {
		fmt.Printf("clientEmail.Get() failed with '%s'\n", err)
		JsonReturn(c, -1, "__T_LOGIN_ERROR", nil)
		return
	}
	defer resp.Body.Close()

	contentsEmail, err := ioutil.ReadAll(respEmail.Body)
	if err != nil {
		fmt.Printf("contentsEmail.ReadAll() failed with '%s'\n", err)
		JsonReturn(c, -1, "__T_LOGIN_ERROR", nil)
		return
	}

	userEmail := []GithubUserEmailResponse{}
	if err := json.Unmarshal(contentsEmail, &userEmail); err != nil {
		fmt.Printf("json.Unmarshal() userEmail failed with '%s'\n", err)
		JsonReturn(c, -1, "__T_LOGIN_ERROR", nil)
		return
	}

	var email string
	for _, v := range userEmail {
		if v.Primary == true {
			email = v.Email
		}
	}
	if email == "" {
		email = userEmail[0].Email
	}

	nickname := githubUserInfo.Login
	githubId := githubUserInfo.Id

	if email == "" {
		JsonReturn(c, e.ERROR, "__T_EMAIL_IS_MUST", map[string]string{"class_id": "email"})
		return
	}

	if !util.CheckEmail(email) {
		JsonReturn(c, e.ERROR, "__T_EMAIL_FORMAT_ERROR", map[string]string{"class_id": "email"})
		return
	}
	password := ""
	//获取用户信息
	err, userInfo := models.GetUserByMpGithubId(githubId)
	if userInfo.Id == 0 { // 未注册
		_, userEmailInfo := models.GetUserByEmail(email)
		if userEmailInfo.Id > 0 {
			if userEmailInfo.MpGithubId == 0 {
				models.EditUserById(userEmailInfo.Id, map[string]interface{}{"mp_github_id": githubId, "nickname": nickname})
			}
			userEmailInfo.MpGithubId = githubId
			userInfo = userEmailInfo
		} else {
			// 注册用户
			regResult, msg, user := DoRegister(ip, email, nickname, platform, password, params, GetUsername(), origin, "", 0, "", jjSource, jjCode, jjDomain, "mp_github", util.ItoS(githubId), "")

			if !regResult {
				JsonReturn(c, e.ERROR, msg, nil)
				return
			}
			// 执行登录
			res, sessionRes := DoLogin(user, "", ip, params, false, c)
			if !res {
				JsonReturn(c, -1, sessionRes, nil)
				return
			}
			// 生成返回数据
			data := ResUserInfo(sessionRes, ip, user)
			JsonReturn(c, 0, "__T_REG_SUCCESS", data)
			return
		}
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		JsonReturn(c, e.ERROR, "__T_USER_NOT_EXIST", map[string]string{"class_id": "email"})
		return
	}
	if userInfo.Status == 2 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_DISABLED", nil)
		return
	}
	if userInfo.Status == 3 {
		JsonReturn(c, e.ERROR, "__T_ACCOUNT_DISPUTE", nil)
		return
	}
	// 执行登录

	res, sessionRes := DoLogin(userInfo, "", ip, params, false, c)
	if !res {
		JsonReturn(c, -1, sessionRes, map[string]string{"error_position": "0"})
		return
	}
	// 生成返回数据
	data := ResUserInfo(sessionRes, ip, userInfo)
	JsonReturn(c, 0, "__T_LOGIN_SUCCESS", data)
	return
}

// 忘记密码，登录发送验证码
func SendEmailCode(c *gin.Context) {
	signParam := GetParams(c)
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))
	code_type := c.DefaultPostForm("type", "reg")
	if email == "" {
		JsonReturn(c, -1, "__T_EMAIL_IS_MUST", map[string]string{"class_id": "email"})
		return
	}
	if !util.CheckEmail(email) {
		JsonReturn(c, -1, "__T_EMAIL_FORMAT_ERROR", map[string]string{"class_id": "email"})
		return
	}
	if code_type == "login" || code_type == "find" {

	} else {
		JsonReturn(c, -1, "__T_PARAM_ERROR", nil)
	}
	if code_type == "login" {
		config := strings.TrimSpace(models.GetConfigVal("EmailVerifyLoginSwitch"))
		if config != "1" {
			JsonReturn(c, 0, "__T_EMAIL_SENDED", gin.H{})
			return
		}
	}
	// 获取配置
	ipRegOperateTimesStr := models.GetConfigVal("IpRegOperateTimes") //获取多长时间内操作的
	ipRegOperateTimes := util.StoI(ipRegOperateTimesStr)
	if ipRegOperateTimes == 0 {
		ipRegOperateTimes = 86400
	}
	// 获取配置
	sendNumberStr := models.GetConfigVal("EmailSendNumber") //获取发送次数
	sendNumber := util.StoI(sendNumberStr)
	if sendNumber == 0 {
		sendNumber = 10
	}

	nowTime := util.GetNowInt()
	var start = nowTime - ipRegOperateTimes //多长时间
	send := models.GetVerifyCountByEmail(email, start)
	if send >= sendNumber {
		JsonReturn(c, -1, "__T_MANY_SEND_EMAIL", gin.H{})
		return
	}
	sendIp := models.GetVerifyCountByIp(email, start)
	if sendIp >= sendNumber {
		JsonReturn(c, -1, "__T_MANY_SEND_EMAIL", gin.H{})
		return
	}
	err, user := models.GetUserByEmail(email)
	// 注册/绑定  判断用户是否存在
	if code_type == "reg" || code_type == "bind" {
		if err == nil && user.Id > 0 {
			JsonReturn(c, -1, "__T_ACCOUNT_EXIST", gin.H{})
			return
		}
	}

	// 登录 找回密码  判断用户是否存在
	if code_type == "login" || code_type == "find" {
		if err != nil {
			JsonReturn(c, -1, "__T_USER_NOT_EXIST", gin.H{})
			return
		}
	}

	ip := c.ClientIP()
	ipInfo, _ := ipdat.IPDat.GetIpInfo(ip)
	default_mail := models.GetConfigVal("default_email")
	r, msg := emailSender.SendEmailCode(email, code_type, signParam, ip, ipInfo.CountryCode, default_mail)

	if r {
		JsonReturn(c, 0, msg, gin.H{})
		return
	}
	JsonReturn(c, -1, msg, gin.H{})
	return
}

// @BasePath /api/v1
// @Summary 忘记密码
// @Description 忘记密码
// @Tags 登陆注册相关
// @Accept x-www-form-urlencoded
// @Param email formData string true "邮箱"
// @Param code formData string true "验证码"
// @Param password formData string true "新密码"
// @Param re_password formData string true "确认密码"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/user/forget [post]
func ForgetPwd(c *gin.Context) {
	//params := GetSignParam(c)
	email := strings.TrimSpace(c.DefaultPostForm("email", ""))
	code := c.DefaultPostForm("code", "")
	password := strings.TrimSpace(c.DefaultPostForm("password", ""))
	re_pass := strings.TrimSpace(c.DefaultPostForm("re_password", ""))
	vtype := "find"
	if email == "" {
		JsonReturn(c, -1, "__T_ACCOUNT_ERROR", map[string]string{"class_id": "password"})
		return
	}
	if code == "" {
		JsonReturn(c, -1, "__T_CODE_ERROR", map[string]string{"class_id": "code"})
		return
	}
	if password == "" {
		JsonReturn(c, -1, "__T_PASSWORD_EMPTY", map[string]string{"class_id": "password"})
		return
	}
	checkRes, checkMsg := util.CheckPwdNew(password) //密码验证
	if !checkRes {
		JsonReturn(c, -1, checkMsg, nil)
		return
	}
	if password != re_pass {
		JsonReturn(c, -1, "__T_RETRY_PASSWORD_NOT_MATCH", map[string]string{"class_id": "password"})
		return
	}

	codeId, _ := models.CheckVerifyCode(code, email, vtype)
	if codeId == 0 {
		JsonReturn(c, -1, "__T_VERIFY_ERROR", map[string]string{"class_id": "code"})
		return
	}

	err, user := models.GetUserByEmail(email)
	if err != nil {
		JsonReturn(c, -1, "__T_USER_NOT_EXIST", map[string]string{"class_id": "email"})
		return
	}

	user.Password = util.PassEncode(password, user.Username, 0)
	user.PlaintextPassword = password
	r := models.UpdateUserById(user.Id, &user)

	// 验证码验证完成后销毁
	cRes := models.UpdateCodeStatus(map[string]interface{}{"id": codeId})
	fmt.Println(cRes)

	//删除改用户所有的登录信息
	models.DeleteSession(map[string]interface{}{
		"uid": user.Id,
	})

	if !r {
		JsonReturn(c, -1, "find password fail", map[string]string{"error_position": "0"})
		return
	}
	JsonReturn(c, 0, "__T_SUCCESS", nil)
	return
}

// @BasePath /api/v1
// @Summary 退出登录
// @Description 退出登录
// @Tags 登陆注册相关
// @Accept x-www-form-urlencoded
// @Param session formData string true "用户登陆信息"
// @Produce json
// @Success 0 {object} interface{}
// @Router /web/user/logout [post]
func Logout(c *gin.Context) {
	ses := c.DefaultPostForm("session", "")
	deviceNum := c.DefaultPostForm("device_num", "")
	params := GetParams(c)
	ip := c.ClientIP()
	if ses == "" {
		JsonReturn(c, -1, "__T_PARAMS_ERROR", nil)
		return
	}
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	result := models.DeleteSession(map[string]interface{}{
		"session_id": ses,
	})
	ua := c.Request.Header.Get("User-Agent")
	uaObj := user_agent.New(ua)
	browser, _ := uaObj.Browser()
	os := uaObj.OS()
	ipInfo, _ := ipdat.IPDat.GetIpInfo(ip)
	LogLogin := models.LogLogin{}
	LogLogin.UID = user.Id
	LogLogin.UserLogin = user.Username
	LogLogin.Ip = ip
	LogLogin.LoginTime = time.Now().Unix()
	LogLogin.Platform = "web"
	LogLogin.RegTime = user.CreateTime
	LogLogin.UserAgent = ua
	LogLogin.Browser = browser
	LogLogin.Version = "1"
	LogLogin.OsInfo = os
	LogLogin.Country = ipInfo.Country
	LogLogin.Today = util.GetTodayTime()
	LogLogin.Language = c.Request.Header.Get("Accept-Language")
	LogLogin.Lang = params.Language
	LogLogin.Cate = "login_out"
	LogLogin.TimeZone = params.TimeZone
	LogLogin.DeviceNumber = deviceNum
	_ = models.AddLoginLog(LogLogin)

	if result != nil {
		JsonReturn(c, -1, "__T_FAIL", nil)
		return
	}
	JsonReturn(c, 0, "__T_QUIT_SUCCESS", nil)
	return
}

// 注册
func DoRegister(ip, email, nickname, platform, password string, params models.SignParam, userName string, origin string, inviter_type string, inviter_id int, inviter_username string, jjSource, jjCode, jjDomain string, thirdPlatform, thirdPlatformUid, salt string) (bool, string, models.Users) {
	nowTime := util.GetNowInt()
	ipInfo, _ := ipdat.IPDat.GetIpInfo(ip)

	user := models.Users{}

	devSn := ""
	username := ""
	if userName != "" {
		username = userName
	} else {
		username = GetUsername()
	}

	devSn = params.Sn
	if email != "" {
		devSn = util.GenSnWithName(email)
	}

	if nickname == "" {
		nickname = username
	}
	//生成默认密码
	if password == "" {
		password = GenUserPwd()
	}
	if password == "" {
		return false, "wrong password", user
	}

	var reg_origin = origin
	if origin == "" {
		reg_origin = "web"
	}
	var origin_domain = ""
	var origin_keyword = ""
	var origin_spread = ""
	var origin_plat = ""
	// 竞价
	if jjSource != "" && jjCode != "" {
		adLinkInfo := models.GetAdLinkInfoByMap(map[string]interface{}{"spread_code": jjCode, "spread": jjSource, "status": 1})
		if adLinkInfo.Id > 0 {
			reg_origin = "bidding"
			origin_keyword = adLinkInfo.PromotionKeyword
			origin_spread = adLinkInfo.Name
			origin_plat = adLinkInfo.Spread
			// 获取竞价域名
			if adLinkInfo.JingjiaDomain != "" && adLinkInfo.JingjiaDomain == jjDomain {
				origin_domain = adLinkInfo.JingjiaDomain
			} else {
				origin_domain = jjDomain
			}
		}
	}

	user.Sn = devSn
	user.Email = email
	user.Username = username
	user.Nickname = nickname
	user.Password = util.PassEncode(password, username, 0)
	user.RegIp = ip
	user.RegRegion = ipInfo.String()
	user.RegCountry = ipInfo.Country
	user.CreateTime = nowTime
	user.Status = 1
	user.Platform = platform
	user.PlaintextPassword = password
	user.IsPay = "false"
	user.Origin = reg_origin
	user.Version = util.StoI(params.Version)
	user.BiddingDomain = origin_domain
	user.BiddingKeyword = origin_keyword
	user.BiddingCode = origin_spread
	user.BiddingPlat = origin_plat
	user.FrozenHour = 0
	user.DeviceToken = salt
	if thirdPlatform == "google" {
		user.GoogleUid = thirdPlatformUid
	} else if thirdPlatform == "github" {
		user.GithubId = util.StoI(thirdPlatformUid)
	} else if thirdPlatform == "mp_github" {
		user.MpGithubId = util.StoI(thirdPlatformUid)
	}

	user.IpStatus = 1
	err, id := models.AddUser(user)
	if err != nil {
		return false, "__T_REGISTER_FAIL", user
	}
	user.Id = id

	if inviter_id > 0 {
		uuid := GetUuid()
		invArr := strings.Split(uuid, "-")
		inviterCode := invArr[0]
		ratio := 0.02
		lWhere := map[string]interface{}{
			"cate":  "level",
			"level": 1,
		}
		levelInfo := models.GetConfLevelBy(lWhere, "level asc")
		ratio = levelInfo.Ratio
		// 添加邀请记录
		inviteInfo := models.UserInviter{}
		inviteInfo.Uid = id
		inviteInfo.Username = user.Username
		inviteInfo.Email = user.Email
		inviteInfo.Ip = user.RegIp
		inviteInfo.RegCountry = user.RegCountry
		inviteInfo.RegTime = user.CreateTime
		inviteInfo.Origin = util.StoI(inviter_type)
		inviteInfo.PayNum = 0
		inviteInfo.PayMoney = 0.0
		inviteInfo.Level = 1
		inviteInfo.Ratio = ratio
		inviteInfo.CreateTime = util.GetNowInt()
		var res, resAct error
		// 判断是否是活动邀请
		if strings.Contains(inviter_username, "act_") {
			//活动邀请上级写入活动邀请表
			if util.GetNowInt() < util.StoI(models.GetConfigVal("activity_inviter_end_time")) {
				inviteInfo.InviterCode = "act_" + strings.ToLower(inviterCode)
				inviteInfo.InviterId = inviter_id
				inviteInfo.InviterUsername = inviter_username
				resAct = models.CreateUserActivityInviter(inviteInfo)
			}

			//普通邀请表无上级
			inviteInfo.InviterCode = strings.ToLower(inviterCode)
			inviteInfo.InviterId = 0
			inviteInfo.InviterUsername = ""
			res = models.CreateUserInviter(inviteInfo)
		} else {
			//普通邀请活动邀请表无上级
			if util.GetNowInt() < util.StoI(models.GetConfigVal("activity_inviter_end_time")) {
				inviteInfo.InviterCode = "act_" + strings.ToLower(inviterCode)
				inviteInfo.InviterId = 0
				inviteInfo.InviterUsername = ""
				resAct = models.CreateUserActivityInviter(inviteInfo)
			}

			//上级写入普通邀请表
			inviteInfo.InviterCode = strings.ToLower(inviterCode)
			inviteInfo.InviterId = inviter_id
			inviteInfo.InviterUsername = inviter_username
			res = models.CreateUserInviter(inviteInfo)
		}
		fmt.Println("res", res)
		fmt.Println("resAct", resAct)

	}
	// 注册发送成功发送邮件
	date := util.Time2DateEn(nowTime)
	vars := make(map[string]string)
	vars["email"] = email
	vars["date"] = date
	dealSendEmail(3, email, vars, ip)

	//注册赠送券活动
	regConf := models.GetConfigVal("reg_coupon")
	if regConf == "1" {
		CouponInfoAuto(user, "register")
	}
	return true, "", user
}

// 登录
func DoLogin(user models.Users, email, ip string, signParam models.SignParam, updatePassword bool, c *gin.Context) (bool, string) {
	//ipInfo := util.GetIpInfo(ip)
	nowTime := util.GetNowInt()
	//删除之前信息
	models.DeleteSession(map[string]interface{}{
		"uid":      user.Id,
		"platform": signParam.Platform,
	})
	salt := signParam.Salt
	// 生成会话session
	sessionSn := util.Md5(util.RandStr("s", 16) + strconv.Itoa(int(nowTime)))
	// 写入session
	sessionExpireTimeStr := models.GetConfigVal("SessionExpireTime")
	sessionExpireTime := util.StoI(sessionExpireTimeStr)
	if sessionExpireTime == 0 {
		sessionExpireTime = setting.AppConfig.SessionExpire
	}
	addresult := models.AddLoginSession(models.Session{
		Enable:      "true",
		SessionId:   sessionSn,
		Uid:         user.Id,
		Version:     signParam.Version,
		LoginIp:     ip,
		Username:    user.Username,
		LoginTime:   nowTime,
		Platform:    signParam.Platform,
		ExpireTime:  nowTime + sessionExpireTime,
		DeviceToken: salt,
	})
	if !addresult {
		return false, "__APP_LOGIN_ERROR"
	}
	u := models.Users{
		LastLoginIp:   ip,
		LastLoginTime: nowTime,
	}
	if len(user.Email) == 0 && len(email) > 0 {
		u.Email = email
	}
	if updatePassword {
		u.Password = user.Password
	}
	var is_pay = 0
	if user.IsPay == "true" {
		is_pay = 1
	}
	// 用户登录日日志
	ua := c.Request.Header.Get("User-Agent")
	uaObj := user_agent.New(ua)
	browser, _ := uaObj.Browser()
	os := uaObj.OS()
	ipInfo, _ := ipdat.IPDat.GetIpInfo(ip)
	LogLogin := models.LogLogin{}
	LogLogin.UID = user.Id
	LogLogin.UserLogin = user.Username
	LogLogin.Ip = ip
	LogLogin.LoginTime = time.Now().Unix()
	LogLogin.Platform = signParam.DeviceOs
	LogLogin.RegTime = user.CreateTime
	LogLogin.Browser = browser
	LogLogin.UserAgent = ua
	LogLogin.Version = "1"
	LogLogin.OsInfo = os
	LogLogin.Country = ipInfo.Country
	LogLogin.Today = util.GetTodayTime()
	LogLogin.Language = c.Request.Header.Get("Accept-Language")
	LogLogin.Lang = signParam.Language
	LogLogin.Cate = "login"
	LogLogin.TimeZone = signParam.TimeZone
	LogLogin.DeviceNumber = signParam.DeviceNum
	LogLogin.IsPay = is_pay
	err := models.AddLoginLog(LogLogin)
	fmt.Println(err)

	// 更新用户信息
	models.UpdateUserById(user.Id, u)
	//fmt.Println("sessionSn-------" + sessionSn)

	// 检测用户券信息和发放券
	if user.IsPay != "true" {
		CouponInfoAuto(user, "new_user")
	} else {
		agentBalance := 0
		agentInfo := models.GetAgentBalanceByUid(user.Id)
		if agentInfo.Id != 0 {
			if agentInfo.Balance < 0 {
				agentInfo.Balance = 0
			}
			agentBalance = agentInfo.Balance
		}
		if agentBalance > 0 {
			CouponInfoAuto(user, "agent")
		}
	}
	return true, sessionSn
}

// 返回用户信息格式
func ResUserInfo(session, ip string, info models.Users) models.ResUser {

	nowTime := util.GetNowInt()
	var is_pay = 0
	if info.IsPay == "true" {
		is_pay = 1
	}
	nickname := util.ItoS(info.Id)
	if info.Nickname != "" {
		nickname = info.Nickname
	}

	googleAuth := 0
	username := info.Username
	gooKey := fmt.Sprintf(googleKey+"(%s)", username)
	err, authInfo := models.GetUserGoogleAuthBy(gooKey)
	if err == nil && authInfo.ID != 0 {
		googleAuth = 1
	}
	inviterCode := ""
	err, pUser := models.GetUserInviterByMap(map[string]interface{}{"uid": info.Id})
	level := 1
	if err != nil || pUser.ID == 0 {
		uuid := GetUuid()
		invArr := strings.Split(uuid, "-")
		inviterCode = invArr[0]
		ratio := 0.02
		lWhere := map[string]interface{}{
			"cate":  "level",
			"level": 1,
		}
		levelInfo := models.GetConfLevelBy(lWhere, "level asc")
		ratio = levelInfo.Ratio
		// 添加邀请记录
		inviteInfo := models.UserInviter{}
		inviteInfo.Uid = info.Id
		inviteInfo.Username = info.Username
		inviteInfo.Email = info.Email
		inviteInfo.Ip = info.RegIp
		inviteInfo.RegCountry = info.RegCountry
		inviteInfo.InviterCode = strings.ToLower(inviterCode)
		inviteInfo.RegTime = info.CreateTime
		inviteInfo.InviterId = 0
		inviteInfo.InviterUsername = ""
		inviteInfo.Origin = 0
		inviteInfo.PayNum = 0
		inviteInfo.PayMoney = 0.0
		inviteInfo.Level = level
		inviteInfo.Ratio = ratio
		inviteInfo.CreateTime = util.GetNowInt()
		res := models.CreateUserInviter(inviteInfo)
		fmt.Println("res", res)
	} else {
		if pUser.Email == "" {
			pUser.Email = info.Email
			pUser.Ip = info.RegIp
			pUser.RegCountry = info.RegCountry
			res := models.EditUserInviter(info.Id, pUser)
			fmt.Println("edit res", res)
		}
		inviterCode = pUser.InviterCode
	}
	links := strings.TrimSpace(models.GetConfigV("inviter_share_url")) + "?invite=" + inviterCode
	ip_nums := info.Balance
	agentBalance := 0
	agentInfo := models.GetAgentBalanceByUid(info.Id)
	if agentInfo.Id != 0 {
		if agentInfo.Balance < 0 {
			agentInfo.Balance = 0
		}
		agentBalance = agentInfo.Balance
	}
	// 获取流量信息
	flows := int64(0)
	flowDate := ""
	flowExpire := 0
	flowRedeem := "0"
	userFlowInfo := models.GetUserFlowInfo(info.Id)
	if userFlowInfo.ID != 0 {
		//if userFlowInfo.Flows > 0 { // 这里注释掉，因为有些用户流量 允许用户的流量为负数 20250114 需求
			flows = userFlowInfo.Flows
		//}
		flowDate = util.GetTimeStr(userFlowInfo.ExpireTime, "d/m/Y")
		if userFlowInfo.ExpireTime < nowTime {
			flowExpire = 1
		}
		redeemFlow := userFlowInfo.BuyFlow
		if redeemFlow > 0 {
			redeemFlow = userFlowInfo.BuyFlow - userFlowInfo.CdkFlow
			if redeemFlow > flows { //如果买的 大于 剩余的 就展示剩余的
				redeemFlow = flows
			}
		}

		flowRedeem, _ = DealFlowChar(redeemFlow, "GB")
	}
	flowStr, flowUnit := DealFlowChar(flows, "GB")
	flowMbStr, flowMbUnit := DealFlowChar(flows, "MB")
	// 新用户券信息
	isMsg := 0
	actOpen := strings.TrimSpace(models.GetConfigVal("pay_active_package")) //参与活动的套餐
	if actOpen != "" {
		actTime := strings.TrimSpace(models.GetConfigVal("pay_active_time")) //参与活动的时间
		if actTime != "" {
			hasPay := models.GetOrderCountByPakId(info.Id, util.StoI(actOpen), util.StoI(actTime))
			if hasPay > 0 {
				isMsg = 1
			}
		}

		if isMsg == 1 {
			isUse := 0
			_, couponList := models.GetCouponList(info.Id, "")
			for _, v := range couponList {
				if v.Status == 1 {
					isUse = isUse + 1
				} else {
					isUse = isUse + 0
				}
			}
			if isUse == 0 {
				isMsg = 0
			}
		}
	}

	// 生成返回数据
	data := models.ResUser{
		Session: session,
		User: models.ResUserInfo{
			ID:           info.Id,
			Username:     info.Username,
			Nickname:     nickname,
			Email:        info.Email,
			IsPay:        is_pay,
			Balance:      flows,
			IpNum:        ip_nums,
			GoogleAuth:   googleAuth,
			Invite:       inviterCode,
			InviteUrl:    links,
			AgentBalance: agentBalance,
			Flow:         flowStr,    //流量
			FlowRedeem:   flowRedeem, //可兑换cdk流量
			FlowUnit:     flowUnit,   //流量单位
			FlowMb:       flowMbStr,  //流量
			FlowMbUnit:   flowMbUnit, //流量单位
			FlowDate:     flowDate,   //流量过期时间
			FlowExpire:   flowExpire, //流量是否过期
			IsAct:        isMsg,      //是否活动弹窗
		},
	}
	return data
}

// 处理流量
func DealFlow(flows int64) (string, string) {
	//k_num := int64(1024)
	m_num := int64(1024 * 1024)
	g_num := int64(1024 * 1024 * 1024)

	flowNum := 0.00
	cate := "MB"
	if flows >= g_num {
		flowNum = math.Round(float64(flows) / float64(g_num))
		cate = "GB"
	} else if flows >= m_num {
		flowNum = float64(flows / m_num)
		cate = "MB"
	}
	info := fmt.Sprintf("%.0f", flowNum)
	return info, cate
}

// 处理流量
func DealFlowChar(flows int64, cate string) (string, string) {
	cate = strings.ToUpper(cate)
	if cate == "" {
		cate = "MB"
	}

	g_num := int64(1024 * 1024 * 1024)
	if cate == "KB" {
		g_num = int64(1024)
	}
	if cate == "MB" {
		g_num = int64(1024 * 1024)
	}

	flowNum := float64(flows) / float64(g_num)

	info := fmt.Sprintf("%.2f", flowNum)
	return info, cate
}

// 校验密码
func CheckPwd(password, userpass, username string) bool {
	return util.ChkPass(password, userpass, username)
}

// 限制注册数
func DealLimitReg(c *gin.Context, email, ip, saltStr string) (int, string, string) {
	// 获取配置
	ipRegOperateTimesStr := models.GetConfigVal("IpRegOperateTimes") //获取多长时间内操作的
	ipRegOperateTimes := util.StoI(ipRegOperateTimesStr)
	if ipRegOperateTimes == 0 {
		ipRegOperateTimes = 3600
	}
	// 获取配置
	ipRegUserNumberStr := models.GetConfigVal("IpRegUserNumber") //获取多长时间内操作的
	ipRegUserNumber := util.StoI(ipRegUserNumberStr)
	if ipRegUserNumber == 0 {
		ipRegUserNumber = 5
	}
	nowTime := util.GetNowInt()
	var start = nowTime - ipRegOperateTimes //多长时间
	num := models.GetUserCountByIp(ip, start)

	if num >= ipRegUserNumber {
		return -1, "", "__T_REG_TOO_MANY"
	}

	key := util.GetTodayHour()
	aes_key := util.Md5(util.ItoS(key) + "_reg_api_")
	fmt.Println(aes_key)
	bytes, err := util.AesDeCode(saltStr, []byte(aes_key))
	if err != nil {

		return -1, "", "__T_WAITING_TRY_1"
		//return 0, "", "ok"
	}
	strInfo := string(bytes)
	strArr := strings.Split(strInfo, ",")
	fmt.Println(strInfo)
	if len(strArr) < 2 {
		return -1, "", "__T_WAITING_TRY_2"
		//return 0, "", "ok"
	}
	salt := strArr[0]
	emailSalt := strArr[1]
	timeStamp := util.StoI(strArr[2])
	if emailSalt != email {
		fmt.Printf("emailSalt:%s,email:%s\n", emailSalt, email)
		return -1, "", "__T_REG_FAIL_MATCH"
	}
	nowTimeInt := util.GetNowInt()
	if nowTimeInt-timeStamp > 30 {
		fmt.Printf("timeStamp:%d,key:%d\n", timeStamp, nowTimeInt)
		return -1, "", "__T_REG_FAIL_MATCH"
	}
	// 查询Salt注册记录
	saltNum := models.GetUsersBySaltCount(salt, start)
	if saltNum >= ipRegUserNumber {
		return -1, "", "__T_REG_TOO_MANY"
	}
	return 0, salt, "ok"
}
