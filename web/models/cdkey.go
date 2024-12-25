package models

import (
	"api-360proxy/web/pkg/util"
	"strings"
)

type MdCdkey struct {
	Id           int    `json:"id"`
	Uid          int    `json:"uid"`
	Cate         string `json:"cate"`
	Cdkey        string `json:"cdkey"`
	Number       int64  `json:"number"`
	BindUsername string `json:"bind_username"`
	BindUid      int    `json:"bind_uid"`
	BindTime     int    `json:"bind_time"`
	Status       int    `json:"status"`
	Create_time  int    `json:"create_time"`
	Locked       int    `json:"locked"` //是否锁定  1 锁定
}

type MdCdkeyJoinUser struct {
	Id               int    `json:"id"`
	Cid              int    `json:"cid"`
	Cate             int    `json:"cate"`
	Uid              int    `json:"uid"`
	Code             string `json:"code"`
	Name             string `json:"name"`
	BindUid          int    `json:"bind_uid"`
	BindUsername     string `json:"bind_username"`
	Status           int    `json:"status"`   // 状态:1生效 2已使用
	UseTime          int    `json:"use_time"` // 使用时间
	Title            string `json:"title"`
	Value            int64  `json:"value"`      //
	UserType         string `json:"user_type"`  // 用户类型 payed 已付费 no_pay 未付费
	UseType          int    `json:"use_type"`   // 使用类型 1单次使用(单个券应单个用户使用)  2重复使用(单个券对应多个用户)
	Expire           int    `json:"expire"`     // 过期时间
	ExpiryDay        int    `json:"expiry_day"` // 过期天数  0永不过期
	UseCycle         int    `json:"use_cycle"`
	UseNumber        int    `json:"use_number"`
	Platform         int    `json:"platform"` //平台ID
	GroupId          int    `json:"group_id"` //分组ID
	CreateTime       int    `json:"create_time"`
	Balance          int64  `json:"balance"` //用户余额
	Email            string `json:"email"`
	GenerateRemark   string `json:"generate_remark"`   //生成备注
	RedemptionRemark string `json:"redemption_remark"` //兑换备注
}

/// 生成cdk返回
type ResGenerateList struct {
	Id             string `json:"id"`
	ExchangeType   string `json:"exchange_type"` // 兑换类型
	CdkKey         string `json:"cdK_key"`
	Email          string `json:"email"`
	Value          string `json:"value"`
	UsageMode      int    `json:"usage_mode"`      // 模式: 0-cdk兑换 1-储值
	Status         int    `json:"status"`          // 状态: 1-未兑换 2-已兑换
	GenerateTime   string `json:"generate_time"`   // 生成时间
	RedemptionTime string `json:"redemption_time"` // 提取时间
	Balance        string `json:"balance"`         //用户余额
	Remark         string `json:"remark"`          //备注

}

/// 兑换cdk返回
type ResRedemptionList struct {
	Id             int    `json:"id"`
	ExchangeType   string `json:"exchange_type"`   // 兑换类型
	RedemptionTime string `json:"redemption_time"` // 提取时间
	Value          string `json:"value"`
	CdkKey         string `json:"cdK_key"`
	UsageMode      int    `json:"usage_mode"` // 模式: 0-cdk兑换 1-储值
	Remark         string `json:"remark"`     //备注
}

// 获取cdk生成列表 根据类型
func GetGenerateListByCate(uid, start, end, cate int, mode, timeType, cdkEmail, status,cateStr string) (data []MdCdkeyJoinUser) {
	fields := "a.*,us.balance,us.email"
	joinTable := CmUserTable
	if cate == 3 {
		fields = "a.*,us.flows as balance,us.email"
		joinTable = userFlowTable
	}
	if cate == 4 {
		fields = "a.*,us.flows as balance,us.email"
		joinTable = userDynamicIspTable
	}
	if cate == 5 {
		fields = "a.*,us.expire_time as balance,us.email"
		joinTable = "cm_user_flow_day"
	}
	//pakId := 0
	exDay := 0
	if cate == 6 {
		packageList := GetStaticPackageList()
		packArr := map[int]int{}
		for _, v := range packageList {
			packArr[v.Value] = v.Id
		}
		exArr := strings.Split(cateStr, "-")
		exDay = util.StoI(exArr[1])
		if  exDay == 0 {
			exDay = 7
		}
		//var ok bool
		//pakId,ok = packArr[exDay]
		//if !ok {
		//	pakId = packageList[0].Id
		//}
		fields = "a.*,us.balance as balance,us.email"
		joinTable = "cm_user_static_ip"
	}

	dbtt := db.Table("cm_exchange_list" + " as a").Select(fields).Joins("left join " + joinTable + " as us on a.bind_uid=us.uid")
	if cate == 2 {
		dbtt = db.Table("cm_exchange_list" + " as a").Select(fields).Joins("left join " + joinTable + " as us on a.bind_uid=us.id")
	}
	if cate == 6 {
		dbtt = db.Table("cm_exchange_list" + " as a").Select(fields).Joins("left join " + joinTable + " as us on a.bind_uid=us.uid and a.region=us.pak_region and a.day=us.expire_day")
	}

	dbtt = dbtt.Where("a.uid = ?", uid)
	dbtt = dbtt.Where("a.mode = ?", mode)
	dbtt = dbtt.Where("a.cate = ?", cate)

	if mode == "balance" {
		dbtt = dbtt.Where("a.name = ?", "cdk")
	}
	if cdkEmail != "" {
		dbtt = dbtt.Where("a.code = ? OR us.email = ?", cdkEmail, cdkEmail)
	}
	if timeType == "0" {
		if start > 0 {
			dbtt = dbtt.Where("a.create_time >= ?", start)
		}
		if end > 0 {
			dbtt = dbtt.Where("a.create_time < ?", end+86400)
		}
	} else {
		if start > 0 {
			dbtt = dbtt.Where("a.use_time >= ?", start)
		}
		if end > 0 {
			dbtt = dbtt.Where("a.use_time < ?", end+86400)
		}
	}
	//静态IP
	if cate == 6 {
		dbtt = dbtt.Where("a.day = ?", exDay)
	}

	if status != "0" && status != "" {
		dbtt = dbtt.Where("a.status = ?", status)
	}
	dbtt = dbtt.Order("id desc").Find(&data)
	return
}

// 获取cdk兑换列表-isp
func GetRedemptionList(uid, start, end, cate int, mode, cdk, status string) (data []MdCdkeyJoinUser) {
	dbtt := db.Table("cm_exchange_list")
	dbtt = dbtt.Where("bind_uid = ?", uid)
	if cate > 0 {
		dbtt = dbtt.Where("cate = ?", cate)
	}

	if cdk != "" {
		dbtt = dbtt.Where("code = ?", cdk)
	}
	if mode != "" {
		dbtt = dbtt.Where("mode = ?", mode)
	}

	if start > 0 {
		dbtt = dbtt.Where("use_time >= ?", start)
	}
	if end > 0 {
		dbtt = dbtt.Where("use_time < ?", end+86400)
	}

	if status == "1" {
		dbtt = dbtt.Where("name = ?", "cdk")
	} else if status == "2" {
		if mode == "balance" {
			dbtt = dbtt.Where("name = ?", "self")
		} else {
			dbtt = dbtt.Where("name = ?", "to_user")
		}
	}
	dbtt = dbtt.Order("id desc").Find(&data)
	return
}

type StCdkeyStatsLog struct {
	Id           int    `json:"id"`
	Uid          int    `json:"uid"`
	Cate         string `json:"cate"`
	Mode         string `json:"mode"`
	Value        int64  `json:"value"`
	Number       int    `json:"number"`
	BindUsername string `json:"bind_username"`
	BindUid      int    `json:"bind_uid"`
	LastTime     int    `json:"last_time"`
	IsCollect    int    `json:"is_collect"`
	Remark       string `json:"remark"`
}
type StCdkeyStatsLogUser struct {
	Id           int    `json:"id"`
	Uid          int    `json:"uid"`
	Cate         string `json:"cate"`
	Value        int64  `json:"value"`
	Number       int    `json:"number"`
	BindUsername string `json:"bind_username"`
	BindUid      int    `json:"bind_uid"`
	LastTime     int    `json:"last_time"`
	IsCollect    int    `json:"is_collect"`
	Remark       string `json:"remark"`
	Email        string `json:"email"`
}

type StCdkeyStatsJoinUser struct {
	Id           int    `json:"id"`
	Uid          int    `json:"uid"`
	Cate         string `json:"cate"`
	Value        int64  `json:"value"`
	Number       int    `json:"number"`
	BindUsername string `json:"bind_username"`
	BindUid      int    `json:"bind_uid"`
	LastTime     int    `json:"last_time"`
	IsCollect    int    `json:"is_collect"`
	Remark       string `json:"remark"`
	Balance      int    `json:"balance"` //用户余额
	Flows        int    `json:"flow"`    //用户流量余额
	Email        string `json:"email"`
}

/// 用户使用日志返回
type ResUserUsageList struct {
	Id                 string `json:"id"`
	Email              string `json:"email"`
	ExchangeType       string `json:"exchange_type"`
	Balance            string `json:"balance"` //用户余额
	Value              string `json:"value"`
	Times              int    `json:"times"`
	LastRedemptionTime string `json:"last_redemption_time"` // 最后兑换时间
	Remark             string `json:"remark"`               //备注
	Collect            int    `json:"collect"`              //是否收藏
}

// 获取信息 By Id
func GetCdkUseStatsById(id int) (info StCdkeyStatsLog, err error) {
	err = db.Table("st_cdkey_use_log").Where("id =?", id).First(&info).Error
	return
}

// 获取信息 By 用户信息
func GetCdkUseStatsByUser(uid, bindUid int, cate, mode string, start, end int) (info []StCdkeyStatsLog) {
	dbtt := db.Table("st_cdkey_use_log")
	if uid > 0 {
		dbtt = dbtt.Where("uid = ?", uid)
	}
	if bindUid > 0 {
		dbtt = dbtt.Where("bind_uid = ?", bindUid)
	}

	if cate != "" {
		dbtt = dbtt.Where("cate = ?", cate)
	}
	if mode != "" {
		dbtt = dbtt.Where("mode = ?", mode)
	}
	if start > 0 {
		dbtt = dbtt.Where("last_time >= ?", start)
	}
	if end > 0 {
		dbtt = dbtt.Where("last_time < ?", end+86400)
	}
	dbtt.Find(&info)
	return
}

// 更新信息
func EditCdkUseStatsById(id int, params interface{}) (err error) {
	err = db.Table("st_cdkey_use_log").Where("id =?", id).Update(params).Error
	return
}

// 获取联表数据分页列表
// code, 使用模式  agent 代理商 balance 余额  score 积分 inviter 邀请
// cate, 类型   isp  flow   unlimited  dynamic_isp
func GetCdkUseStatsJoin(uid int, mode, cate, email string, is_collect, start, end int, sortType, sort string) (data []StCdkeyStatsJoinUser) {
	fields := "a.*,us.balance,us.email"
	joinTable := CmUserTable
	if cate == "flow" {
		fields = "a.*,us.flows as flows"
		joinTable = userFlowTable
	}
	if cate == "dynamic_isp" {
		fields = "a.*,us.flows as flows"
		joinTable = userDynamicIspTable
	}
	if cate == "unlimited" {
		fields = "a.*,us.expire_time as flows"
		joinTable = "cm_user_flow_day"
	}

	dbtt := db.Table("st_cdkey_use_log" + " as a").Select(fields).Joins("left join " + joinTable + " as us on a.bind_uid=us.uid")
	if cate == "isp" {
		dbtt = db.Table("st_cdkey_use_log" + " as a").Select(fields).Joins("left join " + joinTable + " as us on a.bind_uid=us.id")
	}

	dbtt = dbtt.Where("a.uid = ?", uid)
	if mode != "" {
		dbtt = dbtt.Where("a.mode = ?", mode)
	}
	if cate != "" {
		dbtt = dbtt.Where("a.cate = ?", cate)
	}
	if email != "" {
		dbtt = dbtt.Where("a.email = ?", email)
	}
	if is_collect > 0 {
		if is_collect == 2 {
			dbtt = dbtt.Where("a.is_collect = ?", 2)
		} else {
			dbtt = dbtt.Where("a.is_collect <> ?", 2)
		}
	}
	if start > 0 {
		dbtt = dbtt.Where("a.last_time >= ?", start)
	}
	if end > 0 {
		dbtt = dbtt.Where("a.last_time < ?", end+86400)
	}
	if sortType != "" && sort != "" {
		if sortType == "0" {
			if sort == "0" {
				if cate == "flow" {
					dbtt = dbtt.Order("us.flows asc").Find(&data)
				} else {
					dbtt = dbtt.Order("us.balance asc").Find(&data)
				}
			} else if sort == "1" {
				if cate == "flow" {
					dbtt = dbtt.Order("us.flows desc").Find(&data)
				} else {
					dbtt = dbtt.Order("us.balance desc").Find(&data)
				}
			}
		} else if sortType == "1" {
			if sort == "0" {
				dbtt = dbtt.Order("a.value asc").Find(&data)
			} else if sort == "1" {
				dbtt = dbtt.Order("a.value desc").Find(&data)
			}
		} else if sortType == "2" {
			if sort == "0" {
				dbtt = dbtt.Order("a.number asc").Find(&data)
			} else if sort == "1" {
				dbtt = dbtt.Order("a.number desc").Find(&data)
			}
		} else if sortType == "3" {
			if sort == "0" {
				dbtt = dbtt.Order("a.last_time asc").Find(&data)
			} else if sort == "1" {
				dbtt = dbtt.Order("a.last_time desc").Find(&data)
			}
		}
	} else {
		dbtt = dbtt.Order("id desc").Find(&data)
	}

	return
}

func GetCdkUseStatsJoin_bak(uid int, cate, email string, is_collect, start, end int, sortType, sort string) (data []StCdkeyStatsJoinUser) {
	dbtt := db.Table("st_cdkey_use_log as a").Select("a.*,u.balance,u.email").Joins("left join " + CmUserTable + " as u on a.bind_uid=u.id")
	if cate == "flow" {
		dbtt = db.Table("st_cdkey_use_log as a").Select("a.*,us.flows,us.email").Joins("left join " + userFlowTable + " as us on a.bind_uid=us.id")
	}
	dbtt = dbtt.Where("a.uid = ?", uid)
	if cate != "" {
		dbtt = dbtt.Where("a.cate = ?", cate)
	}
	if email != "" {
		dbtt = dbtt.Where("u.email = ?", email)
	}
	if is_collect > 0 {
		if is_collect == 2 {
			dbtt = dbtt.Where("a.is_collect = ?", 2)
		} else {
			dbtt = dbtt.Where("a.is_collect <> ?", 2)
		}
	}
	if start > 0 {
		dbtt = dbtt.Where("a.last_time >= ?", start)
	}
	if end > 0 {
		dbtt = dbtt.Where("a.last_time < ?", end+86400)
	}
	if sortType != "" && sort != "" {
		if sortType == "0" {
			if sort == "0" {
				if cate == "flow" {
					dbtt = dbtt.Order("us.flows asc").Find(&data)
				} else {
					dbtt = dbtt.Order("u.balance asc").Find(&data)
				}
			} else if sort == "1" {
				if cate == "flow" {
					dbtt = dbtt.Order("us.flows desc").Find(&data)
				} else {
					dbtt = dbtt.Order("u.balance desc").Find(&data)
				}
			}
		} else if sortType == "1" {
			if sort == "0" {
				dbtt = dbtt.Order("a.value asc").Find(&data)
			} else if sort == "1" {
				dbtt = dbtt.Order("a.value desc").Find(&data)
			}
		} else if sortType == "2" {
			if sort == "0" {
				dbtt = dbtt.Order("a.number asc").Find(&data)
			} else if sort == "1" {
				dbtt = dbtt.Order("a.number desc").Find(&data)
			}
		} else if sortType == "3" {
			if sort == "0" {
				dbtt = dbtt.Order("a.last_time asc").Find(&data)
			} else if sort == "1" {
				dbtt = dbtt.Order("a.last_time desc").Find(&data)
			}
		}
	} else {
		dbtt = dbtt.Order("id desc").Find(&data)
	}

	return
}

// 获取联表数据分页列表
// code, 使用模式  agent 代理商 balance 余额  score 积分 inviter 邀请
// cate, 类型   isp  flow   unlimited  dynamic_isp
func GetCdkUseStats(uid int, mode, cate, email string, is_collect, start, end int, sortType, sort string) (data []StCdkeyStatsLogUser) {

	fields := "a.*,us.email"
	joinTable := CmUserTable

	dbtt := db.Table("st_cdkey_use_log" + " as a").Select(fields).Joins("left join " + joinTable + " as us on a.bind_uid=us.id")
	dbtt = dbtt.Where("a.uid = ?", uid)
	if mode != "" {
		dbtt = dbtt.Where("a.mode = ?", mode)
	}
	if cate != "" {
		dbtt = dbtt.Where("a.cate = ?", cate)
	}
	if email != "" {
		dbtt = dbtt.Where("us.email = ?", email)
	}
	if is_collect > 0 {
		if is_collect == 2 {
			dbtt = dbtt.Where("a.is_collect = ?", 2)
		} else {
			dbtt = dbtt.Where("a.is_collect <> ?", 2)
		}
	}
	if start > 0 {
		dbtt = dbtt.Where("a.last_time >= ?", start)
	}
	if end > 0 {
		dbtt = dbtt.Where("a.last_time < ?", end+86400)
	}
	if sortType != "" && sort != "" {
		if sortType == "1" {
			if sort == "0" {
				dbtt = dbtt.Order("a.value asc").Find(&data)
			} else if sort == "1" {
				dbtt = dbtt.Order("a.value desc").Find(&data)
			}
		} else if sortType == "2" {
			if sort == "0" {
				dbtt = dbtt.Order("a.number asc").Find(&data)
			} else if sort == "1" {
				dbtt = dbtt.Order("a.number desc").Find(&data)
			}
		} else if sortType == "3" {
			if sort == "0" {
				dbtt = dbtt.Order("a.last_time asc").Find(&data)
			} else if sort == "1" {
				dbtt = dbtt.Order("a.last_time desc").Find(&data)
			}
		}
	} else {
		dbtt = dbtt.Order("a.id desc").Find(&data)
	}

	return
}
