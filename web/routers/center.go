package routers

import (
	"api-360proxy/web/controller"
	"github.com/gin-gonic/gin"
)

func centerRouter(router *gin.Engine) {

	// 购买相关
	mig := router.Group("/migrate")
	mig.POST("/flow_day", controller.DealOldUrl) // 获取用户购买记录

	// 购买相关
	pur := router.Group("/center/purchase")
	pur.POST("/record", controller.UserOrderListBy) // 获取用户购买记录

	pack := router.Group("/center/package")
	pack.POST("/flow", controller.GetPackageFlow)                  // 获取流量列表
	pack.POST("/dynamic_isp", controller.GetDynamicISPPackageList) // 获取动态isp套餐列表
	pack.POST("/flow_day", controller.GetFlowDayPackageList)       // 获取不限量套餐列表
	pack.POST("/socks5", controller.GetSocks5PackageList)          // 获取ip套餐列表
	pack.POST("/static", controller.GetStaticPackage)              // 获取静态长效套餐列表
	pack.POST("/low_price", controller.GetLowPrice)                // 获取各套餐最低价格

	flows := router.Group("/center/flows")
	flows.POST("/get_info", controller.GetAccountInfoV2) // 获取流量信息

	//api提取
	//对外
	router.GET("/api/extract_ip", controller.ExtractIp) // 提取IP
	//router.GET("/api/add_ip", controller.FlowApiAddWhite)     // 添加
	//router.GET("/api/del_ip", controller.FlowApiDelWhite)     // 删除
	//router.GET("/api/lists_ip", controller.FlowApiListsWhite) // 查询列表
	//对内
	white := router.Group("/center/flow_api")
	white.POST("/info", controller.FlowApiInfo)                  //基础信息
	white.POST("/lists", controller.FlowApiWhitelist)            //白名单IP列表
	white.POST("/add", controller.AddFlowApiWhite)               //白名单IP列表--添加
	white.POST("/set_white", controller.SetFlowApiWhite)         //白名单IP列表禁用启用
	white.POST("/delete", controller.DelFlowApiWhite)            //白名单IP列表--删除
	white.POST("/exist_white_list", controller.ExistWhiteList)   // 是否存在白名单IP
	white.POST("/download", controller.FlowApiWhitelistDownload) // 下载白名单IP列表
	white.POST("/domain", controller.ApiDomain)                  // 白名单域名

	//优惠券
	coupon := router.Group("/center/coupon")
	coupon.POST("/my_coupons", controller.GetMyCoupons)                // 获取我的优惠券列表
	coupon.POST("/my_coupons_list", controller.GetMyCouponListByPakId) // 获取优惠券下拉列表接口
	coupon.POST("/popup", controller.GetCouponPopup)                   // 优惠券弹窗
	coupon.POST("/popup_click", controller.ClickCouponPopup)           // 优惠券弹窗点击
	coupon.POST("/get_coupon", controller.GetCoupon)                   // 领取优惠卷

	cdk := router.Group("/center/cdk")
	cdk.POST("/ex/generate_list", controller.NewGenerateList)                  // cdk 生成列表
	cdk.POST("/ex/redemption_list", controller.NewRedemptionList)              // cdk 兑换列表
	cdk.POST("/ex/stats_use", controller.GetUserCdkStats)                      // cdk 统计使用记录
	cdk.POST("/ex/stats_collect", controller.GetUserCdkCollect)                // cdk 统计收藏
	cdk.POST("/ex/stats_remark", controller.GetUserCdkRemark)                  // cdk 统计备注
	cdk.POST("/ex/stats_download", controller.GetUserCdkStatsDownload)         // 统计使用记录下载
	cdk.POST("/ex/generate_list_download", controller.GetGenerateListDownload) // 生成列表记录下载

	cdk.POST("/balance_batch", controller.BalanceGenerateBatchCdk) // 批量生成CDK
	cdk.POST("/balance_self_use", controller.SelfBalanceCdk)       // 自用生成CDK
	cdk.POST("/ex_dynamic", controller.ExchangeDynamicCdk)         // 兑换动态ISP CDK
	cdk.POST("/ex_unlimited", controller.ExchangeUnlimitedCdk)     // 兑换不限量CDK

	// 余额相关
	cb := router.Group("/center/balance")
	cb.POST("/config", controller.BalanceConfigInfo)                      // 获取余额配置信息
	cb.POST("/records", controller.BalanceRecord)                         // 获取余额记录
	cb.POST("/stats_use", controller.GetUserBalanceCdkStats)              // cdk 统计使用记录
	cb.POST("/stats_download", controller.GetUserBalanceCdkStatsDownload) // 统计使用记录下载
	cb.POST("/stats_detail", controller.GetUserBalanceCdkStatsDetail)     // 统计使用记录详情
	// 不限量相关
	cu := router.Group("/center/unlimited")
	cu.POST("/server_record", controller.GetUserUnlimitedLog)  // 获取余额配置信息
	cu.POST("/port_domain", controller.GetUnlimitedPortDomain) // 获取端口白名单地址

	// 不限量机器监控
	cvm := router.Group("/center/monitor")
	cvm.POST("/stats", controller.TencentCvmMonitor)                  //实时监控
	cvm.POST("/restart", controller.TencentCvmRestart)                //重启
	cvm.POST("/status", controller.TencentCvmDescribeStatus)          //获取服务器状态
	cvm.POST("/stats_download", controller.TencentCvmMonitorDownload) //实时监控数据下载
}
