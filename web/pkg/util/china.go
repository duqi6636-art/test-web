package util

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type IP struct {
	ip  uint32
	off uint32
}

var ChinaIP = make([]IP, 0)

func InitChinaIP() {
	file, err := os.Open("conf/CN.txt")
	if err != nil {
		fmt.Println("初始化中国IP段失败", err.Error())
		os.Exit(0)
	}
	buf := make([]byte, 1000000)
	_, err = file.Read(buf)
	if err != nil {
		fmt.Println("初始化中国IP段失败", err.Error())
		os.Exit(0)
	}
	for _, val := range strings.Split(string(buf), "\n") {
		reg1 := regexp.MustCompile(`\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}`)
		if reg1 == nil {
			fmt.Println("初始化中国IP段失败", "文件数据格式错误")
			os.Exit(0)
		}
		result1 := reg1.FindAllStringSubmatch(val, -1)
		if len(result1) >= 2 {
			var bean = IP{
				ip:  InetAtoN(result1[0][0]),
				off: InetAtoN(result1[1][0]),
			}
			ChinaIP = append(ChinaIP, bean)
		}
	}
	if len(ChinaIP) < 0 {
		fmt.Println("初始化中国IP段失败", "")
		os.Exit(0)
	}
}

func FindChinaIP(IPtmp uint32) bool {
	for i := 0; i < len(ChinaIP); i++ {
		if (IPtmp >= ChinaIP[i].ip) &&
			(IPtmp <= ChinaIP[i].off) {
			return true
		}
	}
	return false
}

