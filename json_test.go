package snapshot

import (
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type testStruct struct {
	Member1 string
	Member2 string
}

func mkTestStruct() testStruct {
	return testStruct{
		Member1: "hello",
		Member2: "world",
	}
}

func TestAsJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:  "callable input serialises return value",
			input: mkTestStruct,
			expected: `{
  "Member1": "hello",
  "Member2": "world"
}
`,
		},
		{
			name:  "non-callable input serialises input",
			input: mkTestStruct(),
			expected: `{
  "Member1": "hello",
  "Member2": "world"
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := AsJSON(tt.input)
			if err != nil {
				t.Fatalf("failed to create reader %v", err)
			}
			buf := new(strings.Builder)
			_, err = io.Copy(buf, reader)
			if err != nil {
				t.Fatalf("failed to copy reader: %v", err)
			}
			str := buf.String()
			if diff := cmp.Diff(tt.expected, str); diff != "" {
				t.Fatalf("unexpected reader output: %v", diff)
			}
		})
	}
}
