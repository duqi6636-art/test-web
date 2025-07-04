package models

type Country struct {
	Id             int    `json:"id"`
	Name           string `json:"name"`
	Iso3           string `json:"iso3"`
	NumericCode    string `json:"numeric_code"`
	Iso2           string `json:"iso2"`
	Phonecode      string `json:"phonecode"`
	Capital        string `json:"capital"`
	Currency       string `json:"currency"`
	CurrencyName   string `json:"currency_name"`
	CurrencySymbol string `json:"currency_symbol"`
	Flag           string `json:"flag"`
}

const countriesTableName = "cm_countries"

func GetCountryList(query interface{}, args ...interface{}) (list []Country) {
	db.Table(countriesTableName).Where(query, args...).Find(&list)
	return
}

func GetCountryById(id int) (country Country) {
	db.Table(countriesTableName).Where("id=?", id).First(&country)
	return
}

func UpdateCountryWhere(values interface{}, query interface{}, args ...interface{}) {
	db.Table(countriesTableName).Where(query, args).Updates(values)
}
