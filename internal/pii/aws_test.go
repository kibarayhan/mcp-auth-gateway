package pii

import "testing"

func TestFilter_AWSAccountIDs(t *testing.T) {
	f := NewFilter()
	tests := []struct{ in, want string }{
		{"Account 123456789012 has queues", "Account [AWS_ACCOUNT] has queues"},
		{"No account here", "No account here"},
	}
	for _, tt := range tests {
		got := f.Redact(tt.in)
		if got != tt.want {
			t.Errorf("Redact(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFilter_AWSARNs(t *testing.T) {
	f := NewFilter()
	in := "arn:aws:sqs:eu-west-1:123456789012:agency-completed-tasks-dlq"
	got := f.Redact(in)
	if got == in {
		t.Errorf("ARN should be redacted, got: %s", got)
	}
	t.Logf("Redacted: %s", got)
}

func TestFilter_AWSAccessKeys(t *testing.T) {
	f := NewFilter()
	in := "Key is AKIAIOSFODNN7EXAMPLE"
	got := f.Redact(in)
	if got == in {
		t.Errorf("Access key should be redacted, got: %s", got)
	}
	t.Logf("Redacted: %s", got)
}

func TestFilter_QueueNamesPreserved(t *testing.T) {
	f := NewFilter()
	in := "DLQ depth: 34 messages in agency-completed-tasks-dlq"
	got := f.Redact(in)
	if got != in {
		t.Errorf("Queue name should NOT be redacted.\n  in:  %s\n  got: %s", in, got)
	}
}
