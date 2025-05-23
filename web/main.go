package main

import (
	"api-360proxy/pkg/ipdat"
	"api-360proxy/web/crons"
	"api-360proxy/web/db/clickhousedb"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/setting"
	"api-360proxy/web/routers"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func init() {
	setting.Setup()
	models.Setup()
	models.StatisticsDbSetup()
	//初始化clickhouse数据库
	clickhousedb.InitClickhouseDb()
	clickhousedb.InitClickhouseCherryLogDb()
	models.CreateDatabaseTables()
}

// @title           360 API
// @version         1.0
// @description     This 360 API
// @host      	    textapi.360proxy.net
// @schemes         http https
func main() {
	if setting.AppConfig.AccountFlowSwitch == 1 { //配置开启的时候才开启
		go func() {
			for {
				crons.DealFlowInfo() //处理用户流量信息
			}
		}()
		go func() {
			for {
				crons.DealLongIspFlowInfo() //处理用户长效Isp流量信息
			}
		}()
		go func() {
			for {
				crons.DealInviteFlow() //处理用户自用
			}
		}()
		go func() {
			for {
				crons.HandleCdkInfo() // 处理 生成cdk
			}
		}()
		go func() {
			for {
				crons.ExchangeCdk() // 兑换cdk
			}
		}()
		go func() {
			for {
				crons.HandleScore() // 积分兑换流量
			}
		}()
		go func() {
			for {
				crons.HandleScoreFlowDay() // 积分兑换不限量流量
			}
		}()
		go func() {
			for {
				crons.HandleCdkBalance() // 生成cdk-余额充值
			}
		}()
	}
	go func() {
		crons.GoCron() //执行计划任务
	}()

	// 初始化 IP库
	ipdat.IPDat = ipdat.GetObject(setting.IpDatPath)
	gin.SetMode(setting.RunMode)
	routersInit := routers.InitRouter()
	readTimeout := setting.ReadTimeout
	writeTimeout := setting.WriteTimeout
	endPoint := fmt.Sprintf(":%d", setting.HttpPort)
	maxHeaderBytes := 1 << 20
	server := &http.Server{
		Addr:           endPoint,
		Handler:        routersInit,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	log.Printf("[info] start http server listening %s", endPoint)
	server.ListenAndServe()

}
