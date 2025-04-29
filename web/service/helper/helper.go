package helper

import (
	"fmt"
	"github.com/imroc/req/v3"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func DownloadImage(imageURL, localFilePath,token string) error {
	//client1 := req.Client()
	client := req.C()
	client.SetCommonHeader("Authorization","Token token="+token)
	resp, err := client.R().SetOutputFile(localFilePath).Get(imageURL)
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	return nil
}

func DownloadImage11(imageURL, localFilePath,token string) error {
	resp,err := CurlGetHeader(imageURL,token)
	//resp, err := client.R().SetOutputFile(localFilePath).Get(imageURL)
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}
	fmt.Println(resp)
	//if resp.StatusCode != http.StatusOK {
	//	return fmt.Errorf("bad status code: %d", resp.StatusCode)
	//}
	return nil
}



func GetAppDir() string {
	appDir, err := os.Getwd()
	if err != nil {
		file, _ := exec.LookPath(os.Args[0])
		applicationPath, _ := filepath.Abs(file)
		appDir, _ = filepath.Split(applicationPath)
	}
	return appDir
}


func GetCurrentPath() string {
	dir, _ := os.Executable()
	exPath := filepath.Dir(dir)
	return exPath
}

// 发送Get请求
func CurlGetHeader(apiUrl string, apiToken string) (content string, err error) {
	// 创建 GET 请求
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return "", err
	}

	// 设置请求头
	req.Header.Set("Authorization", "Token token="+apiToken)

	// 使用 http.Client 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return "", err
	}

	// 打印响应内容
	content = string(body)
	return content, nil
}
