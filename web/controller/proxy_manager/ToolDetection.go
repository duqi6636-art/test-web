package proxyManageController

import (
	"api-360proxy/web/controller"
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
	"golang.org/x/net/proxy"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type OnlineProxyCheckerRsp struct {
	Ip       string  `json:"ip"`
	Username string  `json:"username"` // 用户名
	Password string  `json:"password"` // 密码
	Host     string  `json:"host"`     // 主机名
	Port     string  `json:"port"`     // 端口
	Country  string  `json:"country"`  // 地区
	Type     string  `json:"type"`
	Code     int     `json:"code"`
	Duration float32 `json:"duration"`
}

type IpInfoRsp struct {
	Ip       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
	Readme   string `json:"readme"`
}

// 根据错误类型返回适当的HTTP状态码
func getErrorStatusCode(err error) int {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "Forbidden"):
		return http.StatusForbidden
	case strings.Contains(errStr, "Client.Timeout"):
		return http.StatusRequestTimeout
	case strings.Contains(errStr, "proxyconnect tcp: dial tcp: address"):
		return http.StatusBadRequest
	case strings.Contains(errStr, "connection refused"):
		return http.StatusServiceUnavailable
	default:
		return http.StatusBadRequest
	}
}

// parseProxyURL 解析代理URL，处理各种格式
func parseProxyURL(rawURL string) (parsedURL, username, password, host, port string) {
	// 移除协议前缀
	if strings.Contains(rawURL, "://") {
		parts := strings.SplitN(rawURL, "://", 2)
		rawURL = parts[1]
	}

	// 处理包含@的格式 (用户名:密码@主机:端口)
	if strings.Contains(rawURL, "@") {
		parts := strings.Split(rawURL, "@")
		if len(parts) == 2 {
			auth := parts[0]
			hostPort := parts[1]

			// 检查是否需要交换位置（有时格式是 主机:端口@用户名:密码）
			if strings.Contains(auth, ".") && !strings.Contains(hostPort, ".") {
				// 交换位置
				auth, hostPort = hostPort, auth
			}

			// 解析用户名和密码
			if strings.Contains(auth, ":") {
				authParts := strings.Split(auth, ":")
				username = authParts[0]
				if len(authParts) > 1 {
					password = authParts[1]
				}
			}

			// 解析主机和端口
			if strings.Contains(hostPort, ":") {
				hostPortParts := strings.Split(hostPort, ":")
				host = hostPortParts[0]
				if len(hostPortParts) > 1 {
					port = hostPortParts[1]
				}
			}
			parsedURL = auth + "@" + hostPort
			return
		}
	}

	// 处理四段式格式 (可能是 主机:端口:用户名:密码 或 用户名:密码:主机:端口)
	parts := strings.Split(rawURL, ":")
	if len(parts) == 4 {
		// 判断哪部分是主机（包含点）
		if strings.Contains(parts[0], ".") {
			// 格式是 主机:端口:用户名:密码
			host = parts[0]
			port = parts[1]
			username = parts[2]
			password = parts[3]
			parsedURL = username + ":" + password + "@" + host + ":" + port
		} else if strings.Contains(parts[2], ".") {
			// 格式是 用户名:密码:主机:端口
			username = parts[0]
			password = parts[1]
			host = parts[2]
			port = parts[3]
			parsedURL = username + ":" + password + "@" + host + ":" + port
		}
		return
	}
	// 默认情况，直接返回原始URL
	parsedURL = rawURL
	return
}

func parseProxyURLT(proxyURL string) (hostname, port, username, password string, err error) {
	// 移除协议前缀
	if strings.Contains(proxyURL, "://") {
		parts := strings.SplitN(proxyURL, "://", 2)
		proxyURL = parts[1]
	}

	// 处理 username:password@hostname:port 或 hostname:port@username:password 格式
	if strings.Contains(proxyURL, "@") {
		parts := strings.Split(proxyURL, "@")
		if len(parts) != 2 {
			return "", "", "", "", fmt.Errorf("invalid proxy URL format")
		}

		// 判断哪部分包含域名（通过检查是否包含点号）
		if strings.Contains(parts[0], ".") {
			// hostname:port@username:password 格式
			hostParts := strings.Split(parts[0], ":")
			credParts := strings.Split(parts[1], ":")

			if len(hostParts) >= 2 && len(credParts) >= 2 {
				hostname = hostParts[0]
				port = hostParts[1]
				username = credParts[0]
				password = credParts[1]
			} else {
				return "", "", "", "", fmt.Errorf("invalid proxy URL format")
			}
		} else {
			// username:password@hostname:port 格式
			credParts := strings.Split(parts[0], ":")
			hostParts := strings.Split(parts[1], ":")

			if len(credParts) >= 2 && len(hostParts) >= 2 {
				username = credParts[0]
				password = credParts[1]
				hostname = hostParts[0]
				port = hostParts[1]
			} else {
				return "", "", "", "", fmt.Errorf("invalid proxy URL format")
			}
		}
	} else {
		// 处理 hostname:port:username:password 或 username:password:hostname:port 格式
		parts := strings.Split(proxyURL, ":")
		if len(parts) != 4 {
			return "", "", "", "", fmt.Errorf("invalid proxy URL format")
		}

		// 判断哪部分是域名（通过检查是否包含点号）
		if strings.Contains(parts[0], ".") {
			// hostname:port:username:password 格式
			hostname = parts[0]
			port = parts[1]
			username = parts[2]
			password = parts[3]
		} else if strings.Contains(parts[2], ".") {
			// username:password:hostname:port 格式
			username = parts[0]
			password = parts[1]
			hostname = parts[2]
			port = parts[3]
		} else {
			return "", "", "", "", fmt.Errorf("invalid proxy URL format")
		}
	}

	return hostname, port, username, password, nil
}

// checkHTTPProxy 检测HTTP/HTTPS代理
func checkHTTPProxy(proxyURL, username, passwd, host, port, domain, cate string, timeoutMs int) OnlineProxyCheckerRsp {
	result := OnlineProxyCheckerRsp{
		Username: username,
		Password: passwd,
		Host:     host,
		Port:     port,
		Type:     cate,
	}

	start := time.Now()

	// 解析代理地址
	proxyURI, err := url.Parse(proxyURL)
	if err != nil {
		log.Println("url.Parse====", err.Error())
		result.Code = http.StatusBadRequest
		return result
	}
	log.Printf("========proxyURI=====%+v", proxyURI)

	// 创建Transport并设置代理
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURI),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// 创建HTTP客户端
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeoutMs) * time.Millisecond,
	}

	// 创建请求
	req, err := http.NewRequest("GET", domain, nil)
	if err != nil {
		result.Code = http.StatusBadRequest
		return result
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		result.Code = getErrorStatusCode(err)
		return result
	}
	fmt.Println("========", resp.StatusCode)
	defer resp.Body.Close()

	// 计算耗时
	result.Duration = float32(time.Since(start)) / float32(time.Second)

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Code = http.StatusInternalServerError
		return result
	}

	// 解析IP信息
	var ipInfo IpInfoRsp
	if err := json.Unmarshal(body, &ipInfo); err != nil {
		result.Code = http.StatusForbidden
		return result
	}

	// 设置结果
	result.Ip = ipInfo.Ip
	result.Country = ipInfo.Country
	result.Code = http.StatusOK

	return result
}

// 检测SOCKS5代理
func checkSocks5Proxy(proxyURL, username, passwd, host, port, domain, cate string, timeoutMs int) OnlineProxyCheckerRsp {
	result := OnlineProxyCheckerRsp{
		Username: username,
		Password: passwd,
		Host:     host,
		Port:     port,
		Type:     cate,
	}

	start := time.Now()

	// 解析代理地址
	proxyURI, err := url.Parse(proxyURL)
	if err != nil {
		result.Code = http.StatusBadRequest
		return result
	}

	// 创建拨号器
	dialer, err := proxy.FromURL(proxyURI, &net.Dialer{
		Timeout:   time.Duration(timeoutMs) * time.Millisecond,
		KeepAlive: 30 * time.Second,
	})
	if err != nil {
		result.Code = http.StatusBadRequest
		return result
	}

	// 配置Transport
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// 创建HTTP客户端
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeoutMs) * time.Millisecond,
	}

	// 创建请求
	req, err := http.NewRequest("GET", domain, nil)
	if err != nil {
		result.Code = http.StatusBadRequest
		return result
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		result.Code = getErrorStatusCode(err)
		return result
	}
	defer resp.Body.Close()

	// 计算耗时
	result.Duration = float32(time.Since(start)) / float32(time.Second)

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Code = http.StatusInternalServerError
		return result
	}

	// 解析IP信息
	var ipInfo IpInfoRsp
	if err := json.Unmarshal(body, &ipInfo); err != nil {
		result.Code = http.StatusForbidden
		return result
	}

	// 设置结果
	result.Ip = ipInfo.Ip
	result.Country = ipInfo.Country
	result.Code = http.StatusOK

	return result
}

// OnlineProxyChecker 在线代理监测
func OnlineProxyChecker(c *gin.Context) {
	session := c.DefaultPostForm("session", "")
	resCode, msg, user := controller.DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		controller.JsonReturn(c, resCode, msg, nil)
		return
	}
	fmt.Println("=====user=====", user)

	urls := c.DefaultPostForm("urls", "")
	protocolType := com.StrTo(c.DefaultPostForm("protocol_type", "1")).MustInt() // 类型
	duration := com.StrTo(c.DefaultPostForm("duration", "0")).MustInt()          // 延迟

	// 从配置获取检测目标域名，默认使用ipinfo.io
	domain := models.GetConfigV("PROXY_CHECK_DOMAIN")

	if domain == "" {
		domain = "https://ipinfo.io/what-is-my-ip"
	}

	if urls == "" {
		controller.JsonReturn(c, e.ERROR, "__T_URLS_NIL", nil)
		return
	}

	urlList := strings.Split(urls, ",")
	if len(urlList) <= 0 {
		controller.JsonReturn(c, e.ERROR, "__T_URLS_NIL", nil)
		return
	}
	if len(urlList) > 1000 {
		controller.JsonReturn(c, e.ERROR, "__T_URLS_MAX_500", nil)
		return
	}
	maxConcurrent := 20
	if len(urlList) < maxConcurrent {
		maxConcurrent = len(urlList)
	}

	var limitChan = make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var rspList = make([]OnlineProxyCheckerRsp, len(urlList))

	if duration <= 0 {
		duration = 5000
	}

	cate := ""
	switch protocolType {
	case 1:
		cate = "http"
	case 2:
		cate = "https"
	case 3:
		cate = "socks5"
	default:
		controller.JsonReturn(c, e.ERROR, "__T_PROTOCOL_TYPE_ERR", nil)
		return
	}
	fmt.Println("=====cate=====", cate)

	for i, rawURL := range urlList {
		wg.Add(1)
		limitChan <- struct{}{}
		go func(index int, url, cate string) {
			defer wg.Done()
			defer func() { <-limitChan }()

			hostname, port, username, password, _ := parseProxyURLT(rawURL)
			parsedURL := username + ":" + password + "@" + hostname + ":" + port

			switch protocolType {
			case 1:
				rspList[index] = checkHTTPProxy("http://"+parsedURL, username, password, hostname, port, domain, cate, duration)
			case 2:
				rspList[index] = checkHTTPProxy("https://"+parsedURL, username, password, hostname, port, domain, cate, duration)
			case 3:
				rspList[index] = checkSocks5Proxy("socks5://"+parsedURL, username, password, hostname, port, domain, cate, duration)
			}
		}(i, rawURL, cate)
	}
	wg.Wait()

	// 写入检测记录表
	userIp := c.ClientIP()
	ipInfo := controller.GetIpInfo(userIp)
	_, uid := controller.GetUIDbySession(session) //获取用户ID
	var logList = make([]models.LogOnlineChecker, 0, len(rspList))

	for _, rsp := range rspList {
		info := models.LogOnlineChecker{
			Uid:        uid,
			Username:   rsp.Username,
			RspIp:      rsp.Ip,
			RspRegion:  rsp.Country,
			HttpCode:   rsp.Code,
			Duration:   int(rsp.Duration * 1000),
			UserIp:     userIp,
			UserRegion: ipInfo.CountryCode,
			Today:      util.GetTodayTime(),
			Cate:       cate,
			CreateTime: util.GetNowInt(),
		}
		logList = append(logList, info)
	}
	// 异步执行数据库操作，提高接口响应速度
	go func() {
		if err := models.BatchLogOnlineChecker(logList); err != nil {
			log.Printf("Failed to batch log online checker: %v", err)
		}
	}()

	controller.JsonReturn(c, e.SUCCESS, "success", rspList)
	return
}
