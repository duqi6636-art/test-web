package controller

import (
	"api-360proxy/web/models"
	"fmt"
	"github.com/gin-gonic/gin"
)

func DealOldUrl(c *gin.Context) {
	month := c.DefaultPostForm("month", "")
	if month == "" {
		JsonReturn(c, 0, "month empty", nil)
		return
	}

	tableName := "st_url_unlimited" + month
	list, err2 := models.GetOldDataList(tableName)
	fmt.Println(list)
	fmt.Println(err2)
	if err2 != nil {
		JsonReturn(c, 0, "data empty", nil)
		return
	}

	err1 := models.BatchInsertFlowDayLog(list, month)
	fmt.Println(err1)
	if err1 != nil {
		JsonReturn(c, 0, "insert err", nil)
		return
	}
	JsonReturn(c, 200, "success", list)
	return
}
