package util

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type PhoneCode struct {
	PhoneCode string
	phone     string
}

// 获取当前地址
func GetCurrPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

// 检测邮箱格式
func CheckEmail(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// 检测密码格式
func CheckPwd(password string) bool {
	pattern := `^[a-zA-Z0-9-*/+.~!@#$%^&*()]{6,20}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(password)
}

// 检测密码格式
func CheckPwdNew(password string) (bool, string) {
	if len(password) < 6 || len(password) > 20 {
		return false, "__T_PASSWORD_FORMAT"
	}
	pattern := `[a-zA-Z]`
	reg := regexp.MustCompile(pattern)
	has := reg.MatchString(password)
	//fmt.Println("char ====",has)
	if !has {
		return false, "__T_PASSWORD_FORMAT_LETTER"
	}
	//pattern2 := `\d+`
	//reg2 := regexp.MustCompile(pattern2)
	//has2 := reg2.MatchString(password)
	////fmt.Println("num ====",has2)
	//if !has2 {
	//	return false
	//}
	//pattern3 := `[!@#$%&*()_+\-=\[\]{};':"\\|,.<>\/?]`
	//reg3 := regexp.MustCompile(pattern3)
	//has3 := reg3.MatchString(password)
	////fmt.Println("has3 ====",has3)
	//if !has3 {
	//	return false,"__T_PASSWORD_FORMAT_CHAR"
	//}
	pattern3 := `[0-9]`
	reg3 := regexp.MustCompile(pattern3)
	has3 := reg3.MatchString(password)
	if !has3 {
		return false, "__T_PASSWORD_FORMAT_DIGHT"
	}

	pattern4 := `[\p{Han}]+`
	reg4 := regexp.MustCompile(pattern4)
	has4 := reg4.MatchString(password)
	if has4 {
		return false, "__T_PASSWORD_FORMAT_ERROR"
	}

	return true, "OK"
}

// 检测密码格式
func CheckNewPwd(password string) bool {
	pattern := `^[a-zA-Z0-9]{6,20}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(password)
}

// 校验代理账户格式
func CheckUserAccount(account string) bool {
	pattern := `^[a-zA-Z0-9_]{6,20}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(account)
}

// 校验代理账户密码格式
func CheckUserPassword(account string) bool {
	pattern := `^[a-zA-Z0-9_]{6,24}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(account)
}

// 检测邀请码格式
func CheckCode(password string) bool {
	pattern := `^[a-zA-Z0-9]{3,20}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(password)
}

// 检测手机验证码格式
func VerifyPhone(phone string) bool {
	regular := "^1[3-9]\\d{9}$"
	reg := regexp.MustCompile(regular)
	return reg.MatchString(phone)
}

func GetRandomInt(min int, max int) int {
	return rand.Intn(max-min) + min
}

/**
 * 手机号加*
 */
func MobileReplaceRep(str string) string {
	re, _ := regexp.Compile("(\\d{3})(\\d{4})(\\d{4})")
	return re.ReplaceAllString(str, "$1****$3")
}

// 生成指定区间随机数（包括纯数字／纯字母／随机）
func Kand(size int, kind int) []byte {
	i_kind, kinds, result := kind, [][]int{[]int{10, 48}, []int{26, 97}, []int{26, 65}}, make([]byte, size)
	is_all := kind > 2 || kind < 0
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		if is_all { // random ikind
			i_kind = rand.Intn(3)
		}
		scope, base := kinds[i_kind][0], kinds[i_kind][1]
		result[i] = uint8(base + rand.Intn(scope))
	}
	return result
}

// 生成区间随机数
func RandInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return rand.Intn(max-min) + min
}

// md5加密
func Md5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// sha1加密
func SHA1(s string) string {
	o := sha1.New()
	o.Write([]byte(s))
	return hex.EncodeToString(o.Sum(nil))
}

// 获取重定向信息
func HttpGetRedirect(url2 string) string {
	location := ""
	req, err := http.NewRequest("GET", url2, nil)
	if err != nil {
		return location
	}
	c := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 3 * time.Second,
	}
	resp, err := c.Do(req)

	if resp.StatusCode == 302 {
		location = resp.Header.Get("Location")
	}

	return location

}

// 支付使用，获取重定向后的url
func HttpGetReqUrl(url2 string) string {
	location := ""
	req, err := http.NewRequest("GET", url2, nil)
	if err != nil {
		return location
	}
	c := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := c.Do(req)

	if err != nil {
		return ""
	}
	if resp.StatusCode != 200 {
		return ""
	}
	return resp.Request.URL.String()

}

// 发送POST请求
// url:请求地址，data:POST请求提交的数据,
// contentType:
// (1)application/x-www-form-urlencoded  最常见的POST提交数据的方式，浏览器的原生form表单。后面可以跟charset=utf-8
// (2)multipart/form-data
// (3)application/json
// (4)text/xml    XML-RPC远程调用
// content:请求放回的内容
func HttpPost(url string, data interface{}, contentType string) (content string) {
	jsonStr, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Add("content-type", contentType)
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)
	content = string(result)
	return
}

func HttpPostForm(postUrl string, param map[string]string) string {
	data := make(url.Values)
	for k, v := range param {
		data[k] = []string{v}
	}
	resp, err := http.PostForm(postUrl, data)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return string(body)
}

func HttpGET(url string) string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || err != nil {
		panic(err)
	}
	return string(body)
}

// HttpPostMultiPart
func HttpPostMultiPart(url string, path string, file multipart.File, fileName string) (content string, err error) {
	// 创建一个字节缓冲区，用于存储请求体
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// 创建表单中的文件字段
	fileWriter, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return
	}
	// 将文件内容复制到表单中的文件字段
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return
	}
	writer.WriteField("path", path)
	// 必须关闭 writer，以便写入结尾的 boundary
	writer.Close()
	// 创建 POST 请求
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return
	}
	// 设置请求头 Content-Type
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// 发送请求并获取响应
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	content = string(result)
	return
}
