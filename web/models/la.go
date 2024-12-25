package models

const (
	LanguageTable = "cm_la"
)

// 语言包
type Language struct {
	Code   string `json:"code"`
	ZhCn   string `json:"_zh_cn"  gorm:"column:_zh_cn"`
	ZhHk   string `json:"_zh_hk"  gorm:"column:_zh_hk"`
	LEN    string `json:"_en"  gorm:"column:_en_us"`
	LID    string `json:"_id"  gorm:"column:_id"`
	LDE    string `json:"_de"  gorm:"column:_de"`
	LAR    string `json:"_ar"  gorm:"column:_ar"`
	LRU    string `json:"_ru"  gorm:"column:_ru"`
	LJP    string `json:"_jp"  gorm:"column:_jp"`
	LVI    string `json:"_vi"  gorm:"column:_vi"`
	LES    string `json:"_es"  gorm:"column:_es"` // 西班牙语
	LFR    string `json:"_fr"  gorm:"column:_fr"` // 法语
}

func FindLanguageConf() (err error, languages []Language) {
	err = db.Table(LanguageTable).Find(&languages).Error
	return
}

func FindAllLanguages() (languages []Language) {
	db.Table(LanguageTable).Find(&languages).Order(" id asc")
	return
}
func UpdataLanguage(languages Language) {
	db.Table(LanguageTable).Where("code = ?", languages.Code).Update(languages)
	return
}

func AddLanguage(lang Language) bool {
	err := db.Table(LanguageTable).Create(&lang).Error
	return err != nil
}
