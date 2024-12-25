package setting

import (
	"github.com/go-ini/ini"
	"log"
	"time"
)

type Database struct {
	Type        string `ini:"TYPE"`
	Host        string `ini:"HOST"`
	User        string `ini:"USER"`
	Password    string `ini:"PASSWORD"`
	DbName      string `ini:"NAME"`
	TablePrefix string `ini:"TABLE_PREFIX"`
	RmHost      string `ini:"RM_HOST"`
}

type App struct {
	Name                  string `ini:"ProjectName"`
	Oem                   string `ini:"DefaultOem"`
	VerifySignature       bool
	DefaultLanguage       string
	SessionExpire         int `ini:"SessionExpire"`
	BaseDomain            string
	Elastic               int
	AccountFlowSwitch     int
	SendEmailSwitch       int
	resource_domain_local string
	ProjectName           string
}

type Redis struct {
	Select    string
	Host      string
	Password  string
	MaxActive int
	MaxIdle   int
}

var DatabaseConfig = &Database{}
var DatabaseReadConfig = &Database{}
var DatabaseDnsConfig = &Database{}
var DatabaseStatisticsConfig = &Database{}
var AppConfig = &App{}
var RedisConfig = &Redis{}

var (
	Cfg          *ini.File
	RunMode      string
	HttpPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IpDatPath    string
)

func Setup() {
	var err error
	Cfg, err = ini.Load("conf/app.ini")
	if err != nil {
		log.Fatalf("fail to parse %s:%v", Cfg, err)
	}
	LoadBase()
	LoadApp()
	LoadDataBase()
	LoadReadDataBase()
	LoadDnsataBase()
	LoadRedisDb()
	LoadStatisticsDataBase()
}

func LoadBase() {
	sec := Cfg.Section("")
	RunMode = sec.Key("RunMode").MustString("debug")
	HttpPort = sec.Key("HttpPort").MustInt(3000)
	ReadTimeout = sec.Key("ReadTimeout").MustDuration(10) * time.Second
	WriteTimeout = sec.Key("WriteTimeout").MustDuration(10) * time.Second
	IpDatPath = sec.Key("IpDatPath").MustString("/")
}

func LoadApp() {
	err := Cfg.Section("app").MapTo(AppConfig)
	if err != nil {
		log.Fatalf("Cfg.MapTo AppConfig err: %v", err)
	}
}

func LoadDataBase() {
	dsConfig := "database"
	//if RunMode != "" {
	//	dsConfig = dsConfig + "-" + RunMode
	//}
	err := Cfg.Section(dsConfig).MapTo(DatabaseConfig)
	if err != nil {
		log.Fatalf("Cfg.MapTo databaseConfig err: %v", err)
	}
}

func LoadReadDataBase() {
	dsConfig := "read_database"
	//if RunMode != "" {
	//	dsConfig = dsConfig + "-" + RunMode
	//}
	err := Cfg.Section(dsConfig).MapTo(DatabaseReadConfig)
	if err != nil {
		log.Fatalf("Cfg.MapTo read_databaseConfig err: %v", err)
	}
}

func LoadDnsataBase() {
	dsConfig := "dns_database"
	//if RunMode != "" {
	//	dsConfig = dsConfig + "-" + RunMode
	//}
	err := Cfg.Section(dsConfig).MapTo(DatabaseDnsConfig)
	if err != nil {
		log.Fatalf("Cfg.MapTo read_databaseConfig err: %v", err)
	}
}

func LoadStatisticsDataBase() {
	dsConfig := "statistics_database"
	err := Cfg.Section(dsConfig).MapTo(DatabaseStatisticsConfig)
	if err != nil {
		log.Fatalf("Cfg.MapTo read_databaseConfig err: %v", err)
	}
}

func LoadRedisDb() {
	dsConfig := "redis"
	//if RunMode != "" {
	//	dsConfig = dsConfig + "-" + RunMode
	//}
	err := Cfg.Section(dsConfig).MapTo(RedisConfig)
	if err != nil {
		log.Fatalf("Cfg.MapTo redisConfig err: %v", err)
	}
}
