package models

// 字段按字典顺序
type SignParam struct {
	DeviceOs    string `form:"_device_os"`
	Language    string `form:"_language,omitempty"`
	OsVersion   string `form:"_os_version,omitempty"`
	Oem         string `form:"_oem,omitempty"`
	Salt        string `form:"_salt,omitempty"`
	Session     string `form:"_session,omitempty"`
	Sign        string `form:"_sign,omitempty"`
	Sn          string `form:"_sn,omitempty"`
	Timestamp   string `form:"_timestamp,omitempty"`
	Version     string `form:"_version,omitempty"`
	VersionShow string `form:"_version_show,omitempty"`
	Brand       string `form:"_brand,omitempty"`
	Channel     string `form:"_channel,omitempty"`
	DeviceNum   string `form:"_device_num,omitempty"`
	System      string `form:"_system,omitempty"`
	TimeZone    string `form:"_time_zone,omitempty"`
	Platform    string `form:"_platform,omitempty"`
}
