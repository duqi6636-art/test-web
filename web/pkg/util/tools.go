package util

import (
	"math/big"
	"net"
)

// 关联name生成用户名
func GenSnWithName(name string) string {
	return Md5(GetSn() + name)
}

//判断元素是否在数组中存在
func InArray(one interface{}, arr interface{}) bool {
	switch arr.(type) {
	case []string:
		tmp := map[string]string{}
		a := one.(string)
		b := arr.([]string)
		Blen := len(b)
		for i := 0; i < Blen; i++ {
			tmp[b[i]] = "ok"
		}
		if tmp[a] == "ok" {
			return true
		}
		break
	case []int:
		tmp := map[int]string{}
		a := one.(int)
		b := arr.([]int)
		Blen := len(b)
		for i := 0; i < Blen; i++ {
			tmp[b[i]] = "ok"
		}
		if tmp[a] == "ok" {
			return true
		}
		break
	case []int64:
		tmp := map[int64]string{}
		a := one.(int64)
		b := arr.([]int64)
		Blen := len(b)
		for i := 0; i < Blen; i++ {
			tmp[b[i]] = "ok"
		}
		if tmp[a] == "ok" {
			return true
		}
		break
	case []float32:
		tmp := map[float32]string{}
		a := one.(float32)
		b := arr.([]float32)
		Blen := len(b)
		for i := 0; i < Blen; i++ {
			tmp[b[i]] = "ok"
		}
		if tmp[a] == "ok" {
			return true
		}
		break
	case []float64:
		tmp := map[float64]string{}
		a := one.(float64)
		b := arr.([]float64)
		Blen := len(b)
		for i := 0; i < Blen; i++ {
			tmp[b[i]] = "ok"
		}
		if tmp[a] == "ok" {
			return true
		}
		break
	default:
	}
	return false
}

/*
字符串ip转整形IP
*/
func InetAtoN(ip string) uint32 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return uint32(ret.Uint64())
}