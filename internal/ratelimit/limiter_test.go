package ratelimit

import "testing"

func TestParseRate(t *testing.T) {
	tests := []struct {
		input   string
		rate    float64
		wantErr bool
	}{
		{"100/hour", 100.0 / 3600, false},
		{"10/minute", 10.0 / 60, false},
		{"5/second", 5.0, false},
		{"", 0, true},
		{"invalid", 0, true},
		{"abc/hour", 0, true},
	}

	for _, tt := range tests {
		rate, err := ParseRate(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseRate(%q) should error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseRate(%q) error: %v", tt.input, err)
			continue
		}
		if abs(rate-tt.rate) > 0.0001 {
			t.Errorf("ParseRate(%q) = %f, want %f", tt.input, rate, tt.rate)
		}
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func TestLimiter_Allow(t *testing.T) {
	lim := New()
	lim.Configure("akibar", "filesystem", 2.0) // 2 per second

	if !lim.Allow("akibar", "filesystem") {
		t.Error("first call should be allowed")
	}
	if !lim.Allow("akibar", "filesystem") {
		t.Error("second call should be allowed")
	}
}

func TestLimiter_NoConfig(t *testing.T) {
	lim := New()

	if !lim.Allow("anyone", "any-server") {
		t.Error("should allow when no rate limit configured")
	}
}
