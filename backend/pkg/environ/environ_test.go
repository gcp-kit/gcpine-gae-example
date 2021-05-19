package environ

import "testing"

func TestIsTest(t *testing.T) {
	if !IsTest {
		t.Fatal("unexpected, expect=true, actual=false")
	}
}
