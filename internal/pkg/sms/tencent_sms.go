package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"magua-fly-golang/internal/config"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

// Client defines the interface for sending SMS messages.
type Client interface {
	SendCode(ctx context.Context, phoneE164 string, templateID string, params []string) error
}

// TencentClient is an implementation of Client using Tencent Cloud SMS.
// 当内部 client 为空时，SendCode 会降级为仅打印日志（开发环境下使用）。
type TencentClient struct {
	sdkAppID string
	signName string
	client   *sms.Client
}

// NewTencentClient constructs a TencentClient from application config.
// 若未配置必要的短信参数，将返回一个 dev/noop 客户端，不会中断服务启动。
func NewTencentClient(cfg *config.Config) (*TencentClient, error) {
	if cfg.SmsSecretID == "" || cfg.SmsSecretKey == "" || cfg.SmsSdkAppID == "" || cfg.SmsSignName == "" {
		log.Printf("sms: config incomplete, using dev noop client")
		return &TencentClient{}, nil
	}

	cred := common.NewCredential(cfg.SmsSecretID, cfg.SmsSecretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"

	c, err := sms.NewClient(cred, cfg.SmsRegion, cpf)
	if err != nil {
		return nil, err
	}

	return &TencentClient{
		sdkAppID: cfg.SmsSdkAppID,
		signName: cfg.SmsSignName,
		client:   c,
	}, nil
}

// SendCode sends a single verification code SMS to the given E.164 formatted phone number.
func (c *TencentClient) SendCode(ctx context.Context, phoneE164 string, templateID string, params []string) error {
	if c.client == nil {
		// 开发/未配置模式：仅记录日志，便于本地联调
		log.Printf("sms(dev): send code to %s template=%s params=%v", phoneE164, templateID, params)
		return nil
	}

	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = common.StringPtr(c.sdkAppID)
	req.SignName = common.StringPtr(c.signName)
	req.TemplateId = common.StringPtr(templateID)
	req.TemplateParamSet = common.StringPtrs(params)
	req.PhoneNumberSet = common.StringPtrs([]string{phoneE164})

	resp, err := c.client.SendSmsWithContext(ctx, req)
	if err != nil {
		if sdkErr, ok := err.(*tcerrors.TencentCloudSDKError); ok {
			return fmt.Errorf("tencent sms error: %s", sdkErr.Error())
		}
		return err
	}

	b, _ := json.Marshal(resp.Response)
	_ = b // reserved for future structured logging

	if len(resp.Response.SendStatusSet) == 0 {
		return fmt.Errorf("empty sms send status")
	}
	status := resp.Response.SendStatusSet[0]
	if status.Code == nil || *status.Code != "Ok" {
		return fmt.Errorf("sms send failed: code=%s message=%s", ptrToString(status.Code), ptrToString(status.Message))
	}
	return nil
}

func ptrToString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

