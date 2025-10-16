package models

import (
	"api-360proxy/web/pkg/setting"
	"api-360proxy/web/pkg/util"
	"fmt"
	"github.com/coocood/freecache"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"gorm.io/driver/mysql"
	gormV2 "gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"
)

// 创建全局数据库连接对象
var db *gorm.DB
var dbRead *gorm.DB
var dnsDb *gorm.DB
var MasterWriteDb *gormV2.DB

// var logDb *gorm.DB
var RedisPool *redis.Pool

// 国家城市域名端口
var RedisCountryCityPort *redis.Pool

var (
	ConfigCache         = &freecache.Cache{}
	LangMap             = map[string]Language{}
	ConfigMap           = map[string]string{}
	OemVersion          = map[string]OemModel{}
	PackageListMap      = map[string][]CmPackage{}
	PayedPackageListMap = map[string][]CmPackage{}
	PackageTextMap      = map[string]CmPackageInfo{}
	PackageAreaTextMap  = map[string]CmPackageInfo{}
	PackageUnlimitedMap = []PackageUnlimitedModel{}
)

const (
	configTag              = "configTag-"
	oemConfigTag           = "configOemTag-"
	configCacheExpireSecod = 60 * 5
)

// 模型/数据库连接对象 初始化
func Setup() {
	ConfigCache = freecache.NewCache(1024 * 1024)
	var (
		err                                               error
		dbType, dbName, user, password, host, tablePrefix string
	)
	if err != nil {
		log.Fatal(2, "Fail to get section 'database': %v", err)
	}

	dbType = setting.DatabaseConfig.Type
	dbName = setting.DatabaseConfig.DbName
	user = setting.DatabaseConfig.User
	password = setting.DatabaseConfig.Password
	host = setting.DatabaseConfig.Host
	tablePrefix = setting.DatabaseConfig.TablePrefix
	log.Println("开始连接数据库")
	db, err = gorm.Open(dbType, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		user,
		password,
		host,
		dbName))

	if err != nil {
		log.Println(err)
	}
	log.Println("连接数据库成功")
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return tablePrefix + defaultTableName
	}
	db.SingularTable(true)
	if setting.RunMode == "debug" {
		db.LogMode(true)
	}
	//db.DB().SetMaxIdleConns(50)
	//db.DB().SetMaxOpenConns(100)
	//db.DB().SetConnMaxLifetime(time.Minute * 5)
	err = db.DB().Ping()
	if err != nil {
		log.Fatalln(err)
	}
	InitReadDb()
	InitDnsDb()
	InitConfig()
	InitLang()
	InitPackage()
	initRedis()
	// 初始化渠道下载地址
	InitOemVersion()
	// 初始化国家城市端口Redis数据库
	initCountryCityPortRedis()
	MasterWriteDbInit()

}

// WriteDbSetup
func MasterWriteDbInit() {
	var (
		err                          error
		dbName, user, password, host string
	)

	dbName = setting.DatabaseConfig.DbName
	user = setting.DatabaseConfig.User
	password = setting.DatabaseConfig.Password
	host = setting.DatabaseConfig.Host
	MasterWriteDb, err = gormV2.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user,
		password,
		host,
		dbName)), &gormV2.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	if setting.RunMode == "debug" {
		MasterWriteDb.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second, // 慢查询阈值，超过这个阈值的查询将被认为是慢查询
				Colorful:                  true,        // 彩色输出
				IgnoreRecordNotFoundError: true,        // 忽略记录未找到的错误
				LogLevel:                  logger.Info, // 日志级别
			},
		)
	} else {
		MasterWriteDb.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,  // 慢查询阈值，超过这个阈值的查询将被认为是慢查询
				Colorful:                  true,         // 彩色输出
				IgnoreRecordNotFoundError: true,         // 忽略记录未找到的错误
				LogLevel:                  logger.Error, // 日志级别
			},
		)
	}

	if err != nil {
		log.Fatalln(err)
	}
}

// 加载 只读库
func InitReadDb() {
	var (
		err                                               error
		dbType, dbName, user, password, host, tablePrefix string
	)
	if err != nil {
		log.Fatal(2, "Fail to get section 'read_database': %v", err)
	}

	dbType = setting.DatabaseReadConfig.Type
	dbName = setting.DatabaseReadConfig.DbName
	user = setting.DatabaseReadConfig.User
	password = setting.DatabaseReadConfig.Password
	host = setting.DatabaseReadConfig.Host
	tablePrefix = setting.DatabaseReadConfig.TablePrefix

	dbRead, err = gorm.Open(dbType, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		user,
		password,
		host,
		dbName))

	if err != nil {
		log.Println(err)
	}

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return tablePrefix + defaultTableName
	}
	dbRead.SingularTable(true)
	if setting.RunMode == "debug" {
		dbRead.LogMode(true)
	}
}

// 加载 DNS库
func InitDnsDb() {
	var (
		err                                               error
		dbType, dbName, user, password, host, tablePrefix string
	)
	if err != nil {
		log.Fatal(2, "Fail to get section 'read_database': %v", err)
	}

	dbType = setting.DatabaseDnsConfig.Type
	dbName = setting.DatabaseDnsConfig.DbName
	user = setting.DatabaseDnsConfig.User
	password = setting.DatabaseDnsConfig.Password
	host = setting.DatabaseDnsConfig.Host
	tablePrefix = setting.DatabaseDnsConfig.TablePrefix

	dnsDb, err = gorm.Open(dbType, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		user,
		password,
		host,
		dbName))

	if err != nil {
		log.Println(err)
	}

	gorm.DefaultTableNameHandler = func(dnsDb *gorm.DB, defaultTableName string) string {
		return tablePrefix + defaultTableName
	}
	dnsDb.SingularTable(true)
	if setting.RunMode == "debug" {
		dnsDb.LogMode(true)
	}
}

// 加载 redis
func initRedis() {
	RedisPool = &redis.Pool{
		MaxIdle:     setting.RedisConfig.MaxIdle,
		MaxActive:   setting.RedisConfig.MaxActive,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", setting.RedisConfig.Host, redis.DialPassword(setting.RedisConfig.Password))
			if err != nil {
				return c, err
			}
			_, err = c.Do("SELECT", setting.RedisConfig.Select)
			if err != nil {
				return c, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			//if time.Since(t) < time.Minute {
			//	return nil
			//}
			_, err := c.Do("PING")
			return err
		},
	}

	fmt.Println("init redis")
}

// 加载 国家城市域名端口redis
func initCountryCityPortRedis() {
	RedisCountryCityPort = &redis.Pool{
		MaxIdle:     setting.RedisConfig.MaxIdle,
		MaxActive:   setting.RedisConfig.MaxActive,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", setting.RedisConfig.Host, redis.DialPassword(setting.RedisConfig.Password))
			if err != nil {
				return c, err
			}
			_, err = c.Do("SELECT", 1)
			if err != nil {
				return c, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			//if time.Since(t) < time.Minute {
			//	return nil
			//}
			_, err := c.Do("PING")
			return err
		},
	}

	fmt.Println("init countryCityPortRedis")
}
func CloseDB() {
	fmt.Println("db close")
	defer db.Close()
}

func FreshConfigCache() {
	ConfigCache.Clear()
	InitConfig()
	InitOemVersion()
	InitPackage()
}

// 初始化获取APP配置参数, 存入缓存
func InitConfig() {
	fmt.Print("init config")
	err, config := FindConfigs()
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range config {
		ConfigCache.Set([]byte(configTag+v.Code), []byte(v.Value), configCacheExpireSecod)
		//ConfigMap[v.K] = v.V
	}
	ConfigCache.Set([]byte("cacheConfigTMP"), []byte(util.ItoS(util.GetNowInt())), configCacheExpireSecod)
}

// 初始化语言包
func InitLang() {
	err, lang_s := FindLanguageConf()
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range lang_s {
		LangMap[v.Code] = v
	}
}

// 初始化渠道下载地址
func InitOemVersion() {
	oemList, err := GetOemList()
	OemVersion = map[string]OemModel{}
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range oemList {
		str := v.Oem + "_" + v.Client
		OemVersion[str] = v
	}
}

func GetCacheTMP() map[string]string {
	res1 := ""
	res2 := ""
	r1, e1 := ConfigCache.Get([]byte("cacheConfigTMP"))
	if e1 == nil {
		res1 = string(r1)
	}
	r2, e2 := ConfigCache.Get([]byte("cacheOemConfigTMP"))
	if e2 != nil {
		res2 = string(r2)
	}
	result := map[string]string{
		"TMPConfig":    res1,
		"TMPOemConfig": res2,
	}
	return result
}

// 获取config项
func GetConfigVal(v string) string {
	key := configTag + v
	got, err := ConfigCache.Get([]byte(key))
	if err != nil {
		InitConfig()
		got, err = ConfigCache.Get([]byte(key))
		if err != nil {
			return ""
		}
	}
	return string(got)
	//return ConfigMap[v]
}

// 获取oemconfig项
func GetConfigOemVal(v string, oem string) string {
	key := oemConfigTag + v + oem
	got, err := ConfigCache.Get([]byte(key))
	/*if err != nil {
		InitOemConfig()
		got, err = ConfigCache.Get([]byte(key))
		if err != nil {
			return ""
		}
	}*/
	if err != nil {
		return ""
	}
	return string(got)
	//return OemConfigMap[oem+v]
}

func GetLang(code, lang string) string {
	if lang == "" {
		return code
	}
	languageTeam, ok := LangMap[code]
	if !ok {
		return code
	}

	switch lang {
	case "zh-cn": //简体中文
		{
			return languageTeam.ZhCn
		}
	case "zh-tw": //繁体中文
		{
			return languageTeam.ZhHk
		}
	case "tw": //繁体中文
		{
			return languageTeam.ZhHk
		}
	case "_en_us": //英文
		{
			return languageTeam.LEN
		}
	case "en": //英文
		{
			return languageTeam.LEN
		}
	case "id": //印尼
		{
			return languageTeam.LID
		}
	case "ar": //阿拉伯语
		{
			return languageTeam.LAR
		}
	case "de": //德语
		{
			return languageTeam.LDE
		}
	case "ru": //俄语
		{
			return languageTeam.LRU
		}
	case "ja": //日语
		{
			return languageTeam.LJP
		}
	case "vi": //越南语
		{
			return languageTeam.LVI
		}
	case "es": //西班牙语
		{
			return languageTeam.LES
		}
	case "fr": //法语
		{
			return languageTeam.LFR
		}
	default:
		{
			return languageTeam.LEN
		}
	}
}

func RedisSet(key string, val interface{}) {
	redisConn := RedisPool.Get()
	defer redisConn.Close()
	redisConn.Do("set", key, val)
}

func RedisSetEx(key string, val interface{}, expire int) {
	redisConn := RedisPool.Get()
	defer redisConn.Close()
	redisConn.Do("set", key, val, "EX", expire)
}

func RedisGet(key string) string {
	redisConn := RedisPool.Get()
	defer redisConn.Close()
	res, err := redis.String(redisConn.Do("GET", key))
	if err != nil {
		return ""
	}
	return res
}

// 存队列
func RedisLPUSH(key, values string) error {
	redisConn := RedisPool.Get()
	defer redisConn.Close()
	_, err := redisConn.Do("LPUSH", key, values)
	return err
}

// 取队列
func RedisRPop(key string) string {
	redisConn := RedisPool.Get()
	defer redisConn.Close()
	res, err := redis.String(redisConn.Do("RPOP", key))
	if err != nil {
		return ""
	}
	return res
}

// 获取长度
func RedisLLEN(key string) int {
	redisConn := RedisPool.Get()
	defer redisConn.Close()
	res, err := redis.Int(redisConn.Do("LLEN", key))
	if err != nil {
		return 0
	}
	return res
}

func RedisDel(key string) {
	redisConn := RedisPool.Get()
	defer redisConn.Close()
	redisConn.Do("del", key)
}

// / 創建数据库表 - 添加新的，但是不能刪除修改
func CreateDatabaseTables() {

	// 会员表
	db.AutoMigrate(&CmUserMember{})
	// 会员等级表
	db.AutoMigrate(&CmMemberLevel{})

}

// 初始化套餐列表
func InitPackage() {
	packageList, err := AllPackageList()

	PackageListMap = map[string][]CmPackage{}
	PayedPackageListMap = map[string][]CmPackage{}
	PackageTextMap = map[string]CmPackageInfo{}
	PackageUnlimitedMap = []PackageUnlimitedModel{}
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range packageList {
		str := v.PakType
		if v.Pid == 0 {
			PackageListMap[str] = append(PackageListMap[str], v)
		} else {
			PayedPackageListMap[str] = append(PayedPackageListMap[str], v)
		}
	}
	// 套餐文案 // 语言套餐信息配置
	errInfo, packageInfoList := GetPackageInfoList()
	if errInfo != nil {
		log.Fatal(errInfo)
	}
	for _, v := range packageInfoList {
		str := v.Lang + "_" + util.ItoS(v.PackageId)
		PackageTextMap[str] = v
		strArea := v.Lang + "_" + util.ItoS(v.PackageId) + "_" + util.ItoS(v.AreaId)
		PackageAreaTextMap[strArea] = v
	}
	PackageUnlimitedMap = GetPackageUnlimitedList()
}
