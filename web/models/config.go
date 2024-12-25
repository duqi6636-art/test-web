package models

// App基础配置
type Config struct {
	Code  string `json:"code"`
	Value string `json:"value"`
}

func getConfigTableName() string {
	return "cm_config"
}

func FindConfigs() (err error, configs []Config) {
	configTableName := getConfigTableName()
	err = db.Table(configTableName).Find(&configs).Error
	return
}
func GetConfigs(k string) (err error, configs Config) {
	configTableName := getConfigTableName()
	err = db.Table(configTableName).Where("code = ?", k).First(&configs).Error
	return
}
func GetConfigV(k string) (info string) {
	var configs = Config{}
	configTableName := getConfigTableName()
	err := db.Table(configTableName).Where("code = ?", k).First(&configs).Error
	if err == nil {
		info = configs.Value
	} else {
		info = ""
	}
	return info
}
func UpConfigs(k, v string) bool {
	configTableName := getConfigTableName()
	db.Table(configTableName).Where("code = ?", k).Update(map[string]interface{}{
		"value": v,
	})
	return true
}

type JumpConfig struct {
	Cate  string `json:"cate"`
	Code  string `json:"code"`
	Value string `json:"value"`
}

func FindAllJumpHelp(platform string) (err error, configs []JumpConfig) {
	dbx := db.Table("conf_jump_help_center")
	if platform != "" {
		dbx = dbx.Where("platform = ?", platform)
	}
	err = dbx.Find(&configs).Error
	return
}

// 邮箱模板配置
type ConfEmail struct {
	Id          int    `json:"id"`
	TplId       string `json:"tpl_id"`
	Subject     string `json:"subject"`
	FromTo      string `json:"from_to"`
	PackageName string `json:"package_name"`
	PackageUrl  string `json:"package_url"`
}

func GetConfEmail(platform string, cate int) (configs ConfEmail) {
	db.Table("conf_email").Where("platform =?", platform).Where("cate =?", cate).First(&configs)
	return
}

func GetConfEmailBy8(platform string, code string) (configs ConfEmail) {
	db.Table("conf_email").Where("platform =?", platform).Where("cate =?", 8).Where("code =?", code).First(&configs)
	return
}
