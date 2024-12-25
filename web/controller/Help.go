package controller

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"math"
	"strings"
	"time"
)

// blog搜索
// @BasePath /api/v1
// @Summary blog-搜索
// @Description blog-搜索
// @Tags blog相关
// @Accept x-www-form-urlencoded
// @Param cate_id formData string false "分类ID"
// @Param title formData string false "搜索标题"
// @Produce json
// @Success 0 {object} []models.ResBlog{}
// @Router /web/blog/get_search [post]
func GetSearchBlog(c *gin.Context) {
	idStr := strings.TrimSpace(c.DefaultPostForm("cate_id", ""))
	title := strings.TrimSpace(c.DefaultPostForm("title", ""))
	id := util.StoI(idStr)

	blogList, _ := models.GetBlogListBy(id, title)
	blogInfo := []models.ResBlog{}
	for _, va := range blogList {
		views := util.ItoS(va.Views)
		if va.Views > 1000 {
			views = fmt.Sprintf("%.2f", float64(va.Views)/1000) + " K"
		}
		vaInfo := models.ResBlog{}
		vaInfo.Id = va.Id
		vaInfo.Title = va.Title
		vaInfo.Img = va.Img
		vaInfo.CateId = va.CateId
		vaInfo.Abstract = va.Abstract
		vaInfo.Content = va.Content
		vaInfo.Views = views
		vaInfo.Duration = Time2DateEn(va.CreateTime)
		blogInfo = append(blogInfo, vaInfo)
	}
	JsonReturn(c, 0, "__T_SUCCESS", blogInfo)
	return
}

// blog分页加载
// @BasePath /api/v1
// @Summary blog-列表
// @Description blog-列表
// @Tags blog相关
// @Accept x-www-form-urlencoded
// @Param cate_id formData string false "分类ID"
// @Param lang formData string false "语言"
// @Param title formData string false "搜索标题"
// @Param sort formData string false "排序  1 发布时间  2 是热度"
// @Param limit formData string false "每页显示数量"
// @Param page formData string false "页码"
// @Produce json
// @Success 0 {object} map[string]interface{} "total：文章总数，total_page：总页数，lists：文章列表 []models.ResBlog{}"
// @Router /web/blog/list_page [post]
func GetBlogList(c *gin.Context) {
	idStr := strings.TrimSpace(c.DefaultPostForm("cate_id", ""))
	lang := strings.TrimSpace(c.DefaultPostForm("lang", ""))
	search := strings.TrimSpace(c.DefaultPostForm("title", ""))
	sort := strings.TrimSpace(c.DefaultPostForm("sort", "1")) //排序  1 发布时间  2 是热度
	limitStr := c.DefaultPostForm("limit", "10")
	pageStr := c.DefaultPostForm("page", "1")
	id := util.StoI(idStr)
	if limitStr == "" {
		limitStr = "10"
	}
	if pageStr == "" {
		pageStr = "1"
	}
	if sort == "" {
		sort = "1"
	}
	if lang == "zh-tw" || lang == "zh" || lang == "tw" || lang == "zh-cn" || lang == "cn" {
		lang = "zh-tw"
	}
	limit := util.StoI(limitStr)
	page := util.StoI(pageStr)
	offset := (page - 1) * limit
	blogList := models.GetBlogListPage(id, search, offset, limit, sort)

	//分类名称
	blogCateList, _ := models.GetBlogCate()
	cateMap := map[int]string{}
	for _, v := range blogCateList {
		title := v.Title
		if lang == "zh-tw" && v.TitleZh != "" {
			title = v.TitleZh
		}
		cateMap[v.Id] = title
	}

	blogInfo := []models.ResBlog{}
	for _, va := range blogList {
		views := util.ItoS(va.Views)
		lastDate := va.PushTime
		if lastDate == 0 {
			lastDate = va.CreateTime
		}
		cateArr := []string{}
		cateSp := strings.Split(va.CateId, ",")
		for _, v := range cateSp {
			cateId := util.StoI(v)
			cateName, ok := cateMap[cateId]
			if ok {
				cateArr = append(cateArr, cateName)
			}
		}
		title := va.Title
		abstract := va.Abstract
		content := va.Content
		if lang == "zh-tw" && va.TitleZh != "" {
			title = va.TitleZh
			abstract = va.AbstractZh
			content = va.ContentZh
		}
		vaInfo := models.ResBlog{}
		vaInfo.Id = va.Id
		vaInfo.Img = va.Img
		vaInfo.CateArr = cateArr
		vaInfo.CateId = va.CateId
		vaInfo.Title = title
		vaInfo.Abstract = abstract
		vaInfo.Content = content
		vaInfo.Views = views
		vaInfo.Duration = util.GetTimeStr(lastDate, "d-m-Y")
		blogInfo = append(blogInfo, vaInfo)
	}
	totalList, _ := models.GetBlogListBy(id, search)
	totalPage := int(math.Ceil(float64(len(totalList)) / float64(limit)))
	result := map[string]interface{}{
		"total":      len(totalList),
		"total_page": totalPage,
		"lists":      blogInfo,
	}
	JsonReturn(c, 0, "__T_SUCCESS", result)
	return
}

// Blog阅读数
// @BasePath /api/v1
// @Summary blog-获取阅读数
// @Description blog-获取阅读数
// @Tags blog相关
// @Accept x-www-form-urlencoded
// @Param id formData string true "文章ID"
// @Param cate formData string true "文章类型 1 文章 2 教程"
// @Param session formData string false "用户session"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {object} map[string]interface{} "views：阅读数"
// @Router /web/blog/views [post]
func GetViewBlog(c *gin.Context) {
	idStr := strings.TrimSpace(c.DefaultPostForm("id", ""))
	if idStr == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR-- id", nil)
		return
	}
	typeStr := strings.TrimSpace(c.DefaultPostForm("cate", "1"))
	sessionId := c.DefaultPostForm("session", "")
	lang := strings.ToLower(c.DefaultPostForm("lang", "")) //语言
	if lang == "" {
		lang = "en"
	}
	uid := 0
	if sessionId != "" {
		_, uid = GetUIDbySession(sessionId)
	}

	id := util.StoI(idStr)
	cate := util.StoI(typeStr)
	info, err := models.GetBlogById(idStr)
	if err != nil || info.Id == 0 {
		JsonReturn(c, 0, "__T_FAIL", gin.H{})
		return
	}
	ip := c.ClientIP()
	nowTime := util.GetNowInt()
	views := info.Views + 1

	add := models.BlogView{}
	add.Uid = uid
	add.Cate = cate
	add.Lang = lang
	add.BlogId = id
	add.Ip = ip
	add.CreateTime = nowTime
	_ = models.AddBlogView(add)
	upParam := map[string]interface{}{
		"views": views,
	}
	models.EditBlogById(id, upParam)

	res := map[string]interface{}{}
	res["views"] = views
	JsonReturn(c, 0, "__T_SUCCESS", res)
	return
}

// 文章教程阅读数
// @BasePath /api/v1
// @Summary 文章教程-获取阅读数
// @Description 文章教程-获取阅读数
// @Tags 文章教程相关
// @Accept x-www-form-urlencoded
// @Param id formData string true "文章ID"
// @Param cate formData string true "文章类型 1 文章 2 教程"
// @Param session formData string false "用户session"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {object} map[string]interface{} "views：阅读数"
// @Router /web/article/views [post]
func ViewArticle(c *gin.Context) {
	idStr := strings.TrimSpace(c.DefaultPostForm("id", ""))
	if idStr == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR-- id", nil)
		return
	}
	typeStr := strings.TrimSpace(c.DefaultPostForm("cate", "1"))
	sessionId := c.DefaultPostForm("session", "")
	lang := strings.ToLower(c.DefaultPostForm("lang", "")) //语言
	if lang == "" {
		lang = "en"
	}
	uid := 0
	if sessionId != "" {
		_, uid = GetUIDbySession(sessionId)
	}

	id := util.StoI(idStr)
	cate := util.StoI(typeStr)
	info, err := models.GetArticleById(idStr)
	if err != nil || info.Id == 0 {
		JsonReturn(c, 0, "__T_FAIL", gin.H{})
		return
	}
	ip := c.ClientIP()
	nowTime := util.GetNowInt()
	views := info.Views + 1

	add := models.ArticleView{
		ArticleId:  id,
		Cate:       cate,
		Uid:        uid,
		Ip:         ip,
		Lang:       lang,
		CreateTime: nowTime,
	}
	_ = models.AddArticleView(add)
	upParam := map[string]interface{}{
		"views": views,
	}
	models.EditArticleById(id, upParam)

	res := map[string]interface{}{}
	res["views"] = views
	JsonReturn(c, 0, "__T_SUCCESS", res)
	return
}


// 常见问题阅读数
// @BasePath /api/v1
// @Summary 常见问题-获取阅读数
// @Description 常见问题-获取阅读数
// @Tags 常见问题相关
// @Accept x-www-form-urlencoded
// @Param id formData string true "文章ID"
// @Param cate formData string true "文章类型 1 文章 2 教程"
// @Param session formData string false "用户session"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {object} map[string]interface{} "views：阅读数"
// @Router /web/article_video/views [post]
func GetViewVideo(c *gin.Context) {
	idStr := strings.TrimSpace(c.DefaultPostForm("id", ""))
	if idStr == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR-- id", nil)
		return
	}
	id := util.StoI(idStr)
	info, err := models.GetArticleVideoById(idStr)
	if err != nil || info.Id == 0 {
		JsonReturn(c, 0, "__T_FAIL", gin.H{})
		return
	}
	views := info.Views + 1
	upParam := map[string]interface{}{
		"views": views,
	}
	models.EditArticleVideoById(id, upParam)

	res := map[string]interface{}{}
	res["views"] = views
	JsonReturn(c, 0, "__T_SUCCESS", res)
	return
}
// 常见问题阅读数
// @BasePath /api/v1
// @Summary 常见问题-获取阅读数
// @Description 常见问题-获取阅读数
// @Tags 常见问题相关
// @Accept x-www-form-urlencoded
// @Param id formData string true "文章ID"
// @Param cate formData string true "文章类型 1 文章 2 教程"
// @Param session formData string false "用户session"
// @Param lang formData string false "语言"
// @Produce json
// @Success 0 {object} map[string]interface{} "views：阅读数"
// @Router /web/faq/views [post]
func GetViewFAQ(c *gin.Context) {
	idStr := strings.TrimSpace(c.DefaultPostForm("id", ""))
	if idStr == "" {
		JsonReturn(c, -1, "__T_PARAM_ERROR-- id", nil)
		return
	}
	id := util.StoI(idStr)
	info, err := models.GetArticleFAQCateById(idStr)
	if err != nil || info.Id == 0 {
		JsonReturn(c, 0, "__T_FAIL", gin.H{})
		return
	}
	views := info.Views + 1
	upParam := map[string]interface{}{
		"views": views,
	}
	models.EditArticleFAQCateById(id, upParam)

	res := map[string]interface{}{}
	res["views"] = views
	JsonReturn(c, 0, "__T_SUCCESS", res)
	return
}
// 时间戳转 英文日期
func Time2DateEn(t int) string {
	now_time := time.Now()
	now_time = time.Unix(int64(t), 0)
	m := now_time.Month()
	month := m.String()
	year := now_time.Year()
	day := now_time.Day()
	date := util.ItoS(day) + "th-" + month + " " + util.ItoS(year)
	return date
}
