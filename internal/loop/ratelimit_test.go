package loop

import "testing"

func TestIsRateLimitError_TrueCases(t *testing.T) {
	trueCases := []struct {
		name   string
		output string
	}{
		{
			name:   "rate limit lowercase",
			output: `{"type":"error","error":{"type":"overload_error","message":"rate limit exceeded"}}`,
		},
		{
			name:   "quota exceeded",
			output: `{"type":"error","error":{"message":"You have exceeded your quota for this billing period."}}`,
		},
		{
			name:   "overloaded",
			output: `{"type":"error","error":{"type":"overload_error","message":"Claude is currently overloaded with requests."}}`,
		},
		{
			name:   "HTTP 529 status code in message",
			output: `{"type":"error","error":{"message":"API error 529: Service temporarily unavailable"}}`,
		},
		{
			name:   "HTTP 429 status code in message",
			output: `{"type":"error","error":{"message":"API error 429: Too Many Requests"}}`,
		},
		{
			name:   "usage limit",
			output: `{"type":"error","error":{"message":"You have reached your usage limit."}}`,
		},
		{
			name:   "rate_limit underscore variant",
			output: `{"type":"error","error":{"type":"rate_limit_error","message":"Rate limit hit"}}`,
		},
		{
			name:   "uppercase RATE LIMIT",
			output: "Error: RATE LIMIT exceeded, please try again later",
		},
	}

	for _, tc := range trueCases {
		t.Run(tc.name, func(t *testing.T) {
			if !IsRateLimitError(tc.output) {
				t.Errorf("IsRateLimitError(%q) = false, want true", tc.output)
			}
		})
	}
}

func TestIsRateLimitError_FalseCases(t *testing.T) {
	falseCases := []struct {
		name   string
		output string
	}{
		{
			name:   "normal completion",
			output: `{"type":"result","subtype":"success","result":"Work complete"}`,
		},
		{
			name:   "tool use output",
			output: `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"go test ./..."}}]}}`,
		},
		{
			name:   "generic error not rate limit",
			output: `{"type":"error","error":{"message":"file not found: /some/path"}}`,
		},
		{
			name:   "empty string",
			output: "",
		},
	}

	for _, tc := range falseCases {
		t.Run(tc.name, func(t *testing.T) {
			if IsRateLimitError(tc.output) {
				t.Errorf("IsRateLimitError(%q) = true, want false", tc.output)
			}
		})
	}
}
