package studentcontract

import "testing"

func TestParseBenefitsContractSnapshot_learningPointsObject(t *testing.T) {
	raw := []byte(`{"learning_points":{"total":500,"used":120,"remaining":380},"guest_experience":{"quota":10,"used":3}}`)
	s := ParseBenefitsContractSnapshot(raw)
	if s.LearningPoints.Total != 500 || s.LearningPoints.Used != 120 || s.LearningPoints.Remaining != 380 {
		t.Fatalf("learning_points: %+v", s.LearningPoints)
	}
	if s.GuestExperience.Quota != 10 || s.GuestExperience.Used != 3 || s.GuestExperience.Remaining != 7 {
		t.Fatalf("guest: %+v", s.GuestExperience)
	}
}

func TestMergeContractInfoMap_preservesKeys(t *testing.T) {
	base := []byte(`{"foo":1,"learning_points":120}`)
	patch := map[string]interface{}{
		"learning_points": PatchLearningPointsObject(LearningPointsBreakdown{Total: 100, Used: 20, Remaining: 80}),
	}
	out, err := MergeContractInfoMap(base, patch)
	if err != nil {
		t.Fatal(err)
	}
	s := ParseBenefitsContractSnapshot(out)
	if s.LearningPoints.Remaining != 80 {
		t.Fatalf("merged: %+v", s.LearningPoints)
	}
}
