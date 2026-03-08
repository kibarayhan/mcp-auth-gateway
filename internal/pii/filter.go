package pii

import "regexp"

type pattern struct {
	re          *regexp.Regexp
	replacement string
}

// Filter scans text and redacts PII patterns.
type Filter struct {
	patterns []pattern
}

// NewFilter creates a PII filter with default patterns.
func NewFilter() *Filter {
	return &Filter{
		patterns: []pattern{
			// JWT tokens (3 base64 segments separated by dots)
			{regexp.MustCompile(`eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`), "[REDACTED_JWT]"},
			// AWS access keys
			{regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`), "[REDACTED_KEY]"},
			// GitHub PATs
			{regexp.MustCompile(`\bghp_[A-Za-z0-9]{36,}\b`), "[REDACTED_KEY]"},
			// OpenAI-style API keys
			{regexp.MustCompile(`\bsk-proj-[A-Za-z0-9]{12,}\b`), "[REDACTED_KEY]"},
			// Generic long API keys (32+ hex or alphanumeric with prefix)
			{regexp.MustCompile(`\bsk-[A-Za-z0-9]{32,}\b`), "[REDACTED_KEY]"},
			// Credit card numbers (13-19 digits, optional dashes/spaces)
			{regexp.MustCompile(`\b(?:\d[ -]*?){13,19}\b`), "[REDACTED_CC]"},
			// Email addresses
			{regexp.MustCompile(`\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`), "[REDACTED_EMAIL]"},
			// Phone numbers (various formats)
			{regexp.MustCompile(`(?:\+?1[-.\s]?)?(?:\(?[0-9]{3}\)?[-.\s]?)[0-9]{3}[-.\s]?[0-9]{4}\b`), "[REDACTED_PHONE]"},
		},
	}
}

// Redact replaces PII patterns in the input string with redaction placeholders.
func (f *Filter) Redact(s string) string {
	for _, p := range f.patterns {
		s = p.re.ReplaceAllString(s, p.replacement)
	}
	return s
}

// RedactJSON redacts PII from a JSON byte slice by operating on the string representation.
func (f *Filter) RedactJSON(data []byte) []byte {
	return []byte(f.Redact(string(data)))
}
