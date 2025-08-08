package client

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

var auditStatusMap = map[int64]domain.AuditStatus{
	0:  domain.AuditStatusApproved,
	1:  domain.AuditStatusPending,
	2:  domain.AuditStatusPending,
	-1: domain.AuditStatusRejected,
}

var _ SmsClient = (*TencentSmsClient)(nil)

// TencentSmsClient 腾讯云短信客户端实现。
//
// SDK 安装：
//
//	go get -v -u github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111
type TencentSmsClient struct {
	client *sms.Client
	appId  *string
}

// Send 调用腾讯云短信发送接口。
//
// https://cloud.tencent.com/document/product/382/55981
func (tc *TencentSmsClient) Send(req SendReq) (SendResp, error) {
	if len(req.PhoneNumbers) == 0 {
		return SendResp{}, fmt.Errorf("%w: phone number should not be empty", errs.ErrInvalidParam)
	}

	phoneNumberSet := make([]*string, len(req.PhoneNumbers))
	for _, phoneNumber := range req.PhoneNumbers {
		fullPhoneNum := phoneNumber
		if !strings.HasPrefix(phoneNumber, "+") {
			fullPhoneNum = "+86" + phoneNumber
		}
		phoneNumPtr := fullPhoneNum
		phoneNumberSet = append(phoneNumberSet, &phoneNumPtr)
	}

	request := sms.NewSendSmsRequest()
	request.PhoneNumberSet = phoneNumberSet
	request.SmsSdkAppId = tc.appId
	request.TemplateId = &req.TemplateId
	request.SignName = &req.SignName

	if req.TemplateParams != nil {
		templateParamSet := make([]*string, len(req.TemplateParams))
		for _, param := range req.TemplateParams {
			tplParam := param
			templateParamSet = append(templateParamSet, &tplParam)
		}
		request.TemplateParamSet = templateParamSet
	}

	res, err := tc.client.SendSms(request)
	if err != nil {
		return SendResp{}, fmt.Errorf("%w: %w", ErrFailedToSendSms, err)
	}

	if len(res.Response.SendStatusSet) == 0 {
		return SendResp{}, fmt.Errorf("%w: no response from tencent", ErrFailedToSendSms)
	}

	sendResp := SendResp{
		RequestId: *res.Response.RequestId,
		Results:   make(map[string]SendResult),
	}

	for _, status := range res.Response.SendStatusSet {
		phoneNumber := strings.TrimPrefix(*status.PhoneNumber, "+86")
		sendResp.Results[phoneNumber] = SendResult{
			Code:    *status.Code,
			Message: *status.Message,
		}
	}
	return sendResp, nil
}

// CreateTemplate 调用腾讯云短信创建模板接口。
//
// https://cloud.tencent.com/document/product/382/55974
func (tc *TencentSmsClient) CreateTemplate(req CreateTplReq) (CreateTplResp, error) {
	request := sms.NewAddSmsTemplateRequest()

	// 模板名称
	request.TemplateName = &req.TplName
	// 模板内容
	request.TemplateContent = &req.TplContent

	// 模板类型
	smsType := uint64(req.TplType)
	request.SmsType = &smsType

	// 国际 / 港澳台短信
	request.International = &req.International

	// 模板备注
	request.Remark = &req.Remark

	response, err := tc.client.AddSmsTemplate(request)
	if err != nil {
		return CreateTplResp{}, fmt.Errorf("%w: %w", ErrFailedToCreateTpl, err)
	}

	return CreateTplResp{
		RequestId:  *response.Response.RequestId,
		TemplateId: *response.Response.AddTemplateStatus.TemplateId,
	}, nil
}

// QueryTemplateStatus 调用腾讯云短信查询模板状态接口。
//
// https://cloud.tencent.com/document/product/382/52067
func (tc *TencentSmsClient) QueryTemplateStatus(req QueryTplStatusReq) (QueryTplStatusResp, error) {
	request := sms.NewDescribeSmsTemplateListRequest()

	request.International = &req.International
	request.TemplateIdSet = make([]*uint64, len(req.TemplateIds))

	for i := range req.TemplateIds {
		templateId, err := strconv.ParseUint(req.TemplateIds[i], 10, 64)
		if err != nil {
			return QueryTplStatusResp{}, fmt.Errorf("%w: %w", ErrInvalidTemplateId, err)
		}
		request.TemplateIdSet = append(request.TemplateIdSet, &templateId)
	}

	resp, err := tc.client.DescribeSmsTemplateList(request)
	if err != nil {
		return QueryTplStatusResp{}, fmt.Errorf("%w: %w", ErrFailedToQueryTplStatus, err)
	}

	if len(resp.Response.DescribeTemplateStatusSet) == 0 {
		return QueryTplStatusResp{}, fmt.Errorf("%w: no template status found", ErrFailedToQueryTplStatus)
	}

	results := make(map[string]TplStatus)
	for i := range resp.Response.DescribeTemplateStatusSet {
		templateId := strconv.FormatInt(int64(*resp.Response.DescribeTemplateStatusSet[i].TemplateId), 10)
		results[templateId] = TplStatus{
			RequestId:   *resp.Response.RequestId,
			TemplateId:  templateId,
			AuditStatus: auditStatusMap[(*resp.Response.DescribeTemplateStatusSet[i].StatusCode)],
			Reason:      *resp.Response.DescribeTemplateStatusSet[i].ReviewReply,
		}
	}

	return QueryTplStatusResp{
		Results: results,
	}, nil
}

func NewTencentSmsClient(regionId, secretId, secretKey, appId string) *TencentSmsClient {
	client, err := sms.NewClient(common.NewCredential(secretId, secretKey), regionId, profile.NewClientProfile())
	if err != nil {
		panic(err)
	}
	return &TencentSmsClient{
		client: client,
		appId:  &appId,
	}
}
