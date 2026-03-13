package usecase

import (
	"context"
	"testing"
	"time"

	"magua-fly-golang/internal/domain"
)

type fakeSmsRepo struct {
	lastCode *domain.SmsVerificationCode
}

func (f *fakeSmsRepo) Create(ctx context.Context, v *domain.SmsVerificationCode) error {
	f.lastCode = v
	return nil
}

func (f *fakeSmsRepo) FindLatestByPhoneScene(ctx context.Context, phone string, scene domain.SmsScene) (*domain.SmsVerificationCode, error) {
	return f.lastCode, nil
}

func (f *fakeSmsRepo) FindValidCode(ctx context.Context, phone string, scene domain.SmsScene, code string, now time.Time) (*domain.SmsVerificationCode, error) {
	if f.lastCode != nil && f.lastCode.Code == code {
		return f.lastCode, nil
	}
	return nil, ErrInvalidCredentials
}

func (f *fakeSmsRepo) MarkUsed(ctx context.Context, id int64, usedAt time.Time) error {
	return nil
}

type fakeSmsClient struct {
	sent bool
}

func (f *fakeSmsClient) SendCode(ctx context.Context, phoneE164 string, templateID string, params []string) error {
	f.sent = true
	return nil
}

// Test basic send path to ensure we generate a code and call client.
func TestSendSmsCode_Basic(t *testing.T) {
	smsRepo := &fakeSmsRepo{}
	smsClient := &fakeSmsClient{}

	uc := &AuthUseCase{
		smsRepo:   smsRepo,
		smsClient: smsClient,
	}

	err := uc.SendSmsCode(context.Background(), SendSmsCodeRequest{
		Phone: "13800138000",
		Scene: SmsSceneLogin,
	})
	if err != nil {
		t.Fatalf("SendSmsCode error: %v", err)
	}
	if smsRepo.lastCode == nil {
		t.Fatalf("expected code record created")
	}
	if !smsClient.sent {
		t.Fatalf("expected sms client to be called")
	}
}

