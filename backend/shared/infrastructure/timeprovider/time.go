package timeprovider

import "time"

type TimeProvider interface {
	Now() time.Time
}

type RealTimeProvider struct{}

func (RealTimeProvider) Now() time.Time { return time.Now() }
