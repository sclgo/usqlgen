package fi

import "testing"

// SkipLongTest skips the current test when -short is set
func SkipLongTest(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
}
