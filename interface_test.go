package csvbuddy

import "testing"

func TestNewReader(t *testing.T) {
	_ = NewReader(nil)
}
