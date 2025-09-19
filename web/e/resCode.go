package e

const (
	SUCCESS            = 0    //成功 返回码
	ERROR              = -1   //错误 返回码
	OUT_SERVICE_AREA   = 1000 //不在指出地区 返回码
	SESSION_EXPIRED    = 2000 // 用户登录过期 返回码
	ERROR_KYC_REQUIRED = 3000 //需要实名认证
)
