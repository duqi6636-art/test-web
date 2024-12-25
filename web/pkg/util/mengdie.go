//梦蝶加解密算法
package util

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

//生成不通类型的随机数
func RandStr(k string, n int) string {
	// r, pattern := "", ""
	length := 0
	pattern := ""
	switch k {
	case "n":
		pattern = "1234567890"
		length = 9
		break
	case "s":
		pattern = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLOMNOPQRSTUVWXYZ"
		length = 52
		break
	case "r":
		pattern = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLOMNOPQRSTUVWXYZ"
		length = 62
		break
	case "y":
		pattern = "1234567890abcdefghijklmnopqrstuvwxyz"
		length = 35
		break
	case "a":
		pattern = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLOMNOPQRSTUVWXYZ"
		length = 62
		break
	default:
		pattern = "1234567890"
		length = 9
		break
	}
	str := ""
	for i := 0; i < n; i++ {
		//num := rand.Intn(len(pattern)-1)
		num := rand.Intn(length) //
		str = str + pattern[num:num+1]
	}
	return str
}


//校验用户密码是否正确
func ChkPass(passwordDecode, password ,username string) bool {
	return PassEncode(password,username,0) == passwordDecode
}


//解密密码
func DePass(password, mdToken string) string {
	return MdDecode(password, mdToken)
}

//获取加密后的密码
func GetEncodePass(password, token string) string {
	return MdEncode(password, token)
}

func PassEncode(password string, key string, expire int) string {
	var exp string = "0000000000"
	if expire != 0 {
		exp = fmt.Sprintf("%d%d", expire, time.Now().Unix())
	}
	var r = Md5(key)
	var c = 0
	var v = ""
	var str = exp + password

	for i := 0; i < len(str); i++ {
		if c == len(r) {
			c = 0
		}
		v += string(r[c])
		v += string(str[i] ^ r[c])
		c += 1
	}
	return base64.StdEncoding.EncodeToString([]byte(ed(v, key)))
}

func ed(str string, key string) string {
	keyMd5 := Md5(key)
	var v string
	i := 0
	for _, value := range str {
		if len(keyMd5) == i {
			i = 0
		}
		v += string(byte(value) ^ keyMd5[i])
		i++
	}
	return v
}

////获取加密后的radius密码
//func GetEncodeRadiusPass(password string) string {
//	return GetSnPass(password+"radius", 8)
//}

//获取加密后的radius密码
func GetEncodeRadiusPass(password string) string {
	//return  GetSnPass(password+"zhimadaili", 0)
	return Md5(password + "zhimadaili")
}

//随机生成区间范围的随机数
func MtRand(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Intn(max-min) + min
	return randNum
}

// 生成16位唯一SN
func GetSn() string {
	return GetSnPass("", 16)
}

// 根据传入值获取指定长度SN
func GetSnPass(word string, leng int) string {

	if word != "" {
		word = Md5(word)
	} else {
		word = Md5(string(time.Now().Unix()) + RandStr("r", 32))
	}

	if leng > 0 {
		word = Substr(word, len(word)-leng, len(word))
	}
	return word
}

//截取字符串 start 起点下标 end 终点下标(不包括)
func Substr(str string, start int, end int) string {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		return ""
		//panic("start is wrong")
	}

	if end < 0 || end > length {
		return ""
		//panic("end is wrong")
	}

	return string(rs[start:end])
}

//异或算法
func MdXor(str, key string) string {
	if str == "" || key == "" {
		return ""
	}
	//获取加密字符串的byte长度
	str_byte := []byte(str)
	str_lens := len(str_byte) //<-

	//获取key的字符串长度
	my_key := []byte(key)
	k_lens := len(my_key) //<-

	j := 0
	rt := []byte{}
	for i := 0; i < str_lens; i++ {
		j = i % k_lens
		rt = append(rt, str_byte[i]^my_key[j])
	}
	return string(rt)
}

//梦蝶跨平台加密算法
func MdEncode(str string, key string) string {
	//本密钥有效期时间戳
	time_str := GetTime("13")
	//随机13位字符串
	rand_str13 := RandStr("r", 13)
	//base64转码待加密内容
	str_byte := []byte(str)
	str_64 := base64.StdEncoding.EncodeToString(str_byte)
	//生成随机md5key
	rand_key_md5 := Md5(time_str + rand_str13 + key)
	//用随机字符加密,并且在尾部拼上异或的时间戳,格式为:解密内容+13位时间戳加密内容(随机13位字符串异或加密)+随机13位字符串(交换key异或加密)
	str_xor := MdXor(str_64, rand_key_md5) + MdXor(time_str, rand_str13) + MdXor(rand_str13, key)
	return base64.StdEncoding.EncodeToString([]byte(str_xor))
}

//梦蝶跨平台解密算法
func MdDecode(str string, key string) string {
	//找到空格替换为+号
	str = strings.Replace(str, " ", "+", -1)
	//base64解码待加密内容
	dec64_str, err := base64.StdEncoding.DecodeString(str)
	dec_str := string(dec64_str)

	if err != nil {
		fmt.Println(err)
	}
	//分离加密内容,随机密钥,随机字符串
	len_s := len(dec64_str)

	//13位字符串分离
	rand_str := MdXor(Substr(dec_str, len_s-13, len_s), key)
	//13位时间戳
	now_time := MdXor(Substr(dec_str, len_s-26, len_s-13), rand_str)

	//生成随机key
	rand_key_md5 := Md5(now_time + rand_str + key)

	//还原数据
	str_xor := MdXor(Substr(dec_str, 0, len_s-26), rand_key_md5)
	str_xor_b, _ := base64.StdEncoding.DecodeString(str_xor)
	return string(str_xor_b)
}

