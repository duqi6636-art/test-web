package controller

import (
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

/// 刷新所有用户的会员等级
func FreshUserMemberLevel(c *gin.Context) {

	// 查询所有消费金额高于会员1的用户
	memberLevel := models.GetMemberLevelById(2)
	// 查询所有用户消费金额大于等于会员金额的
	err, users := models.GetUserListByPayMoney(memberLevel.MaxMoney)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
	}
	valueArgs := []string{}
	// 编译用户如果没在会员列表里面就插入
	for _, user := range users {

			//// 查询用户会员当前等级
			level := models.GetMemberLevelByMoney(user.PayMoney)
			if level.Id != 0 {

				str := fmt.Sprintf("('%s',%d,%d,'%s','%s',%f,%d,%d,%d,%d)", level.Name, level.Id, user.Id, user.Email, user.Username, user.PayMoney, user.PayNumber, 0, util.GetNowInt(), util.GetNowInt())
				valueArgs = append(valueArgs, str)
			}
	}

	if len(valueArgs) > 0 {

		err := models.AddUserMembers(valueArgs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
		}
	}


	JsonReturn(c, 0, "success", nil)
	return
}