package routers

import (
	"api-360proxy/web/controller"

	"github.com/gin-gonic/gin"
)

func webRouter(router *gin.Engine) {
	router.GET("/ip", controller.DealIp)               //获取IP
	router.GET("/test", controller.Test)               //获取IP
	router.POST("/ip", controller.DealIp)              //获取IP
	router.POST("/uuid", controller.DealUuid)          //发送邮件demo
	router.POST("/init_plugin", controller.InitPlugin) //初始化浏览器插件配置
	router.POST("/init_oem", controller.InitOem)       //初始化oem下载地址
	//图片上传
	router.POST("/upload_img", controller.UploadImage)                      //上传单个图片
	router.POST("/upload_multiple_images", controller.UploadMultipleImages) //上传多个图片

	router.POST("/web/ip_source", controller.IpSource) //IP关键词
	router.POST("/web/check_cn", controller.CheckIsCn) //检查是否 中国大陆

	web := router.Group("/web")
	web.POST("/auth/sign", controller.GetAuthSign)           // 滑块验证
	web.POST("/user/login", controller.Login)                // 邮箱登录
	web.POST("/auth/login", controller.GetAuthLogin)         // 登录验证				//center
	web.POST("/user/login_verify", controller.LoginVerify)   // 人机验证				//center
	web.POST("/user/google_login", controller.GoogleLogin)   // 谷歌登录
	web.POST("/user/github_login", controller.GithubLogin)   // GitHub登录
	web.POST("/user/email_reg", controller.WebReg)           // 邮箱注册
	web.POST("/user/sending", controller.SendEmailCode)      // 发送邮件验证码
	web.POST("/user/user_info", controller.GetInfo)          // 用户信息
	web.POST("/user/auto_login", controller.LoginAuto)       // 自动登录
	web.POST("/user/forget", controller.ForgetPwd)           // 忘记密码
	web.POST("/user/logout", controller.Logout)              // 退出登录
	web.POST("/user/bind_email", controller.BindEmail)       // 绑定邮箱
	web.POST("/user/reset_pwd", controller.ResetPass)        // 修改 /重置密码
	web.POST("/user/pay_list", controller.UserOrderList)     // 用户购买记录
	web.POST("/user/used_ip_list", controller.UsedIp)        // 用户IP消耗记录
	web.POST("/user/used_ip_data", controller.GetUserUseIp)  // 用户IP消耗日期和数量
	web.POST("/user/used_data", controller.GetUserUseIpNew)  // 用户IP消耗日期和数量
	web.POST("/user/email_info", controller.GetInfoByEmail)  // 根据邮箱获取是否购买过信息
	web.POST("/user/bind_wallet", controller.BindWallet)     // 绑定钱包
	web.POST("/user/unbind_wallet", controller.UnBindWallet) // 解绑钱包
	web.POST("/user/up_case", controller.UserCase)           // 用户案例
	web.POST("/user/pack_status", controller.PackStatus)     // 用户套餐状态

	web.POST("/msg/feedback", controller.FeedbackWeb)                   // 提交反馈
	web.POST("/msg/feedback_assistance", controller.FeedbackAssistance) //申请协助反馈
	web.POST("/msg/feedback_unlimited", controller.UnlimitedFeedback)   //不限量套餐定制反馈
	web.POST("/msg/popup", controller.GetPopup)                         // 获取弹窗
	web.POST("/msg/banner", controller.GetBanner)                       // 获取banner
	web.POST("/msg/notice", controller.GetNotice)                       // 获取公告信息
	web.POST("/msg/read_notice", controller.ReadNotice)                 // 读取公告
	web.POST("/msg/del_notice", controller.DelNotice)                   // 删除公告
	web.POST("/msg/notice_pop", controller.GetNoticePop)                // 获取公告弹窗
	web.POST("/msg/close_notice_pop", controller.CloseNoticePop)        // 关闭公告弹窗

	web.POST("/msg/batch_read", controller.MoreReadNotice) // 批量读取公告
	web.POST("/msg/notice_msg", controller.GetNoticeMsg)   // 消息通知

	// google 验证器接口
	web.POST("/auth/auth_info", controller.GetGoogleAuth)      // 创建信息
	web.POST("/auth/verify_code", controller.VerifyCode)       // 验证
	web.POST("/auth/bing_auth", controller.VerifyCodeBind)     // 绑定信息
	web.POST("/auth/unbind_auth", controller.VerifyCodeUnBind) // 解绑信息

	//web.POST("/invite/stats_click", controller.StatsClick) // 活动点击统计

	// invite 邀请返佣接口
	web.POST("/invite/info", controller.GetUserInviteInfo)           // 获取信息
	web.POST("/invite/record", controller.InviterList)               // 邀请记录
	web.POST("/invite/money_record", controller.MoneyList)           // 佣金记录
	web.POST("/invite/pay_record", controller.InviterOrderList)      // 邀请购买记录
	web.POST("/invite/withdrawal", controller.Withdrawal)            // 申请提现
	web.POST("/invite/withdrawal_record", controller.WithdrawalLog)  // 提现记录
	web.POST("/invite/withdrawal_info", controller.WithdrawalDetail) // 提现记录详细信息
	web.POST("/invite/set_code", controller.SetInviteCode)           // 设置邀请码
	web.POST("/invite/ex_balance", controller.ExBalance)             // 佣金余额兑换余额
	web.POST("/invite/ex_flow", controller.ExFlow)                   // 佣金余额兑换余额

	// inviteV1 邀请返佣接口 todo :: 新改版  2023-11-01
	web.POST("/invite_v1/info", controller.GetUserInviteInfoV1)           // 获取信息
	web.POST("/invite_v1/ex_balance", controller.ExBalanceV1)             // 佣金余额兑换
	web.POST("/invite_v1/withdrawal", controller.WithdrawalV1)            // 申请提现
	web.POST("/invite_v1/get_user_money_log", controller.GetUserMoneyLog) // 邀请记录   /  兑换记录   /   提现记录

	//邀请返佣活动相关接口
	web.POST("/activity/invite/get_info", controller.GetActivityUserInfo)             // 获取信息
	web.POST("/activity/invite/record", controller.ActivityInviterList)               // 邀请记录
	web.POST("/activity/invite/money_record", controller.ActivityMoneyList)           // 佣金记录
	web.POST("/activity/invite/pay_record", controller.ActivityInviterOrderList)      // 邀请购买记录
	web.POST("/activity/invite/withdrawal", controller.ActivityWithdrawal)            // 申请提现
	web.POST("/activity/invite/withdrawal_record", controller.ActivityWithdrawalLog)  // 提现记录
	web.POST("/activity/invite/withdrawal_info", controller.ActivityWithdrawalDetail) // 提现记录详细信息
	web.POST("/activity/invite/ex_flow", controller.ActivityExFlow)                   // 佣金余额兑换流量

	web.POST("/roll/order", controller.GetRollOrderList)   // 获取滚动订单信息
	web.POST("/roll/withdrawal", controller.GetRollTxList) // 获取滚动提现信息

	web.POST("/blog/get_search", controller.GetSearchBlog)    // blog-搜索
	web.POST("/blog/list_page", controller.GetBlogList)       // blog-列表
	web.POST("/blog/views", controller.GetViewBlog)           // blog-获取阅读数
	web.POST("/article/views", controller.ViewArticle)        // 文章教程-获取阅读数
	web.POST("/faq/views", controller.GetViewFAQ)             // 常见问题(faq)分类-获取阅读数
	web.POST("/article_video/views", controller.GetViewVideo) // 视频指南-获取阅读数

	web.POST("/coupon/get_list", controller.GetCouponList)          // 优惠券，获取优惠券列表
	web.POST("/coupon/coupon", controller.GetCouponOne)             // 获取优惠
	web.POST("/coupon/card_holder", controller.CardHolder)          // 卡包
	web.POST("/coupon/redeem_coupons", controller.RedeemCoupons)    // 兑换优惠券
	web.POST("/coupon/use_coupon", controller.GetCouponListByPakId) // 优惠券下拉列表接口

	web.POST("/stats/active", controller.StatsActiveClick) // 活动点击统计

	web.POST("/ex/generate", controller.Generate)                   // 生成兑换
	web.POST("/ex/generate_batch", controller.BatchGenerate)        // 批量生成兑换
	web.POST("/ex/get_exchange", controller.ExchangeCdk)            // 兑换
	web.POST("/ex/direct_conversion", controller.DirectConversion)  // 直接提现到本账户ip余额
	web.POST("/ex/generate_list", controller.ExchangeList)          // 兑换列表
	web.POST("/ex/forbid", controller.ForbidEx)                     // 禁用
	web.POST("/ex/flow_cdk", controller.GenerateFlowCdk)            // 生成flow cdk
	web.POST("/ex/flow_cdk_batch", controller.BatchGenerateFlowCdk) // 批量生成flow cdk
	web.POST("/ex/flow_recharge", controller.FlowRechargeUser)      // 给他人充值
	web.POST("/ex/flow_record", controller.ExchangeFlowList)        // 给他人充值

	web.POST("/get_country_code", controller.GetCountryCode) // 国家或国家代码搜索

	web.POST("/user/user_account", controller.AddUserAccount)              // 用户账号
	web.POST("/user/set_pass", controller.SetUserAccount)                  // 修改用户密码
	web.POST("/user/get_country", controller.GetCountry)                   // 国家
	web.POST("/user/get_country_V2", controller.GetCountryV2)              // 国家
	web.POST("/user/get_state", controller.GetState)                       // 州/省
	web.POST("/user/get_city", controller.GetCity)                         // 城市
	web.POST("/user/get_isp", controller.GetCountryIsp)                    // 城市
	web.POST("/user/get_domain", controller.GetUserDomain)                 // 获取域名列表
	web.POST("/user/country_domain_list", controller.GetCountryDomainList) // 获取国家域名列表
	web.POST("/user/country_info", controller.GetUserCountryInfo)          // 获取当前用户的国家信息

	// 流量帐密子账号
	web.POST("/account/get_info", controller.GetAccountInfo)                          // 获取流量信息
	web.POST("/account/set_send", controller.SetSendFlows)                            // 设置流量预警
	web.POST("/account/all_lists", controller.GetUserAccountAllList)                  // 获取子账号帐密列表 包含主账户
	web.POST("/account/all_lists_download", controller.GetUserAccountAllListDownload) // 获取子账号帐密列表 包含主账户
	web.POST("/account/lists", controller.GetUserAccountList)                         // 获取子账号帐密列表
	web.POST("/account/lists_available", controller.GetUserAccountListAvailable)      // 获取当前可用账号列表
	web.POST("/account/add_edit", controller.AddUserFlowAccount)                      // 添加 / 修改账号
	web.POST("/account/set_pass", controller.SetUserAccountPass)                      // 修改账号名称及密码
	web.POST("/account/detail", controller.UserFlowAccountDetail)                     // 账号信息详情
	web.POST("/account/enable_disable", controller.AccountEnableOrDisable)            // 账号启用  / 禁用
	web.POST("/account/del", controller.DelUserAccount)                               // 删除账号
	web.POST("/account/chart_data", controller.GetFlowStats)                          // 统计图数据
	web.POST("/account/url_list", controller.GetUrlStats)                             // 统计图 筛选-Url
	web.POST("/account/used_download", controller.FlowStatsDownload)                  // 使用数控下载
	web.POST("/account/ip_chart_data", controller.GetIpCharData)                      // 获取IP提取统计图数据
	web.POST("/account/ip_chart_data_download", controller.GetIpCharDataDownload)     // 下载IP提取统计图数据

	white := router.Group("/web/white")
	white.POST("/lists", controller.IpWhitelists)            //白名单IP列表
	white.POST("/add", controller.AddWhitelist)              //白名单IP列表--添加
	white.POST("/detail", controller.GetWhitelist)           //白名单IP列表--详细信息
	white.POST("/edit", controller.EditWhitelist)            //白名单IP列表--编辑
	white.POST("/set_status", controller.SetWhitelistStatus) //白名单IP列表--编辑
	white.POST("/delete", controller.DelWhitelist)           //白名单IP列表--删除
	white.POST("/download", controller.WhitelistDownload)    //白名单IP列表--下载

	score := router.Group("/web/score")
	score.POST("/info", controller.GetUserScore)                //积分信息
	score.POST("/record", controller.ScoreRecord)               //积分记录
	score.POST("/exchange", controller.ExScoreFlow)             //积分兑换
	score.POST("/exchange_flow_day", controller.ExScoreFlowDay) //积分兑换不限量
	score.POST("/free", controller.UpFreeWeb)                   //免费获取积分
	score.POST("/feedback", controller.UpFeedbackWeb)           //反馈信息

	packages := router.Group("/web/package")
	packages.POST("/flow", controller.GetPackageFlow)                             //获取流量套餐
	packages.POST("/custom_flow", controller.GetPackageCustomFlow)                //获取自定义流量套餐
	packages.POST("/coupon", controller.GetPackageCustomCoupons)                  //获取自定义流量优惠卷
	packages.POST("/custom_flow_new", controller.GetPackageCustomFlowNew)         //获取自定义流量套餐
	packages.POST("/static_num", controller.GetStaticRegionNum)                   // 获取静态地区数量
	packages.POST("/country_ip_num", controller.GetCountryIpNumber)               // 获取国家IP数
	packages.POST("/halloween_activity", controller.GetHalloweenActivityPackages) //获取万圣节活动套餐
	packages.POST("/flow_list_new", controller.GetPackageNewFlowList)             //获取新用户5G流量套餐列表
	packages.POST("/halloween_enabled", controller.GetHalloweenEnabled)           //判断万圣节套餐是否可以使用

	flowDay := router.Group("/flow_day")
	flowDay.POST("/user/get_country", controller.GetFlowDayCountry) // 获取国家列表

	statics := router.Group("/web/static")
	statics.POST("/ip_list", controller.GetUserStaticIp)               // 长效余额记录
	statics.POST("/region", controller.GetRegion)                      // 国家地区城市
	statics.POST("/pools", controller.StaticIpList)                    // 静态IP池列表
	statics.POST("/use", controller.UseStatic)                         // 提取使用
	statics.POST("/batch_use", controller.BatchUseStatic)              // 批量提取使用
	statics.POST("/check_info", controller.BeforeRecharge)             // 提取续费
	statics.POST("/batch_check_info", controller.BatchBeforeRecharge)  // 批量提取续费
	statics.POST("/recharge", controller.IpRecharge)                   // 提取续费
	statics.POST("/batch_recharge", controller.BatchIpRecharge)        // 批量提取续费
	statics.POST("/set_account", controller.SetIPAccount)              // 编辑信息
	statics.POST("/del", controller.DelStatic)                         // 删除过期静态IP
	statics.POST("/use_list", controller.GetUsedStaticIpList)          // 获取已提取使用列表
	statics.POST("/use_record", controller.GetUsedStaticRecord)        // 统计使用记录
	statics.POST("/use_download", controller.UsedStaticRecordDownload) // 统计使用记录下载
	statics.POST("/change", controller.ChangeStaticIp)                 //  更换的静态IP
	statics.POST("/check_replace", controller.CheckReplace)            // 检测是否可以更换IP
	statics.POST("/check_kyc", controller.CheckKyc)                    // 检测是否可以更换IP

	// 长效ISP流量
	web.POST("/user/long_isp/get_main_user_account", controller.GetLongIspMainAccount)                      // 获取长效Isp用户主账号
	web.POST("/user/long_isp/set_main_pass", controller.SetLongIspMainAccount)                              // 修改主账号用户密码
	web.POST("/account/long_isp/set_send", controller.LongIspSetSendFlows)                                  // 长效Isp设置用户低于流量阀值 发送邮件
	web.POST("/account/long_isp/sub_account_lists", controller.GetLongIspUserChildAccountList)              // 获取子账号帐密列表
	web.POST("/account/long_isp/add_edit", controller.SaveLongIspFlowChildAccount)                          // 添加修改长效Isp子账号
	web.POST("/account/long_isp/sub_account_enable_disable", controller.LongIspChildAccountEnableOrDisable) // 子账号账号启用/禁用
	web.POST("/account/long_isp/sub_account_del", controller.DelLongIspUserChildAccount)                    // 删除子账号
	web.POST("/account/long_isp_url_list", controller.GetLongIspUrlStats)                                   // 统计图 长效Isp筛选-Url
	web.POST("/user/long_isp/get_country_list", controller.GetLongIspCountryCityPort)                       // 获取长效ISP国家城市域名列表
	web.POST("/account/long_isp/detail", controller.LongIspUserFlowAccountDetail)                           // 长效Isp账号信息详情

	// 绑定手机号
	web.POST("/country_list", controller.GetCountryList)                 // 获取国家对应的phone code
	web.POST("/phone/bind", controller.BindPhone)                        // 绑定手机号
	web.POST("/phone/cancel_bind", controller.CancelBindPhone)           // 取消绑定
	web.POST("/phone/bind_info", controller.GetBindPhone)                // 获取绑定手机信息
	web.POST("/country/update/image", controller.CountryListUpdateImage) // 更新国家图标url

	mp := router.Group("/mp")                             //代理管理器接口
	mp.POST("/version", controller.VersionCheck)          // 版本检查
	mp.POST("/feedback", controller.FeedbackAgentManager) // 代理管理器用户反馈
	mp.POST("/github_login", controller.MPGithubLogin)    // GitHub登录

	//实名认证 接口
	kyc := router.Group("/user_kyc")
	kyc.POST("/verify/step_one", controller.IdVerifyStepOne)     // 认证第一步
	kyc.POST("/verify/step_two", controller.IdVerifyStepTwo)     // 认证第二步
	kyc.POST("/verify/step_three", controller.IdVerifyStepThree) // 检测验证结果
	kyc.POST("/verify/get_face_url", controller.GetFaceUrl)      // 获取腾讯人脸核验链接
	kyc.POST("/verify/get_country", controller.GetKycCountry)    // 获取国家列表
	kyc.POST("/verify/all_status", controller.CheckKycStatus)
	kyc.POST("/verify/operator", controller.CheckKycOperator)

	domain := router.Group("/domain")

	domain.POST("/apply", controller.AddDomainWhiteApply)                          // 添加域名白名单申请
	domain.POST("/get_apply_domain", controller.DomainWhiteList)                   // 获取域名白名单申请列表
	domain.POST("/notify/domain_white_review", controller.DomainWhiteReviewNotify) // 域名白名单审核回调
	domain.POST("/check_kyc", controller.CheckDomainKyc)                           // 域名白名单审核回调

	// KYC人工审核接口
	kyc.POST("/upload_document", controller.UploadKycDocument)          // 上传证明材料
	kyc.POST("/submit_manual_review", controller.SubmitKycManualReview) // 提交人工审核
	kyc.POST("/review_status", controller.GetKycReviewStatus)           // 查询审核状态

	kyc.POST("/kyc_review", controller.EnterpriseKycNotify) // 人工审核回调

	// 企业认证相关接口
	enterpriseKyc := router.Group("/enterprise_kyc")
	enterpriseKyc.POST("/submit", controller.SubmitEnterpriseKyc)    // 提交企业认证
	enterpriseKyc.POST("/status", controller.GetEnterpriseKycStatus) // 查询企业认证状态

	chat := router.Group("/chat")
	chat.GET("/time", controller.GetChatTime)
}
