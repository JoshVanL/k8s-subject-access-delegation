package utils

import (
	"testing"
)

func Test_ValidName(t *testing.T) {
	invalidNames := []string{"%&^", "-hello", "hello-", "hello.", ".hello"}
	validNames := []string{"hello", "h12345ello", "*hello", "1234*", "*-hello", "*-h.*", "h", "*"}

	for _, name := range invalidNames {
		testValidName(name, false, t)
	}

	for _, name := range validNames {
		testValidName(name, true, t)
	}
}

func Test_MatchName(t *testing.T) {
	name := "foo-bar.1234"
	matchRegex := []string{"foo-bar.1234", "foo*", "foo-bar*.1234", "foo*234", "*-*", "*1234", "*"}
	noMatchRegex := []string{"foo-bar.1434", "foa*", "foo-bar*..1234", "foo*2134", "*.-*", "*123", ""}

	for _, match := range matchRegex {
		testMatchName(name, match, true, t)
	}

	for _, match := range noMatchRegex {
		testMatchName(name, match, false, t)
	}

}

func testValidName(name string, valid bool, t *testing.T) {
	if ValidName(name) != valid {
		t.Errorf("expected=%t got=%t, '%s'", valid, !valid, name)
	}
}

func testMatchName(name, regex string, valid bool, t *testing.T) {
	b, err := MatchName(name, regex)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if b != valid {
		t.Errorf("expected=%t got=%t, name='%s' regexp='%s'", valid, !valid, name, regex)
	}
}
