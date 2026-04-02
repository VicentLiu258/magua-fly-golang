package usecase

import (
	"math"
	"testing"

	"magua-fly-golang/internal/domain"
)

func TestValidateContractPayment_full_ok(t *testing.T) {
	pt := 100000.0
	err := ValidateContractPayment(&pt, domain.PaymentModeFull, nil)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	err = ValidateContractPayment(&pt, domain.PaymentModeFull, []domain.PaymentInstallment{})
	if err != nil {
		t.Fatalf("expected nil for empty slice, got %v", err)
	}
}

func TestValidateContractPayment_full_with_schedule(t *testing.T) {
	pt := 100000.0
	err := ValidateContractPayment(&pt, domain.PaymentModeFull, []domain.PaymentInstallment{
		{Amount: 100000, DueDate: "2026-01-01"},
	})
	if err != ErrInvalidPaymentSchedule {
		t.Fatalf("expected ErrInvalidPaymentSchedule, got %v", err)
	}
}

func TestValidateContractPayment_installment_ok(t *testing.T) {
	pt := 150000.0
	err := ValidateContractPayment(&pt, domain.PaymentModeInstallment, []domain.PaymentInstallment{
		{Amount: 50000, DueDate: "2026-04-01"},
		{Amount: 50000, DueDate: "2026-07-01"},
		{Amount: 50000, DueDate: "2026-10-01"},
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateContractPayment_installment_too_few(t *testing.T) {
	pt := 100000.0
	err := ValidateContractPayment(&pt, domain.PaymentModeInstallment, []domain.PaymentInstallment{
		{Amount: 100000, DueDate: "2026-04-01"},
	})
	if err != ErrInvalidPaymentSchedule {
		t.Fatalf("expected ErrInvalidPaymentSchedule, got %v", err)
	}
}

func TestValidateContractPayment_installment_no_price(t *testing.T) {
	err := ValidateContractPayment(nil, domain.PaymentModeInstallment, []domain.PaymentInstallment{
		{Amount: 50000, DueDate: "2026-04-01"},
		{Amount: 50000, DueDate: "2026-07-01"},
	})
	if err != ErrPriceTotalRequired {
		t.Fatalf("expected ErrPriceTotalRequired, got %v", err)
	}
}

func TestValidateContractPayment_sum_mismatch(t *testing.T) {
	pt := 150000.0
	err := ValidateContractPayment(&pt, domain.PaymentModeInstallment, []domain.PaymentInstallment{
		{Amount: 50000, DueDate: "2026-04-01"},
		{Amount: 40000, DueDate: "2026-07-01"},
	})
	if err != ErrPaymentScheduleSumMismatch {
		t.Fatalf("expected ErrPaymentScheduleSumMismatch, got %v", err)
	}
}

func TestValidateContractPayment_sum_within_epsilon(t *testing.T) {
	pt := 100.0 + paymentSumEpsilon/2
	err := ValidateContractPayment(&pt, domain.PaymentModeInstallment, []domain.PaymentInstallment{
		{Amount: 50 + paymentSumEpsilon/4, DueDate: "2026-04-01"},
		{Amount: 50 + paymentSumEpsilon/4, DueDate: "2026-07-01"},
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateContractPayment_invalid_mode(t *testing.T) {
	pt := 100.0
	err := ValidateContractPayment(&pt, "bogus", nil)
	if err != ErrInvalidPaymentMode {
		t.Fatalf("expected ErrInvalidPaymentMode, got %v", err)
	}
}

func TestNormalizePaymentMode(t *testing.T) {
	if NormalizePaymentMode("") != domain.PaymentModeFull {
		t.Fatal()
	}
	if NormalizePaymentMode("  INSTALLMENT  ") != domain.PaymentModeInstallment {
		t.Fatal()
	}
}

func TestValidateContractPayment_bad_date(t *testing.T) {
	pt := 100000.0
	err := ValidateContractPayment(&pt, domain.PaymentModeInstallment, []domain.PaymentInstallment{
		{Amount: 50000, DueDate: "not-a-date"},
		{Amount: 50000, DueDate: "2026-07-01"},
	})
	if err != ErrInvalidPaymentSchedule {
		t.Fatalf("expected ErrInvalidPaymentSchedule, got %v", err)
	}
}

func TestValidateContractPayment_installment_too_many(t *testing.T) {
	pt := float64(25 * 10000)
	var sched []domain.PaymentInstallment
	for i := 0; i < 25; i++ {
		sched = append(sched, domain.PaymentInstallment{Amount: 10000, DueDate: "2026-04-01"})
	}
	err := ValidateContractPayment(&pt, domain.PaymentModeInstallment, sched)
	if err != ErrInvalidPaymentSchedule {
		t.Fatalf("expected ErrInvalidPaymentSchedule, got %v", err)
	}
}

func TestPaymentSumEpsilon(t *testing.T) {
	if math.Abs(paymentSumEpsilon-0.01) > 1e-9 {
		t.Fatal("paymentSumEpsilon should be 0.01 for tests")
	}
}
