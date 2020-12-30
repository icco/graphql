package neralie

import (
	"testing"
	"time"
)

func TestFromTime(t *testing.T) {
	tests := map[string]struct {
		Hour   int
		Minute int
		Second int
		Want   string
	}{
		"16:00": {
			Hour: 6,
			Want: "250:000",
		},
		"12:00": {
			Hour: 12,
			Want: "500:000",
		},
		"18:00": {
			Hour: 18,
			Want: "750:000",
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			have := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), tc.Hour, tc.Minute, tc.Second, 0, time.UTC)
			if got := FromTime(have); got.String() != tc.Want {
				t.Errorf("FromTime(%v) = %v; want %s", have, got, tc.Want)
			}
		})
	}
}
