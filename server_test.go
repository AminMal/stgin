package stgin

import "testing"

func TestPathPatternNormalizing(t *testing.T) {
	pattern := "///test/$username///"
	expected := "/test/$username/"
	normalized := normalizePath(pattern)
	if normalized != expected {
		t.Errorf("normalize function did not act as expected, expected: %s, got: %s", expected, normalized)
	}
}
