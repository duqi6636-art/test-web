package models

// Blog
type ResBlog struct {
	Id       int      `json:"id"`
	Title    string   `json:"title"`
	Img      string   `json:"img"`
	Abstract string   `json:"abstract"`
	Content  string   `json:"content"`
	CateId   string   `json:"cate_id"`
	CateArr  []string `json:"cate_arr"`
	Views    string   `json:"views"`
	Duration string   `json:"duration"`
}

// Blog
type Blog struct {
	Id         int    `json:"id"`
	Img        string `json:"img"`
	Title      string `json:"title"`
	TitleZh    string `json:"title_zh"`
	Abstract   string `json:"abstract"`    // 摘要
	AbstractZh string `json:"abstract_zh"` // 摘要-繁体
	Content    string `json:"content"`
	ContentZh  string `json:"content_zh"`
	CateId     string `json:"cate_id"`
	Views      int    `json:"views"`
	PushTime   int    `json:"push_time"`
	CreateTime int    `json:"create_time"`
}

// 查询Blog列表
func GetBlogListBy(cate_id int, title string) (lists []Blog, err error) {
	dbs := db.Table("cm_blog").Where("is_show=?", 1).Where("status=?", 1)
	if cate_id > 0 {
		dbs = dbs.Where("FIND_IN_SET(? ,cate_id)", cate_id)
		//dbs = dbs.Where("cate_id=?", cate_id)
	}
	if title != "" {
		dbs = dbs.Where("title like ?", "%"+title+"%")
	}
	err = dbs.Order("sort desc,id desc").Find(&lists).Error
	return
}

// 获取分页列表
func GetBlogListPage(cate_id int, title string, offset, limit int, sort string) (data []Blog) {
	dbs := db.Table("cm_blog").Where("is_show=?", 1).Where("status=?", 1)
	if cate_id > 0 {
		dbs = dbs.Where("FIND_IN_SET(? ,cate_id)", cate_id)
		//dbs = dbs.Where("cate_id=?", cate_id)
	}
	if title != "" {
		dbs = dbs.Where("title like ?", "%"+title+"%")
	}
	order := "push_time desc,id desc"
	if sort == "1" {
		order = "push_time desc,id desc"
	} else {
		order = "views desc,id desc"
	}
	dbs.Offset(offset).Limit(limit).Order(order).Find(&data)
	return
}

// Blog分类
type BlogCate struct {
	Id      int    `json:"id"`
	Title   string `json:"title"`
	TitleZh string `json:"title_zh"`
}

// 查询Blog分类
func GetBlogCate() (lists []BlogCate, err error) {
	err = db.Table("cm_blog_cate").Where("status = ?", 1).Order("sort desc").Find(&lists).Error
	return
}

// 查询Blog详情
func GetBlogById(id string) (info Blog, err error) {
	err = db.Table("cm_blog").Where("id = ?", id).First(&info).Error
	return
}

type BlogView struct {
	Id         int    `json:"id"`
	BlogId     int    `json:"blog_id"`
	Cate       int    `json:"cate"`
	Uid        int    `json:"uid"`
	Ip         string `json:"ip"`
	Lang       string `json:"lang"`
	CreateTime int    `json:"create_time"`
}

func AddBlogView(data BlogView) (err error) {
	err = db.Table("cm_blog_views").Create(&data).Error
	return
}

func EditBlogById(id int, info interface{}) error {
	err := db.Table("cm_blog").Where("id = ?", id).Updates(info).Error
	return err
}

// 用户指南
type Article struct {
	Id      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	CateId  int    `json:"cate_id"`
	Views   int    `json:"views"` //浏览数
}

// 查询用户指南分类
func GetArticleById(id string) (info Article, err error) {
	err = db.Table("cm_article").Where("status = ? and is_show = ?", 1, 1).Where("id = ?", id).Order("sort desc").Find(&info).Error
	return
}

func EditArticleById(id int, info interface{}) error {
	err := db.Table("cm_article").Where("id = ?", id).Updates(info).Error
	return err
}

// 查询FAQCate
func GetArticleFAQCateById(id string) (info Article, err error) {
	err = db.Table("cm_article_faq_cate").Where("status = ?", 1).Where("id = ?", id).Order("sort desc").Find(&info).Error
	return
}

func EditArticleFAQCateById(id int, info interface{}) error {
	err := db.Table("cm_article_faq_cate").Where("id = ?", id).Updates(info).Error
	return err
}

// 查询FAQ
func GetArticleVideoById(id string) (info Article, err error) {
	err = db.Table("cm_article_video").Where("status = ? and is_show = ?", 1, 1).Where("id = ?", id).Order("sort desc").Find(&info).Error
	return
}

func EditArticleVideoById(id int, info interface{}) error {
	err := db.Table("cm_article_video").Where("id = ?", id).Updates(info).Error
	return err
}

type ArticleView struct {
	Id         int    `json:"id"`
	ArticleId  int    `json:"article_id"`
	Cate       int    `json:"cate"`
	Uid        int    `json:"uid"`
	Num        int    `json:"num"`
	Ip         string `json:"ip"`
	Lang       string `json:"lang"`
	IsDel      int    `json:"id_del"`
	CreateTime int    `json:"create_time"`
}

func AddArticleView(data ArticleView) (err error) {
	err = db.Table("cm_article_views").Create(&data).Error
	return
}

func GetArticleView(articleId int) (data []ArticleView, err error) {
	err = db.Table("cm_article_views").Where("article_id =?", articleId).Find(&data).Error
	return
}
