package pii

import "testing"

func TestFilter_Emails(t *testing.T) {
	f := NewFilter()
	input := "Contact john.doe@example.com for help"
	got := f.Redact(input)
	if got == input {
		t.Error("should redact email address")
	}
	if !contains(got, "[REDACTED_EMAIL]") {
		t.Errorf("got %q, want [REDACTED_EMAIL] placeholder", got)
	}
}

func TestFilter_PhoneNumbers(t *testing.T) {
	f := NewFilter()
	tests := []string{
		"Call +1-555-123-4567 now",
		"Phone: (555) 123-4567",
		"Reach us at 555.123.4567",
	}
	for _, input := range tests {
		got := f.Redact(input)
		if got == input {
			t.Errorf("should redact phone in %q", input)
		}
	}
}

func TestFilter_CreditCards(t *testing.T) {
	f := NewFilter()
	input := "Card: 4111-1111-1111-1111 expires 12/25"
	got := f.Redact(input)
	if got == input {
		t.Error("should redact credit card number")
	}
	if !contains(got, "[REDACTED_CC]") {
		t.Errorf("got %q, want [REDACTED_CC]", got)
	}
}

func TestFilter_APIKeys(t *testing.T) {
	f := NewFilter()
	tests := []struct {
		input string
		desc  string
	}{
		{"key: sk-proj-abc123def456ghi789", "OpenAI-style key"},
		{"token: ghp_1234567890abcdef1234567890abcdef12345678", "GitHub PAT"},
		{"secret: AKIA1234567890ABCDEF", "AWS access key"},
	}
	for _, tt := range tests {
		got := f.Redact(tt.input)
		if got == tt.input {
			t.Errorf("should redact %s in %q", tt.desc, tt.input)
		}
	}
}

func TestFilter_JWTs(t *testing.T) {
	f := NewFilter()
	input := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
	got := f.Redact(input)
	if got == input {
		t.Error("should redact JWT")
	}
	if !contains(got, "[REDACTED_JWT]") {
		t.Errorf("got %q, want [REDACTED_JWT]", got)
	}
}

func TestFilter_NoFalsePositives(t *testing.T) {
	f := NewFilter()
	safe := []string{
		"Hello world",
		"The file is at /tmp/data.txt",
		"SELECT * FROM users WHERE id = 42",
		"version 1.2.3",
	}
	for _, input := range safe {
		got := f.Redact(input)
		if got != input {
			t.Errorf("false positive: %q became %q", input, got)
		}
	}
}

func TestFilter_MultiplePatterns(t *testing.T) {
	f := NewFilter()
	input := "Email john@test.com, card 4111111111111111, key sk-proj-abcdef123456"
	got := f.Redact(input)
	if contains(got, "john@test.com") {
		t.Error("email should be redacted")
	}
	if contains(got, "4111111111111111") {
		t.Error("credit card should be redacted")
	}
	if contains(got, "sk-proj-abcdef123456") {
		t.Error("API key should be redacted")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
