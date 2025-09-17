package controller

import (
	"api-360proxy/web/e"
	"api-360proxy/web/models"
	"api-360proxy/web/pkg/util"
	"api-360proxy/web/service/onfido"
	"api-360proxy/web/service/tencent"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// UserKYC is the controller for user KYC
// 实名认证相关接口

// 实名有效期
const CertValidTime = 190

type RealNameAuthForm struct {
	FirstName string `json:"first_name" form:"first_name" binding:"required"`
	LastName  string `json:"last_name" form:"last_name" binding:"required"`
	Lang      string `json:"lang" form:"lang"`
}

// IdVerifyStepOne 实名认证第一步,创建申请人(applicant_id为空时调用)
func IdVerifyStepOne(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	// 滑块验证 -- start
	captchaSwitch := strings.TrimSpace(models.GetConfigVal("CaptchaIdVerifySwitch")) // 滑块验证认证开关
	if captchaSwitch == "1" {
		ticket := c.DefaultPostForm("ticket", "")
		randStr := c.DefaultPostForm("randstr", "")
		if ticket == "" || randStr == "" {
			JsonReturn(c, e.ERROR, "__T_CAPTCHA_FAIL", nil)
			return
		}

		res, Msg := CaptchaHandle(c, ticket, randStr)
		if !res {
			JsonReturn(c, e.ERROR, Msg, nil)
			return
		}
	}
	// 滑块验证 -- end

	first_name := strings.TrimSpace(c.DefaultPostForm("first_name", ""))
	last_name := strings.TrimSpace(c.DefaultPostForm("last_name", ""))
	if first_name == "" {
		JsonReturn(c, e.ERROR, "__T_FIRST_NAME_ERR", nil)
		return
	}
	if last_name == "" {
		JsonReturn(c, e.ERROR, "__T_LAST_NAME_ERR", nil)
		return
	}
	uid := userInfo.Id

	todayVerifyCount := models.GetKycHistoryByCount(uid)
	numStr := models.GetConfigVal("onfido_limit_count")
	num := util.StoI(numStr)
	if num == 0 {
		num = 2
	}
	if todayVerifyCount >= num {
		JsonReturn(c, e.ERROR, "__T_USER_KYC_TODAY_LIMIT", nil)
		return
	}

	//判断是否已创建过申请人
	userKycInfo := models.GetUserKycByUid(uid)
	if userKycInfo.Status == "1" {
		JsonReturn(c, e.ERROR, "__T_USER_KYC_REPEATEDLY", nil)
		return
	}
	var person = onfido.Person{
		FirstName: first_name,
		LastName:  last_name,
	}
	applicantId, err := onfido.CreateApplicant(uid, person)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_FAIL-- "+err.Error(), nil)
		return
	}
	if userKycInfo.Uid > 0 {
		//记录历史数据
		models.AddKycHistoryData(models.UserKycHistory{
			Uid:               userKycInfo.Uid,
			ApplicantID:       userKycInfo.ApplicantID,
			FirstName:         userKycInfo.FirstName,
			LastName:          userKycInfo.LastName,
			LastWorkflowRunId: userKycInfo.LastWorkflowRunId,
			LinkUrl:           userKycInfo.LinkUrl,
			CertIssuePlace:    userKycInfo.CertIssuePlace,
			CertNumber:        userKycInfo.CertNumber,
			CertCate:          userKycInfo.CertCate,
			CertImages:        userKycInfo.CertImages,
			Status:            userKycInfo.Status,
			CreateTime:        int64(util.GetNowInt()),
		})
		//删除原有数据
		models.DeleteUserKycByUid(userInfo.Id)
	}

	nowTime := time.Now().Unix()
	userKycModels := models.UserKyc{}
	userKycModels.Uid = uid
	userKycModels.FirstName = first_name
	userKycModels.LastName = last_name
	userKycModels.ApplicantID = applicantId
	userKycModels.Status = "0"
	userKycModels.LinkUrl = ""
	userKycModels.CreateTime = nowTime
	userKycModels.UpdateTime = nowTime
	err = models.AddUserKycData(userKycModels)
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_FAIL-- "+err.Error(), nil)
		return
	}
	ret := map[string]string{"applicant_id": applicantId}
	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", ret)
	return
}

// IdVerifyStepTwo 实名认证第二步,创建工作流(获取实名地址)
func IdVerifyStepTwo(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id

	userKycInfo := models.GetUserKycByUid(uid)
	data := map[string]interface{}{}
	if userKycInfo.ApplicantID == "" {
		data["step"] = "step_one"
		JsonReturn(c, e.ERROR, "__T_FIRST_NAME_ERR", data)
		return
	}

	todayVerifyCount := models.GetKycHistoryByCount(uid)
	numStr := models.GetConfigVal("onfido_limit_count")
	num := util.StoI(numStr)
	if num == 0 {
		num = 2
	}
	if todayVerifyCount >= num {
		JsonReturn(c, e.ERROR, "__T_USER_KYC_TODAY_LIMIT", nil)
		return
	}
	if userKycInfo.Status == "1" {
		JsonReturn(c, e.ERROR, "__T_USER_KYC_REPEATEDLY", nil)
		return
	}
	if userKycInfo.LinkUrl != "" {
		// 生成二维码
		faceUrlNew := strings.TrimRight(models.GetConfigV("API_DOMAIN_URL"), "/") + "/kyc_qrcode?id=" + util.MdEncode(util.ItoS(uid), MdKey)
		faceQrcode := strings.TrimRight(models.GetConfigV("API_DOMAIN_URL"), "/") + "/qrcode?data=" + faceUrlNew
		data["face_url"] = faceUrlNew
		data["qrcode"] = faceQrcode
		JsonReturn(c, e.SUCCESS, "__T_SUCCESS", data)
		return
	}
	workflowRun, err := onfido.RunWorkflow(userKycInfo.ApplicantID)
	if err != nil {
		JsonReturn(c, e.ERROR, err.Error(), nil)
		return
	}
	url := workflowRun.Link.URL

	AddLogs("onfido_workflowRun", workflowRun.ID+"|||"+url) //写日志
	//redis.RedisClient.SetEx(cacheTag, url, 600)
	err = models.UpdateUserKycByUid(uid, map[string]interface{}{
		"last_workflow_run_id": workflowRun.ID,
		"link_url":             url,
		"update_time":          time.Now().Unix(),
	})
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_FAIL-- "+err.Error(), nil)
		return
	}

	// 生成二维码
	faceUrlNew := strings.TrimRight(models.GetConfigV("API_DOMAIN_URL"), "/") + "/kyc_qrcode?id=" + util.MdEncode(util.ItoS(uid), MdKey)
	faceQrcode := strings.TrimRight(models.GetConfigV("API_DOMAIN_URL"), "/") + "/qrcode?data=" + faceUrlNew

	data["face_url"] = faceUrlNew
	data["qrcode"] = faceQrcode

	JsonReturn(c, e.SUCCESS, "__T_SUCCESS", data)
	return
}

// IdVerifyStepThree 实名认证第三步,获取实名信息
func IdVerifyStepThree(c *gin.Context) {
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil)
		return
	}
	uid := userInfo.Id
	userKycInfo := models.GetUserKycByUid(uid)
	data := map[string]interface{}{}
	data["email"] = userInfo.Email
	data["cert_cate"] = ""
	data["cert_number"] = ""
	data["cert_images"] = ""
	data["first_name"] = ""
	data["last_name"] = ""
	data["cert_issue_place"] = ""
	data["expire_date"] = ""
	status := 0
	if userKycInfo.Status == "0" {
		status = 0
	} else if userKycInfo.Status == "1" {
		status = 1
		data["email"] = userInfo.Email
		data["cert_cate"] = userKycInfo.CertCate
		data["cert_number"] = userKycInfo.CertNumber
		data["cert_images"] = userKycInfo.CertImages
		data["first_name"] = userKycInfo.FirstName
		data["last_name"] = userKycInfo.LastName
		data["cert_issue_place"] = userKycInfo.CertIssuePlace
		data["expire_date"] = util.GetTime64Str(userKycInfo.ExpireTime, "Y-m-d")
	} else {
		status = 2
	}
	kycManualStatus := models.GetKycManualReviewByUid(uid)
	kycEnterpriseStatus := models.GetEnterpriseKycByUid(uid)
	data["kyc_enterprise"] = kycEnterpriseStatus.ReviewStatus
	data["kyc_individual"] = kycManualStatus.ReviewStatus

	data["status"] = status
	JsonReturn(c, 0, msg, data)
	return
}

// IdVerifyNotify 异步通知
func IdVerifyNotify(c *gin.Context) {
	defer func() {
		c.String(200, "ok")
	}()
	//验签
	signature := c.GetHeader("x-sha2-signature")
	reqBody, _ := io.ReadAll(c.Request.Body) //获取post的数据

	AddLogs("RealNameAuth_start", "info "+signature+string(reqBody)) //写日志
	ofdWebhookToken := models.GetConfigVal("onfido_webhook_token")
	calcSignature := GenerateHMACSHA256(reqBody, ofdWebhookToken)
	fmt.Println(calcSignature)
	if calcSignature != signature {
		AddLogs("RealNameAuth_step1", "Failed to check the sign"+calcSignature) //写日志
		return
	}
	//解析参数
	var reqDataCommon onfido.WebhookReqCommonParam
	err := json.Unmarshal(reqBody, &reqDataCommon)
	if err != nil {
		AddLogs("RealNameAuth_step2", "json unmarshal failed"+"") //写日志
		return
	}
	//判断结果
	if reqDataCommon.Payload.Action == "workflow_task.completed" {
		var reqDataTask onfido.WebhookReqTaskCommonParam
		err = json.Unmarshal(reqBody, &reqDataTask)
		if err != nil {
			AddLogs("RealNameAuth_step3", "json unmarshal failed2"+"") //写日志
			return
		}
		//根据工作流id查找认证信息

		workflowRunId := reqDataTask.Payload.Object.WorkflowRunId
		userKycInfo := models.GetUserKycBy(map[string]interface{}{"last_workflow_run_id": workflowRunId})
		if userKycInfo.Uid == 0 {
			AddLogs("RealNameAuth_step4", "not found workflowRunId"+workflowRunId) //写日志
			return
		}
		AddLogs("RealNameAuth_step5-------------------", "workflowRunId"+workflowRunId)
		switch reqDataTask.Payload.Object.TaskDefId {
		case "profile_data": //个人信息
			var reqDataTaskProfile onfido.WebhookReqTaskProfileDataParam
			err = json.Unmarshal(reqBody, &reqDataTaskProfile)
			if err != nil {
				AddLogs("RealNameAuth_step5", "json unmarshal failed3"+util.ItoS(userKycInfo.Uid)) //写日志
				return
			}
			//获取姓名
			firstName := reqDataTaskProfile.Payload.Resource.Output.FirstName
			lastName := reqDataTaskProfile.Payload.Resource.Output.LastName
			//更新数据
			err = models.UpdateUserKycByUid(userKycInfo.Uid, map[string]interface{}{
				"first_name":  firstName,
				"last_name":   lastName,
				"update_time": time.Now().Unix(),
			})
			if err != nil {
				AddLogs("RealNameAuth_step6", util.ItoS(userKycInfo.Uid)+" error5"+err.Error()) //写日志
				return
			}
			AddLogs("RealNameAuth_step7", util.ItoS(userKycInfo.Uid)+" info:profile_data"+"success") //写日志

		case "document_check": //证件信息
			var reqDataTaskDocument onfido.WebhookReqTaskDocumentParam
			err = json.Unmarshal(reqBody, &reqDataTaskDocument)
			if err != nil {
				AddLogs("RealNameAuth_step8", util.ItoS(userKycInfo.Uid)+" error6"+"json unmarshal failed4.") //写日志
				return
			}
			//下载证件信息
			var documentUrls []string
			documentIds := reqDataTaskDocument.Payload.Resource.Input.DocumentIds
			for _, documentId := range documentIds {
				documentUrl, err := onfido.DownloadDocument(documentId.Id, "")
				if err != nil {
					AddLogs("RealNameAuth_step9", util.ItoS(userKycInfo.Uid)+" error:"+err.Error()) //写日志
				}
				documentUrls = append(documentUrls, documentUrl)
			}
			documentUrlsJson, _ := json.Marshal(documentUrls)
			//获取个人信息
			documentType := reqDataTaskDocument.Payload.Resource.Output.Properties.DocumentType
			issuingCountry := reqDataTaskDocument.Payload.Resource.Output.Properties.IssuingCountry
			certNumber := reqDataTaskDocument.Payload.Resource.Output.Properties.DocumentNumber
			firstName := reqDataTaskDocument.Payload.Resource.Output.Properties.FirstName
			lastName := reqDataTaskDocument.Payload.Resource.Output.Properties.LastName
			////判断证件号是否已被认证
			//status := "0"
			//if models.CheckCertNumber(certNumber) {
			//	status = "4"
			//	//logging.Error("RealNameAuthNotify error:The certificate has been authenticated by other users")
			//	AddLogs("RealNameAuth_step10", util.ItoS(userKycInfo.Uid)+" error", "error:The certificate has been authenticated by other users") //写日志
			//}
			//更新数据
			err = models.UpdateUserKycByUid(userKycInfo.Uid, map[string]interface{}{
				"cert_issue_place": issuingCountry,
				"cert_cate":        documentType,
				"cert_number":      certNumber,
				"first_name":       firstName,
				"last_name":        lastName,
				"cert_images":      string(documentUrlsJson),
				"update_time":      time.Now().Unix(),
				//"status":           status,
			})
			if err != nil {
				AddLogs("RealNameAuth_step11", util.ItoS(userKycInfo.Uid)+" error7"+err.Error()) //写日志
				return
			}
			AddLogs("RealNameAuth_step12", util.ItoS(userKycInfo.Uid)+" info:document_check"+"success") //写日志

		case "face_check_motion": //人脸验证
			var reqDataTaskMotion onfido.WebhookReqTaskMotionParam
			err = json.Unmarshal(reqBody, &reqDataTaskMotion)
			if err != nil {
				log.Println("RealNameAuthNotify error6:", "json unmarshal failed4.")
				return
			}
			//下载证件信息
			var motionUrls []string
			motionIds := reqDataTaskMotion.Payload.Resource.Input.MotionVideoId
			for _, motionId := range motionIds {
				motionUrl, err := onfido.DownloadDocument(motionId.Id, "mp4")
				AddLogs("face_check_motion DownloadDocument", motionUrl)
				if err != nil {
					log.Println("RealNameAuthNotify error6:", "json unmarshal failed4.")
				}
				motionUrls = append(motionUrls, motionUrl)
			}
			log.Println("face_check_motion", motionUrls)
			motionUrlsJson, _ := json.Marshal(motionUrls)
			//更新数据
			AddLogs("face_check_motion", string(motionUrlsJson))
			err = models.UpdateUserKycByUid(userKycInfo.Uid, map[string]interface{}{
				"cert_video":  string(motionUrlsJson),
				"update_time": time.Now().Unix(),
			})
			if err != nil {
				AddLogs("FaceCheck_step1 UpdateUserKycByUid", util.ItoS(userKycInfo.Uid)+" error7") //写日志
				return
			}

		default:
			AddLogs("RealNameAuth_step13", util.ItoS(userKycInfo.Uid)+" unknown TaskDefId:"+reqDataTask.Payload.Object.TaskDefId) //写日志
			return
		}
	} else if reqDataCommon.Payload.Action == "workflow_run.completed" {
		var reqData onfido.WebhookReqRunParam
		err = json.Unmarshal(reqBody, &reqData)
		if err != nil {
			AddLogs("RealNameAuth_step14", "error8"+"json unmarshal failed4.") //写日志
			return
		}

		applicantId := reqData.Payload.Resource.ApplicantId
		userKycInfo := models.GetUserKycBy(map[string]interface{}{"applicant_id": applicantId})
		if userKycInfo.Uid == 0 {
			AddLogs("RealNameAuth_step15", "not found applicantId"+applicantId) //写日志
			return
		}
		if userKycInfo.Status == "4" {
			AddLogs("RealNameAuth_step16", util.ItoS(userKycInfo.Uid)+" error"+"The certificate has been authenticated by other users") //写日志
			return
		}
		if reqData.Payload.Resource.Status != "approved" {
			status := "2"
			_, userInfo := models.GetUserById(userKycInfo.Uid)
			_ = models.UpdateUserKycByUid(userKycInfo.Uid, map[string]interface{}{"status": status, "update_time": time.Now().Unix()})
			ddMsg := fmt.Sprintf("【cherryProxy】%s提交了实名认证，当前状态[%s]，请前往后台审核", userInfo.Username, reqData.Payload.Resource.Status)
			fmt.Println(ddMsg)
			AddLogs("RealNameAuth_step17", "info3 status="+reqData.Payload.Resource.Status+ddMsg) //写日志
			// 认证失败，向质量部传递第三方信息
			go func() {
				err := submitThirdPartyKycInfo(userKycInfo, "failed", reqData.Payload.Resource.Status)
				if err != nil {
					AddLogs("SubmitThirdParty_Failed", err.Error())
				} else {
					AddLogs("SubmitThirdParty_Failed", "success")
				}
			}()
			return
		}

		//修改用户实名状态为正常
		nowTime := time.Now().Unix()
		_ = models.UpdateUserKycByUid(userKycInfo.Uid, map[string]interface{}{
			"status":      "1",
			"update_time": nowTime,
			"expire_time": nowTime + CertValidTime*86400,
		})
		//清除缓存
		AddLogs("RealNameAuth_step18", "info"+"success") //写日志
		// 认证成功，向质量部传递第三方信息
		go func() {
			err := submitThirdPartyKycInfo(userKycInfo, "approved", reqData.Payload.Resource.Status)
			if err != nil {
				AddLogs("SubmitThirdParty_Success", err.Error())
			} else {
				AddLogs("SubmitThirdParty_Success", "success")
			}
		}()

		return
	} else {
		AddLogs("RealNameAuth_step19", "unknown action"+reqDataCommon.Payload.Action) //写日志
		return
	}

	return
}

// 返回结果结构体
type FaceVerificationResult struct {
	Code     string `json:"code"`     // 返回码
	OrderNo  string `json:"orderNo"`  // 订单号
	H5FaceId string `json:"h5faceId"` // 唯一标识
	NewSign  string `json:"newSign"`  // 签名
}

// TencentKycNotify 异步通知
func TencentKycNotify(c *gin.Context) {
	code := c.Query("code")
	orderNo := c.Query("orderNo")
	h5FaceId := c.Query("h5faceId")
	newSignature := c.Query("newSignature")
	AddLogs("TencentRealNameAuth_start", "code="+code+" orderNo="+orderNo+" h5FaceId="+h5FaceId+" newSignature="+newSignature) //写日志

	kycAppId := models.GetConfigVal("tencent_kyc_app_id")
	param := []string{kycAppId, orderNo, code}
	res := tencent.VerifyFace(param, newSignature)
	if res == false {
		AddLogs("TencentRealNameAuth_step3", "verify failed") //写日志
		return
	}
	//根据订单号查找认证信息
	userKycInfo := models.GetUserKycBy(map[string]interface{}{"applicant_id": h5FaceId, "last_workflow_run_id": orderNo})
	if userKycInfo.Uid == 0 {
		AddLogs("TencentRealNameAuth_step4", "not found orderNo"+orderNo) //写日志
		return
	}

	status := "1"
	thirdPartyStatus := "approved"
	if code != "0" {
		status = "-3"
		thirdPartyStatus = "failed"
	}
	//修改用户实名状态
	nowTime := time.Now().Unix()
	_ = models.UpdateUserKycByUid(userKycInfo.Uid, map[string]interface{}{
		"status":      status,
		"update_time": nowTime,
		"expire_time": nowTime + CertValidTime*86400,
	})
	AddLogs("RealNameAuth_step5", "success") //写日志
	// 向质量部传递第三方信息
	go func() {
		err := submitThirdPartyKycInfo(userKycInfo, thirdPartyStatus, code)
		if err != nil {
			AddLogs("SubmitThirdParty_Tencent", err.Error())
		} else {
			AddLogs("SubmitThirdParty_Tencent", "success")
		}
	}()

	return
}

func GenerateHMACSHA256(data []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// submitThirdPartyKycInfo 向质量部传递第三方KYC信息
func submitThirdPartyKycInfo(userKycInfo models.UserKyc, verifyStatus string, thirdPartyResult string) error {
	// 获取用户信息
	err, userInfo := models.GetUserById(userKycInfo.Uid)
	if err != nil || userInfo.Id == 0 {
		return fmt.Errorf("用户不存在")
	}

	// 获取认证相关的文件信息
	var certImages string
	var certVideo string
	var photoURL string
	var tencentNonce, tencentOrderNo string
	// 判断认证类型（国内/国外）
	isdomestic := isDomesticKyc(userKycInfo.LinkUrl)

	// 根据认证类型设置source值
	sourceValue := 2 // 默认为onfido
	if isdomestic {
		sourceValue = 4 // 腾讯认证
	}
	if isdomestic {
		// 腾讯认证：解析链接参数并获取照片
		tencentNonce, tencentOrderNo, err = parseTencentKycUrl(userKycInfo.LinkUrl)
		if err != nil {
			log.Printf("解析腾讯KYC链接失败: %v", err)
			return fmt.Errorf("解析腾讯KYC链接失败")
		}

		// 获取照片路由地址
		kycAppId := models.GetConfigVal("tencent_kyc_app_id")
		// 第一步：获取 KYC 访问令牌
		accessToken, err := getKycAccessToken()
		if err != nil {
			log.Printf("获取KYC访问令牌失败: %v", err)
			// 继续执行，不影响提交
		} else {
			// 第二步：获取 API 签名票据
			ticketList, err := getApiTicket(accessToken, "SIGN", "")
			if err != nil || len(ticketList) == 0 {
				log.Printf("获取API签名票据失败: %v", err)
				// 继续执行，不影响提交
			} else {
				ticket := ticketList[0].Value

				// 使用解析出的参数获取照片
				param := []string{tencentNonce, tencentOrderNo, "1.0.0", kycAppId}
				sign, err := getKycSign(param, ticket)
				if err == nil {
					VerifyReq := VerifyRequest{
						AppId:   kycAppId,
						OrderNo: tencentOrderNo,
						Version: "1.0.0",
						Nonce:   tencentNonce,
						Sign:    sign,
						GetFile: "2",
					}
					err, response := sendRequest(VerifyReq)
					if err == nil {
						photoURL, err = SaveTencentKycPhoto(response.Result.Photo, tencentOrderNo)
						if err != nil {
							log.Printf("保存照片失败: %v", err)
							return err
						} else {
							log.Printf("成功获取照片URL: %s", photoURL)
						}
					} else {
						log.Printf("查询照片失败: %v", err)
						return err
					}
				}
			}
		}
	} else {
		// Onfido认证：从数据库获取已下载的证件图片和视频
		// 解析JSON格式的URL数组并转换为逗号分割的字符串
		if userKycInfo.CertImages != "" {
			var imageUrls []string
			if err := json.Unmarshal([]byte(userKycInfo.CertImages), &imageUrls); err == nil {
				certImages = strings.Join(imageUrls, ",")
			} else {
				// 如果解析失败，直接使用原始值
				certImages = userKycInfo.CertImages
				log.Printf("解析证件图片JSON失败: %v", err)
			}
		}

		if userKycInfo.CertVideo != "" {
			var videoUrls []string
			if err := json.Unmarshal([]byte(userKycInfo.CertVideo), &videoUrls); err == nil {
				certVideo = strings.Join(videoUrls, ",")
			} else {
				// 如果解析失败，直接使用原始值
				certVideo = userKycInfo.CertVideo
				log.Printf("解析人脸视频JSON失败: %v", err)
			}
		}

		log.Printf("Onfido认证 - 证件图片: %s, 人脸视频: %s", certImages, certVideo)
	}

	// 获取用户历史购买记录类型
	orderTypes := getUserOrderTypes(userKycInfo.Uid)

	// 准备提交数据
	threeInfo := map[string]interface{}{
		"id":           userKycInfo.ApplicantID,
		"audit_status": getAuditStatus(verifyStatus), // 三方审核状态：1=通过，2=未通过
		"audit_time":   util.GetNowInt(),             // 审核时间
		"remark":       thirdPartyResult,             // 三方审核备注
	}
	// 根据认证类型添加相应的文件信息
	if isdomestic {
		// 腾讯认证：添加人脸照片
		if photoURL != "" {
			threeInfo["face_photo"] = photoURL
		}
		threeInfo["document_photo"] = ""
	} else {
		// Onfido认证：添加证件图片和人脸视频
		if certImages != "" {
			threeInfo["document_photo"] = certImages // 证件图片
		}
		if certVideo != "" {
			threeInfo["face_photo"] = certVideo // 人脸视频
		}
	}
	submitData := map[string]interface{}{
		"oa_platform_name":   "922proxy",
		"uid":                userKycInfo.Uid,
		"account":            userInfo.Username,
		"account_type":       1, // 1：普通账号，2：代理商
		"reg_time":           userInfo.CreateTime,
		"apply_time":         util.GetNowInt(),
		"verify_type":        1,           // 1：个人，2：企业
		"source":             sourceValue, // 2：onfido，3：平台自审，4：腾讯
		"order_type":         orderTypes,  // 用户历史购买记录类型
		"three_parties_info": threeInfo,
	}

	// 生成签名
	departmentId := models.GetConfigVal("third_party_department_id")
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signKey := models.GetConfigVal("third_party_sign_key")

	sign := generateThirdPartySign(departmentId, timestamp, signKey)

	// 调用第三方API
	apiUrl := models.GetConfigVal("third_party_kyc_api_url")
	if apiUrl == "" {
		return fmt.Errorf("第三方KYC接口URL未配置")
	}

	jsonData, err := json.Marshal(submitData)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %v", err)
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("departmentId", departmentId)
	req.Header.Set("timestamp", timestamp)
	req.Header.Set("sign", sign)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 处理响应
	if resp.StatusCode != 200 {
		return fmt.Errorf("API调用失败: %d", resp.StatusCode)
	}
	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应体失败: %v", err)
	}

	// 记录请求和响应日志
	AddLogs("submitThirdPartyKycInfo_Request", string(jsonData))
	AddLogs("submitThirdPartyKycInfo_Response", string(respBody))

	// 解析响应
	var response ThirdPartyKycResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if response.Code != 0 {
		return fmt.Errorf("API调用失败 code为: %v", response.Code)
	}
	return nil
}

// getUserOrderTypes 获取用户历史购买记录类型
func getUserOrderTypes(uid int) string {
	// 套餐类型映射表
	pakTypeNames := map[string]string{
		"normal":      "ISP Proxies",
		"agent":       "ISP Proxies (Enterprise)",
		"long_term":   "Static Residential Proxies",
		"flow":        "Residential Proxies",
		"flow_agent":  "Residential Proxies (Enterprise)",
		"flow_day":    "Unlimited Residential Proxies",
		"dynamic_isp": "Rotating ISP Proxies",
		"balance":     "Balance Recharge",
	}

	// 获取用户所有已支付订单的套餐类型
	orderList := models.GetOrderListBy(uid, 0, 0, 3, "") // pay_status=3表示已支付

	// 使用map去重
	pakTypeMap := make(map[string]bool)
	for _, order := range orderList {
		if order.PakType != "" {
			pakTypeMap[order.PakType] = true
		}
	}

	// 转换为英文描述数组
	var pakTypes []string
	for pakType := range pakTypeMap {
		if typeName, exists := pakTypeNames[pakType]; exists {
			pakTypes = append(pakTypes, typeName)
		} else {
			// 如果映射表中没有对应的类型，使用原始值
			pakTypes = append(pakTypes, pakType)
		}
	}

	// 如果没有购买记录，返回空字符串
	if len(pakTypes) == 0 {
		return ""
	}

	// 用逗号连接所有套餐类型
	return strings.Join(pakTypes, ",")
}

// getAuditStatus 根据验证状态获取审核状态
func getAuditStatus(verifyStatus string) int {
	if verifyStatus == "approved" {
		return 1 // 通过
	}
	return 2 // 未通过
}

// 获取腾讯人脸核验链接
func GetFaceUrl(c *gin.Context) {
	//获取用户信息
	resCode, msg, userInfo := DealUser(c) //处理用户信息
	if resCode != e.SUCCESS {
		JsonReturn(c, resCode, msg, nil) //返回错误信息
		return
	}
	name := c.DefaultPostForm("name", "")
	idNo := c.DefaultPostForm("id_no", "")

	if idNo == "" || name == "" {
		JsonReturn(c, e.ERROR, "__T_PARAM_ERROR", nil) //返回错误信息
		return
	}

	todayVerifyCount := models.GetKycHistoryByCount(userInfo.Id)
	numStr := models.GetConfigVal("onfido_limit_count")
	num := util.StoI(numStr)
	if num == 0 {
		num = 2
	}
	if todayVerifyCount >= num {
		JsonReturn(c, e.ERROR, "__T_USER_KYC_TODAY_LIMIT", nil)
		return
	}

	//获取用户人脸信息
	orderNo := "cherrykyx" + util.GetOrderId()
	userId := util.ItoS(userInfo.Id)

	faceUrl, err, faceId := tencent.GetH5FaceId(orderNo, name, idNo, userId)
	if err != nil {
		JsonReturn(c, e.ERROR, err.Error(), nil) //返回错误信息
		return
	}

	userKycInfo := models.GetUserKycByUid(userInfo.Id)
	if userKycInfo.Status == "1" {
		JsonReturn(c, e.ERROR, "__T_USER_KYC_REPEATEDLY", nil)
		return
	}

	if userKycInfo.Uid > 0 {
		//记录历史数据
		models.AddKycHistoryData(models.UserKycHistory{
			Uid:               userKycInfo.Uid,
			ApplicantID:       userKycInfo.ApplicantID,
			FirstName:         userKycInfo.FirstName,
			LastName:          userKycInfo.LastName,
			LastWorkflowRunId: userKycInfo.LastWorkflowRunId,
			LinkUrl:           userKycInfo.LinkUrl,
			CertIssuePlace:    userKycInfo.CertIssuePlace,
			CertNumber:        userKycInfo.CertNumber,
			CertCate:          userKycInfo.CertCate,
			CertImages:        userKycInfo.CertImages,
			Status:            userKycInfo.Status,
			CreateTime:        int64(util.GetNowInt()),
		})
		//删除原有数据
		models.DeleteUserKycByUid(userInfo.Id)
	}

	//写入认证记录
	models.AddUserKycData(models.UserKyc{
		Uid:               userInfo.Id,
		ApplicantID:       faceId,
		FirstName:         name,
		LastWorkflowRunId: orderNo,
		LinkUrl:           faceUrl,
		CertNumber:        idNo,
		CertCate:          "id_card",
		Status:            "0",
		CreateTime:        int64(util.GetNowInt()),
	})
	if err != nil {
		JsonReturn(c, e.ERROR, "__T_FAIL-- "+err.Error(), nil)
		return
	}
	faceUrlNew := strings.TrimRight(models.GetConfigV("API_DOMAIN_URL"), "/") + "/kyc_qrcode?id=" + util.MdEncode(userId, MdKey)
	faceQrcode := strings.TrimRight(models.GetConfigV("API_DOMAIN_URL"), "/") + "/qrcode?data=" + faceUrlNew
	data := map[string]interface{}{
		"face_url": faceUrlNew,
		"qrcode":   faceQrcode,
	}
	JsonReturn(c, e.SUCCESS, "", data) //返回人脸信息
	return
}

func GetKycCountry(c *gin.Context) {
	countryList := models.GetAllCountry(0, "")
	countryList = append(countryList, models.ExtractCountry{
		Name:    "China",
		Country: "CN",
		Sort:    0,
	})
	JsonReturn(c, e.SUCCESS, "", countryList) //返回人脸信息
	return
}

func KycQrcode(c *gin.Context) {
	uidStrCode := c.Query("id")
	uidStr := util.MdDecode(uidStrCode, MdKey)
	uid := util.StoI(uidStr)
	kyeInfo := models.GetUserKycByUid(uid)
	if kyeInfo.Uid > 0 {
		//跳转到人脸核验页面
		c.Redirect(http.StatusFound, kyeInfo.LinkUrl)
		return
	}
}
