package services

import (
	"testing"
)

func TestIsCronDue_ValidExpression(t *testing.T) {
	svc := &ScheduleService{}

	// Test with a cron that should match every minute
	// We can't easily test exact timing, but we can verify parsing doesn't error
	result := svc.isCronDue("* * * * *", "UTC")
	// This should match since * * * * * means every minute
	if !result {
		t.Error("expected * * * * * to match current time")
	}
}

func TestIsCronDue_InvalidExpression(t *testing.T) {
	svc := &ScheduleService{}

	result := svc.isCronDue("invalid-cron", "UTC")
	if result {
		t.Error("expected invalid cron to not be due")
	}
}

func TestIsCronDue_SpecificHour(t *testing.T) {
	svc := &ScheduleService{}

	// A cron for a specific hour that likely doesn't match now
	// This test may be flaky if run at exactly 3am, but that's acceptable
	result := svc.isCronDue("0 3 31 2 *", "UTC") // Feb 31 doesn't exist, so this should never match
	if result {
		t.Error("expected impossible cron to not match")
	}
}
