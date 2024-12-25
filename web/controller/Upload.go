package controller

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"path"
	"strings"
	"sync"
)

type Uploads struct {
	Code int                    `json:"code"`
	Data map[string]interface{} `json:"data"`
	Msg  string                 `json:"msg"`
}

// @BasePath
// @Summary 上传单个图片
// @Description 上传单个图片
// @Tags 图片上传
// @Accept x-www-form-urlencoded
// @Produce json
// @Success 0 {object} interface{} ""fileUrl": """
// @Router /upload_img [post]
func UploadImage(c *gin.Context) {

	// 获取上传文件
	f, err := c.FormFile("file")
	fmt.Println("err1 = ", err)
	if err != nil {
		JsonReturn(c, -1, "获取上传文件失败!", nil)
		return
	}

	fd, err := f.Open()
	fmt.Println("err2 = ", err)
	if err != nil {

		JsonReturn(c, -1, "上传文件失败!", nil)
		return
	} else {
		// 获取并校验文件扩展名
		fileExt := strings.ToLower(path.Ext(f.Filename))
		if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".gif" && fileExt != ".jpeg" {

			JsonReturn(c, -1, "上传失败, 文件类型仅支持png,jpg,gif,jpeg", nil)
			return
		}

		resourceDomainLocal := models.GetConfigVal("resource_domain_local")
		url := resourceDomainLocal + "/upload_img"
		rep, _ := util.HttpPostMultiPart(url, "image", fd, f.Filename)
		res := Uploads{}

		json.Unmarshal([]byte(rep), &res)
		if res.Code == 0 {

			JsonReturn(c, 0, "success", map[string]interface{}{

				"fileUrl": res.Data["path_url"],
			})
			return
		}
		JsonReturn(c, -1, "上传文件失败!", nil)
		return
	}
}

// @BasePath
// @Summary 上传多个图片
// @Description 上传多个图片
// @Tags 图片上传
// @Accept x-www-form-urlencoded
// @Produce json
// @Success 0 {object} interface{} "fileUrls": """
// @Router /upload_multiple_images [post]
func UploadMultipleImages(c *gin.Context) {

	// 获取上传文件
	form, err := c.MultipartForm()
	if err != nil {
		JsonReturn(c, -1, "获取上传文件失败!", nil)
		return
	}
	files := form.File["file"]
	// 使用 WaitGroup 等待所有 goroutine 完成
	var wg sync.WaitGroup
	// 创建通道
	results := make(chan string, len(files))
	var fileUrls []string

	for _, f := range files {

		wg.Add(1)
		uploadImageAndGetFileUrl(results, f, &wg)
	}
	// 等待所有 worker goroutine 完成
	wg.Wait()

	close(results)
	for res := range results {

		if len(res) > 0 {
			fileUrls = append(fileUrls, res)
		}
	}
	JsonReturn(c, 0, "success", map[string]interface{}{

		"fileUrls": strings.Join(fileUrls, ","),
	})
}

func uploadImageAndGetFileUrl(results chan<- string, f *multipart.FileHeader, wg *sync.WaitGroup) {

	defer wg.Done()
	fd, err := f.Open()
	fmt.Println("err2 = ", err)
	if err != nil {

		results <- ""
		return
	} else {
		// 获取并校验文件扩展名
		fileExt := strings.ToLower(path.Ext(f.Filename))
		if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".gif" && fileExt != ".jpeg" {

			results <- ""
			return
		}
		resourceDomainLocal := models.GetConfigVal("resource_domain_local")
		url := resourceDomainLocal + "/upload_img"
		rep, _ := util.HttpPostMultiPart(url, "image", fd, f.Filename)
		res := Uploads{}

		json.Unmarshal([]byte(rep), &res)
		if res.Code == 0 {

			results <- fmt.Sprintf("%v", res.Data["path_url"])
			return
		}
		results <- ""
		return
	}
}
