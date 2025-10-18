package domain

import (
	"testing"
)

func TestAccount(t *testing.T) {
	a := &Account{}

	if err := a.SetPassword("abcdefg"); err != nil {
		t.Fatal(err)
	} else if !a.ComparePassword("abcdefg") {
		t.Fatal("compare password should be true")
	} else if a.ComparePassword("abcdef") {
		t.Fatal("compare password should be false")
	}
}
