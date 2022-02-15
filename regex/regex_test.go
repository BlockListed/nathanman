package regex_test

import (
	"nathanman/regex"
	"testing"
)

func TestRegex(t *testing.T) {
	r := regex.Regex.FindStringSubmatch("üqweüfqsefnathan")[0]
	if r == "" || r != "üqweüfqsefnathan" {
		t.Errorf("x = üqweüfqsefnathan; Regex.FindStringSubmatch = %v; want üqweüfqsefnathan", r)
	}
	r = regex.Regex.FindStringSubmatch("adjf asjdfnathan")[0]
	if r == "" || r != "asjdfnathan" {
		t.Errorf("x = adjf asjdfnathan; Regex.FindStringSubmatch = %v; want asjdfnathan", r)
	}
	r2 := regex.Regex.FindStringSubmatch("nathan")
	if r2 != nil {
		t.Errorf("x = nathan; Regex.FindStringSubmatch = %v; want nil", r2)
	}
}
