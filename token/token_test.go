package token

import (
	"testing"
)

func TestLookupIdent(t *testing.T) {

	for k, v := range keywords {
		if v != LookupIdent(k) {
			t.Errorf("LookupIdent gave %s, wanted %s", LookupIdent(k), v)
		}
	}

	if "IDENT" != LookupIdent("notreallyanidentifier") {
		t.Errorf("Wanted IDENT, got %s", LookupIdent("notreallyanidentifier"))
	}
}
