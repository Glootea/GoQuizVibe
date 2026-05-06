package services_test

import (
	"testing"
	"time"

	"github.com/goquizvibe/services"
)

func TestRealTimeProvider_Now(t *testing.T) {
	t.Parallel()

	tp := services.RealTimeProvider{}
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

	var tp services.TimeProvider = services.RealTimeProvider{}
	now := tp.Now()

	if now.IsZero() {
		t.Error("TimeProvider.Now() returned zero time")
	}
}
