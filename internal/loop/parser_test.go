package loop

import (
	"testing"
)

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventUnknown, "Unknown"},
		{EventIterationStart, "IterationStart"},
		{EventAssistantText, "AssistantText"},
		{EventToolStart, "ToolStart"},
		{EventToolResult, "ToolResult"},
		{EventStoryStarted, "StoryStarted"},
		{EventStoryCompleted, "StoryCompleted"},
		{EventComplete, "Complete"},
		{EventMaxIterationsReached, "MaxIterationsReached"},
		{EventError, "Error"},
		{EventRetrying, "Retrying"},
		{EventWatchdogTimeout, "WatchdogTimeout"},
		{EventFrontPressure, "FrontPressure"},
		{EventFrontPressureResolved, "FrontPressureResolved"},
		{EventFrontPressureScrap, "FrontPressureScrap"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.eventType.String(); got != tt.expected {
				t.Errorf("EventType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseLineEmpty(t *testing.T) {
	tests := []string{
		"",
		"   ",
		"\t",
		"\n",
	}

	for _, line := range tests {
		if got := ParseLine(line); got != nil {
			t.Errorf("ParseLine(%q) = %v, want nil", line, got)
		}
	}
}

func TestParseLineInvalidJSON(t *testing.T) {
	tests := []string{
		"not json",
		"{invalid}",
		"[1, 2, 3]",
	}

	for _, line := range tests {
		if got := ParseLine(line); got != nil {
			t.Errorf("ParseLine(%q) = %v, want nil", line, got)
		}
	}
}

func TestParseLineSystemInit(t *testing.T) {
	line := `{"type":"system","subtype":"init","cwd":"/test","session_id":"abc123","tools":["Read","Write"]}`

	event := ParseLine(line)
	if event == nil {
		t.Fatal("ParseLine returned nil, want event")
	}
	if event.Type != EventIterationStart {
		t.Errorf("event.Type = %v, want EventIterationStart", event.Type)
	}
}

func TestParseLineSystemOtherSubtype(t *testing.T) {
	line := `{"type":"system","subtype":"other"}`

	event := ParseLine(line)
	if event != nil {
		t.Errorf("ParseLine returned %v, want nil for non-init system message", event)
	}
}

func TestParseLineAssistantText(t *testing.T) {
	line := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello, I will help you."}]}}`

	event := ParseLine(line)
	if event == nil {
		t.Fatal("ParseLine returned nil, want event")
	}
	if event.Type != EventAssistantText {
		t.Errorf("event.Type = %v, want EventAssistantText", event.Type)
	}
	if event.Text != "Hello, I will help you." {
		t.Errorf("event.Text = %q, want %q", event.Text, "Hello, I will help you.")
	}
}

func TestParseLineChiefComplete(t *testing.T) {
	tests := []string{
		`{"type":"assistant","message":{"content":[{"type":"text","text":"All stories complete! <chief-complete/>"}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"text","text":"<chief-complete/>"}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"text","text":"Done\n<chief-complete/>\nGoodbye"}]}}`,
	}

	for _, line := range tests {
		event := ParseLine(line)
		if event == nil {
			t.Fatalf("ParseLine(%q) returned nil, want event", line)
		}
		if event.Type != EventComplete {
			t.Errorf("ParseLine(%q): event.Type = %v, want EventComplete", line, event.Type)
		}
	}
}

func TestParseLineToolUse(t *testing.T) {
	line := `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"toolu_123","name":"Read","input":{"file_path":"/test/file.go"}}]}}`

	event := ParseLine(line)
	if event == nil {
		t.Fatal("ParseLine returned nil, want event")
	}
	if event.Type != EventToolStart {
		t.Errorf("event.Type = %v, want EventToolStart", event.Type)
	}
	if event.Tool != "Read" {
		t.Errorf("event.Tool = %q, want %q", event.Tool, "Read")
	}
	if event.ToolInput == nil {
		t.Fatal("event.ToolInput is nil")
	}
	if filePath, ok := event.ToolInput["file_path"].(string); !ok || filePath != "/test/file.go" {
		t.Errorf("event.ToolInput[\"file_path\"] = %v, want %q", event.ToolInput["file_path"], "/test/file.go")
	}
}

func TestParseLineToolResult(t *testing.T) {
	line := `{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_123","content":"File contents here"}]}}`

	event := ParseLine(line)
	if event == nil {
		t.Fatal("ParseLine returned nil, want event")
	}
	if event.Type != EventToolResult {
		t.Errorf("event.Type = %v, want EventToolResult", event.Type)
	}
	if event.Text != "File contents here" {
		t.Errorf("event.Text = %q, want %q", event.Text, "File contents here")
	}
}

func TestParseLineStoryStarted(t *testing.T) {
	line := `{"type":"assistant","message":{"content":[{"type":"text","text":"Working on the next story.\n<ralph-status>US-003</ralph-status>\nLet me implement this."}]}}`

	event := ParseLine(line)
	if event == nil {
		t.Fatal("ParseLine returned nil, want event")
	}
	if event.Type != EventStoryStarted {
		t.Errorf("event.Type = %v, want EventStoryStarted", event.Type)
	}
	if event.StoryID != "US-003" {
		t.Errorf("event.StoryID = %q, want %q", event.StoryID, "US-003")
	}
}

func TestParseLineResultMessage(t *testing.T) {
	line := `{"type":"result","subtype":"success","is_error":false,"result":"Done"}`

	event := ParseLine(line)
	// Result messages are not emitted as events
	if event != nil {
		t.Errorf("ParseLine returned %v, want nil for result message", event)
	}
}

func TestParseLineUnknownType(t *testing.T) {
	line := `{"type":"unknown_type"}`

	event := ParseLine(line)
	if event != nil {
		t.Errorf("ParseLine returned %v, want nil for unknown type", event)
	}
}

func TestParseLineEmptyContent(t *testing.T) {
	line := `{"type":"assistant","message":{"content":[]}}`

	event := ParseLine(line)
	if event != nil {
		t.Errorf("ParseLine returned %v, want nil for empty content", event)
	}
}

func TestParseLineRealWorldSystemInit(t *testing.T) {
	// Real example from Claude output
	line := `{"type":"system","subtype":"init","cwd":"/Users/codemonkey/projects/chief","session_id":"7cdf33f6-72ec-4c0e-94fb-e2b637e109da","tools":["Task","Bash","Read"],"model":"claude-opus-4-5-20251101","permissionMode":"default"}`

	event := ParseLine(line)
	if event == nil {
		t.Fatal("ParseLine returned nil, want event")
	}
	if event.Type != EventIterationStart {
		t.Errorf("event.Type = %v, want EventIterationStart", event.Type)
	}
}

func TestParseLineRealWorldToolUse(t *testing.T) {
	// Real example from Claude output
	line := `{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","id":"msg_01XPBzHqPaCuQMUFm77eHe4D","type":"message","role":"assistant","content":[{"type":"tool_use","id":"toolu_01MiR5Ps9inHigemS2gxEz5R","name":"Read","input":{"file_path":"/Users/codemonkey/projects/chief/go.mod"}}]}}`

	event := ParseLine(line)
	if event == nil {
		t.Fatal("ParseLine returned nil, want event")
	}
	if event.Type != EventToolStart {
		t.Errorf("event.Type = %v, want EventToolStart", event.Type)
	}
	if event.Tool != "Read" {
		t.Errorf("event.Tool = %q, want %q", event.Tool, "Read")
	}
}

func TestExtractStoryID(t *testing.T) {
	tests := []struct {
		text     string
		startTag string
		endTag   string
		expected string
	}{
		{
			text:     "<ralph-status>US-001</ralph-status>",
			startTag: "<ralph-status>",
			endTag:   "</ralph-status>",
			expected: "US-001",
		},
		{
			text:     "Some text <ralph-status>US-002</ralph-status> more text",
			startTag: "<ralph-status>",
			endTag:   "</ralph-status>",
			expected: "US-002",
		},
		{
			text:     "<ralph-status> US-003 </ralph-status>",
			startTag: "<ralph-status>",
			endTag:   "</ralph-status>",
			expected: "US-003",
		},
		{
			text:     "no tags here",
			startTag: "<ralph-status>",
			endTag:   "</ralph-status>",
			expected: "",
		},
		{
			text:     "<ralph-status>unclosed",
			startTag: "<ralph-status>",
			endTag:   "</ralph-status>",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := extractStoryID(tt.text, tt.startTag, tt.endTag)
			if got != tt.expected {
				t.Errorf("extractStoryID(%q, %q, %q) = %q, want %q", tt.text, tt.startTag, tt.endTag, got, tt.expected)
			}
		})
	}
}

func TestParseLineMultipleContentBlocks(t *testing.T) {
	// When there are multiple content blocks, we return the first meaningful one
	// This tests that text comes before tool_use in the content array
	line := `{"type":"assistant","message":{"content":[{"type":"text","text":"First"},{"type":"tool_use","name":"Read","input":{}}]}}`

	event := ParseLine(line)
	if event == nil {
		t.Fatal("ParseLine returned nil, want event")
	}
	// Should return the text event since it comes first
	if event.Type != EventAssistantText {
		t.Errorf("event.Type = %v, want EventAssistantText", event.Type)
	}
	if event.Text != "First" {
		t.Errorf("event.Text = %q, want %q", event.Text, "First")
	}
}

func TestParseLineToolUseFirst(t *testing.T) {
	// When tool_use comes first
	line := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Write","input":{"file_path":"/test"}},{"type":"text","text":"Second"}]}}`

	event := ParseLine(line)
	if event == nil {
		t.Fatal("ParseLine returned nil, want event")
	}
	// Should return the tool_use event since it comes first
	if event.Type != EventToolStart {
		t.Errorf("event.Type = %v, want EventToolStart", event.Type)
	}
	if event.Tool != "Write" {
		t.Errorf("event.Tool = %q, want %q", event.Tool, "Write")
	}
}

func TestParseLineFrontPressurePresent(t *testing.T) {
	line := `{"type":"assistant","message":{"content":[{"type":"text","text":"I found a problem.\n<front-pressure>The data model in stories 8-12 is wrong.</front-pressure>"}]}}`

	event := ParseLine(line)
	if event == nil {
		t.Fatal("ParseLine returned nil, want event")
	}
	if event.Type != EventFrontPressure {
		t.Errorf("event.Type = %v, want EventFrontPressure", event.Type)
	}
	if event.Text != "The data model in stories 8-12 is wrong." {
		t.Errorf("event.Text = %q, want %q", event.Text, "The data model in stories 8-12 is wrong.")
	}
}

func TestParseLineFrontPressureAbsent(t *testing.T) {
	line := `{"type":"assistant","message":{"content":[{"type":"text","text":"Everything looks fine, proceeding with implementation."}]}}`

	event := ParseLine(line)
	if event == nil {
		t.Fatal("ParseLine returned nil, want event")
	}
	if event.Type != EventAssistantText {
		t.Errorf("event.Type = %v, want EventAssistantText (no front pressure tag)", event.Type)
	}
}

func TestParseLineFrontPressureMalformed(t *testing.T) {
	// No closing tag — should not emit EventFrontPressure
	line := `{"type":"assistant","message":{"content":[{"type":"text","text":"Something wrong <front-pressure>no closing tag here"}]}}`

	event := ParseLine(line)
	if event == nil {
		t.Fatal("ParseLine returned nil, want event")
	}
	if event.Type != EventAssistantText {
		t.Errorf("event.Type = %v, want EventAssistantText for malformed front-pressure tag", event.Type)
	}
}

func TestParseLineFrontPressureMultiline(t *testing.T) {
	// Multiline concern text — in stream-json, newlines are encoded as \n
	line := `{"type":"assistant","message":{"content":[{"type":"text","text":"<front-pressure>The authentication model assumes session tokens,\nbut the specified DB schema uses JWT-only.\nThis breaks stories 5 through 9.</front-pressure>"}]}}`

	expectedConcern := "The authentication model assumes session tokens,\nbut the specified DB schema uses JWT-only.\nThis breaks stories 5 through 9."

	event := ParseLine(line)
	if event == nil {
		t.Fatal("ParseLine returned nil, want event")
	}
	if event.Type != EventFrontPressure {
		t.Errorf("event.Type = %v, want EventFrontPressure", event.Type)
	}
	if event.Text != expectedConcern {
		t.Errorf("event.Text = %q, want %q", event.Text, expectedConcern)
	}
}

func TestParseLineFrontPressureAtStart(t *testing.T) {
	// Tag appears at the very beginning of the text block
	line := `{"type":"assistant","message":{"content":[{"type":"text","text":"<front-pressure>The DB schema is missing a critical foreign key.</front-pressure>\nThis blocks all downstream stories."}]}}`

	event := ParseLine(line)
	if event == nil {
		t.Fatal("ParseLine returned nil, want event")
	}
	if event.Type != EventFrontPressure {
		t.Errorf("event.Type = %v, want EventFrontPressure", event.Type)
	}
	if event.Text != "The DB schema is missing a critical foreign key." {
		t.Errorf("event.Text = %q, want %q", event.Text, "The DB schema is missing a critical foreign key.")
	}
}

func TestParseLineFrontPressureMidText(t *testing.T) {
	// Tag appears in the middle of a text block, surrounded by other content
	line := `{"type":"assistant","message":{"content":[{"type":"text","text":"I have reviewed the story carefully. <front-pressure>The API contract in story 5 contradicts the schema defined in story 3.</front-pressure> I cannot proceed without resolving this conflict."}]}}`

	event := ParseLine(line)
	if event == nil {
		t.Fatal("ParseLine returned nil, want event")
	}
	if event.Type != EventFrontPressure {
		t.Errorf("event.Type = %v, want EventFrontPressure", event.Type)
	}
	if event.Text != "The API contract in story 5 contradicts the schema defined in story 3." {
		t.Errorf("event.Text = %q, want %q", event.Text, "The API contract in story 5 contradicts the schema defined in story 3.")
	}
}

func TestParseLineStandardEventsUnaffectedByFrontPressure(t *testing.T) {
	// Verify standard events still parse correctly after EventFrontPressure was added
	tests := []struct {
		name      string
		line      string
		wantType  EventType
		wantTool  string
	}{
		{
			name:     "assistant text",
			line:     `{"type":"assistant","message":{"content":[{"type":"text","text":"Normal assistant message without any tags."}]}}`,
			wantType: EventAssistantText,
		},
		{
			name:     "chief-complete tag",
			line:     `{"type":"assistant","message":{"content":[{"type":"text","text":"All done! <chief-complete/>"}]}}`,
			wantType: EventComplete,
		},
		{
			name:     "ralph-status tag",
			line:     `{"type":"assistant","message":{"content":[{"type":"text","text":"Starting work.\n<ralph-status>US-007</ralph-status>"}]}}`,
			wantType: EventStoryStarted,
		},
		{
			name:     "tool use",
			line:     `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"toolu_01","name":"Read","input":{"file_path":"/test"}}]}}`,
			wantType: EventToolStart,
			wantTool: "Read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := ParseLine(tt.line)
			if event == nil {
				t.Fatal("ParseLine returned nil, want event")
			}
			if event.Type != tt.wantType {
				t.Errorf("event.Type = %v, want %v", event.Type, tt.wantType)
			}
			if tt.wantTool != "" && event.Tool != tt.wantTool {
				t.Errorf("event.Tool = %q, want %q", event.Tool, tt.wantTool)
			}
		})
	}
}
