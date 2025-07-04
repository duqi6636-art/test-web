package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
)

// 国家或国家代码搜索
// @BasePath /api/v1
// @Summary 国家或国家代码搜索
// @Description 国家或国家代码搜索
// @Tags 帮助中心
// @Accept x-www-form-urlencoded
// @Param code formData string false "国家 / 国家代码"
// @Produce json
// @Success 0 {object} map[string]interface{} "lists：国家列表"
// @Router /web/get_country_code [post]
func GetCountryCode(c *gin.Context) {
	code := strings.TrimSpace(c.DefaultPostForm("code", ""))
	//limitStr := c.DefaultPostForm("limit","20")
	//pageStr := c.DefaultPostForm("page","1")
	//if limitStr == "" {
	//	limitStr = "20"
	//}
	//if pageStr == "" {
	//	pageStr = "1"
	//}
	code = strings.ToLower(code)
	//limit := util.StoI(limitStr)
	//page := util.StoI(pageStr)
	//offset := (page - 1) * limit
	//lists := models.GetCountryPage(offset,limit,code)
	lists := models.GetCountry(code)
	//total := models.GetCountryCount(code)
	//totalPage := int(math.Ceil(float64(total)/float64(limit)))
	result := map[string]interface{}{
		//"total" 		: total,
		//"total_page" 	: totalPage,
		"lists": lists,
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", result)
	return
}

func GetCountryList(c *gin.Context) {
	list := models.GetCountryList("is_verify = ?", 1)
	for i, country := range list {
		if !strings.Contains(country.Phonecode, "+") {
			country.Phonecode = "+" + country.Phonecode
		}
		list[i] = country
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", list)
}

// 上传文件

func CountryListUpdateImage(c *gin.Context) {
	list := models.GetCountryList("is_verify = ?", 1)
	rsp := map[int]interface{}{}
	for _, country := range list {
		rqt, err := http.NewRequest("GET", country.Flag, nil)
		if err != nil {
			JsonReturn(c, e.ERROR, "创建请求失败", err)
			return
		}
		client := &http.Client{}
		resp, err := client.Do(rqt)
		if err != nil {
			JsonReturn(c, e.ERROR, "请求失败", err)
			return
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			JsonReturn(c, e.ERROR, "读取body失败", err)
			return
		}
		log.Println(string(b))
		resourceDomainLocal := models.GetConfigVal("resource_domain_local")
		url := resourceDomainLocal + "/upload_img"
		lu := strings.Split(country.Flag, "/")
		fileName := lu[len(lu)-1]

		//rep, _ := util.HttpPostMultiPart(url, "image", tmpFile, fileName)

		var requestBody bytes.Buffer
		writer := multipart.NewWriter(&requestBody)

		// 创建表单中的文件字段
		fileWriter, err := writer.CreateFormFile("file", fileName)
		if err != nil {
			JsonReturn(c, e.ERROR, "CreateFormFile", err)
			return
		}
		// 将文件内容复制到表单中的文件字段
		fileWriter.Write(b)
		// 必须关闭 writer，以便写入结尾的 boundary
		writer.Close()
		// 创建 POST 请求
		req, err := http.NewRequest("POST", url, &requestBody)
		if err != nil {
			JsonReturn(c, e.ERROR, "创建post请求11", err)
			return
		}
		// 设置请求头 Content-Type
		req.Header.Set("Content-Type", writer.FormDataContentType())
		// 发送请求并获取响应
		client = &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			JsonReturn(c, e.ERROR, "client.Do", err)
			return
		}
		defer resp.Body.Close()

		result, err := ioutil.ReadAll(resp.Body)

		res := Uploads{}
		json.Unmarshal(result, &res)
		if res.Code == 0 {
			// 更新url
			if res.Data["path_url"] != "" {
				models.UpdateCountryWhere(map[string]interface{}{"flag": res.Data["path_url"]}, "id = ?", country.Id)
				//rsp[country.Id] = res.Data["path_url"]
			}

		}
		rsp[country.Id] = res.Data["path_url"]
	}
	JsonReturn(c, e.SUCCESS, "__SUCCESS", rsp)

}
