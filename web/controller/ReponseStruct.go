package controller

// GetAccountInfoResponse 获取账户信息返回结构
type GetAccountInfoResponse struct {
	FlowDay    []GetUnlimitedStruct `json:"flow_day"`    // 账户不限量信息
	Flows      GetFlowsStruct       `json:"flows"`       // 流量类型信息
	DynamicIsp GetFlowsStruct       `json:"dynamic_isp"` // 动态ISP流量
	Isp        GetBalanceResponse   `json:"isp"`         // ISP流量
	Static     []GetStaticResponse  `json:"static"`      // 静态IP信息
	IspAgent   GetBalanceResponse   `json:"isp_agent"`   // ISP代理 企业余额
	FloWAgent  GetFowAgentResponse  `json:"flow_agent"`  // 住宅代理 企业余额
	Balance    GetBalanceInfo       `json:"balance"`     // 账户余额
}

type GetFowAgentResponse struct {
	Flow       string `json:"flow"`         //用户剩余流量 GB
	FlowRedeem string `json:"flow_redeem"`  //可兑换 流量余额 GB
	FlowUnit   string `json:"flow_unit"`    //用户剩余流量单位 GB
	FlowMb     string `json:"flow_mb"`      //用户剩余流量 MB
	FlowMbUnit string `json:"flow_mb_unit"` //用户剩余流量单位 MB
	FlowDate   string `json:"flow_date"`    //用户剩余流量到期时间
	FlowExpire int    `json:"flow_expire"`  //用户剩余流量是否到期
}

type GetStaticResponse struct {
	Id        int    `json:"id"`
	PakName   string `json:"pak_name"`   // 套餐类型
	Balance   int    `json:"balance"`    // 剩余IP
	ExpireDay int    `json:"expire_day"` // 过期时间
}

type GetBalanceResponse struct {
	Balance int `json:"balance"` //余额
}

type GetBalanceInfo struct {
	Balance string `json:"balance"` //余额
	Status  int    `json:"status"`  //状态
}

type GetIspResponse struct {
	IspNum int `json:"isp_num"` // ip数量
}

// GetUnlimitedStruct 获取账户不限量信息返回
type GetUnlimitedStruct struct {
	Day       int    `json:"day"`        // 不限量时长账户剩余天数
	DayExpire string `json:"day_expire"` // 不限量到期时间
	DayUnit   string `json:"day_unit"`   // 不限量到期时间单位
	DayUse    int    `json:"day_use"`    // 是否能用不限量时长
	DayStatus int    `json:"day_status"` // 是否冻结
}

// GetFlowsStruct 获取账户 流量信息返回
type GetFlowsStruct struct {
	Flow       int64  `json:"flow"`        // 账户流量
	ExpireDate string `json:"expire_date"` // 账户流量到期日
	Expired    int    `json:"expired"`     // 流量到期时间戳
	FlowGb     string `json:"flow_gb"`     // 流量单位为GB
	FlowMb     string `json:"flow_mb"`     // 流量单位为MB
	SendFlow   string `json:"send_flow"`   // 已发送邮件
	Status     int    `json:"status"`      // 是否冻结流量
	SendOpen   int    `json:"send_open"`   // 是否开通邮件提醒
	SendUnit   string `json:"send_unit"`   // 邮件提醒单位
	UnitGb     string `json:"unit_gb"`     // 单位为GB
	UnitMb     string `json:"unit_mb"`     // 单位为MB
}

// SetSendFlowsResponse 设置邮件提醒流量返回结构
type SetSendFlowsResponse struct {
	SendFlow string `json:"send_flow"` // 邮件提醒流量值
	SendOpen int    `json:"send_open"` // 是否开通邮件提醒
	SendUnit string `json:"send_unit"` // 邮件提醒流量单位
}

// GetChartDataResponse 获取流量日志返回结构
type GetChartDataResponse struct {
	Cate  []string             `json:"cate"`   //类型
	Unit  string               `json:"unit"`   //单位
	XData []string             `json:"x_data"` //X轴 时间列表
	YData map[string][]float64 `json:"y_data"` //Y轴 各类型流量列表
}

type GetChartIspDataResponse struct {
	Cate  []string         `json:"cate"`
	Unit  string           `json:"unit"`
	XData []string         `json:"x_data"`
	YData map[string][]int `json:"y_data"`
}

type GetUserEmailExclusiveResponse struct {
	Email  string `json:"email"`
	ExInfo struct {
		Code     string `json:"code"`
		Ratio    int    `json:"ratio"`
		Percent  string `json:"percent"`
		Img      string `json:"img"`
		Money    int    `json:"money"`
		Discount string `json:"discount"`
	} `json:"ex_info"`
	IsEx   int    `json:"is_ex"`
	IsPay  int    `json:"is_pay"`
	Wallet string `json:"wallet"`
}

// StaticCheckRepayResponse   统计用户静态续费前检测数据
type StaticCheckRepayResponse struct {
	Balance int                   `json:"balance"` // 账户余额
	IP      string                `json:"ip"`      // IP
	Lists   []StaticCheckRepayPak `json:"lists"`   // 余额列表
	Region  string                `json:"region"`  // 地区
}

type StaticCheckRepayPak struct {
	Id         int    `json:"id"`
	PakName    string `json:"pak_name"`    // 套餐类型
	Balance    int    `json:"balance"`     // 剩余IP
	ExpireDay  int    `json:"expire_day"`  // 过期天数
	ExpireTime string `json:"expire_time"` // 过期时间
}
