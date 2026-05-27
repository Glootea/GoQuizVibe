package services_test

import (
	"testing"
	"time"

	"github.com/goquizvibe/backend/shared/infrastructure/timeprovider"
)

func TestRealTimeProvider_Now(t *testing.T) {
	t.Parallel()

	tp := timeprovider.RealTimeProvider{}
	before := time.Now()

	result := tp.Now()

	after := time.Now()

	if result.Before(before) {
		t.Errorf("Now() = %v, should be >= %v", result, before)
	}
	if result.After(after) {
		t.Errorf("Now() = %v, should be <= %v", result, after)
	}
}

func TestTimeProviderInterface(t *testing.T) {
	t.Parallel()

	var tp timeprovider.TimeProvider = timeprovider.RealTimeProvider{}
	now := tp.Now()

	if now.IsZero() {
		t.Error("TimeProvider.Now() returned zero time")
	}
}
