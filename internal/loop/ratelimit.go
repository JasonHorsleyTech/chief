package loop

import "strings"

// rateLimitPatterns contains substrings that indicate a rate-limit or quota error
// from Claude's stream-json output.
var rateLimitPatterns = []string{
	"rate limit",
	"rate_limit",
	"quota",
	"overloaded",
	"529",
	"429",
	"usage limit",
}

// IsRateLimitError returns true if the given output string contains a known
// rate-limit or quota error pattern from Claude's output.
func IsRateLimitError(output string) bool {
	lower := strings.ToLower(output)
	for _, pattern := range rateLimitPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}
