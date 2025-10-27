package constants

// 登录安全配置键名常量
const (
	// 全局登录安全配置
	ConfigKeyGlobalTimeWindow   = "login_security_global_time_window"   // 全局时间窗口
	ConfigKeyGlobalFailureLimit = "login_security_global_failure_limit" // 全局失败次数限制

	// 用户登录安全配置
	ConfigKeyUserTimeWindow   = "login_security_user_time_window"   // 用户时间窗口
	ConfigKeyUserFailureLimit = "login_security_user_failure_limit" // 用户失败次数限制

	// IP登录安全配置
	ConfigKeyIPTimeWindow   = "login_security_ip_time_window"   // IP时间窗口
	ConfigKeyIPFailureLimit = "login_security_ip_failure_limit" // IP失败次数限制

	// 登录重置配置
	ConfigKeyLoginResetTimeWindow = "login_reset_time_window" // 登录重置时间窗口
)

// 登录安全默认值常量
const (
	// 默认时间窗口（秒）
	DefaultGlobalTimeWindow = 600  // 10分钟
	DefaultUserTimeWindow   = 300  // 5分钟
	DefaultIPTimeWindow     = 900  // 15分钟
	DefaultResetTimeWindow  = 1200 // 10分钟

	// 默认失败次数限制
	DefaultGlobalFailureLimit = 50 // 50次
	DefaultGlobalMLimit       = 20 // 20分钟
	DefaultUserFailureLimit   = 3  // 3次
	DefaultIPFailureLimit     = 10 // 10次
)
