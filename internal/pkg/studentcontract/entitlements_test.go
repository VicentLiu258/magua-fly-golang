package studentcontract

import (
	"testing"
)

func TestParseEntitlementsFromContractInfo_empty(t *testing.T) {
	e := ParseEntitlementsFromContractInfo(nil)
	if e.LearningPoints != 0 || e.GuestQuota != 0 || e.GuestUsed != 0 || e.GuestRemaining != 0 {
		t.Fatalf("expected zeros, got %+v", e)
	}
	e = ParseEntitlementsFromContractInfo([]byte("{}"))
	if e.LearningPoints != 0 || e.GuestRemaining != 0 {
		t.Fatalf("expected zeros, got %+v", e)
	}
}

func TestParseEntitlementsFromContractInfo_full(t *testing.T) {
	raw := []byte(`{"learning_points":120,"guest_experience":{"quota":5,"used":2}}`)
	e := ParseEntitlementsFromContractInfo(raw)
	if e.LearningPoints != 120 {
		t.Fatalf("learning_points: %d", e.LearningPoints)
	}
	if e.GuestQuota != 5 || e.GuestUsed != 2 || e.GuestRemaining != 3 {
		t.Fatalf("guest: %+v", e)
	}
}

func TestParseEntitlementsFromContractInfo_invalidJSON(t *testing.T) {
	e := ParseEntitlementsFromContractInfo([]byte(`not json`))
	if e.LearningPoints != 0 || e.GuestRemaining != 0 {
		t.Fatalf("expected zeros, got %+v", e)
	}
}

func TestParseEntitlementsFromContractInfo_negativeAndOverflowUsed(t *testing.T) {
	raw := []byte(`{"learning_points":-10,"guest_experience":{"quota":3,"used":10}}`)
	e := ParseEntitlementsFromContractInfo(raw)
	if e.LearningPoints != 0 {
		t.Fatalf("learning_points clamp: %d", e.LearningPoints)
	}
	if e.GuestQuota != 3 || e.GuestUsed != 10 || e.GuestRemaining != 0 {
		t.Fatalf("guest remaining floor: %+v", e)
	}
}
