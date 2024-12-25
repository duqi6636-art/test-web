package util

import (
	"fmt"
	"strconv"
	"strings"
)

func MatchIp(ip, iprange string) bool {
	ipb := ip2binary(ip)
	ipr := strings.Split(iprange, "/")
	masklen, err := strconv.ParseUint(ipr[1], 10, 32)
	if err != nil {
		fmt.Println(err)
		return false
	}
	iprb := ip2binary(ipr[0])
	return strings.EqualFold(ipb[0:masklen], iprb[0:masklen])
}

//将IP地址转化为二进制String
func ip2binary(ip string) string {
	str := strings.Split(ip, ".")
	var ipstr string
	for _, s := range str {
		i, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			fmt.Println(err)
		}
		ipstr = ipstr + fmt.Sprintf("%08b", i)
	}
	return ipstr
}
