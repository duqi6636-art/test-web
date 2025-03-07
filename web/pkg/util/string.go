package util

import (
	"github.com/unknwon/com"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 分割字符串
func Split(str, sep string) []string {
	return strings.Split(str, sep)
}

// 手机号加*
func PhoneFmt(phone string) string {
	slice := []byte(phone)
	return string(slice[0:3]) + "****" + string(slice[7:])
}

// 去除字符串两端空格
func Trim(str string) string {
	return strings.Trim(str, " ")
}

// string 包含
func StrContain(target, char string) bool {
	return strings.Contains(target, char)
}

// 检查字符串切片是否包含某个值
func InArrayString(elem string, list []string) bool {
	for _, val := range list {
		if val == elem {
			return true
		}
	}
	return false
}

// 检查字符串切片是否包含某个值
func InArrayInt(elem int, list []int) bool {
	for _, val := range list {
		if val == elem {
			return true
		}
	}
	return false
}

// 获取全局唯一order_id
func GetOrderId() string {
	formatLayout := "20060102150405"
	orderNo := time.Now().Format(formatLayout)
	r := RandInt(100000, 999999)
	return orderNo + com.ToStr(r)
}

func Sif(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

func IfInt(condition bool, falseVal, trueVal int) int {
	if condition {
		return trueVal
	}
	return falseVal
}

func StoF(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return v
	}
	return 0.0
}

func FtoS(s float64) string {
	str := ""
	str = strconv.FormatFloat(s, 'f', 0, 64)
	return str
}

func FtoS2(s float64, num int) string {
	str := ""
	str = strconv.FormatFloat(s, 'f', num, 64)
	return str
}

func ItoS(s int) string {
	v := strconv.Itoa(s)
	return v
}
func StoI(s string) int {
	v, err := strconv.Atoi(s)
	if err == nil {
		return v
	}
	return 0
}

// 去除字符串中的html标签
func TrimHtml(src string) string {
	//将HTML标签全转换成小写
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllStringFunc(src, strings.ToLower)
	//去除STYLE
	re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	src = re.ReplaceAllString(src, "")
	//去除SCRIPT
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	src = re.ReplaceAllString(src, "")
	//去除所有尖括号内的HTML代码，并换成换行符
	re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllString(src, "\n")
	//去除连续的换行符
	re, _ = regexp.Compile("\\s{2,}")
	src = re.ReplaceAllString(src, "\n")
	return strings.TrimSpace(src)
}

// 三元 实现
func Mif(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

func RemoveParentheses(s string) string {
	// 定义正则表达式模式，用于匹配括号及其内部的内容
	re := regexp.MustCompile(`\([^)]*\)`)
	// 使用空字符串替换匹配到的内容
	return strings.TrimSpace(re.ReplaceAllString(s, ""))
}
