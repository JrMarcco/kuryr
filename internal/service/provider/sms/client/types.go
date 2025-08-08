package client

import (
	"fmt"

	"github.com/JrMarcco/kuryr/internal/domain"
)

//go:generate mockgen -source=./types.go -destination=./mock/sms_client.mock.go -package=clientmock -typed SmsClient

var (
	ErrInvalidTemplateId = fmt.Errorf("[kuryr] invalid template id")

	ErrFailedToSendSms        = fmt.Errorf("[kuryr] failed to send sms")
	ErrFailedToCreateTpl      = fmt.Errorf("[kuryr] failed to create sms template")
	ErrFailedToQueryTplStatus = fmt.Errorf("[kuryr] failed to query sms template status")
)

type SmsClient interface {
	Send(req SendReq) (SendResp, error)
	CreateTemplate(req CreateTplReq) (CreateTplResp, error)
	QueryTemplateStatus(req QueryTplStatusReq) (QueryTplStatusResp, error)
}

type SendStatus int32

// SendReq 发送短信请求。
type SendReq struct {
	PhoneNumbers   []string
	SignName       string
	TemplateId     string
	TemplateParams map[string]string
}

// SendResp 发送短信响应。
type SendResp struct {
	RequestId string
	Results   map[string]SendResult // receiver -> SendResult
}

// SendResult 短信发送结果。
//
// 每个手机号对应一个结果。
type SendResult struct {
	Code    string
	Message string
}

type TemplateType int32

// CreateTplReq 创建短信模板请求。
type CreateTplReq struct {
	TplType       TemplateType
	TplName       string
	TplContent    string
	International uint64 // 国际 / 港澳台短信: 0 -> 大陆短信; 1 -> 国际 / 港澳台短信
	Remark        string
}

// CreateTplResp 创建短信模板响应。
type CreateTplResp struct {
	RequestId  string
	TemplateId string
}

// QueryTplStatusReq 查询短信模板状态请求。
type QueryTplStatusReq struct {
	International uint64 // 国际 / 港澳台短信: 0 -> 大陆短信; 1 -> 国际 / 港澳台短信
	TemplateIds   []string
}

// QueryTplStatusResp 查询短信模板状态响应。
type QueryTplStatusResp struct {
	Results map[string]TplStatus
}

// TplStatus 短信模板状态。
type TplStatus struct {
	RequestId   string
	TemplateId  string
	AuditStatus domain.AuditStatus
	Reason      string
}
