package java_test

import (
	"testing"

	"github.com/nedpals/errgoengine/error_templates/java"
	testutils "github.com/nedpals/errgoengine/error_templates/test_utils"
)

func TestGetSpaceBoundary(t *testing.T) {
	// generate test cases for getSpaceBoundary

	testCases := []struct {
		input    string
		expected string
		start    int
		end      int
	}{
		{"", "", 0, 0},
		{" ", " ", 0, 0},
		{"  ", "  ", 0, 0},
		{"   ", "   ", 0, 0},
		{"    ", "    ", 0, 0},
		{"     aaaa", "     ", 0, 0},
		{"   aaa   ", "   ", 0, 4},
		{"   aaa   bbb", "   ", 9, 0},
		{"a     bbb   ", "     ", 1, 6},
		{"   aaa   bbb", "   ", 6, 9},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			from, to := java.GetSpaceBoundary(tc.input, tc.start, tc.end)
			if from == 0 && to == 0 && tc.expected != "" {
				t.Errorf("expected %q, got %q (from=%d, to=%d)", tc.expected, "", from, to)
			} else if input := tc.input[from:to]; input != tc.expected {
				t.Errorf("expected %q, got %q (from=%d, to=%d)", tc.expected, input, from, to)
			}
		})
	}
}

func TestJavaErrorTemplates(t *testing.T) {
	testutils.SetupTest(t, testutils.SetupTestConfig{
		DirName:        "java",
		TemplateLoader: java.LoadErrorTemplates,
	}).Execute(t)
}
