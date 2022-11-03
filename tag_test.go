package goconf

import (
	"fmt"
	"testing"
)

func TestNormalizeEnvKey(t *testing.T) {
	tests := []struct {
		original string
		expected string
	}{
		{
			original: "simple",
			expected: "SIMPLE",
		},
		{
			original: "with%symbol",
			expected: "WITH_SYMBOL",
		},
		{
			original: "with%two%%symbols",
			expected: "WITH_TWO_SYMBOLS",
		},
		{
			original: "_trim_",
			expected: "TRIM",
		},
	}

	t.Log("Given the need to normalize env var keys")
	for i, tt := range tests {
		fn := func(t *testing.T) {
			t.Logf("TEST %d: normalizing %s", i+1, tt.original)
			actual := normalizeEnvKey(tt.original)
			if tt.expected != actual {
				t.Fatalf("Failed: expected %s, got %s", tt.expected, actual)
			}
			t.Log("Success: got the expected key")
		}
		t.Run(fmt.Sprintf("Normalizing %s", tt.original), fn)
	}
}
