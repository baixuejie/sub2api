package dsh

import "testing"

func TestAtLeastNode(t *testing.T) {
	t.Parallel()
	minimum := [3]int{22, 19, 0}
	for _, test := range []struct {
		version string
		ok      bool
	}{
		{"v22.19.0", true},
		{"22.19.1", true},
		{"v23.0.0", true},
		{"v22.18.9", false},
		{"v21.99.99", false},
		{"garbage", false},
	} {
		if got := AtLeastNode(test.version, minimum); got != test.ok {
			t.Errorf("AtLeastNode(%q) = %v, want %v", test.version, got, test.ok)
		}
	}
}

func TestParseWebURL(t *testing.T) {
	t.Parallel()
	if got, ok := ParseWebURL("dsh web: http://127.0.0.1:54321"); !ok || got != "http://127.0.0.1:54321" {
		t.Fatalf("unexpected result: %q %v", got, ok)
	}
	for _, line := range []string{
		"dsh web: https://127.0.0.1:1",
		"dsh web: http://0.0.0.0:1",
		"prefix dsh web: http://127.0.0.1:1",
	} {
		if _, ok := ParseWebURL(line); ok {
			t.Fatalf("accepted invalid line %q", line)
		}
	}
}
