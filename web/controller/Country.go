package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"github.com/gin-gonic/gin"
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
