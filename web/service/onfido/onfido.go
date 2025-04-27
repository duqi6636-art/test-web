package onfido

import (
	"api-360proxy/web/models"
	"api-360proxy/web/service/helper"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"
)

// CreateApplicant 创建申请人
func CreateApplicant(uid int, person Person) (string, error) {
	//判断申请人是否存在，不存在则创建
	userKycInfo := models.GetUserKycByUid(uid)
	applicantId := userKycInfo.ApplicantID
	if applicantId == "" {
		//创建申请人
		var uri = "https://api.eu.onfido.com/v3.6/applicants/"
		var data = map[string]interface{}{
			"first_name": person.FirstName,
			"last_name":  person.LastName,
			//"dob":        person.DOB,
			//"address": map[string]string{
			//	"building_number": person.Address.BuildingNumber,
			//	"street":          person.Address.Street,
			//	"town":            person.Address.Town,
			//	"postcode":        person.Address.Postcode,
			//	"country":         person.Address.Country,
			//},
		}

		ofdToken := models.GetConfigVal("onfido_token")
		resp, err := HttpPostByAuth(uri, ofdToken, data)
		fmt.Println("resp1:", resp)
		if err != nil {
			return "", err
		}

		var respData = Applicant{}
		err = json.Unmarshal([]byte(resp), &respData)
		if err != nil {
			return "", err
		}
		applicantId = respData.ID
	}
	return applicantId, nil
}

func HttpPostByAuth(url string, authToken string, data interface{}) (content string, err error) {
	jsonStr, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Token token="+authToken)

	defer req.Body.Close()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)
	content = string(result)
	return content, nil
}

// RunWorkflow 运行工作流
func RunWorkflow(applicantId string) (WorkflowRun, error) {
	var respData = WorkflowRun{}
	//创建工作流运行
	var uri = "https://api.eu.onfido.com/v3.6/workflow_runs"
	ofdWorkflowId := models.GetConfigVal("onfido_workflow_id")
	ofdToken := models.GetConfigVal("onfido_token")
	redirectUrl := models.GetConfigVal("onfido_redirect_url")
	var data = map[string]interface{}{
		"workflow_id":  ofdWorkflowId,
		"applicant_id": applicantId,
		"link": map[string]string{
			"completed_redirect_url": redirectUrl + "?IdVerifyStatus=success",
		},
	}
	resp, err := HttpPostByAuth(uri, ofdToken, data)
	if err != nil {
		return respData, err
	}
	err = json.Unmarshal([]byte(resp), &respData)
	if err != nil {
		return respData, err
	}

	return respData, nil
}

// DownloadDocument 下载证件图片
func DownloadDocument(documentId string) (string, error) {
	var uri = "https://api.eu.onfido.com/v3.6/documents/" + documentId + "/download"
	// 创建一个新的文件
	appDir := helper.GetAppDir()
	if runtime.GOOS != "windows" {
		appDir = helper.GetCurrentPath()
	}
	date := time.Now().Format("2006_01_02")
	filePath := appDir + "/static/kyc/" + date
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err = os.Mkdir(filePath, 0755)
		if err != nil {
			// 处理创建文件夹失败的错误
			return "", errors.New(err.Error())
		}
	}
	fileName := fmt.Sprintf("%d.jpeg", time.Now().Unix())
	ofdToken := models.GetConfigVal("onfido_token")
	err := helper.DownloadImage(uri, filePath+"/"+fileName, ofdToken)

	if err != nil {
		fmt.Println("下载图片:", err)
		return "", err
	}
	//访问地址
	apiHref := models.GetConfigVal("onfido_api_href")
	url := apiHref + "/static/kyc/" + date + "/" + fileName

	return url, nil
}

// 创建申请人-请求参数结构体
type Person struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	DOB       string  `json:"dob"`
	Address   Address `json:"address"`
}
type Address struct {
	BuildingNumber string `json:"building_number"`
	Street         string `json:"street"`
	Town           string `json:"town"`
	Postcode       string `json:"postcode"`
	Country        string `json:"country"`
}

// 创建申请人-返回参数结构体
type AddressRes struct {
	FlatNumber     interface{} `json:"flat_number"`
	BuildingNumber interface{} `json:"building_number"`
	BuildingName   interface{} `json:"building_name"`
	Street         string      `json:"street"`
	SubStreet      interface{} `json:"sub_street"`
	Town           string      `json:"town"`
	State          interface{} `json:"state"`
	Postcode       string      `json:"postcode"`
	Country        string      `json:"country"`
	Line1          interface{} `json:"line1"`
	Line2          interface{} `json:"line2"`
	Line3          interface{} `json:"line3"`
}

type Location struct {
	IPAddress          string `json:"ip_address"`
	CountryOfResidence string `json:"country_of_residence"`
}

type Applicant struct {
	ID          string      `json:"id"`
	CreatedAt   string      `json:"created_at"`
	Sandbox     bool        `json:"sandbox"`
	FirstName   string      `json:"first_name"`
	LastName    string      `json:"last_name"`
	Email       interface{} `json:"email"`
	DOB         string      `json:"dob"`
	DeleteAt    interface{} `json:"delete_at"`
	Href        string      `json:"href"`
	IDNumbers   []string    `json:"id_numbers"`
	PhoneNumber string      `json:"phone_number"`
	Address     AddressRes  `json:"address"`
	Location    Location    `json:"location"`
}

// 创建工作流运行-返回参数结构体
type Link struct {
	CompletedRedirectURL string `json:"completed_redirect_url"`
	ExpiredRedirectURL   string `json:"expired_redirect_url"`
	ExpiresAt            string `json:"expires_at"`
	Language             string `json:"language"`
	URL                  string `json:"url"`
}

type WorkflowRun struct {
	ID                string                 `json:"id"`
	ApplicantID       string                 `json:"applicant_id"`
	WorkflowID        string                 `json:"workflow_id"`
	WorkflowVersionID int                    `json:"workflow_version_id"`
	Status            string                 `json:"status"`
	DashboardURL      string                 `json:"dashboard_url"`
	Output            map[string]interface{} `json:"output"`
	Reasons           []string               `json:"reasons"`
	Error             interface{}            `json:"error"`
	CreatedAt         string                 `json:"created_at"`
	UpdatedAt         string                 `json:"updated_at"`
	Link              Link                   `json:"link"`
}

type WebhookReqCommonParam struct {
	Payload struct {
		ResourceType string `json:"resource_type"`
		Action       string `json:"action"`
	} `json:"payload"`
}

type WebhookReqRunParam struct {
	Payload struct {
		ResourceType string `json:"resource_type"`
		Action       string `json:"action"`
		Object       struct {
			Id                 string    `json:"id"`
			Status             string    `json:"status"`
			CompletedAtIso8601 time.Time `json:"completed_at_iso8601"`
			Href               string    `json:"href"`
		} `json:"object"`
		Resource struct {
			DashboardUrl      string        `json:"dashboard_url"`
			WorkflowVersionId int           `json:"workflow_version_id"`
			ApplicantId       string        `json:"applicant_id"`
			CreatedAt         time.Time     `json:"created_at"`
			Reasons           []interface{} `json:"reasons"`
			Link              struct {
				Language             interface{} `json:"language"`
				CompletedRedirectUrl interface{} `json:"completed_redirect_url"`
				ExpiresAt            interface{} `json:"expires_at"`
				ExpiredRedirectUrl   interface{} `json:"expired_redirect_url"`
				Url                  string      `json:"url"`
			} `json:"link"`
			Error      interface{} `json:"error"`
			Output     interface{} `json:"output"`
			Id         string      `json:"id"`
			UpdatedAt  time.Time   `json:"updated_at"`
			Status     string      `json:"status"`
			WorkflowId string      `json:"workflow_id"`
		} `json:"resource"`
	} `json:"payload"`
}

type WebhookReqTaskCommonParam struct {
	Payload struct {
		ResourceType string `json:"resource_type"`
		Action       string `json:"action"`
		Object       struct {
			Id                 string    `json:"id"`
			TaskSpecId         string    `json:"task_spec_id"`
			TaskDefId          string    `json:"task_def_id"`
			WorkflowRunId      string    `json:"workflow_run_id"`
			Status             string    `json:"status"`
			CompletedAtIso8601 time.Time `json:"completed_at_iso8601"`
			Href               string    `json:"href"`
		} `json:"object"`
	}
}

type WebhookReqTaskDocumentParam struct {
	Payload struct {
		ResourceType string `json:"resource_type"`
		Action       string `json:"action"`
		Object       struct {
			Id                 string    `json:"id"`
			TaskSpecId         string    `json:"task_spec_id"`
			TaskDefId          string    `json:"task_def_id"`
			WorkflowRunId      string    `json:"workflow_run_id"`
			Status             string    `json:"status"`
			CompletedAtIso8601 time.Time `json:"completed_at_iso8601"`
			Href               string    `json:"href"`
		} `json:"object"`
		Resource struct {
			Input struct {
				DocumentIds []struct {
					ChecksumSha256 string `json:"checksum_sha256"`
					Id             string `json:"id"`
					Type           string `json:"type"`
				} `json:"document_ids"`
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
			} `json:"input"`
			CreatedAt      time.Time   `json:"created_at"`
			UpdatedAt      time.Time   `json:"updated_at"`
			TaskDefVersion interface{} `json:"task_def_version"`
			Id             string      `json:"id"`
			TaskDefId      string      `json:"task_def_id"`
			Output         struct {
				Status     string `json:"status"`
				Result     string `json:"result"`
				SubResult  string `json:"sub_result"`
				Uuid       string `json:"uuid"`
				Properties struct {
					DocumentType   string `json:"document_type"`
					IssuingCountry string `json:"issuing_country"`
					FirstName      string `json:"first_name"`
					LastName       string `json:"last_name"`
					DocumentNumber string `json:"document_number"`
				} `json:"properties"`
			} `json:"output"`
			WorkflowRunId string `json:"workflow_run_id"`
		} `json:"resource"`
	} `json:"payload"`
}

type WebhookReqTaskProfileDataParam struct {
	Payload struct {
		ResourceType string `json:"resource_type"`
		Action       string `json:"action"`
		Object       struct {
			Id                 string    `json:"id"`
			TaskSpecId         string    `json:"task_spec_id"`
			TaskDefId          string    `json:"task_def_id"`
			WorkflowRunId      string    `json:"workflow_run_id"`
			Status             string    `json:"status"`
			CompletedAtIso8601 time.Time `json:"completed_at_iso8601"`
			Href               string    `json:"href"`
		} `json:"object"`
		Resource struct {
			CreatedAt      time.Time   `json:"created_at"`
			TaskDefId      string      `json:"task_def_id"`
			Id             string      `json:"id"`
			TaskDefVersion interface{} `json:"task_def_version"`
			Output         struct {
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
				Address   struct {
				} `json:"address"`
			} `json:"output"`
			WorkflowRunId string `json:"workflow_run_id"`
			Input         struct {
			} `json:"input"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"resource"`
	} `json:"payload"`
}
