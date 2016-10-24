package invitro

import (
	"testing"
	// "github.com/CaLYpSEDamir/InfoTechGo/invitro"
)

func TestPalindrome(t *testing.T) {
	if !lsPalindrome("detartrated") {
		t.Error("IsPalindrome(\"detartrated\") = false")
	}
	if !lsPalindrome("kayak") {
		t.Error("IsPalindrome(\"kayak\") = false")
	}
}
