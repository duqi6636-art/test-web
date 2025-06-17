package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
	"strings"
	"time"
)

// 绑定手机号

func BindPhone(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}

	phone := c.DefaultPostForm("phone", "")
	countryId := com.StrTo(c.DefaultPostForm("country_id", "0")).MustInt()
	if phone == "" {
		JsonReturn(c, e.ERROR, "__T_PARAM_PHONE", nil)
		return
	}

	if countryId <= 0 {
		JsonReturn(c, e.ERROR, "__T_PARAM_COUNTRY_ID", nil)
		return
	}
	now := time.Now().Unix()
	var bp = models.UserBindPhone{Uid: user.Id}
	bp.Get()
	if bp.Id <= 0 {
		// 创建
		bp = models.UserBindPhone{
			Uid:        user.Id,
			Phone:      phone,
			Status:     1,
			CountryId:  countryId,
			UpdateTime: now,
			CreateTime: now,
		}
		bp.Create()
	} else {
		// 更新
		bp.UpdateTime = now
		bp.Phone = phone
		bp.CountryId = countryId
		bp.Status = 1
		bp.Update(map[string]interface{}{"phone": phone, "country_id": countryId, "status": 1, "update_time": now})
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", bp)
}

func CancelBindPhone(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	var bp = models.UserBindPhone{Uid: user.Id}
	now := time.Now().Unix()
	bp.Update(map[string]interface{}{"status": -1, "update_time": now})
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", nil)
}

func GetBindPhone(c *gin.Context) {
	resCode, msg, user := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	var bp = models.UserBindPhone{Uid: user.Id}
	bp.Get()
	var country = models.Country{}
	if bp.Id > 0 && bp.CountryId > 0 {
		country = models.GetCountryById(bp.CountryId)
		if !strings.Contains(country.Phonecode, "+") {
			country.Phonecode = "+" + country.Phonecode
		}
	}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", map[string]interface{}{"bind_info": bp, "country": country})
}
